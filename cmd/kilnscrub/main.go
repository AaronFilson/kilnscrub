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
)

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
	case "daemon":
		return runDaemon(ctx, args[1:], stdout, stderr)
	case "ui":
		return runUI(ctx, args[1:], stdout, stderr)
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
  daemon   Run the closed-loop PID draft & quench safety controller
  ui       Launch the interactive web dashboard (default :8080)`)
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

func (f sizingFlags) apply(base model.DesignInput) model.DesignInput {
	input := base
	input.KilnVolumeCuFt = *f.kilnVol
	input.PeakExitTempF = *f.peakTemp
	input.SaltRateLbsPerHr = *f.saltRate
	input.FuelType = *f.fuel
	input.VenturiThroatVelocity = model.MetersPerSec(*f.throatVel)
	input.VenturiLGRatio = *f.lgRatio
	input.TargetHClRemoval = *f.hclRemoval
	input.TargetKilnDraftInWC = *f.draft
	return input
}

func runDesign(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("design", flag.ContinueOnError)
	fs.SetOutput(stderr)

	defaults := model.DefaultSaltKilnInput()
	sf := bindSizingFlags(fs, defaults)
	asJSON := fs.Bool("json", false, "Output complete ScrubberDesign as indented JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	input := sf.apply(defaults)
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

	input := sf.apply(defaults)
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

	srv := web.NewServer(scrubber.NewDesigner(), cad.NewExporter())

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
