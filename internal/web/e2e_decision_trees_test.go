package web_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/internal/control"
	"github.com/AaronFilson/kilnscrub/internal/web"
	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
	"github.com/AaronFilson/kilnscrub/pkg/studio"
)

type realStudioBackendAdapter struct {
	eval *studio.Evaluator
}

func (a realStudioBackendAdapter) PresetsJSON() ([]byte, error) {
	return json.Marshal(struct {
		Presets []studio.StudioPreset `json:"presets"`
	}{Presets: studio.Presets()})
}

func (a realStudioBackendAdapter) EvaluateJSON(reqBody []byte) ([]byte, error) {
	proj, err := studio.ParseProject(reqBody)
	if err != nil {
		return nil, err
	}
	res, err := a.eval.Evaluate(proj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(res)
}

func (a realStudioBackendAdapter) BuildPacketZip(reqBody []byte) ([]byte, string, error) {
	proj, err := studio.ParseProject(reqBody)
	if err != nil {
		return nil, "", err
	}
	zipBytes, _, err := a.eval.BuildPacketZip(proj)
	if err != nil {
		return nil, "", err
	}
	return zipBytes, studio.PacketFilename(proj), nil
}

func newRealFullStackServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()
	designer := scrubber.NewDesigner()
	exporter := cad.NewExporter()
	studioEval := studio.NewEvaluator(designer, exporter)

	cfg := model.DefaultControlConfig()
	cfg.LoopInterval = 40 * time.Millisecond
	hal := control.NewSimHAL(control.DefaultSimHALConfig())
	daemon, err := control.NewDaemon(cfg, hal, control.WithHALMode("sim"))
	if err != nil {
		t.Fatalf("NewDaemon: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = daemon.Run(ctx)
	}()

	// Wait briefly for at least one telemetry tick.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !daemon.Status().UpdatedAt.IsZero() {
			break
		}
		time.Sleep(15 * time.Millisecond)
	}

	srv := web.NewServer(designer, exporter)
	srv.SetControlBackend(daemon)
	srv.SetStudioBackend(realStudioBackendAdapter{eval: studioEval})
	ts := httptest.NewServer(srv.Handler())
	cleanup := func() {
		ts.Close()
		cancel()
		<-done
	}
	return ts, cleanup
}

func postStudioEvaluate(t *testing.T, baseURL string, proj studio.StudioProject) (int, studio.StudioEvaluation, string) {
	t.Helper()
	raw, err := json.Marshal(proj)
	if err != nil {
		t.Fatalf("marshal project: %v", err)
	}
	resp, err := http.Post(baseURL+"/api/studio/evaluate", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("POST /api/studio/evaluate: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	var eval studio.StudioEvaluation
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, &eval); err != nil {
			t.Fatalf("unmarshal StudioEvaluation: %v\nbody=%s", err, string(body))
		}
	}
	return resp.StatusCode, eval, string(body)
}

