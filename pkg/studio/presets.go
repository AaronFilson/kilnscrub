package studio

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// Studio configuration mode constants.
const (
	VaporGlazeNone = "none"
	VaporGlazeSoda = "soda"
	VaporGlazeSalt = "salt"

	BuildPathCOTSNoWeld = "cots_no_weld"
	BuildPathWeldedDXF  = "welded_dxf"

	CorrosionTierHobby316L    = "hobby_316l"
	CorrosionTierHeavyC276FRP = "heavy_c276_frp"
)

var coneFahrenheitMap = map[string]float64{
	"06": 1830.0,
	"6":  2232.0,
	"8":  2305.0,
	"10": 2381.0,
	"11": 2400.0,
}

// ConeToFahrenheit returns the approximate Orton pyrometric cone firing temperature
// in degrees Fahrenheit (Cone 06=1830 F, Cone 6=2232 F, Cone 8=2305 F,
// Cone 10=2381 F, Cone 11=2400 F). Accepts both "Cone 10" and "10" forms.
func ConeToFahrenheit(cone string) (float64, bool) {
	norm := strings.ToLower(strings.TrimSpace(cone))
	norm = strings.TrimPrefix(norm, "cone")
	norm = strings.TrimSpace(norm)
	temp, ok := coneFahrenheitMap[norm]
	return temp, ok
}

// SupportedCones returns the canonical pyrometric cone labels supported by ConeToFahrenheit.
func SupportedCones() []string {
	return []string{"Cone 06", "Cone 6", "Cone 8", "Cone 10", "Cone 11"}
}

// StudioProject wraps a kiln engineering DesignInput with studio-friendly build,
// vapor-glaze, corrosion, and urban site readiness parameters.
type StudioProject struct {
	PresetID            string            `json:"preset_id"`
	StudioName          string            `json:"studio_name"`
	FiringCone          string            `json:"firing_cone"`
	FiringDurationHours float64           `json:"firing_duration_hours"`
	VaporGlazeMode      string            `json:"vapor_glaze_mode"`
	BuildPath           string            `json:"build_path"`
	CorrosionTier       string            `json:"corrosion_tier"`
	SnapToCOTS          bool              `json:"snap_to_cots"`
	IncludeSilencer     bool              `json:"include_silencer"`
	IncludePlumeDiluter bool              `json:"include_plume_diluter"`
	PropertyLineFt      float64           `json:"property_line_ft"`
	AmbientWinterF      float64           `json:"ambient_winter_f"`
	AvailableVoltage    int               `json:"available_voltage"`
	GardenHoseGPM       float64           `json:"garden_hose_gpm"`
	Input               model.DesignInput `json:"input"`
}

// StudioPreset describes a curated, ready-to-evaluate studio kiln profile.
type StudioPreset struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Tagline  string        `json:"tagline"`
	Audience string        `json:"audience"`
	Badge    string        `json:"badge"`
	Project  StudioProject `json:"project"`
}

