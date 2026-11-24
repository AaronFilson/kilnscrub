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

	BuildPathCOTSNoWeld      = "cots_no_weld"
	BuildPathWeldedDXF       = "welded_dxf"
	BuildPathCustomFabDeluxe = "custom_fab_deluxe"
	BuildPathLocalCeramic    = "local_ceramic_masonry"

	// Phase 10 regional sourcing & local alkali constants.
	LocalRegionNairobi = "nairobi"
	LocalRegionMumbai  = "mumbai"
	LocalRegionGeneral = "general_local"

	LocalAlkaliSodaAsh    = "soda_ash"
	LocalAlkaliSlakedLime = "slaked_lime"
	LocalAlkaliWoodAshLye = "wood_ash_lye"
	LocalAlkaliBakingSoda = "baking_soda"

	// Phase 10B/10C local ceramic draft / zero-electricity steam-jet ejector mode constants.
	LocalDraftNaturalChimney       = "natural_chimney"
	LocalDraftSteamEjectorSubsonic = "steam_ejector_subsonic"
	LocalDraftSteamEjectorLP       = "steam_ejector_lp"
	LocalDraftSteamEjectorHP       = "steam_ejector_hp"

	// Phase 13/14 local ceramic boiler heat source & off-grid steam water pump / sump-pool constants.
	LocalBoilerDedicatedHearth          = "dedicated_hearth"
	LocalBoilerKilnExhaustCoil          = "kiln_exhaust_coil"
	LocalBoilerStage1FeedwaterPreheater = "stage1_feedwater_preheater"
	LocalBoilerStage1WaterJacket        = "stage1_water_jacket"

	LocalPumpManualTreadle  = "manual_treadle"
	LocalPumpSteamPPCP      = "steam_ppcp"
	LocalPumpSteamJetSyphon = "steam_jet_syphon"
	LocalPumpSumpPoolNoPump = "sump_pool_no_pump"

	// CorrosionTierStandard316L is the canonical light/occasional 316L SS tier.
	CorrosionTierStandard316L = "standard_316l"
	// CorrosionTierHobby316L is accepted as a backward-compatible alias for CorrosionTierStandard316L.
	CorrosionTierHobby316L    = "hobby_316l"
	CorrosionTierHeavyC276FRP = "heavy_c276_frp"

	// Phase 7 selectable stage alignment, parallel layout, and control tier constants.
	StageAlignmentInlineIDBlower = "inline_id_blower"
	StageAlignmentEductorDryFan  = "eductor_dry_fan"

	PackedBedLayoutSingleColumn     = "single_column"
	PackedBedLayoutParallelManifold = "parallel_manifold"

	VenturiLayoutSingleThroat         = "single_throat"
	VenturiLayoutParallelMiniVenturis = "parallel_mini_venturis"

	ControlTierRPiModbus             = "rpi_modbus"
	ControlTierElectromechanicalOnly = "electromechanical_only"

	// Phase 8 selectable blower/motive drive and packing media constants.
	BlowerDriveIndustrialRadialVFD  = "industrial_radial_vfd"
	BlowerDriveRegenSideChannel     = "regen_side_channel"
	BlowerDriveDualBypassBlowers    = "dual_bypass_blowers"
	BlowerDriveCompressorAirAtomize = "compressor_air_atomize"
	BlowerDriveCompressorEductor    = "compressor_eductor_only"

	PackingMediaPPPallRings         = "pp_pall_rings"
	PackingMediaAquacultureBioballs = "aquaculture_bioballs"
	PackingMediaDIYCeramicRings     = "diy_ceramic_rings"
)