func assertUnitConsistency(t *testing.T, label string, ev studio.StudioEvaluation) {
	t.Helper()
	d := ev.Design

	// 1. Mole fraction conservation across Inlet, Quenched, and Outlet gas streams.
	for name, g := range map[string]model.GasComposition{
		"inlet":    d.InletGas,
		"quenched": d.QuenchedGas,
		"outlet":   d.OutletGas,
	} {
		if math.Abs(g.SumMoleFractions()-1.0) > 1e-6 {
			t.Errorf("%s %s SumMoleFractions = %.9f, want 1.0", label, name, g.SumMoleFractions())
		}
		if g.VolumetricFlow <= 0 || g.VolumetricFlowACFM <= 0 {
			t.Errorf("%s %s invalid gas flow: m3/s=%v acfm=%v", label, name, g.VolumetricFlow, g.VolumetricFlowACFM)
		}
		wantACFM := g.VolumetricFlow.CFM()
		if math.Abs(g.VolumetricFlowACFM-wantACFM) > 1e-2 {
			t.Errorf("%s %s ACFM mismatch: %.4f vs %.4f", label, name, g.VolumetricFlowACFM, wantACFM)
		}
	}

	// 2. Pressure Pa <-> inWC exact NIST 4 C conversion (249.08891 Pa/inWC).
	if math.Abs(d.Blower.TotalStaticPressureInWC-d.Blower.TotalStaticPressure.InWC()) > 1e-3 {
		t.Errorf("%s blower Pa/inWC mismatch: %.4f vs %.4f", label, d.Blower.TotalStaticPressureInWC, d.Blower.TotalStaticPressure.InWC())
	}
	sumPa := d.Blower.Breakdown.TotalPa()
	if math.Abs(float64(sumPa-d.Blower.TotalStaticPressure)) > 1.0 {
		t.Errorf("%s pressure breakdown sum %.2f Pa != total %.2f Pa", label, sumPa, d.Blower.TotalStaticPressure)
	}

	// 3. Packed-bed flooding limit <= 75% (both design and COTS snapped).
	if d.PackedBed.PercentOfFlood <= 0 || d.PackedBed.PercentOfFlood > 75.01 {
		t.Errorf("%s design PercentOfFlood = %.2f out of (0, 75]", label, d.PackedBed.PercentOfFlood)
	}
	if ev.COTS.ActualPercentOfFlood <= 0 || ev.COTS.ActualPercentOfFlood > 75.01 {
		t.Errorf("%s COTS ActualPercentOfFlood = %.2f out of (0, 75]", label, ev.COTS.ActualPercentOfFlood)
	}

	// 4. Local ceramic unit parity: ChimneyHeightM <-> ChimneyHeightFt, TankLiters <-> TankGal,
	// SprayFlowGPM <-> SprayFlowLPM, and Frugal Cost == COTS Total Cost.
	if ev.Frugal != nil && ev.Frugal.Enabled {
		fr := ev.Frugal
		if math.Abs(fr.EstimatedLocalCostUSD-ev.COTS.EstimatedTotalUSD) > 0.02 {
			t.Errorf("%s Frugal cost %.2f != COTS cost %.2f", label, fr.EstimatedLocalCostUSD, ev.COTS.EstimatedTotalUSD)
		}
		wantChimFt := math.Round(model.Meters(fr.ChimneyHeightM).Feet()*10.0) / 10.0
		if math.Abs(fr.ChimneyHeightFt-wantChimFt) > 0.11 {
			t.Errorf("%s chimney m->ft mismatch: %.1f m -> %.1f ft (want %.1f ft)", label, fr.ChimneyHeightM, fr.ChimneyHeightFt, wantChimFt)
		}
		const litersPerGal = model.CubicMetersPerUSGallon * 1000.0
		if fr.HasOverheadTank && fr.GravityTankLiters > 0 {
			wantGal := math.Round((fr.GravityTankLiters/litersPerGal)*10.0) / 10.0
			if math.Abs(fr.GravityTankGal-wantGal) > 0.15 {
				t.Errorf("%s tank L->gal mismatch: %.1f L -> %.1f gal (want %.1f gal)", label, fr.GravityTankLiters, fr.GravityTankGal, wantGal)
			}
		}
		wantLPM := math.Round(fr.SprayFlowGPM*litersPerGal*100.0) / 100.0
		if math.Abs(fr.SprayFlowLPM-wantLPM) > 0.03 {
			t.Errorf("%s spray GPM->LPM mismatch: %.2f GPM -> %.2f LPM (want %.2f LPM)", label, fr.SprayFlowGPM, fr.SprayFlowLPM, wantLPM)
		}
	}
}

