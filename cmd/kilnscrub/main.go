package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/AaronFilson/kilnscrub/internal/control"
	"github.com/AaronFilson/kilnscrub/internal/web"
	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
	"github.com/AaronFilson/kilnscrub/pkg/studio"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "kilnscrub: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	switch args[0] {
	case "design":
		return runDesign(args[1:], stdout, stderr)
	case "export":
		return runExport(args[1:], stdout, stderr)
	case "presets":
		return runPresets(args[1:], stdout, stderr)
	case "guide":
		return runGuide(args[1:], stdout, stderr)
	case "packet":
		return runPacket(args[1:], stdout, stderr)
	case "daemon":
		return runDaemon(ctx, args[1:], stdout, stderr)
	case "ui":
		return runUI(ctx, args[1:], stdout, stderr)
	case "version", "--version", "-version":
		fmt.Fprintf(stdout, "kilnscrub %s\n", version)
		return nil
	case "-h", "--help", "help":
		printUsage(stdout)
		return nil
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `kilnscrub - Pure-Go Kiln Exhaust Scrubber Designer & Active Controller

Usage:
  kilnscrub <command> [flags]

Commands:
  design   Size a 4-stage kiln exhaust scrubber from firing parameters
  export   Export fabrication .stl mesh, unfolded .dxf patterns, and BOM
  presets  List curated studio kiln presets (ID, profile, audience, badge)
  guide    Evaluate COTS hardware, urban readiness, and Build/Firing Guides
  packet   Generate a complete 11-file Studio Build Packet (.zip)
  daemon   Run the closed-loop PID draft & quench safety controller
  ui       Launch the interactive 4-tab web dashboard (default :8080)
  version  Print the kilnscrub version`)
}

type sizingFlags struct {
	kilnVol    *float64
	peakTemp   *float64
	saltRate   *float64
	fuel       *string
	throatVel  *float64
	lgRatio    *float64
	hclRemoval *float64
	draft      *float64
}

func bindSizingFlags(fs *flag.FlagSet, defaults model.DesignInput) sizingFlags {
	return sizingFlags{
		kilnVol:    fs.Float64("kiln-volume", defaults.KilnVolumeCuFt, "Kiln interior volume (cu.ft.)"),
		peakTemp:   fs.Float64("peak-temp", defaults.PeakExitTempF, "Peak kiln exit gas temperature (F)"),
		saltRate:   fs.Float64("salt-rate", defaults.SaltRateLbsPerHr, "NaCl salt introduction rate (lbs/hr)"),
		fuel:       fs.String("fuel", defaults.FuelType, "Fuel stoichiometry: wood_salt, wood, gas_salt, gas"),
		throatVel:  fs.Float64("throat-vel", float64(defaults.VenturiThroatVelocity), "Venturi throat gas velocity (m/s)"),
		lgRatio:    fs.Float64("lg", defaults.VenturiLGRatio, "Venturi liquid-to-gas volumetric ratio (m3/m3)"),
		hclRemoval: fs.Float64("hcl-removal", defaults.TargetHClRemoval, "Target fractional HCl removal efficiency in (0, 1)"),
		draft:      fs.Float64("draft", defaults.TargetKilnDraftInWC, "Target kiln draft setpoint (inWC, negative = suction)"),
	}
}

func (f sizingFlags) applyExplicit(base model.DesignInput, setFlags map[string]bool) model.DesignInput {
	input := base
	if setFlags["kiln-volume"] {
		input.KilnVolumeCuFt = *f.kilnVol
	}
	if setFlags["peak-temp"] {
		input.PeakExitTempF = *f.peakTemp
	}
	if setFlags["salt-rate"] {
		input.SaltRateLbsPerHr = *f.saltRate
	}
	if setFlags["fuel"] {
		input.FuelType = *f.fuel
	}
	if setFlags["throat-vel"] {
		input.VenturiThroatVelocity = model.MetersPerSec(*f.throatVel)
	}
	if setFlags["lg"] {
		input.VenturiLGRatio = *f.lgRatio
	}
	if setFlags["hcl-removal"] {
		input.TargetHClRemoval = *f.hclRemoval
	}
	if setFlags["draft"] {
		input.TargetKilnDraftInWC = *f.draft
	}
	return input
}

func resolveStudioProject(presetID, projectPath string) (studio.StudioProject, error) {
	presetTrim := strings.TrimSpace(presetID)
	projectTrim := strings.TrimSpace(projectPath)
	if presetTrim != "" && projectTrim != "" {
		return studio.StudioProject{}, fmt.Errorf("cannot specify both --preset and --project")
	}
	if projectTrim != "" {
		return studio.LoadProject(projectTrim)
	}
	if presetTrim != "" {
		preset, ok := studio.PresetByID(presetTrim)
		if !ok {
			presets := studio.Presets()
			ids := make([]string, 0, len(presets))
			for _, p := range presets {
				ids = append(ids, p.ID)
			}
			return studio.StudioProject{}, fmt.Errorf("unknown preset %q (available: %s)", presetTrim, strings.Join(ids, ", "))
		}
		return preset.Project, nil
	}
	return studio.DefaultProject(), nil
}