// CorrosionTierLabel returns the user-facing display label for a corrosion tier.
func CorrosionTierLabel(tier string) string {
	if tier == CorrosionTierHeavyC276FRP {
		return "Heavy Duty (Alloy C-276 / Vinyl-Ester FRP)"
	}
	return "Standard 316L SS (Light / Occasional)"
}

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
// vapor-glaze, corrosion, urban site readiness, and Phase 7 stage-design parameters.
type StudioProject struct {
	PresetID                  string            `json:"preset_id"`
	StudioName                string            `json:"studio_name"`
	FiringCone                string            `json:"firing_cone"`
	FiringDurationHours       float64           `json:"firing_duration_hours"`
	SaltingHours              float64           `json:"salting_hours"`
	VaporGlazeMode            string            `json:"vapor_glaze_mode"`
	BuildPath                 string            `json:"build_path"`
	CorrosionTier             string            `json:"corrosion_tier"`
	SnapToCOTS                bool              `json:"snap_to_cots"`
	IncludeSilencer           bool              `json:"include_silencer"`
	IncludePlumeDiluter       bool              `json:"include_plume_diluter"`
	StageAlignment            string            `json:"stage_alignment"`
	PackedBedLayout           string            `json:"packed_bed_layout"`
	VenturiLayout             string            `json:"venturi_layout"`
	CompactFootprint          bool              `json:"compact_footprint"`
	SharedBaffledSump         bool              `json:"shared_baffled_sump"`
	ControlTier               string            `json:"control_tier"`
	BlowerDriveOption         string            `json:"blower_drive_option,omitempty"`
	ParallelPipeSizeIn        float64           `json:"parallel_pipe_size_in,omitempty"`
	PackingMediaOption        string            `json:"packing_media_option,omitempty"`
	ShopCompressorSCFM        float64           `json:"shop_compressor_scfm,omitempty"`
	ZeroElectricity           bool              `json:"zero_electricity,omitempty"`
	LocalRegion               string            `json:"local_region,omitempty"`
	LocalAlkali               string            `json:"local_alkali,omitempty"`
	LocalDraftMode            string            `json:"local_draft_mode,omitempty"`
	LocalBoilerSource         string            `json:"local_boiler_source,omitempty"`
	LocalPumpMode             string            `json:"local_pump_mode,omitempty"`
	DeluxeCondensingReclaimer bool              `json:"deluxe_condensing_reclaimer,omitempty"`
	DeluxeVariableThroatVFD   bool              `json:"deluxe_variable_throat_vfd,omitempty"`
	DeluxeStage1HeatRecovery  bool              `json:"deluxe_stage1_heat_recovery,omitempty"`
	PropertyLineFt            float64           `json:"property_line_ft"`
	AmbientWinterF            float64           `json:"ambient_winter_f"`
	AvailableVoltage          int               `json:"available_voltage"`
	GardenHoseGPM             float64           `json:"garden_hose_gpm"`
	Input                     model.DesignInput `json:"input"`
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
				SaltingHours:        0.0,
				VaporGlazeMode:      VaporGlazeNone,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierStandard316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				StageAlignment:      StageAlignmentInlineIDBlower,
				PackedBedLayout:     PackedBedLayoutSingleColumn,
				VenturiLayout:       VenturiLayoutSingleThroat,
				CompactFootprint:    false,
				SharedBaffledSump:   true,
				ControlTier:         ControlTierRPiModbus,
				BlowerDriveOption:   BlowerDriveIndustrialRadialVFD,
				ParallelPipeSizeIn:  4.0,
				PackingMediaOption:  PackingMediaPPPallRings,
				ShopCompressorSCFM:  15.0,
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
				SaltingHours:        0.0,
				VaporGlazeMode:      VaporGlazeSoda,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierStandard316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				StageAlignment:      StageAlignmentInlineIDBlower,
				PackedBedLayout:     PackedBedLayoutSingleColumn,
				VenturiLayout:       VenturiLayoutSingleThroat,
				CompactFootprint:    false,
				SharedBaffledSump:   true,
				ControlTier:         ControlTierRPiModbus,
				BlowerDriveOption:   BlowerDriveIndustrialRadialVFD,
				ParallelPipeSizeIn:  4.0,
				PackingMediaOption:  PackingMediaPPPallRings,
				ShopCompressorSCFM:  15.0,
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
				SaltingHours:        3.0,
				VaporGlazeMode:      VaporGlazeSalt,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierStandard316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				StageAlignment:      StageAlignmentInlineIDBlower,
				PackedBedLayout:     PackedBedLayoutSingleColumn,
				VenturiLayout:       VenturiLayoutSingleThroat,
				CompactFootprint:    false,
				SharedBaffledSump:   true,
				ControlTier:         ControlTierRPiModbus,
				BlowerDriveOption:   BlowerDriveIndustrialRadialVFD,
				ParallelPipeSizeIn:  4.0,
				PackingMediaOption:  PackingMediaPPPallRings,
				ShopCompressorSCFM:  15.0,
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
				SaltingHours:        4.0,
				VaporGlazeMode:      VaporGlazeSalt,
				BuildPath:           BuildPathWeldedDXF,
				CorrosionTier:       CorrosionTierHeavyC276FRP,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				StageAlignment:      StageAlignmentInlineIDBlower,
				PackedBedLayout:     PackedBedLayoutSingleColumn,
				VenturiLayout:       VenturiLayoutSingleThroat,
				CompactFootprint:    false,
				SharedBaffledSump:   true,
				ControlTier:         ControlTierRPiModbus,
				BlowerDriveOption:   BlowerDriveIndustrialRadialVFD,
				ParallelPipeSizeIn:  4.0,
				PackingMediaOption:  PackingMediaPPPallRings,
				ShopCompressorSCFM:  15.0,
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
				SaltingHours:        0.0,
				VaporGlazeMode:      VaporGlazeNone,
				BuildPath:           BuildPathCOTSNoWeld,
				CorrosionTier:       CorrosionTierStandard316L,
				SnapToCOTS:          true,
				IncludeSilencer:     true,
				StageAlignment:      StageAlignmentInlineIDBlower,
				PackedBedLayout:     PackedBedLayoutSingleColumn,
				VenturiLayout:       VenturiLayoutSingleThroat,
				CompactFootprint:    false,
				SharedBaffledSump:   true,
				ControlTier:         ControlTierRPiModbus,
				BlowerDriveOption:   BlowerDriveIndustrialRadialVFD,
				ParallelPipeSizeIn:  4.0,
				PackingMediaOption:  PackingMediaPPPallRings,
				ShopCompressorSCFM:  15.0,
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

// OffGridCeramicProject returns a pre-configured 24 cu.ft. Local Ceramic / Masonry
// zero-electricity passive scrubber project for the requested region ("nairobi",
// "mumbai", or "general_local"; defaults to "nairobi").
func OffGridCeramicProject(region string) StudioProject {
	p := DefaultProject()
	p.BuildPath = BuildPathLocalCeramic
	p.ZeroElectricity = true
	p.PackingMediaOption = PackingMediaDIYCeramicRings
	p.ControlTier = ControlTierElectromechanicalOnly
	p.IncludeSilencer = false
	p.IncludePlumeDiluter = false
	switch strings.ToLower(strings.TrimSpace(region)) {
	case LocalRegionMumbai:
		p.StudioName = "Mumbai Off-Grid Ceramic Wood-Salt Studio"
		p.LocalRegion = LocalRegionMumbai
		p.LocalAlkali = LocalAlkaliSodaAsh
		p.AmbientWinterF = 75.0
	case LocalRegionGeneral:
		p.StudioName = "Local Pottery Off-Grid Ceramic Wood-Salt Studio"
		p.LocalRegion = LocalRegionGeneral
		p.LocalAlkali = LocalAlkaliSodaAsh
		p.AmbientWinterF = 55.0
	default:
		p.StudioName = "Nairobi Off-Grid Ceramic Wood-Salt Studio"
		p.LocalRegion = LocalRegionNairobi
		p.LocalAlkali = LocalAlkaliSodaAsh
		p.AmbientWinterF = 65.0
	}
	return p.Normalize()
}

// OffGridSteamCeramicProject returns a pre-configured 24 cu.ft. Local Ceramic / Masonry
// zero-electricity scrubber project equipped with a certified steam-jet ejector
// ("steam_ejector_subsonic", "steam_ejector_lp", "steam_ejector_hp", or "natural_chimney";
// empty defaults to "steam_ejector_lp").
func OffGridSteamCeramicProject(region, draftMode string) StudioProject {
	p := OffGridCeramicProject(region)
	switch strings.ToLower(strings.TrimSpace(draftMode)) {
	case LocalDraftNaturalChimney:
		p.LocalDraftMode = LocalDraftNaturalChimney
	case LocalDraftSteamEjectorSubsonic:
		p.LocalDraftMode = LocalDraftSteamEjectorSubsonic
		p.StudioName += " (Steam-Jet Subsonic 6 psig)"
	case LocalDraftSteamEjectorHP:
		p.LocalDraftMode = LocalDraftSteamEjectorHP
		p.StudioName += " (Steam-Jet HP 60 psig)"
	default:
		p.LocalDraftMode = LocalDraftSteamEjectorLP
		p.StudioName += " (Steam-Jet LP 20 psig)"
	}
	return p.Normalize()
}

// IsLocalCeramic reports whether the project runs in Phase 10 Local Ceramic / Masonry
// or 100% Zero-Electricity passive/steam-ejector draft mode.
func (p StudioProject) IsLocalCeramic() bool {
	return p.BuildPath == BuildPathLocalCeramic || p.ZeroElectricity
}

// EffectiveLocalDraftMode returns the canonical Phase 10B/10C local draft mode
// ("natural_chimney", "steam_ejector_subsonic", "steam_ejector_lp", or "steam_ejector_hp"),
// defaulting to "natural_chimney".
func (p StudioProject) EffectiveLocalDraftMode() string {
	switch p.LocalDraftMode {
	case LocalDraftSteamEjectorSubsonic, LocalDraftSteamEjectorLP, LocalDraftSteamEjectorHP, LocalDraftNaturalChimney:
		return p.LocalDraftMode
	default:
		return LocalDraftNaturalChimney
	}
}

// IsSteamEjector reports whether a Phase 10B/10C steam-jet ejector draft mode is selected.
func (p StudioProject) IsSteamEjector() bool {
	return p.LocalDraftMode == LocalDraftSteamEjectorSubsonic ||
		p.LocalDraftMode == LocalDraftSteamEjectorLP ||
		p.LocalDraftMode == LocalDraftSteamEjectorHP
}

// HasPressurizedSteam20Plus reports whether >= 20 psig steam (LP or HP) is active,
// which is required to drive a mechanical PPCP, steam-jet syphon, or submerged sump-pool scrubber.
func (p StudioProject) HasPressurizedSteam20Plus() bool {
	return p.LocalDraftMode == LocalDraftSteamEjectorLP ||
		p.LocalDraftMode == LocalDraftSteamEjectorHP
}

// EffectiveLocalBoilerSource returns the canonical Phase 13/14 boiler heat source.
func (p StudioProject) EffectiveLocalBoilerSource() string {
	switch p.LocalBoilerSource {
	case LocalBoilerKilnExhaustCoil, LocalBoilerStage1FeedwaterPreheater, LocalBoilerStage1WaterJacket, LocalBoilerDedicatedHearth:
		return p.LocalBoilerSource
	default:
		return LocalBoilerDedicatedHearth
	}
}

// EffectiveLocalPumpMode returns the canonical Phase 13 off-grid water pump / sump-pool mode.
func (p StudioProject) EffectiveLocalPumpMode() string {
	switch p.LocalPumpMode {
	case LocalPumpSteamPPCP, LocalPumpSteamJetSyphon, LocalPumpSumpPoolNoPump, LocalPumpManualTreadle:
		return p.LocalPumpMode
	default:
		return LocalPumpManualTreadle
	}
}

// EffectiveLocalRegion returns the canonical regional sourcing key ("nairobi",
// "mumbai", or "general_local"), defaulting to "nairobi".
func (p StudioProject) EffectiveLocalRegion() string {
	switch p.LocalRegion {
	case LocalRegionMumbai, LocalRegionGeneral, LocalRegionNairobi:
		return p.LocalRegion
	default:
		return LocalRegionNairobi
	}
}

// EffectiveLocalAlkali returns the canonical local alkali key ("soda_ash",
// "slaked_lime", "wood_ash_lye", or "baking_soda"), defaulting to "soda_ash".
func (p StudioProject) EffectiveLocalAlkali() string {
	switch p.LocalAlkali {
	case LocalAlkaliSlakedLime, LocalAlkaliWoodAshLye, LocalAlkaliBakingSoda, LocalAlkaliSodaAsh:
		return p.LocalAlkali
	default:
		return LocalAlkaliSodaAsh
	}
}

// Normalize fills default Phase 7 option values, normalizes legacy "hobby_316l"
// to canonical "standard_316l", and clamps salting_hours to firing_duration_hours.
func (p StudioProject) Normalize() StudioProject {
	out := p
	if out.CorrosionTier == "" || out.CorrosionTier == CorrosionTierHobby316L {
		out.CorrosionTier = CorrosionTierStandard316L
	}
	if out.StageAlignment == "" {
		out.StageAlignment = StageAlignmentInlineIDBlower
	}
	if out.PackedBedLayout == "" {
		out.PackedBedLayout = PackedBedLayoutSingleColumn
	}
	if out.VenturiLayout == "" {
		out.VenturiLayout = VenturiLayoutSingleThroat
	}
	if out.ControlTier == "" {
		out.ControlTier = ControlTierRPiModbus
	}
	if out.BlowerDriveOption == "" {
		out.BlowerDriveOption = BlowerDriveIndustrialRadialVFD
	}
	if out.ParallelPipeSizeIn == 0 {
		out.ParallelPipeSizeIn = 4.0
	}
	if out.PackingMediaOption == "" {
		out.PackingMediaOption = PackingMediaPPPallRings
	}
	if out.ShopCompressorSCFM == 0 {
		out.ShopCompressorSCFM = 15.0
	}
	if out.LocalDraftMode == "" {
		out.LocalDraftMode = LocalDraftNaturalChimney
	}
	if out.LocalBoilerSource == "" {
		out.LocalBoilerSource = LocalBoilerDedicatedHearth
	}
	if out.LocalPumpMode == "" {
		out.LocalPumpMode = LocalPumpManualTreadle
	}
	if out.BuildPath == "" && (out.IsSteamEjector() || out.ZeroElectricity) {
		out.BuildPath = BuildPathLocalCeramic
	}
	if out.IsLocalCeramic() {
		if out.LocalRegion == "" {
			out.LocalRegion = LocalRegionNairobi
		}
		if out.LocalAlkali == "" {
			out.LocalAlkali = LocalAlkaliSodaAsh
		}
	}
	if out.VaporGlazeMode != VaporGlazeSalt {
		out.SaltingHours = 0.0
	} else {
		if out.SaltingHours <= 0 {
			out.SaltingHours = math.Min(3.0, out.FiringDurationHours)
		}
		if out.FiringDurationHours > 0 && out.SaltingHours > out.FiringDurationHours {
			out.SaltingHours = out.FiringDurationHours
		}
	}
	return out
}

// EffectiveSaltingHours returns the active peak salting window in hours
// (0 for non-salt modes; clamped to (0, FiringDurationHours] for salt mode).
func (p StudioProject) EffectiveSaltingHours() float64 {
	if p.VaporGlazeMode != VaporGlazeSalt {
		return 0.0
	}
	hrs := p.SaltingHours
	if hrs <= 0 {
		hrs = 3.0
	}
	if p.FiringDurationHours > 0 && hrs > p.FiringDurationHours {
		hrs = p.FiringDurationHours
	}
	return round1(hrs)
}

// EffectiveDesignInput returns the physics DesignInput reconciled with VaporGlazeMode.
// In "soda" or "none" modes, HCl salt loading is forced to 0 lbs/hr and fuel type is
// switched from *_salt to the corresponding clean combustion mode.
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
	// Phase 8: Twin-fluid air-atomizing venturi nozzle produces 25 um SMD droplets,
	// unlocking reduced 48 m/s throat velocity with d_50 <= 0.22 um parity and ~45% lower dP.
	if !p.IsLocalCeramic() && p.StageAlignment != StageAlignmentEductorDryFan && p.BlowerDriveOption == BlowerDriveCompressorAirAtomize {
		in.VenturiDropletDiameter = model.MicronsToMeters(25.0)
		if in.VenturiThroatVelocity > 48.0 {
			in.VenturiThroatVelocity = 48.0
		}
	}
	// Phase 8: Propagate alternative packing media hydraulics into the core packed-bed solver
	// so Eckert flooding velocity, column diameter, and blower static pressure stay physically consistent.
	if !p.IsLocalCeramic() {
		switch p.PackingMediaOption {
		case PackingMediaDIYCeramicRings:
			in.PackingName = `3/4" Potter-Extruded Thin-Wall Stoneware Raschig Rings`
			in.PackingDiameter = model.InchesToMeters(0.75)
			in.PackingVoidage = 0.72
			in.PackingFactor = 620.0
		case PackingMediaAquacultureBioballs:
			in.PackingName = `1.125" High-Surface Aquaculture PP Bio-Balls`
			in.PackingDiameter = model.InchesToMeters(1.125)
			in.PackingVoidage = 0.85
			in.PackingFactor = 210.0
		}
	}
	return in
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// IsDeluxe reports whether the project uses the Phase 9 Custom Fabrication Deluxe build path.
func (p StudioProject) IsDeluxe() bool {
	return p.BuildPath == BuildPathCustomFabDeluxe
}

// Validate checks that all StudioProject parameters and its effective DesignInput are valid.
func (p StudioProject) Validate() error {
	if !isFinite(p.FiringDurationHours) || p.FiringDurationHours <= 0 || p.FiringDurationHours > 168.0 {
		return fmt.Errorf("firing_duration_hours must be finite and in (0, 168], got %v", p.FiringDurationHours)
	}
	if !isFinite(p.SaltingHours) || p.SaltingHours < 0 || p.SaltingHours > 168.0 {
		return fmt.Errorf("salting_hours must be finite and in [0, 168], got %v", p.SaltingHours)
	}
	switch p.VaporGlazeMode {
	case VaporGlazeNone, VaporGlazeSoda, VaporGlazeSalt:
	default:
		return fmt.Errorf("vapor_glaze_mode must be %q, %q, or %q, got %q",
			VaporGlazeNone, VaporGlazeSoda, VaporGlazeSalt, p.VaporGlazeMode)
	}
	switch p.BuildPath {
	case BuildPathCOTSNoWeld, BuildPathWeldedDXF, BuildPathCustomFabDeluxe, BuildPathLocalCeramic:
	case "":
		if !p.IsSteamEjector() && !p.ZeroElectricity {
			return fmt.Errorf("build_path must be %q, %q, %q, or %q, got %q",
				BuildPathCOTSNoWeld, BuildPathWeldedDXF, BuildPathCustomFabDeluxe, BuildPathLocalCeramic, p.BuildPath)
		}
	default:
		return fmt.Errorf("build_path must be %q, %q, %q, or %q, got %q",
			BuildPathCOTSNoWeld, BuildPathWeldedDXF, BuildPathCustomFabDeluxe, BuildPathLocalCeramic, p.BuildPath)
	}
	// Reject mutually contradictory cross-phase architecture combinations cleanly.
	if p.ZeroElectricity && p.BuildPath != "" && p.BuildPath != BuildPathLocalCeramic {
		return fmt.Errorf("zero_electricity=true requires build_path=%q (got %q)", BuildPathLocalCeramic, p.BuildPath)
	}
	if p.IsSteamEjector() && p.BuildPath != "" && p.BuildPath != BuildPathLocalCeramic {
		return fmt.Errorf("local_draft_mode=%q requires build_path=%q (got %q)", p.LocalDraftMode, BuildPathLocalCeramic, p.BuildPath)
	}
	if p.BuildPath == BuildPathCustomFabDeluxe && p.PackedBedLayout == PackedBedLayoutParallelManifold {
		return fmt.Errorf("build_path=%q specifies a single filament-wound FRP column with bolted manway and cannot be combined with packed_bed_layout=%q", BuildPathCustomFabDeluxe, PackedBedLayoutParallelManifold)
	}
	if p.StageAlignment == StageAlignmentEductorDryFan {
		switch p.BlowerDriveOption {
		case BlowerDriveCompressorAirAtomize, BlowerDriveCompressorEductor, BlowerDriveDualBypassBlowers:
			return fmt.Errorf("stage_alignment=%q cannot be combined with blower_drive_option=%q (use %q or %q)",
				StageAlignmentEductorDryFan, p.BlowerDriveOption, BlowerDriveIndustrialRadialVFD, BlowerDriveRegenSideChannel)
		}
		if p.VenturiLayout == VenturiLayoutParallelMiniVenturis {
			return fmt.Errorf("stage_alignment=%q replaces Stage 2 with a single downflow static eductor and cannot be combined with venturi_layout=%q",
				StageAlignmentEductorDryFan, VenturiLayoutParallelMiniVenturis)
		}
	}
	switch p.LocalRegion {
	case "", LocalRegionNairobi, LocalRegionMumbai, LocalRegionGeneral:
	default:
		return fmt.Errorf("local_region must be %q, %q, or %q, got %q",
			LocalRegionNairobi, LocalRegionMumbai, LocalRegionGeneral, p.LocalRegion)
	}
	switch p.LocalAlkali {
	case "", LocalAlkaliSodaAsh, LocalAlkaliSlakedLime, LocalAlkaliWoodAshLye, LocalAlkaliBakingSoda:
	default:
		return fmt.Errorf("local_alkali must be %q, %q, %q, or %q, got %q",
			LocalAlkaliSodaAsh, LocalAlkaliSlakedLime, LocalAlkaliWoodAshLye, LocalAlkaliBakingSoda, p.LocalAlkali)
	}
	switch p.LocalDraftMode {
	case "", LocalDraftNaturalChimney, LocalDraftSteamEjectorSubsonic, LocalDraftSteamEjectorLP, LocalDraftSteamEjectorHP:
	default:
		return fmt.Errorf("local_draft_mode must be %q, %q, %q, or %q, got %q",
			LocalDraftNaturalChimney, LocalDraftSteamEjectorSubsonic, LocalDraftSteamEjectorLP, LocalDraftSteamEjectorHP, p.LocalDraftMode)
	}
	switch p.LocalBoilerSource {
	case "", LocalBoilerDedicatedHearth, LocalBoilerKilnExhaustCoil, LocalBoilerStage1FeedwaterPreheater, LocalBoilerStage1WaterJacket:
	default:
		return fmt.Errorf("local_boiler_source must be %q, %q, %q, or %q, got %q",
			LocalBoilerDedicatedHearth, LocalBoilerKilnExhaustCoil, LocalBoilerStage1FeedwaterPreheater, LocalBoilerStage1WaterJacket, p.LocalBoilerSource)
	}
	if p.LocalBoilerSource != "" && p.LocalBoilerSource != LocalBoilerDedicatedHearth {
		if p.BuildPath != "" && p.BuildPath != BuildPathLocalCeramic {
			return fmt.Errorf("local_boiler_source=%q requires build_path=%q (got %q)", p.LocalBoilerSource, BuildPathLocalCeramic, p.BuildPath)
		}
		if !p.IsSteamEjector() {
			return fmt.Errorf("local_boiler_source=%q requires a steam-jet local_draft_mode (steam_ejector_subsonic, steam_ejector_lp, or steam_ejector_hp)", p.LocalBoilerSource)
		}
		if p.LocalBoilerSource == LocalBoilerKilnExhaustCoil && p.LocalDraftMode != LocalDraftSteamEjectorLP {
			return fmt.Errorf("local_boiler_source=%q (<2 L seamless monotube coil) is restricted strictly to 20 psig local_draft_mode=%q (banned at 60 psig per certified baby-boiler rule; use %q for 6 psig/60 psig)",
				LocalBoilerKilnExhaustCoil, LocalDraftSteamEjectorLP, LocalBoilerStage1FeedwaterPreheater)
		}
		if p.LocalBoilerSource == LocalBoilerStage1WaterJacket && !p.HasPressurizedSteam20Plus() {
			return fmt.Errorf("local_boiler_source=%q (ASME/IBR-certified <=25 L shop pressurized jacket) requires >=20 psig steam (steam_ejector_lp or steam_ejector_hp); use %q (0 psig vented economizer) for 6 psig subsonic mode",
				LocalBoilerStage1WaterJacket, LocalBoilerStage1FeedwaterPreheater)
		}
	}
	switch p.LocalPumpMode {
	case "", LocalPumpManualTreadle, LocalPumpSteamPPCP, LocalPumpSteamJetSyphon, LocalPumpSumpPoolNoPump:
	default:
		return fmt.Errorf("local_pump_mode must be %q, %q, %q, or %q, got %q",
			LocalPumpManualTreadle, LocalPumpSteamPPCP, LocalPumpSteamJetSyphon, LocalPumpSumpPoolNoPump, p.LocalPumpMode)
	}
	if p.LocalPumpMode != "" && p.LocalPumpMode != LocalPumpManualTreadle {
		if p.BuildPath != "" && p.BuildPath != BuildPathLocalCeramic {
			return fmt.Errorf("local_pump_mode=%q requires build_path=%q (got %q)", p.LocalPumpMode, BuildPathLocalCeramic, p.BuildPath)
		}
		if p.LocalPumpMode == LocalPumpSumpPoolNoPump {
			if p.LocalDraftMode != LocalDraftSteamEjectorHP {
				return fmt.Errorf("local_pump_mode=%q (co-current drenched trough + 2-stage ceramic sieve tray) requires 60 psig local_draft_mode=%q (~4.3 inWC draft; 20 psig LP only yields ~1.0 inWC)",
					LocalPumpSumpPoolNoPump, LocalDraftSteamEjectorHP)
			}
		} else if !p.HasPressurizedSteam20Plus() {
			return fmt.Errorf("local_pump_mode=%q requires >=20 psig steam (local_draft_mode=steam_ejector_lp or steam_ejector_hp; 0/6 psig cannot overcome 3.5 m head)", p.LocalPumpMode)
		}
	}
	if (p.DeluxeCondensingReclaimer || p.DeluxeVariableThroatVFD || p.DeluxeStage1HeatRecovery) && p.BuildPath != BuildPathCustomFabDeluxe {
		return fmt.Errorf("deluxe energy/water-saving upgrades require build_path=%q (got %q)", BuildPathCustomFabDeluxe, p.BuildPath)
	}
	if p.DeluxeVariableThroatVFD {
		if p.StageAlignment == StageAlignmentEductorDryFan {
			return fmt.Errorf("deluxe_variable_throat_vfd requires stage_alignment=%q (incompatible with %q)", StageAlignmentInlineIDBlower, StageAlignmentEductorDryFan)
		}
		if p.VenturiLayout == VenturiLayoutParallelMiniVenturis {
			return fmt.Errorf("deluxe_variable_throat_vfd requires venturi_layout=%q (incompatible with %q)", VenturiLayoutSingleThroat, VenturiLayoutParallelMiniVenturis)
		}
		if p.BlowerDriveOption != "" && p.BlowerDriveOption != BlowerDriveIndustrialRadialVFD && p.BlowerDriveOption != BlowerDriveCompressorAirAtomize {
			return fmt.Errorf("deluxe_variable_throat_vfd requires a VFD blower drive (%q or %q, got %q)",
				BlowerDriveIndustrialRadialVFD, BlowerDriveCompressorAirAtomize, p.BlowerDriveOption)
		}
	}
	switch p.CorrosionTier {
	case CorrosionTierStandard316L, CorrosionTierHobby316L, CorrosionTierHeavyC276FRP:
	default:
		return fmt.Errorf("corrosion_tier must be %q or %q, got %q",
			CorrosionTierStandard316L, CorrosionTierHeavyC276FRP, p.CorrosionTier)
	}
	switch p.StageAlignment {
	case "", StageAlignmentInlineIDBlower, StageAlignmentEductorDryFan:
	default:
		return fmt.Errorf("stage_alignment must be %q or %q, got %q",
			StageAlignmentInlineIDBlower, StageAlignmentEductorDryFan, p.StageAlignment)
	}
	switch p.PackedBedLayout {
	case "", PackedBedLayoutSingleColumn, PackedBedLayoutParallelManifold:
	default:
		return fmt.Errorf("packed_bed_layout must be %q or %q, got %q",
			PackedBedLayoutSingleColumn, PackedBedLayoutParallelManifold, p.PackedBedLayout)
	}
	switch p.VenturiLayout {
	case "", VenturiLayoutSingleThroat, VenturiLayoutParallelMiniVenturis:
	default:
		return fmt.Errorf("venturi_layout must be %q or %q, got %q",
			VenturiLayoutSingleThroat, VenturiLayoutParallelMiniVenturis, p.VenturiLayout)
	}
	switch p.ControlTier {
	case "", ControlTierRPiModbus, ControlTierElectromechanicalOnly:
	default:
		return fmt.Errorf("control_tier must be %q or %q, got %q",
			ControlTierRPiModbus, ControlTierElectromechanicalOnly, p.ControlTier)
	}
	switch p.BlowerDriveOption {
	case "", BlowerDriveIndustrialRadialVFD, BlowerDriveRegenSideChannel, BlowerDriveDualBypassBlowers, BlowerDriveCompressorAirAtomize, BlowerDriveCompressorEductor:
	default:
		return fmt.Errorf("blower_drive_option must be %q, %q, %q, %q, or %q, got %q",
			BlowerDriveIndustrialRadialVFD, BlowerDriveRegenSideChannel, BlowerDriveDualBypassBlowers,
			BlowerDriveCompressorAirAtomize, BlowerDriveCompressorEductor, p.BlowerDriveOption)
	}
	if !isFinite(p.ParallelPipeSizeIn) || (p.ParallelPipeSizeIn != 0 && p.ParallelPipeSizeIn != 2.0 && p.ParallelPipeSizeIn != 3.0 && p.ParallelPipeSizeIn != 4.0) {
		return fmt.Errorf("parallel_pipe_size_in must be 0, 2, 3, or 4, got %v", p.ParallelPipeSizeIn)
	}
	switch p.PackingMediaOption {
	case "", PackingMediaPPPallRings, PackingMediaAquacultureBioballs, PackingMediaDIYCeramicRings:
	default:
		return fmt.Errorf("packing_media_option must be %q, %q, or %q, got %q",
			PackingMediaPPPallRings, PackingMediaAquacultureBioballs, PackingMediaDIYCeramicRings, p.PackingMediaOption)
	}
	if !isFinite(p.ShopCompressorSCFM) || (p.ShopCompressorSCFM != 0 && (p.ShopCompressorSCFM < 1.0 || p.ShopCompressorSCFM > 200.0)) {
		return fmt.Errorf("shop_compressor_scfm must be 0 (default 15) or finite in [1, 200], got %v", p.ShopCompressorSCFM)
	}
	if !isFinite(p.PropertyLineFt) || p.PropertyLineFt < 3.0 || p.PropertyLineFt > 5000.0 {
		return fmt.Errorf("property_line_ft must be finite and in [3, 5000], got %v", p.PropertyLineFt)
	}
	if !isFinite(p.AmbientWinterF) || p.AmbientWinterF < -60.0 || p.AmbientWinterF > 120.0 {
		return fmt.Errorf("ambient_winter_f must be finite and in [-60, 120], got %v", p.AmbientWinterF)
	}
	if p.AvailableVoltage != 120 && p.AvailableVoltage != 230 && p.AvailableVoltage != 240 && !(p.ZeroElectricity && p.AvailableVoltage == 0) {
		return fmt.Errorf("available_voltage must be 120, 230, or 240 (or 0 when zero_electricity=true), got %d", p.AvailableVoltage)
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
	proj = proj.Normalize()
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
	proj = proj.Normalize()
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