func TestE2EDecisionTreesAndUnitConsistency(t *testing.T) {
	ts, cleanup := newRealFullStackServer(t)
	defer cleanup()

	base := studio.DefaultProject()

	// --- Tree A: All 5 Curated Studio Presets ---
	for _, pr := range studio.Presets() {
		code, ev, body := postStudioEvaluate(t, ts.URL, pr.Project)
		if code != http.StatusOK {
			t.Fatalf("preset %s: status=%d body=%s", pr.ID, code, body)
		}
		assertUnitConsistency(t, "preset/"+pr.ID, ev)
	}

	// --- Tree B: All 10 Fuel Types x Vapor Glaze Modes ---
	fuelCases := []struct {
		fuel  string
		glaze string
		salt  float64
	}{
		{model.FuelWoodSalt, studio.VaporGlazeSalt, 2.0},
		{model.FuelWood, studio.VaporGlazeNone, 0.0},
		{model.FuelGasSalt, studio.VaporGlazeSalt, 1.5},
		{model.FuelGas, studio.VaporGlazeSoda, 0.0},
		{model.FuelCoalSalt, studio.VaporGlazeSalt, 2.0},
		{model.FuelCoal, studio.VaporGlazeNone, 0.0},
		{model.FuelPeatSalt, studio.VaporGlazeSalt, 1.5},
		{model.FuelPeat, studio.VaporGlazeNone, 0.0},
		{model.FuelBiomassAgSalt, studio.VaporGlazeSalt, 1.5},
		{model.FuelBiomassAg, studio.VaporGlazeNone, 0.0},
	}
	for _, fc := range fuelCases {
		p := base
		p.VaporGlazeMode = fc.glaze
		p.Input.FuelType = fc.fuel
		p.Input.SaltRateLbsPerHr = fc.salt
		code, ev, body := postStudioEvaluate(t, ts.URL, p)
		if code != http.StatusOK {
			t.Fatalf("fuel %s/%s: status=%d body=%s", fc.fuel, fc.glaze, code, body)
		}
		assertUnitConsistency(t, "fuel/"+fc.fuel, ev)
	}

	// --- Tree C: Industrial Build Paths, Stage Alignments, Drives, Packing & Deluxe Eco ---
	industrialCases := []struct {
		name   string
		mutate func(*studio.StudioProject)
	}{
		{"cots_budget_combo", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCOTSNoWeld
			p.PackedBedLayout = studio.PackedBedLayoutParallelManifold
			p.ParallelPipeSizeIn = 4
			p.PackingMediaOption = studio.PackingMediaDIYCeramicRings
			p.BlowerDriveOption = studio.BlowerDriveCompressorAirAtomize
			p.CompactFootprint = true
			p.ControlTier = studio.ControlTierElectromechanicalOnly
		}},
		{"welded_eductor_dry_fan", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathWeldedDXF
			p.StageAlignment = studio.StageAlignmentEductorDryFan
			p.BlowerDriveOption = studio.BlowerDriveRegenSideChannel
			p.PackingMediaOption = studio.PackingMediaAquacultureBioballs
			p.IncludePlumeDiluter = true
		}},
		{"cots_parallel_mini_venturis_dual_bypass", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCOTSNoWeld
			p.VenturiLayout = studio.VenturiLayoutParallelMiniVenturis
			p.BlowerDriveOption = studio.BlowerDriveDualBypassBlowers
			p.PackedBedLayout = studio.PackedBedLayoutParallelManifold
			p.ParallelPipeSizeIn = 3
		}},
		{"deluxe_all_eco_upgrades", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCustomFabDeluxe
			p.CorrosionTier = studio.CorrosionTierHeavyC276FRP
			p.DeluxeCondensingReclaimer = true
			p.DeluxeVariableThroatVFD = true
			p.DeluxeStage1HeatRecovery = true
		}},
		{"deluxe_eductor_without_var_throat", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCustomFabDeluxe
			p.StageAlignment = studio.StageAlignmentEductorDryFan
			p.DeluxeCondensingReclaimer = true
			p.DeluxeStage1HeatRecovery = true
		}},
	}
	for _, ic := range industrialCases {
		p := base
		ic.mutate(&p)
		code, ev, body := postStudioEvaluate(t, ts.URL, p)
		if code != http.StatusOK {
			t.Fatalf("industrial %s: status=%d body=%s", ic.name, code, body)
		}
		assertUnitConsistency(t, "industrial/"+ic.name, ev)
	}

	// --- Tree D: Global Local Ceramic / Off-Grid Matrix (10 Regions x Draft x Boiler x Pump x Alkali) ---
	localCases := []struct {
		name string
		proj studio.StudioProject
	}{
		{"nairobi_natural_soda", studio.OffGridCeramicProject(studio.LocalRegionNairobi)},
		{"mumbai_6psig_preheater", func() studio.StudioProject {
			p := studio.OffGridSteamCeramicProject(studio.LocalRegionMumbai, studio.LocalDraftSteamEjectorSubsonic)
			p.LocalBoilerSource = studio.LocalBoilerStage1FeedwaterPreheater
			return p
		}()},
		{"khurja_coal_60psig_sump_pool", func() studio.StudioProject {
			p := studio.GlobalUrbanSolidFuelProject(studio.LocalRegionKhurjaDhaka, model.FuelCoalSalt, studio.LocalDraftSteamEjectorHP)
			p.LocalPumpMode = studio.LocalPumpSumpPoolNoPump
			p.LocalBoilerSource = studio.LocalBoilerStage1WaterJacket
			return p
		}()},
		{"cusco_3400m_peat_20psig_coil", func() studio.StudioProject {
			p := studio.GlobalUrbanSolidFuelProject(studio.LocalRegionCuscoAndes, model.FuelPeat, studio.LocalDraftSteamEjectorLP)
			p.LocalBoilerSource = studio.LocalBoilerKilnExhaustCoil
			p.LocalPumpMode = studio.LocalPumpSteamPPCP
			return p
		}()},
		{"lagos_pks_trona_syphon", func() studio.StudioProject {
			p := studio.GlobalUrbanSolidFuelProject(studio.LocalRegionLagosAccra, model.FuelBiomassAgSalt, studio.LocalDraftSteamEjectorLP)
			p.LocalAlkali = studio.LocalAlkaliTronaKanwa
			p.LocalPumpMode = studio.LocalPumpSteamJetSyphon
			return p
		}()},
		{"kathmandu_coal_6psig", studio.GlobalUrbanSolidFuelProject(studio.LocalRegionKathmandu, model.FuelCoal, studio.LocalDraftSteamEjectorSubsonic)},
		{"tonala_wood_afterburner_trona", studio.GlobalUrbanSolidFuelProject(studio.LocalRegionTonalaPuebla, model.FuelWood, studio.LocalDraftNaturalChimney)},
		{"cairo_fez_pomace_hybrid_230v", func() studio.StudioProject {
			p := studio.GlobalUrbanSolidFuelProject(studio.LocalRegionCairoFez, model.FuelBiomassAg, studio.LocalDraftNaturalChimney)
			p.ZeroElectricity = false
			p.AvailableVoltage = 230
			return p
		}()},
		{"highland_peat_salt_woodash", studio.GlobalUrbanSolidFuelProject(studio.LocalRegionHighlandPeat, model.FuelPeatSalt, studio.LocalDraftNaturalChimney)},
		{"general_local_baking_soda", func() studio.StudioProject {
			p := studio.OffGridCeramicProject(studio.LocalRegionGeneral)
			p.LocalAlkali = studio.LocalAlkaliBakingSoda
			return p
		}()},
	}
	for _, lc := range localCases {
		code, ev, body := postStudioEvaluate(t, ts.URL, lc.proj)
		if code != http.StatusOK {
			t.Fatalf("local %s: status=%d body=%s", lc.name, code, body)
		}
		assertUnitConsistency(t, "local/"+lc.name, ev)
	}

	// --- Mutual-Exclusivity Rejection Rules R1-R10 Must Return HTTP 400 ---
	invalidCases := []struct {
		name   string
		mutate func(*studio.StudioProject)
	}{
		{"R1_zero_elec_on_cots", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCOTSNoWeld
			p.ZeroElectricity = true
		}},
		{"R2_zero_voltage_without_zero_elec", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathLocalCeramic
			p.ZeroElectricity = false
			p.AvailableVoltage = 0
		}},
		{"R3_deluxe_parallel_manifold", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCustomFabDeluxe
			p.PackedBedLayout = studio.PackedBedLayoutParallelManifold
		}},
		{"R4_deluxe_eco_on_cots", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCOTSNoWeld
			p.DeluxeCondensingReclaimer = true
		}},
		{"R5_deluxe_var_throat_on_eductor", func(p *studio.StudioProject) {
			p.BuildPath = studio.BuildPathCustomFabDeluxe
			p.StageAlignment = studio.StageAlignmentEductorDryFan
			p.DeluxeVariableThroatVFD = true
		}},
		{"R6_eductor_with_mini_venturis", func(p *studio.StudioProject) {
			p.StageAlignment = studio.StageAlignmentEductorDryFan
			p.VenturiLayout = studio.VenturiLayoutParallelMiniVenturis
		}},
		{"R7_coil_banned_on_60psig", func(p *studio.StudioProject) {
			*p = studio.OffGridSteamCeramicProject(studio.LocalRegionMumbai, studio.LocalDraftSteamEjectorHP)
			p.LocalBoilerSource = studio.LocalBoilerKilnExhaustCoil
		}},
		{"R8_water_jacket_banned_on_6psig", func(p *studio.StudioProject) {
			*p = studio.OffGridSteamCeramicProject(studio.LocalRegionMumbai, studio.LocalDraftSteamEjectorSubsonic)
			p.LocalBoilerSource = studio.LocalBoilerStage1WaterJacket
		}},
		{"R9_ppcp_banned_on_6psig", func(p *studio.StudioProject) {
			*p = studio.OffGridSteamCeramicProject(studio.LocalRegionMumbai, studio.LocalDraftSteamEjectorSubsonic)
			p.LocalPumpMode = studio.LocalPumpSteamPPCP
		}},
		{"R10_sump_pool_banned_on_20psig", func(p *studio.StudioProject) {
			*p = studio.OffGridSteamCeramicProject(studio.LocalRegionMumbai, studio.LocalDraftSteamEjectorLP)
			p.LocalPumpMode = studio.LocalPumpSumpPoolNoPump
		}},
	}
	for _, bad := range invalidCases {
		p := base
		bad.mutate(&p)
		code, _, body := postStudioEvaluate(t, ts.URL, p)
		if code != http.StatusBadRequest {
			t.Errorf("invalid case %s: got HTTP %d, want 400; body=%s", bad.name, code, body)
		}
	}

	// --- End-to-End Build Packet (.zip) File Counts (12 electric, 11 electromechanical, 8 local) ---
	packetChecks := []struct {
		name      string
		proj      studio.StudioProject
		wantFiles int
	}{
		{"electric_rpi", base, 12},
		{"electric_hardwired", func() studio.StudioProject {
			p := base
			p.ControlTier = studio.ControlTierElectromechanicalOnly
			return p
		}(), 11},
		{"local_khurja", studio.GlobalUrbanSolidFuelProject(studio.LocalRegionKhurjaDhaka, model.FuelCoalSalt, studio.LocalDraftSteamEjectorHP), 8},
	}
	for _, pc := range packetChecks {
		raw, _ := json.Marshal(pc.proj)
		resp, err := http.Post(ts.URL+"/api/studio/packet", "application/json", bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("packet %s: %v", pc.name, err)
		}
		zipBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("packet %s: status=%d body=%s", pc.name, resp.StatusCode, string(zipBytes))
		}
		zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
		if err != nil {
			t.Fatalf("packet %s invalid zip: %v", pc.name, err)
		}
		if len(zr.File) != pc.wantFiles {
			t.Errorf("packet %s has %d files, want %d", pc.name, len(zr.File), pc.wantFiles)
		}
	}

	// --- Live Control Telemetry & Fault Injection / Reset Round-Trip ---
	resp, err := http.Post(ts.URL+"/api/control/fault", "application/json", bytes.NewReader([]byte(`{"action":"water_loss"}`)))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("control water_loss fault failed: err=%v status=%d", err, resp.StatusCode)
	}
	resp.Body.Close()
	time.Sleep(120 * time.Millisecond)

	stResp, err := http.Get(ts.URL + "/api/control/status")
	if err != nil {
		t.Fatalf("GET /api/control/status: %v", err)
	}
	var stPayload struct {
		Enabled bool                `json:"enabled"`
		Status  model.ControlStatus `json:"status"`
	}
	_ = json.NewDecoder(stResp.Body).Decode(&stPayload)
	stResp.Body.Close()
	if !stPayload.Enabled || !stPayload.Status.InterlockActive {
		t.Errorf("expected live control interlock trip after water_loss, got enabled=%v interlock=%v", stPayload.Enabled, stPayload.Status.InterlockActive)
	}

	// Clear fault and reset interlock.
	resp, _ = http.Post(ts.URL+"/api/control/fault", "application/json", bytes.NewReader([]byte(`{"action":"clear_all"}`)))
	resp.Body.Close()
	time.Sleep(120 * time.Millisecond)
	resp, err = http.Post(ts.URL+"/api/control/reset", "application/json", nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("control reset failed: err=%v status=%d", err, resp.StatusCode)
	}
	resp.Body.Close()
}