func runDesign(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("design", flag.ContinueOnError)
	fs.SetOutput(stderr)

	defaults := model.DefaultSaltKilnInput()
	sf := bindSizingFlags(fs, defaults)
	presetID := fs.String("preset", "", "Optional studio preset ID to seed sizing parameters (see 'kilnscrub presets')")
	projectPath := fs.String("project", "", "Optional path to a saved StudioProject project.json file")
	asJSON := fs.Bool("json", false, "Output complete ScrubberDesign as indented JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	base := defaults
	if strings.TrimSpace(*presetID) != "" || strings.TrimSpace(*projectPath) != "" {
		proj, err := resolveStudioProject(*presetID, *projectPath)
		if err != nil {
			return err
		}
		base = proj.EffectiveDesignInput()
	}

	input := sf.applyExplicit(base, setFlags)
	if err := input.Validate(); err != nil {
		return err
	}

	design, err := scrubber.NewDesigner().Design(input)
	if err != nil {
		return err
	}

	if *asJSON {
		out, err := json.MarshalIndent(design, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal design json: %w", err)
		}
		fmt.Fprintln(stdout, string(out))
		return nil
	}

	printDesignSummary(stdout, design)
	return nil
}

func printDesignSummary(w io.Writer, d model.ScrubberDesign) {
	fmt.Fprintf(w, "Kiln Scrub 4-Stage Engineering Summary\n")
	fmt.Fprintf(w, "======================================\n")
	fmt.Fprintf(w, "Kiln Input : %.1f cu.ft. | Peak %.0f F | Salt %.2f lbs/hr | Fuel %s | Draft %.3f inWC\n",
		d.Input.KilnVolumeCuFt, d.Input.PeakExitTempF, d.Input.SaltRateLbsPerHr, d.Input.FuelType, d.Input.TargetKilnDraftInWC)
	fmt.Fprintf(w, "Inlet Gas  : %.1f ACFM (%.4f kg/s) @ %.0f F | Quenched: %.1f ACFM @ %.1f F\n\n",
		d.InletGas.VolumetricFlowACFM, d.InletGas.MassFlow, d.InletGas.Temperature.Fahrenheit(),
		d.QuenchedGas.VolumetricFlowACFM, d.QuenchedGas.Temperature.Fahrenheit())

	fmt.Fprintf(w, "Stage 1 - Adiabatic Quench Tower (%s)\n", d.Quench.Material)
	fmt.Fprintf(w, "  Tsat / Outlet / Dew Point : %.1f F (%.1f K) / %.1f F / %.1f F\n",
		d.Quench.AdiabaticSatTemp.Fahrenheit(), d.Quench.AdiabaticSatTemp,
		d.Quench.OutletTemp.Fahrenheit(), d.Quench.DewPoint.Fahrenheit())
	fmt.Fprintf(w, "  Spray Water               : %.2f GPM evap | %.2f GPM total (%.1fx overspray, %d nozzles)\n",
		d.Quench.EvaporatedWaterGPM, d.Quench.TotalSprayWaterGPM, d.Quench.OversprayFactor, d.Quench.NozzleCount)
	fmt.Fprintf(w, "  Geometry & dP             : D=%.3f m (%.1f in) x H=%.3f m (%.2f ft) | dP=%.2f inWC (%.0f Pa)\n\n",
		d.Quench.TowerDiameter, d.Quench.TowerDiameter.Inches(),
		d.Quench.TowerHeight, d.Quench.TowerHeight.Feet(),
		d.Quench.PressureDropInWC, d.Quench.PressureDrop)

	fmt.Fprintf(w, "Stage 2 - Wet Venturi Scrubber (%s)\n", d.Venturi.Material)
	fmt.Fprintf(w, "  Throat & Spray            : D_th=%.3f m (%.2f in) | v_th=%.1f m/s | L/G=%.4f (%.2f GPM)\n",
		d.Venturi.ThroatDiameter, d.Venturi.ThroatDiameter.Inches(),
		d.Venturi.ThroatVelocity, d.Venturi.LGRatio, d.Venturi.LiquidFlowGPM)
	fmt.Fprintf(w, "  Calvert d50               : NaCl fume %.3f um | Wood ash %.3f um\n",
		d.Venturi.CutDiameterD50Microns, d.Venturi.WoodAshCutDiameterD50Microns)
	fmt.Fprintf(w, "  Overall Capture & dP      : NaCl %.1f%% | Wood ash %.1f%% | dP=%.2f inWC (%.0f Pa)\n\n",
		d.Venturi.OverallNaClFumeEfficiency*100.0, d.Venturi.OverallWoodAshEfficiency*100.0,
		d.Venturi.PressureDropInWC, d.Venturi.PressureDrop)

	fmt.Fprintf(w, "Stage 3 - Packed-Bed HCl Absorber & Demister (%s)\n", d.PackedBed.Material)
	fmt.Fprintf(w, "  Column & Packing          : D=%.3f m (%.1f in) x H=%.3f m (%.2f ft) | Z_pack=%.3f m (%s)\n",
		d.PackedBed.ColumnDiameter, d.PackedBed.ColumnDiameter.Inches(),
		d.PackedBed.ColumnHeight, d.PackedBed.ColumnHeight.Feet(),
		d.PackedBed.PackingHeight, d.PackedBed.PackingSpec)
	fmt.Fprintf(w, "  Hydraulics & HCl Removal  : %.1f%% of flood | HCl %.2f%% (%.0f -> %.1f ppm) | NaHCO3 %.2f kg/hr | Recirc %.2f GPM\n",
		d.PackedBed.PercentOfFlood, d.PackedBed.HClRemovalEfficiency*100.0,
		d.PackedBed.HClInletPPM, d.PackedBed.HClOutletPPM,
		d.PackedBed.NaHCO3ConsumptionKgPerHr, d.PackedBed.RecirculationFlowGPM)
	fmt.Fprintf(w, "  Pressure Drop             : Bed %.2f inWC + Demister %.2f inWC = %.2f inWC (%.0f Pa)\n\n",
		d.PackedBed.BedPressureDrop.InWC(), d.PackedBed.DemisterPressureDrop.InWC(),
		d.PackedBed.TotalPressureDropInWC, d.PackedBed.TotalPressureDrop)

	fmt.Fprintf(w, "Stage 4 - Induced-Draft Blower (%s)\n", d.Blower.Material)
	fmt.Fprintf(w, "  Duty Point                : %.1f ACFM @ %.2f inWC (%.0f Pa) | Kiln draft setpoint %.3f inWC\n",
		d.Blower.FlowACFM, d.Blower.TotalStaticPressureInWC, d.Blower.TotalStaticPressure, d.Blower.TargetKilnDraftInWC)
	fmt.Fprintf(w, "  dP Breakdown (inWC)       : Duct %.2f | Quench %.2f | Venturi %.2f | Bed %.2f | Demister %.2f | Draft %.2f\n",
		d.Blower.Breakdown.DuctInWC, d.Blower.Breakdown.QuenchInWC, d.Blower.Breakdown.VenturiInWC,
		d.Blower.Breakdown.PackedBedInWC, d.Blower.Breakdown.DemisterInWC, d.Blower.Breakdown.KilnDraftInWC)
	fmt.Fprintf(w, "  Power & Wheel             : %.2f shaft HP -> %.2f HP motor (eta %.0f%%) | Wheel %.3f m (%.1f in) @ %.0f RPM\n\n",
		d.Blower.ShaftPowerHP, d.Blower.RecommendedMotorHP, d.Blower.Efficiency*100.0,
		d.Blower.WheelDiameter, d.Blower.WheelDiameter.Inches(), d.Blower.OperatingRPM)

	fmt.Fprintf(w, "Skid Envelope               : L=%.2f m x W=%.2f m x H=%.2f m (%.1f x %.1f x %.1f ft) | Duct D=%.3f m (%.1f in) | Area=%.2f m2\n",
		d.Envelope.Length, d.Envelope.Width, d.Envelope.Height,
		d.Envelope.Length.Feet(), d.Envelope.Width.Feet(), d.Envelope.Height.Feet(),
		d.Envelope.DuctDiameter, d.Envelope.DuctDiameter.Inches(), d.Envelope.FootprintArea)

	if len(d.Warnings) == 0 {
		fmt.Fprintf(w, "Warnings                    : none\n")
	} else {
		fmt.Fprintf(w, "Warnings (%d):\n", len(d.Warnings))
		for _, warn := range d.Warnings {
			fmt.Fprintf(w, "  - %s\n", warn)
		}
	}
}

func runExport(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(stderr)

	defaults := model.DefaultSaltKilnInput()
	sf := bindSizingFlags(fs, defaults)
	presetID := fs.String("preset", "", "Optional studio preset ID to seed sizing parameters (see 'kilnscrub presets')")
	projectPath := fs.String("project", "", "Optional path to a saved StudioProject project.json file")
	outDir := fs.String("out", ".", "Output directory for scrubber.stl, scrubber.dxf, bom.csv, and bom.md")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if strings.TrimSpace(*outDir) == "" {
		return fmt.Errorf("out directory must not be empty")
	}

	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	base := defaults
	if strings.TrimSpace(*presetID) != "" || strings.TrimSpace(*projectPath) != "" {
		proj, err := resolveStudioProject(*presetID, *projectPath)
		if err != nil {
			return err
		}
		base = proj.EffectiveDesignInput()
	}

	input := sf.applyExplicit(base, setFlags)
	if err := input.Validate(); err != nil {
		return err
	}

	design, err := scrubber.NewDesigner().Design(input)
	if err != nil {
		return err
	}

	exporter := cad.NewExporter()
	stlBytes, err := exporter.ExportSTL(design)
	if err != nil {
		return fmt.Errorf("export stl: %w", err)
	}
	dxfBytes, err := exporter.ExportDXF(design)
	if err != nil {
		return fmt.Errorf("export dxf: %w", err)
	}
	bom, err := exporter.GenerateBOM(design)
	if err != nil {
		return fmt.Errorf("generate bom: %w", err)
	}
	bomCSV, err := exporter.FormatBOMCSV(bom)
	if err != nil {
		return fmt.Errorf("format bom csv: %w", err)
	}
	bomMD, err := exporter.FormatBOMMarkdown(bom)
	if err != nil {
		return fmt.Errorf("format bom markdown: %w", err)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", *outDir, err)
	}

	outputs := []struct {
		name string
		data []byte
	}{
		{name: "scrubber.stl", data: stlBytes},
		{name: "scrubber.dxf", data: dxfBytes},
		{name: "bom.csv", data: bomCSV},
		{name: "bom.md", data: bomMD},
	}

	for _, item := range outputs {
		path := filepath.Join(*outDir, item.name)
		if err := os.WriteFile(path, item.data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Fprintf(stdout, "wrote %s (%d bytes)\n", path, len(item.data))
	}
	fmt.Fprintf(stdout, "BOM summary: %d items | %.1f kg total | $%.2f est. USD\n",
		len(bom.Items), bom.TotalWeightKg, bom.TotalCostUSD)
	return nil
}

func runPresets(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("presets", flag.ContinueOnError)
	fs.SetOutput(stderr)

	asJSON := fs.Bool("json", false, "Output studio presets as indented JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	presets := studio.Presets()
	if *asJSON {
		payload := struct {
			Presets []studio.StudioPreset `json:"presets"`
		}{Presets: presets}
		out, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal presets json: %w", err)
		}
		fmt.Fprintln(stdout, string(out))
		return nil
	}

	fmt.Fprintln(stdout, "Kiln Scrub Curated Studio Presets")
	fmt.Fprintln(stdout, "=================================")
	for _, p := range presets {
		fmt.Fprintf(stdout, "\n%s  [%s]\n", p.ID, p.Badge)
		fmt.Fprintf(stdout, "  Name     : %s\n", p.Name)
		fmt.Fprintf(stdout, "  Audience : %s\n", p.Audience)
		fmt.Fprintf(stdout, "  Tagline  : %s\n", p.Tagline)
	}
	return nil
}

type studioOptionFlags struct {
	buildPath       *string
	saltingHours    *float64
	alignment       *string
	bedLayout       *string
	parallelPipeIn  *float64
	venturiLayout   *string
	compact         *bool
	controls        *string
	blowerDrive     *string
	packingMedia    *string
	compressorSCFM  *float64
	zeroElectricity *bool
	localRegion     *string
	localAlkali     *string
	localDraftMode  *string
}

func bindStudioOptionFlags(fs *flag.FlagSet) studioOptionFlags {
	return studioOptionFlags{
		buildPath:       fs.String("build-path", "", "Fabrication build path: cots_no_weld, welded_dxf, custom_fab_deluxe, or local_ceramic_masonry"),
		saltingHours:    fs.Float64("salting-hours", 0, "Active peak salting window in hours (e.g. 2-4 hrs for salt mode)"),
		alignment:       fs.String("alignment", "", "Stage alignment: inline_id_blower or eductor_dry_fan"),
		bedLayout:       fs.String("bed-layout", "", "Stage 3 packed-bed layout: single_column or parallel_manifold"),
		parallelPipeIn:  fs.Float64("parallel-pipe-in", 0, "Shell-less parallel manifold pipe size in inches: 2, 3, or 4"),
		venturiLayout:   fs.String("venturi-layout", "", "Stage 2 venturi layout: single_throat or parallel_mini_venturis"),
		compact:         fs.Bool("compact", false, "Use compact U-turn folded skid footprint"),
		controls:        fs.String("controls", "", "Control tier: rpi_modbus or electromechanical_only"),
		blowerDrive:     fs.String("blower-drive", "", "Stage 4 blower drive: industrial_radial_vfd, regen_side_channel, dual_bypass_blowers, compressor_air_atomize, or compressor_eductor_only"),
		packingMedia:    fs.String("packing-media", "", "Stage 3 packing media: pp_pall_rings, aquaculture_bioballs, or diy_ceramic_rings"),
		compressorSCFM:  fs.Float64("compressor-scfm", 0, "Existing studio shop air compressor capacity (SCFM @ 90 PSI, default 15)"),
		zeroElectricity: fs.Bool("zero-electricity", false, "Zero electricity on site (passive thermal-reheat chimney draft + manual gravity tank; implies local_ceramic_masonry)"),
		localRegion:     fs.String("local-region", "", "Local sourcing region: nairobi, mumbai, or general_local"),
		localAlkali:     fs.String("local-alkali", "", "Local neutralizing alkali: soda_ash, slaked_lime, wood_ash_lye, or baking_soda"),
		localDraftMode:  fs.String("local-draft-mode", "", "Passive velocity/draft driver for local_ceramic_masonry (natural_chimney|steam_ejector_subsonic|steam_ejector_lp|steam_ejector_hp)"),
	}
}

func (f studioOptionFlags) applyExplicit(base studio.StudioProject, setFlags map[string]bool) studio.StudioProject {
	proj := base
	if setFlags["build-path"] {
		proj.BuildPath = strings.ToLower(strings.TrimSpace(*f.buildPath))
	}
	if setFlags["salting-hours"] {
		proj.SaltingHours = *f.saltingHours
	}
	if setFlags["alignment"] {
		proj.StageAlignment = strings.ToLower(strings.TrimSpace(*f.alignment))
	}
	if setFlags["bed-layout"] {
		proj.PackedBedLayout = strings.ToLower(strings.TrimSpace(*f.bedLayout))
	}
	if setFlags["parallel-pipe-in"] || *f.parallelPipeIn != 0 {
		proj.ParallelPipeSizeIn = *f.parallelPipeIn
	}
	if setFlags["venturi-layout"] {
		proj.VenturiLayout = strings.ToLower(strings.TrimSpace(*f.venturiLayout))
	}
	if setFlags["compact"] {
		proj.CompactFootprint = *f.compact
	}
	if setFlags["controls"] {
		proj.ControlTier = strings.ToLower(strings.TrimSpace(*f.controls))
	}
	if setFlags["blower-drive"] || strings.TrimSpace(*f.blowerDrive) != "" {
		proj.BlowerDriveOption = strings.ToLower(strings.TrimSpace(*f.blowerDrive))
	}
	if setFlags["packing-media"] || strings.TrimSpace(*f.packingMedia) != "" {
		proj.PackingMediaOption = strings.ToLower(strings.TrimSpace(*f.packingMedia))
	}
	if setFlags["compressor-scfm"] || *f.compressorSCFM != 0 {
		proj.ShopCompressorSCFM = *f.compressorSCFM
	}
	if setFlags["zero-electricity"] {
		proj.ZeroElectricity = *f.zeroElectricity
		if proj.ZeroElectricity && !setFlags["build-path"] {
			proj.BuildPath = studio.BuildPathLocalCeramic
		}
	}
	if setFlags["local-region"] {
		proj.LocalRegion = strings.ToLower(strings.TrimSpace(*f.localRegion))
	}
	if setFlags["local-alkali"] {
		proj.LocalAlkali = strings.ToLower(strings.TrimSpace(*f.localAlkali))
	}
	if setFlags["local-draft-mode"] {
		proj.LocalDraftMode = strings.ToLower(strings.TrimSpace(*f.localDraftMode))
		if (proj.LocalDraftMode == studio.LocalDraftSteamEjectorSubsonic ||
			proj.LocalDraftMode == studio.LocalDraftSteamEjectorLP ||
			proj.LocalDraftMode == studio.LocalDraftSteamEjectorHP) && !setFlags["build-path"] {
			proj.BuildPath = studio.BuildPathLocalCeramic
		}
	}
	return proj
}

func runGuide(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("guide", flag.ContinueOnError)
	fs.SetOutput(stderr)

	presetID := fs.String("preset", "", "Studio preset ID (default: wood_salt_24; see 'kilnscrub presets')")
	projectPath := fs.String("project", "", "Path to a saved StudioProject project.json file")
	outDir := fs.String("out", "", "Optional output directory to write BUILD_GUIDE.md and OPERATIONS_CHECKLIST.md")
	optFlags := bindStudioOptionFlags(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	if setFlags["out"] && strings.TrimSpace(*outDir) == "" {
		return fmt.Errorf("out directory must not be empty")
	}

	proj, err := resolveStudioProject(*presetID, *projectPath)
	if err != nil {
		return err
	}
	proj = optFlags.applyExplicit(proj, setFlags)

	eval, err := studio.NewEvaluator(scrubber.NewDesigner(), cad.NewExporter()).Evaluate(proj)
	if err != nil {
		return err
	}

	printStudioSummary(stdout, eval)

	if strings.TrimSpace(*outDir) != "" {
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			return fmt.Errorf("create output directory %q: %w", *outDir, err)
		}
		files := []struct {
			name string
			data string
		}{
			{name: "BUILD_GUIDE.md", data: eval.BuildGuideMD},
			{name: "OPERATIONS_CHECKLIST.md", data: eval.OperationsChecklistMD},
		}
		for _, item := range files {
			path := filepath.Join(*outDir, item.name)
			if err := os.WriteFile(path, []byte(item.data), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
			fmt.Fprintf(stdout, "wrote %s (%d bytes)\n", path, len(item.data))
		}
		return nil
	}

	fmt.Fprintln(stdout)
	fmt.Fprint(stdout, eval.BuildGuideMD)
	return nil
}

func printStudioSummary(w io.Writer, eval studio.StudioEvaluation) {
	p := eval.Project
	c := eval.COTS
	u := eval.Urban
	a := eval.ArchitectureOptions
	r := eval.Recirculation
	sp := eval.ShopPlan
	fmt.Fprintf(w, "Kiln Scrub Studio Readiness & COTS Summary\n")
	fmt.Fprintf(w, "==========================================\n")
	fmt.Fprintf(w, "Studio          : %s (preset=%s)\n", p.StudioName, p.PresetID)
	fmt.Fprintf(w, "Kiln & Firing   : %.1f cu.ft. | %s (%.0f F peak, %.0f hrs, %.1f hr salting) | Glaze: %s\n",
		eval.Design.Input.KilnVolumeCuFt, p.FiringCone, eval.Design.Input.PeakExitTempF, p.FiringDurationHours, p.EffectiveSaltingHours(), p.VaporGlazeMode)
	fmt.Fprintf(w, "Build & Alloy   : path=%s | tier=%s (%s) | snap_to_cots=%v\n",
		p.BuildPath, p.CorrosionTier, studio.CorrosionTierLabel(p.CorrosionTier), c.Snapped)
	if p.IsDeluxe() && sp.DeluxeSpec != nil {
		fmt.Fprintf(w, "Build Path / Deluxe : %s | delta=%+.2f USD | acoustic %.1f -> %.1f dBA @ 3ft (%.1f dBA @ %.0f ft)\n",
			p.BuildPath, sp.DeluxeSpec.CostDeltaUSD,
			sp.AcousticSummary.BareDBA3Ft, sp.AcousticSummary.TreatedDBA3Ft,
			sp.AcousticSummary.PropertyLineDBA, sp.AcousticSummary.PropertyLineFt)
	}
	if f := eval.Frugal; f != nil && f.Enabled {
		powerLabel := fmt.Sprintf("%.0f W Local Hybrid", f.PowerWatts)
		laborLabel := f.ManualPumpType
		if f.ZeroElectricity {
			powerLabel = "0 W Off-Grid"
			laborLabel = fmt.Sprintf("%d jerrycans/hr or %.1f treadle min/hr (%s)",
				f.Jerrycans20LPerHr, f.TreadleMinutesPerHr, f.ManualPumpType)
		}
		fmt.Fprintf(w, "Local Ceramic / Off-Grid : %s | %s | Power: %s\n",
			f.RegionLabel, f.LocalCurrencyEstimate, powerLabel)
		fmt.Fprintf(w, "  Chimney & Draft : %.1f m (%.1f ft) @ %.0f F reheat | %.3f vs %.3f inWC (feasible=%v)\n",
			f.ChimneyHeightM, f.ChimneyHeightFt, f.StackReheatTempF,
			f.AvailableDraftInWC, f.RequiredDraftInWC, f.DraftFeasible)
		if f.SteamEnabled {
			barStr := fmt.Sprintf("%.2f barg", f.SteamPressureBar)
			if f.DraftMode == studio.LocalDraftSteamEjectorSubsonic {
				barStr = fmt.Sprintf("%.3f barg", f.SteamPressureBar)
			}
			fmt.Fprintf(w, "  Steam-Jet Ejector : %.0f psig (%s) | %.1f m/s (Mach %.2f) | throat %.2f mm | %.1f kg/hr | %.1f vs %.1f kW waste heat | %.3f inWC (ER=%.2f) | %s\n",
				f.SteamPressurePSIG, barStr, f.SteamNozzleVelMS, f.SteamNozzleMach,
				f.SteamNozzleThroatMM, f.SteamFlowKgHr, f.SteamThermalKW, f.KilnWasteHeatKW,
				f.EjectorDraftInWC, f.EjectorEntrainmentRatio, f.SteamSourceType)
		}
		fmt.Fprintf(w, "  Ceramic Bed     : %.0f\" wide stoneware (v_s=%.2f m/s, Ergun dP=%.3f inWC)\n",
			f.BedDiameterIn, f.BedSuperficialVelMS, f.BedDPInWC)
		fmt.Fprintf(w, "  Gravity & Labor : %.0f L (%.0f gal) @ %.1f ft (%.2f PSI, %.2f L/min) | %s\n",
			f.GravityTankLiters, f.GravityTankGal, f.TankElevationFt, f.GravityHeadPSI, f.SprayFlowLPM, laborLabel)
		fmt.Fprintf(w, "  Local Alkali    : %s (%s) | %.2f kg/firing ($%.2f)\n",
			f.AlkaliName, f.AlkaliFormula, f.AlkaliKgPerFiring, f.AlkaliCostUSD)
		fmt.Fprintf(w, "  Honest Capture  : HCl %.1f%% | Coarse Ash %.1f%% | Submicron NaCl %.1f%%\n",
			f.HClRemovalPct, f.CoarseAshCapturePct, f.SubmicronNaClCapturePct)
	}
	fmt.Fprintf(w, "Stage Options   : alignment=%s | bed=%s (%.0f\" pipe) | venturi=%s | compact=%v | controls=%s | delta=%+.2f USD\n",
		a.StageAlignment, a.PackedBedLayout, a.ParallelBed.PipeNominalIn, a.VenturiLayout, a.CompactFootprint, a.ControlTier, a.CostDeltaUSD)
	fmt.Fprintf(w, "Blower Drive    : %s (%s) | %.1f inWC | $%.2f (%+.2f USD) | feasible=%v\n",
		a.BlowerDrive.Option, a.BlowerDrive.Label, a.BlowerDrive.RequiredStaticInWC, a.BlowerDrive.FanCostUSD, a.BlowerDrive.CostDeltaUSD, a.BlowerDrive.Feasible)
	fmt.Fprintf(w, "Packing Media   : %s (%s) | %.0f%% voidage | %.2f inWC bed dP | $%.2f (%+.2f USD)\n",
		a.PackingMedia.Option, a.PackingMedia.Label, a.PackingMedia.Voidage*100.0, a.PackingMedia.BedDPInWC, a.PackingMedia.MediaCostUSD, a.PackingMedia.CostDeltaUSD)
	fmt.Fprintf(w, "Shop Compressor (Eductor vs Air-Atomize) : %.1f SCFM avail | Eductor: %.1f SCFM (~%.1f HP, feasible=%v) | Air-Atomize: %.1f SCFM @ 40 PSI (%.0f um, -%.1f inWC, saves $%.0f, feasible=%v)\n",
		a.Compressor.AvailableSCFM, a.Compressor.EductorSCFMNeeded, a.Compressor.EductorHPNeeded, a.Compressor.EductorFeasible,
		a.Compressor.AtomizeSCFMNeeded, a.Compressor.AtomizeDropletUm, a.Compressor.AtomizeDPSavedInWC, a.Compressor.AtomizeCostSavedUSD, a.Compressor.AtomizeFeasible)
	fmt.Fprintf(w, "COTS Hardware   : $%.2f est. USD | Quench %.1f\" | Column %.1f\" (%.1f%% of flood) | %s\n",
		c.EstimatedTotalUSD, c.QuenchVessel.NominalIn, c.PackedColumn.NominalIn, c.ActualPercentOfFlood, c.BlowerVFD.Name)
	fmt.Fprintf(w, "Recirculation   : %.1f%% water saved | dump every %d firing(s) (SG >= %.3f)\n",
		r.WaterSavedPct, r.FiringsBeforeFullDump, r.DumpSpecificGravity)
	fmt.Fprintf(w, "Shop Plan       : %d parts, %d cut-sheet items (%.1f\" L x %.1f\" W x %.1f\" H)\n",
		len(sp.Parts), len(sp.CutList),
		sp.OverallLengthIn, sp.OverallWidthIn, sp.OverallHeightIn)
	fmt.Fprintf(w, "Shop Plan Systems : %d electrical circuits, %d plumbing lines, %d bracing members, %d acoustic zones, NEC 110.26 %.0f\"x%.0f\" panel zone\n",
		len(sp.Electrical), len(sp.Plumbing), len(sp.Bracing), len(sp.Acoustics),
		sp.Clearances.NECPanelDepthIn, sp.Clearances.NECPanelWidthIn)
	fmt.Fprintf(w, "Urban Readiness : %s\n", strings.ToUpper(u.OverallStatus))
	fmt.Fprintf(w, "  Electrical    : [%s] %s (~%.1f A)\n", u.Electrical.Status, u.Electrical.CircuitLabel, u.Electrical.RequiredAmps)
	fmt.Fprintf(w, "  Water         : [%s] %.2f GPM peak (~%.0f gal/firing)\n", u.Water.Status, u.Water.RequiredGPM, u.Water.TotalFiringGallons)
	fmt.Fprintf(w, "  Noise         : [%s] %.1f dBA @ %.0f ft property line (limit %.0f dBA)\n",
		u.Noise.Status, u.Noise.PropertyLineDBA, u.Noise.PropertyLineFt, u.Noise.UrbanLimitDBA)
	fmt.Fprintf(w, "  Winter Plume  : [%s] %.0f ft bare -> %.0f ft w/ %.0f CFM HV squirrel-cage (vs %.0f ft fence; 0-ft reheat %.1f kW)\n",
		u.Plume.Status, u.Plume.BarePlumeLengthFt, u.Plume.DilutedPlumeLengthFt, u.Plume.DilutionCFMNeeded, p.PropertyLineFt, u.Plume.DilutionReheatKW)
	fmt.Fprintf(w, "  Drain / Soda  : [%s] %s | target pH %s\n",
		u.Drain.Status, u.Drain.BulkPackagingLabel, u.Drain.TargetDrainPH)
	fmt.Fprintf(w, "  Corrosion     : [%s] %s\n", u.Corrosion.Status, u.Corrosion.InspectionInterval)
}

func runPacket(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("packet", flag.ContinueOnError)
	fs.SetOutput(stderr)

	presetID := fs.String("preset", "", "Studio preset ID (default: wood_salt_24; see 'kilnscrub presets')")
	projectPath := fs.String("project", "", "Path to a saved StudioProject project.json file")
	outPath := fs.String("out", "", "Output .zip file path (default: kilnscrub-<preset>-packet.zip)")
	optFlags := bindStudioOptionFlags(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	if setFlags["out"] && strings.TrimSpace(*outPath) == "" {
		return fmt.Errorf("out path must not be empty")
	}

	proj, err := resolveStudioProject(*presetID, *projectPath)
	if err != nil {
		return err
	}
	proj = optFlags.applyExplicit(proj, setFlags)

	targetPath := strings.TrimSpace(*outPath)
	if targetPath == "" {
		targetPath = studio.PacketFilename(proj)
	}

	zipBytes, eval, err := studio.NewEvaluator(scrubber.NewDesigner(), cad.NewExporter()).BuildPacketZip(proj)
	if err != nil {
		return err
	}

	if parent := filepath.Dir(targetPath); parent != "" && parent != "." {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("create output directory %q: %w", parent, err)
		}
	}
	if err := os.WriteFile(targetPath, zipBytes, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", targetPath, err)
	}

	fmt.Fprintf(stdout, "wrote %s (%d bytes, 11 files) | %s | Est. COTS $%.2f | Urban Readiness: %s\n",
		targetPath, len(zipBytes), eval.Project.StudioName, eval.COTS.EstimatedTotalUSD, strings.ToUpper(eval.Urban.OverallStatus))
	return nil
}

type studioBackendAdapter struct {
	eval *studio.Evaluator
}

func newStudioBackendAdapter(designer model.Designer, exporter model.CADExporter) *studioBackendAdapter {
	return &studioBackendAdapter{
		eval: studio.NewEvaluator(designer, exporter),
	}
}

func (a *studioBackendAdapter) PresetsJSON() ([]byte, error) {
	payload := struct {
		Presets []studio.StudioPreset `json:"presets"`
	}{
		Presets: studio.Presets(),
	}
	return json.Marshal(payload)
}

func (a *studioBackendAdapter) EvaluateJSON(reqBody []byte) ([]byte, error) {
	proj, err := studio.ParseProject(reqBody)
	if err != nil {
		return nil, err
	}
	eval, err := a.eval.Evaluate(proj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(eval)
}

func (a *studioBackendAdapter) BuildPacketZip(reqBody []byte) ([]byte, string, error) {
	proj, err := studio.ParseProject(reqBody)
	if err != nil {
		return nil, "", err
	}
	zipBytes, eval, err := a.eval.BuildPacketZip(proj)
	if err != nil {
		return nil, "", err
	}
	return zipBytes, studio.PacketFilename(eval.Project), nil
}

func resolveDaemonConfig(
	configPath string,
	halMode string,
	halExplicit bool,
	draft *float64,
	quenchTempF *float64,
	interval *time.Duration,
) (control.DaemonConfig, error) {
	var (
		dcfg control.DaemonConfig
		err  error
	)
	if strings.TrimSpace(configPath) != "" {
		dcfg, err = control.LoadDaemonConfig(configPath)
		if err != nil {
			return control.DaemonConfig{}, err
		}
	} else {
		dcfg = control.DefaultDaemonConfig()
	}

	if halExplicit {
		dcfg.HALMode = strings.ToLower(strings.TrimSpace(halMode))
	}
	if draft != nil {
		dcfg.Control.SetTargetDraftInWC(*draft)
	}
	if quenchTempF != nil {
		dcfg.Control.QuenchSetpoint = model.FahrenheitToKelvin(*quenchTempF)
	}
	if interval != nil {
		dcfg.Control.LoopInterval = *interval
	}
	dcfg.Control.Sync()
	if err := dcfg.Validate(); err != nil {
		return control.DaemonConfig{}, err
	}
	return dcfg, nil
}

func openControlHAL(dcfg control.DaemonConfig) (model.HardwareHAL, io.Closer, error) {
	switch dcfg.HALMode {
	case control.HALModeSim:
		simCfg := dcfg.Sim
		if simCfg.AutoStepInterval == 0 {
			simCfg.AutoStepInterval = dcfg.Control.LoopInterval
		}
		return control.NewSimHAL(simCfg), nil, nil
	case control.HALModeModbus:
		mbHAL, err := control.OpenModbusHAL(dcfg.Modbus)
		if err != nil {
			return nil, nil, fmt.Errorf("open modbus HAL (%s): %w", dcfg.Modbus.Device, err)
		}
		return mbHAL, mbHAL, nil
	default:
		return nil, nil, fmt.Errorf("unsupported hal mode %q (expected %q or %q)", dcfg.HALMode, control.HALModeSim, control.HALModeModbus)
	}
}

func runDaemon(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("daemon", flag.ContinueOnError)
	fs.SetOutput(stderr)

	defaults := model.DefaultControlConfig()
	configPath := fs.String("config", "", "Path to JSON daemon configuration file")
	halMode := fs.String("hal", control.HALModeSim, "Hardware abstraction layer backend: sim or modbus")
	draft := fs.Float64("draft", defaults.TargetDraftInWC, "Target kiln draft (inWC, negative = suction)")
	quenchTempF := fs.Float64("quench-temp-f", defaults.QuenchSetpoint.Fahrenheit(), "Target quench outlet temperature (F)")
	interval := fs.Duration("interval", defaults.LoopInterval, "Closed-loop control tick interval")
	duration := fs.Duration("duration", 0, "Optional bounded run duration (0 = run until SIGINT/SIGTERM)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if *duration < 0 {
		return fmt.Errorf("duration must be >= 0, got %v", *duration)
	}

	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	var draftOpt, quenchOpt *float64
	var intervalOpt *time.Duration
	if setFlags["draft"] {
		draftOpt = draft
	}
	if setFlags["quench-temp-f"] {
		quenchOpt = quenchTempF
	}
	if setFlags["interval"] {
		intervalOpt = interval
	}

	dcfg, err := resolveDaemonConfig(*configPath, *halMode, setFlags["hal"], draftOpt, quenchOpt, intervalOpt)
	if err != nil {
		return err
	}

	hal, closer, err := openControlHAL(dcfg)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}

	cfg := dcfg.Control
	var tickNum atomic.Uint64
	daemon, err := control.NewDaemon(
		cfg,
		hal,
		control.WithHALMode(dcfg.HALMode),
		control.WithTelemetryCallback(func(st model.ControlStatus) {
			n := tickNum.Add(1)
			fmt.Fprintf(stdout, "tick=%d state=%s draft=%.4f inWC (%.2f Pa) quench_out=%.1f F water=%.2f GPM vfd=%.1f%% valve=%.1f%% bypass=%v\n",
				n, st.State, st.FilteredDraftInWC, st.FilteredDraftPa,
				st.Readings.QuenchOutletTemp.Fahrenheit(), st.Readings.WaterFlowGPM,
				st.Commands.BlowerVFD*100.0, st.Commands.QuenchSprayValve*100.0,
				st.Commands.BypassDamperOpen)
		}),
		control.WithErrorCallback(func(err error) {
			fmt.Fprintf(stderr, "daemon hal error: %v\n", err)
		}),
	)
	if err != nil {
		return err
	}

	sigCtx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	runCtx := sigCtx
	var cancelTimeout context.CancelFunc
	if *duration > 0 {
		runCtx, cancelTimeout = context.WithTimeout(sigCtx, *duration)
		defer cancelTimeout()
	}

	mode := "until SIGINT/SIGTERM"
	if *duration > 0 {
		mode = fmt.Sprintf("for %v", *duration)
	}
	fmt.Fprintf(stdout, "kilnscrub daemon starting (%s, hal=%s): target_draft=%.3f inWC (%.2f Pa) quench_sp=%.1f F interval=%v\n",
		mode, dcfg.HALMode, cfg.TargetDraftInWC, cfg.TargetDraftPa, cfg.QuenchSetpoint.Fahrenheit(), cfg.LoopInterval)

	if err := daemon.Run(runCtx); err != nil {
		return err
	}

	final := daemon.Status()
	fmt.Fprintf(stdout, "kilnscrub daemon stopped cleanly after %d ticks (final state=%s, interlock=%v)\n",
		daemon.Ticks(), final.State, final.InterlockActive)
	return nil
}

func runUI(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	fs.SetOutput(stderr)

	addr := fs.String("addr", ":8080", "HTTP listen address")
	configPath := fs.String("config", "", "Path to JSON daemon configuration file for embedded live control")
	halMode := fs.String("hal", control.HALModeSim, "Hardware abstraction layer backend for embedded control: sim or modbus")
	noControl := fs.Bool("no-control", false, "Disable embedded live control daemon and telemetry backend")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if strings.TrimSpace(*addr) == "" {
		return fmt.Errorf("addr must not be empty")
	}

	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	designer := scrubber.NewDesigner()
	exporter := cad.NewExporter()
	srv := web.NewServer(designer, exporter)
	srv.SetStudioBackend(newStudioBackendAdapter(designer, exporter))

	sigCtx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	var (
		daemonCancel context.CancelFunc
		daemonDone   chan struct{}
		halCloser    io.Closer
		controlDesc  = "disabled"
	)

	if !*noControl {
		dcfg, err := resolveDaemonConfig(*configPath, *halMode, setFlags["hal"], nil, nil, nil)
		if err != nil {
			return err
		}
		hal, closer, err := openControlHAL(dcfg)
		if err != nil {
			return err
		}
		halCloser = closer

		daemon, err := control.NewDaemon(
			dcfg.Control,
			hal,
			control.WithHALMode(dcfg.HALMode),
			control.WithErrorCallback(func(err error) {
				fmt.Fprintf(stderr, "ui control hal error: %v\n", err)
			}),
		)
		if err != nil {
			if halCloser != nil {
				_ = halCloser.Close()
			}
			return err
		}

		srv.SetControlBackend(daemon)
		controlDesc = dcfg.HALMode

		var dCtx context.Context
		dCtx, daemonCancel = context.WithCancel(sigCtx)
		daemonDone = make(chan struct{})
		go func() {
			defer close(daemonDone)
			if err := daemon.Run(dCtx); err != nil {
				fmt.Fprintf(stderr, "ui control daemon error: %v\n", err)
			}
		}()
	}

	defer func() {
		if daemonCancel != nil {
			daemonCancel()
		}
		if daemonDone != nil {
			<-daemonDone
		}
		if halCloser != nil {
			_ = halCloser.Close()
		}
	}()

	displayURL := formatListenURL(*addr)
	fmt.Fprintf(stdout, "kilnscrub ui listening on %s (addr %s, control=%s)\n", displayURL, *addr, controlDesc)

	return srv.ListenAndServe(sigCtx, *addr)
}

func formatListenURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	if strings.HasPrefix(addr, "0.0.0.0:") {
		return "http://localhost:" + strings.TrimPrefix(addr, "0.0.0.0:")
	}
	return "http://" + addr
}
