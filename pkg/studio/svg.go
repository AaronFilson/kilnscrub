package studio

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

func svgEscape(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func truncateSVGText(s string, maxLen int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= maxLen {
		return string(r)
	}
	if maxLen <= 3 {
		return string(r[:maxLen])
	}
	return string(r[:maxLen-3]) + "..."
}

// RenderShopPlanSVG renders a standalone, printer-friendly architectural blueprint SVG
// from a structured ShopPlan so every Build Packet (including masonry/off-grid builds
// without sheet-metal STL/DXF files) includes a crisp visual schematic.
func RenderShopPlanSVG(plan ShopPlan) []byte {
	const (
		width  = 1120
		height = 760
	)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace">`+"\n", width, height, width, height)
	b.WriteString(`<defs>
  <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
    <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#1e3a5f" stroke-width="0.5" opacity="0.35"/>
  </pattern>
  <marker id="arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 1 L 10 5 L 0 9 z" fill="#38bdf8"/>
  </marker>
</defs>
`)
	// Blueprint background & border.
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#0b192c"/>`+"\n", width, height)
	fmt.Fprintf(&b, `<rect x="16" y="16" width="%d" height="%d" fill="url(#grid)" stroke="#38bdf8" stroke-width="1.5"/>`+"\n", width-32, height-32)

	// Header Title Block.
	b.WriteString(`<rect x="28" y="28" width="1064" height="62" fill="#102542" stroke="#38bdf8" stroke-width="1"/>` + "\n")
	fmt.Fprintf(&b, `<text x="44" y="54" fill="#f8fafc" font-size="18" font-weight="bold">%s</text>`+"\n", svgEscape(truncateSVGText(plan.Title, 82)))
	modeSub := fmt.Sprintf("Build Path: %s | Overall Envelope: %.1f\" L x %.1f\" W x %.1f\" H (%.2f m x %.2f m x %.2f m)",
		plan.BuildPath,
		plan.OverallLengthIn, plan.OverallWidthIn, plan.OverallHeightIn,
		plan.OverallLengthIn*0.0254, plan.OverallWidthIn*0.0254, plan.OverallHeightIn*0.0254)
	if plan.LocalDraftMode != "" {
		modeSub += fmt.Sprintf(" | Draft: %s", plan.LocalDraftMode)
	}
	fmt.Fprintf(&b, `<text x="44" y="76" fill="#93c5fd" font-size="11">%s</text>`+"\n", svgEscape(truncateSVGText(modeSub, 128)))

	// Process Flow Schematic Panel (Top).
	b.WriteString(`<rect x="28" y="104" width="1064" height="250" fill="#0f213a" stroke="#1e4976" stroke-width="1"/>` + "\n")
	b.WriteString(`<text x="42" y="124" fill="#38bdf8" font-size="12" font-weight="bold">PROCESS ELEVATION &amp; STAGE SCHEMATIC</text>` + "\n")

	isLocal := plan.BuildPath == BuildPathLocalCeramic
	type stageBox struct {
		x, y, w, h int
		tag        string
		title      string
		detail     string
		accent     string
	}
	var boxes []stageBox
	if isLocal {
		stage4Title := "Tall Brick Chimney"
		if plan.SteamEjectorEnabled {
			stage4Title = "Steam-Jet & Stack"
		}
		boxes = []stageBox{
			{52, 162, 150, 96, "[1] INLET", "Refractory Flue", "Firebrick (>=36\" clr)", "#f97316"},
			{238, 148, 175, 118, "[2-3] STAGE 1-2", "Masonry Quench", "Gravity Spray & Baffles", "#f59e0b"},
			{452, 138, 210, 140, "[4] STAGE 3", "Wide Stoneware Bed", "1\" Ceramic Rings + Mist", "#22c55e"},
			{702, 146, 175, 124, "[6-7] GRAVITY", "3.5m Overhead Tank", "Closed-Loop Sump & Pump", "#38bdf8"},
			{912, 134, 152, 144, "[8] EXHAUST", stage4Title, "Outdoor Dispersion", "#a78bfa"},
		}
	} else {
		stage2Title := "Wet Venturi"
		if plan.StageAlignment == StageAlignmentEductorDryFan {
			stage2Title = "Downflow Eductor"
		} else if plan.VenturiLayout == VenturiLayoutParallelMiniVenturis {
			stage2Title = "Parallel Venturis"
		}
		stage3Title := "Packed Column"
		if plan.PackedBedLayout == PackedBedLayoutParallelManifold {
			stage3Title = "Uniseal Pipe Bundle"
		}
		boxes = []stageBox{
			{52, 162, 145, 96, "[1] INLET", "Hot Flue & Bypass", "Fail-Open Damper (200 F)", "#f97316"},
			{232, 148, 168, 118, "[2] STAGE 1", "Quench Tower", "Co-Current Spray Cooling", "#f59e0b"},
			{434, 162, 168, 96, "[3] STAGE 2", stage2Title, "Submicron NaCl Capture", "#eab308"},
			{636, 138, 185, 140, "[4-5] STAGE 3", stage3Title, "HCl Neutralization + Sump", "#22c55e"},
			{856, 146, 208, 128, "[6-8] STAGE 4", "Blower & Stack", "Draft Control + Outdoor Stack", "#38bdf8"},
		}
	}

	for i, bx := range boxes {
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" rx="4" fill="#132b4c" stroke="%s" stroke-width="1.8"/>`+"\n", bx.x, bx.y, bx.w, bx.h, bx.accent)
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="%s" font-size="10" font-weight="bold">%s</text>`+"\n", bx.x+10, bx.y+20, bx.accent, svgEscape(bx.tag))
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="#f8fafc" font-size="12" font-weight="bold">%s</text>`+"\n", bx.x+10, bx.y+44, svgEscape(bx.title))
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="#cbd5e1" font-size="10">%s</text>`+"\n", bx.x+10, bx.y+66, svgEscape(bx.detail))
		if i < len(boxes)-1 {
			next := boxes[i+1]
			x1 := bx.x + bx.w
			y1 := bx.y + bx.h/2
			x2 := next.x - 4
			y2 := next.y + next.h/2
			fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#38bdf8" stroke-width="2" marker-end="url(#arrow)"/>`+"\n", x1, y1, x2, y2)
		}
	}

	// Clearance Callout Banner inside schematic.
	clrText := fmt.Sprintf("CLEARANCES: Flue->Combustibles >= %.0f\" | Property Line >= %.0f ft | Stack > Roof >= %.0f ft | Service Aisle >= %.0f\" | Pump Service >= %.0f\"",
		plan.Clearances.FlueToCombustiblesIn, plan.Clearances.PropertyLineMinFt,
		plan.Clearances.StackAboveRoofFt, plan.Clearances.ServiceAisleIn, plan.Clearances.PumpServiceIn)
	b.WriteString(`<rect x="44" y="302" width="1032" height="36" rx="3" fill="#0b192c" stroke="#334155" stroke-width="1"/>` + "\n")
	fmt.Fprintf(&b, `<text x="58" y="324" fill="#fde68a" font-size="11">%s</text>`+"\n", svgEscape(clrText))

	// Bottom-Left Panel: Parts Legend.
	b.WriteString(`<rect x="28" y="368" width="600" height="360" fill="#0f213a" stroke="#1e4976" stroke-width="1"/>` + "\n")
	b.WriteString(`<text x="42" y="390" fill="#38bdf8" font-size="12" font-weight="bold">NUMBERED PARTS LEGEND</text>` + "\n")
	y := 412
	maxParts := len(plan.Parts)
	if maxParts > 12 {
		maxParts = 12
	}
	for i := 0; i < maxParts; i++ {
		pt := plan.Parts[i]
		line := fmt.Sprintf("[%d] %s — %s ($%.0f)", pt.Num, truncateSVGText(pt.Label, 40), truncateSVGText(pt.DimensionsIn, 32), pt.CostUSD)
		fmt.Fprintf(&b, `<text x="42" y="%d" fill="#e2e8f0" font-size="10.5">%s</text>`+"\n", y, svgEscape(truncateSVGText(line, 92)))
		y += 25
	}

	// Bottom-Right Panel: Electrical, Plumbing, Acoustics & Bracing Summary.
	b.WriteString(`<rect x="644" y="368" width="448" height="360" fill="#0f213a" stroke="#1e4976" stroke-width="1"/>` + "\n")
	b.WriteString(`<text x="658" y="390" fill="#38bdf8" font-size="12" font-weight="bold">SYSTEMS &amp; UTILITIES SUMMARY</text>` + "\n")

	sysLines := []string{
		fmt.Sprintf("Pump Station: %s", truncateSVGText(plan.PumpStation.PumpType, 46)),
		fmt.Sprintf("  Flow: %.2f GPM @ %.1f ft head (%.2g HP, %dV)", plan.PumpStation.FlowGPM, plan.PumpStation.HeadFt, plan.PumpStation.HP, plan.PumpStation.Voltage),
		fmt.Sprintf("  Strainer: %s", truncateSVGText(plan.PumpStation.Strainer, 48)),
		fmt.Sprintf("Acoustics: %.1f dBA bare -> %.1f dBA @ 3 ft -> %.1f dBA @ %.0f ft",
			plan.AcousticSummary.BareDBA3Ft, plan.AcousticSummary.TreatedDBA3Ft,
			plan.AcousticSummary.PropertyLineDBA, plan.AcousticSummary.PropertyLineFt),
		fmt.Sprintf("Electrical Circuits: %d | Plumbing Lines: %d | Bracing: %d | Cut Items: %d",
			len(plan.Electrical), len(plan.Plumbing), len(plan.Bracing), len(plan.CutList)),
		"--- PLUMBING LINES ---",
	}
	for i, pl := range plan.Plumbing {
		if i >= 7 {
			break
		}
		sysLines = append(sysLines, fmt.Sprintf("%s: %.2f\" %s (%.2f GPM @ %.1f PSI)", pl.ID, pl.SizeIn, truncateSVGText(pl.Label, 30), pl.FlowGPM, pl.PressurePSI))
	}
	sysLines = append(sysLines, "--- WIRING / CODE ACCESS ---")
	sysLines = append(sysLines, truncateSVGText(plan.Clearances.WiringAccessNote, 68))

	sy := 412
	for _, sl := range sysLines {
		fill := "#cbd5e1"
		if strings.HasPrefix(sl, "---") {
			fill = "#38bdf8"
		}
		fmt.Fprintf(&b, `<text x="658" y="%d" fill="%s" font-size="10.2">%s</text>`+"\n", sy, fill, svgEscape(truncateSVGText(sl, 70)))
		sy += 21
		if sy > 716 {
			break
		}
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}
