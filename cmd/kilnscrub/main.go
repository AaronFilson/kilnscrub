package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "kilnscrub: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "design":
		return runDesign(args[1:])
	case "export":
		return runExport(args[1:])
	case "daemon":
		return runDaemon(args[1:])
	case "ui":
		return runUI(args[1:])
	case "-h", "--help", "help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func printUsage() {
	fmt.Println(`kilnscrub - Pure-Go Kiln Exhaust Scrubber Designer & Active Controller

Usage:
  kilnscrub <command> [flags]

Commands:
  design   Size a 4-stage kiln exhaust scrubber from firing parameters
  export   Export fabrication .stl mesh, unfolded .dxf patterns, and BOM
  daemon   Run the closed-loop PID draft & quench safety controller
  ui       Launch the interactive web dashboard (default :8080)`)
}

func runDesign(args []string) error {
	fs := flag.NewFlagSet("design", flag.ContinueOnError)
	defaults := model.DefaultSaltKilnInput()
	kilnVol := fs.Float64("kiln-volume", defaults.KilnVolumeCuFt, "Kiln interior volume (cu.ft.)")
	peakTemp := fs.Float64("peak-temp", defaults.PeakExitTempF, "Peak kiln exit gas temperature (F)")
	saltRate := fs.Float64("salt-rate", defaults.SaltRateLbsPerHr, "NaCl salt introduction rate (lbs/hr)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	input := defaults
	input.KilnVolumeCuFt = *kilnVol
	input.PeakExitTempF = *peakTemp
	input.SaltRateLbsPerHr = *saltRate
	if err := input.Validate(); err != nil {
		return err
	}

	out, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("design stub (Phase 1 — engineering solver wired in Phase 3)\n%s\n", out)
	return nil
}

func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	outDir := fs.String("out", ".", "Output directory for .stl, .dxf, and BOM files")
	if err := fs.Parse(args); err != nil {
		return err
	}
	fmt.Printf("export stub (Phase 1 — CAD/BOM exporters wired in Phase 3; out=%s)\n", *outDir)
	return nil
}

func runDaemon(args []string) error {
	fs := flag.NewFlagSet("daemon", flag.ContinueOnError)
	defaults := model.DefaultControlConfig()
	draft := fs.Float64("draft", defaults.TargetDraftInWC, "Target kiln draft (inWC)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	fmt.Printf("daemon stub (Phase 1 — active PID controller wired in Phase 3; target draft=%.3f inWC)\n", *draft)
	return nil
}

func runUI(args []string) error {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "HTTP listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	fmt.Printf("ui stub (Phase 1 — embedded web dashboard wired in Phase 3; addr=%s)\n", *addr)
	return nil
}