// Presets returns the 5 curated studio and maker presets.
func Presets() []StudioPreset {
	rakuIn := model.DefaultSaltKilnInput()
	rakuIn.KilnVolumeCuFt = 6.0
	rakuIn.PeakExitTempF = 1830.0
	rakuIn.SaltRateLbsPerHr = 0.0
	rakuIn.FuelType = model.FuelGas
	rakuIn.VenturiThroatVelocity = 60.0

	sodaIn := model.DefaultSaltKilnInput()
	sodaIn.KilnVolumeCuFt = 14.0
	sodaIn.PeakExitTempF = 2350.0
	sodaIn.SaltRateLbsPerHr = 0.0
	sodaIn.FuelType = model.FuelGas
	sodaIn.VenturiThroatVelocity = 65.0

	woodSaltIn := model.DefaultSaltKilnInput()
	woodSaltIn.KilnVolumeCuFt = 24.0
	woodSaltIn.PeakExitTempF = 2400.0
	woodSaltIn.SaltRateLbsPerHr = 2.0
	woodSaltIn.FuelType = model.FuelWoodSalt

	coopIn := model.DefaultSaltKilnInput()
	coopIn.KilnVolumeCuFt = 48.0
	coopIn.PeakExitTempF = 2400.0
	coopIn.SaltRateLbsPerHr = 3.5
	coopIn.FuelType = model.FuelWoodSalt

	foundryIn := model.DefaultSaltKilnInput()
	foundryIn.KilnVolumeCuFt = 16.0
	foundryIn.PeakExitTempF = 2200.0
	foundryIn.SaltRateLbsPerHr = 0.0
	foundryIn.FuelType = model.FuelGas
	foundryIn.VenturiThroatVelocity = 75.0

	return []StudioPreset{
		{
			ID:       "electric_raku_6",
			Name:     "Small Electric / Test / Raku Hood",
			Tagline:  "Compact 6 cu.ft. low-fire & raku smoke hood built with no-weld hardware on a standard 120V garage circuit.",
			Audience: "Home potters, test kilns & raku studios",
			Badge:    "120V Friendly · No-Weld",
			Project: StudioProject{
				PresetID:            "electric_raku_6",
				StudioName:          "Home Raku & Test Kiln",
				FiringCone:          "Cone 06",
				FiringDurationHours: 6.0,
				VaporGlazeMode:      VaporGlazeNone,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierHobby316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				PropertyLineFt:      25.0,
				AmbientWinterF:      40.0,
				AvailableVoltage:    120,
				GardenHoseGPM:       5.0,
				Input:               rakuIn,
			},
		},
		{
			ID:       "gas_soda_14",
			Name:     "Backyard Gas / Soda Kiln",
			Tagline:  "Clean 14 cu.ft. Cone 10 soda-vapor kiln with zero HCl acid gas, compact footprint, and neighbor-friendly noise.",
			Audience: "Urban & suburban soda-firing ceramicists",
			Badge:    "Zero HCl · Backyard Ready",
			Project: StudioProject{
				PresetID:            "gas_soda_14",
				StudioName:          "Backyard Gas & Soda Kiln",
				FiringCone:          "Cone 10",
				FiringDurationHours: 10.0,
				VaporGlazeMode:      VaporGlazeSoda,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierHobby316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				PropertyLineFt:      30.0,
				AmbientWinterF:      35.0,
				AvailableVoltage:    240,
				GardenHoseGPM:       5.0,
				Input:               sodaIn,
			},
		},
		{
			ID:       "wood_salt_24",
			Name:     "Reference Studio Wood / Salt Kiln",
			Tagline:  "Validated 24 cu.ft. Cone 11 wood-salt reference scrubber neutralizing 2.0 lbs/hr NaCl at 99% HCl capture.",
			Audience: "Dedicated studio potters & wood-salt artists",
			Badge:    "Reference Design",
			Project: StudioProject{
				PresetID:            "wood_salt_24",
				StudioName:          "Reference Studio Wood-Salt Kiln",
				FiringCone:          "Cone 11",
				FiringDurationHours: 16.0,
				VaporGlazeMode:      VaporGlazeSalt,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierHobby316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				PropertyLineFt:      40.0,
				AmbientWinterF:      35.0,
				AvailableVoltage:    240,
				GardenHoseGPM:       6.0,
				Input:               woodSaltIn,
			},
		},
		{
			ID:       "coop_wood_48",
			Name:     "Community / Co-op Wood-Salt Kiln",
			Tagline:  "Heavy-duty 48 cu.ft. Cone 11 wood-salt train upgraded to Alloy C-276 & vinyl-ester FRP for frequent firings.",
			Audience: "Community arts centers, colleges & co-op kilns",
			Badge:    "Heavy Duty · C-276 / FRP",
			Project: StudioProject{
				PresetID:            "coop_wood_48",
				StudioName:          "Community Co-op Wood-Salt Kiln",
				FiringCone:          "Cone 11",
				FiringDurationHours: 24.0,
				VaporGlazeMode:      VaporGlazeSalt,
				BuildPath:           BuildPathWeldedDXF,
				CorrosionTier:       CorrosionTierHeavyC276FRP,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				PropertyLineFt:      60.0,
				AmbientWinterF:      30.0,
				AvailableVoltage:    240,
				GardenHoseGPM:       8.0,
				Input:               coopIn,
			},
		},
		{
			ID:       "foundry_glass_16",
			Name:     "Home Foundry / Glass / Enamel Furnace",
			Tagline:  "High-capture 16 cu.ft. 2200 F particulate & fume scrubber tuned for backyard foundry, glass, and enameling.",
			Audience: "Home foundry, glassblowing & metal arts makers",
			Badge:    "Particulate Focused",
			Project: StudioProject{
				PresetID:            "foundry_glass_16",
				StudioName:          "Home Foundry & Glass Furnace",
				FiringCone:          "Cone 6",
				FiringDurationHours: 8.0,
				VaporGlazeMode:      VaporGlazeNone,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierHobby316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				PropertyLineFt:      35.0,
				AmbientWinterF:      40.0,
				AvailableVoltage:    240,
				GardenHoseGPM:       5.0,
				Input:               foundryIn,
			},
		},
	}
}

// PresetByID looks up a curated studio preset by its stable ID.
func PresetByID(id string) (StudioPreset, bool) {
	needle := strings.TrimSpace(id)
	for _, p := range Presets() {
		if p.ID == needle {
			return p, true
		}
	}
	return StudioPreset{}, false
}

// DefaultProject returns the reference 24 cu.ft. wood/salt studio project.
func DefaultProject() StudioProject {
	p, _ := PresetByID("wood_salt_24")
	return p.Project
}

// EffectiveDesignInput returns the physics DesignInput reconciled with VaporGlazeMode.
// In "soda" or "none" modes, HCl salt loading is forced to 0 lbs/hr and fuel type is
// switched from *_salt to the corresponding clean combustion mode, while preserving
// StudioProject.VaporGlazeMode so Soda-vs-Salt comparisons remain exact.
func (p StudioProject) EffectiveDesignInput() model.DesignInput {
	in := p.Input
	switch p.VaporGlazeMode {
	case VaporGlazeNone, VaporGlazeSoda:
		in.SaltRateLbsPerHr = 0.0
		switch in.FuelType {
		case "", model.FuelGasSalt:
			in.FuelType = model.FuelGas
		case model.FuelWoodSalt:
			in.FuelType = model.FuelWood
		}
	case VaporGlazeSalt:
		switch in.FuelType {
		case "":
			in.FuelType = model.FuelWoodSalt
		case model.FuelWood:
			in.FuelType = model.FuelWoodSalt
		case model.FuelGas:
			in.FuelType = model.FuelGasSalt
		}
	}
	return in
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// Validate checks that all StudioProject parameters and its effective DesignInput are valid.
func (p StudioProject) Validate() error {
	if !isFinite(p.FiringDurationHours) || p.FiringDurationHours <= 0 || p.FiringDurationHours > 168.0 {
		return fmt.Errorf("firing_duration_hours must be finite and in (0, 168], got %v", p.FiringDurationHours)
	}
	switch p.VaporGlazeMode {
	case VaporGlazeNone, VaporGlazeSoda, VaporGlazeSalt:
	default:
		return fmt.Errorf("vapor_glaze_mode must be %q, %q, or %q, got %q",
			VaporGlazeNone, VaporGlazeSoda, VaporGlazeSalt, p.VaporGlazeMode)
	}
	switch p.BuildPath {
	case BuildPathCOTSNoWeld, BuildPathWeldedDXF:
	default:
		return fmt.Errorf("build_path must be %q or %q, got %q",
			BuildPathCOTSNoWeld, BuildPathWeldedDXF, p.BuildPath)
	}
	switch p.CorrosionTier {
	case CorrosionTierHobby316L, CorrosionTierHeavyC276FRP:
	default:
		return fmt.Errorf("corrosion_tier must be %q or %q, got %q",
			CorrosionTierHobby316L, CorrosionTierHeavyC276FRP, p.CorrosionTier)
	}
	if !isFinite(p.PropertyLineFt) || p.PropertyLineFt < 3.0 || p.PropertyLineFt > 5000.0 {
		return fmt.Errorf("property_line_ft must be finite and in [3, 5000], got %v", p.PropertyLineFt)
	}
	if !isFinite(p.AmbientWinterF) || p.AmbientWinterF < -60.0 || p.AmbientWinterF > 120.0 {
		return fmt.Errorf("ambient_winter_f must be finite and in [-60, 120], got %v", p.AmbientWinterF)
	}
	if p.AvailableVoltage != 120 && p.AvailableVoltage != 240 {
		return fmt.Errorf("available_voltage must be 120 or 240, got %d", p.AvailableVoltage)
	}
	if !isFinite(p.GardenHoseGPM) || p.GardenHoseGPM <= 0 || p.GardenHoseGPM > 100.0 {
		return fmt.Errorf("garden_hose_gpm must be finite and in (0, 100], got %v", p.GardenHoseGPM)
	}
	if err := p.EffectiveDesignInput().Validate(); err != nil {
		return fmt.Errorf("input: %w", err)
	}
	return nil
}

// ParseProject parses a complete or partial StudioProject JSON payload, seeding omitted
// fields from the referenced preset (if preset_id matches a known preset) or DefaultProject().
func ParseProject(data []byte) (StudioProject, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return DefaultProject(), nil
	}

	var probe struct {
		PresetID string `json:"preset_id"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return StudioProject{}, fmt.Errorf("parse studio project JSON: %w", err)
	}

	proj := DefaultProject()
	if probe.PresetID != "" {
		if preset, ok := PresetByID(probe.PresetID); ok {
			proj = preset.Project
		}
	}

	if err := json.Unmarshal(data, &proj); err != nil {
		return StudioProject{}, fmt.Errorf("unmarshal studio project JSON: %w", err)
	}

	if strings.TrimSpace(proj.StudioName) == "" {
		proj.StudioName = "Studio Kiln Scrubber"
	}
	if err := proj.Validate(); err != nil {
		return StudioProject{}, fmt.Errorf("invalid studio project: %w", err)
	}
	return proj, nil
}

// LoadProject reads and parses a StudioProject JSON file from disk.
func LoadProject(path string) (StudioProject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return StudioProject{}, fmt.Errorf("read project file %q: %w", path, err)
	}
	return ParseProject(data)
}

// SaveProject validates and writes a StudioProject as formatted JSON to disk.
func SaveProject(path string, proj StudioProject) error {
	if err := proj.Validate(); err != nil {
		return fmt.Errorf("invalid studio project: %w", err)
	}
	data, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal studio project: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write project file %q: %w", path, err)
	}
	return nil
}
