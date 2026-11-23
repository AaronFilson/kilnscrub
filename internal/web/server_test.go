package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/internal/web"
	"github.com/AaronFilson/kilnscrub/pkg/model"
)

type mockDesigner struct {
	mu        sync.Mutex
	lastInput model.DesignInput
	callCount int
	err       error
}

func (m *mockDesigner) Design(input model.DesignInput) (model.ScrubberDesign, error) {
	m.mu.Lock()
	m.lastInput = input
	m.callCount++
	err := m.err
	m.mu.Unlock()

	if err != nil {
		return model.ScrubberDesign{}, err
	}

	draftMagInWC := -input.TargetKilnDraftInWC
	breakdown := model.PressureDropBreakdown{
		Duct:          model.InWCToPascal(0.8),
		DuctInWC:      0.8,
		Quench:        model.InWCToPascal(1.2),
		QuenchInWC:    1.2,
		Venturi:       model.InWCToPascal(18.5),
		VenturiInWC:   18.5,
		PackedBed:     model.InWCToPascal(2.4),
		PackedBedInWC: 2.4,
		Demister:      model.InWCToPascal(0.6),
		DemisterInWC:  0.6,
		KilnDraft:     model.InWCToPascal(draftMagInWC),
		KilnDraftInWC: draftMagInWC,
	}
	totalPa := breakdown.TotalPa()
	totalInWC := totalPa.InWC()

	return model.ScrubberDesign{
		Input: input,
		InletGas: model.GasComposition{
			Temperature:        model.FahrenheitToKelvin(input.PeakExitTempF),
			VolumetricFlowACFM: 950.0,
		},
		QuenchedGas: model.GasComposition{
			Temperature:        model.FahrenheitToKelvin(165.0),
			VolumetricFlowACFM: 420.0,
		},
		OutletGas: model.GasComposition{
			Temperature:        model.FahrenheitToKelvin(160.0),
			VolumetricFlowACFM: 410.0,
		},
		Quench: model.QuenchResult{
			InletTemp:          model.FahrenheitToKelvin(input.PeakExitTempF),
			AdiabaticSatTemp:   model.FahrenheitToKelvin(164.5),
			OutletTemp:         model.FahrenheitToKelvin(165.2),
			DewPoint:           model.FahrenheitToKelvin(158.0),
			EvaporatedWaterGPM: 1.15,
			TotalSprayWaterGPM: 2.30,
			TowerDiameter:      0.45,
			TowerHeight:        1.35,
			NozzleCount:        2,
			PressureDrop:       breakdown.Quench,
			PressureDropInWC:   breakdown.QuenchInWC,
			Material:           "316L Stainless Steel",
		},
		Venturi: model.VenturiResult{
			InletDiameter:                0.22,
			ThroatDiameter:               0.085,
			OutletDiameter:               0.24,
			ConvergingLength:             0.28,
			ThroatLength:                 0.18,
			DivergingLength:              0.55,
			ThroatVelocity:               input.VenturiThroatVelocity,
			LGRatio:                      input.VenturiLGRatio,
			LiquidFlowGPM:                3.6,
			CutDiameterD50:               model.MicronsToMeters(0.42),
			CutDiameterD50Microns:        0.42,
			WoodAshCutDiameterD50:        model.MicronsToMeters(0.39),
			WoodAshCutDiameterD50Microns: 0.39,
			EfficiencyCurve: []model.EfficiencyPoint{
				{ParticleDiameter: 0.1e-6, ParticleDiameterMicrons: 0.1, WoodAshEfficiency: 0.12, NaClFumeEfficiency: 0.10},
				{ParticleDiameter: 0.5e-6, ParticleDiameterMicrons: 0.5, WoodAshEfficiency: 0.64, NaClFumeEfficiency: 0.58},
				{ParticleDiameter: 1.0e-6, ParticleDiameterMicrons: 1.0, WoodAshEfficiency: 0.94, NaClFumeEfficiency: 0.91},
				{ParticleDiameter: 2.0e-6, ParticleDiameterMicrons: 2.0, WoodAshEfficiency: 0.99, NaClFumeEfficiency: 0.98},
			},
			OverallWoodAshEfficiency:  0.93,
			OverallNaClFumeEfficiency: 0.87,
			PressureDrop:              breakdown.Venturi,
			PressureDropInWC:          breakdown.VenturiInWC,
			Material:                  "316L Stainless Steel",
		},
		PackedBed: model.PackedBedResult{
			ColumnDiameter:         0.52,
			ColumnHeight:           2.35,
			FloodingVelocity:       2.10,
			SuperficialGasVelocity: 1.28,
			PercentOfFlood:         60.95, // 0..100 scale
			PackingHeight:          1.15,
			NTU:                    4.61,
			HTU:                    0.25,
			PackingSpec:            input.PackingName,
			PackingDiameter:        input.PackingDiameter,
			PackingVoidage:         input.PackingVoidage,
			HClInletPPM:            1400.0,
			HClOutletPPM:           14.0,
			HClRemovalEfficiency:   input.TargetHClRemoval, // 0..1 scale
			RecirculationFlowGPM:   8.5,
			SumpPH:                 input.SumpPHTarget,
			SumpVolume:             0.18,
			BedPressureDrop:        breakdown.PackedBed,
			DemisterPressureDrop:   breakdown.Demister,
			DemisterHeight:         0.15,
			TotalPressureDrop:      breakdown.PackedBed + breakdown.Demister,
			TotalPressureDropInWC:  breakdown.PackedBedInWC + breakdown.DemisterInWC,
			Material:               "CPVC / Polypropylene",
		},
		Blower: model.BlowerResult{
			Breakdown:               breakdown,
			TargetKilnDraftInWC:     input.TargetKilnDraftInWC,
			TotalStaticPressure:     totalPa,
			TotalStaticPressureInWC: totalInWC,
			Flow:                    model.CFMToCubicMetersPerSec(410.0),
			FlowACFM:                410.0,
			Efficiency:              0.62, // 0..1 scale
			ShaftPowerHP:            2.42,
			RecommendedMotorHP:      3.0,
			WheelDiameter:           0.38,
			OperatingRPM:            3450.0,
			OperatingPoint: model.CurvePoint{
				Flow:         model.CFMToCubicMetersPerSec(410.0),
				FlowACFM:     410.0,
				Pressure:     totalPa,
				PressureInWC: totalInWC,
			},
			FanCurve: []model.CurvePoint{
				{FlowACFM: 0, PressureInWC: totalInWC * 1.25},
				{FlowACFM: 410, PressureInWC: totalInWC},
				{FlowACFM: 650, PressureInWC: 0},
			},
			SystemCurve: []model.CurvePoint{
				{FlowACFM: 0, PressureInWC: draftMagInWC},
				{FlowACFM: 410, PressureInWC: totalInWC},
				{FlowACFM: 550, PressureInWC: totalInWC * 1.8},
			},
			Material: "FRP / Coated Carbon Steel",
		},
		Envelope: model.EnvelopeDimensions{
			Length:       2.8,
			Width:        1.1,
			Height:       2.65,
			DuctDiameter: 0.20,
		},
	}, nil
}

type mockExporter struct {
	stlData []byte
	dxfData []byte
	csvData []byte
	mdData  []byte
	bom     model.BOM
	err     error
}

func newMockExporter() *mockExporter {
	bom := model.BOM{
		Title: "kilnscrub Mock BOM",
		Items: []model.BOMItem{
			{
				Category:    "Stage 1 Quench",
				Name:        "Quench Shell",
				Material:    "316L SS",
				Quantity:    1,
				Unit:        "ea",
				UnitCostUSD: 450.0,
				WeightKg:    32.5,
			},
		},
	}
	bom.RecalculateTotals()

	return &mockExporter{
		stlData: append([]byte("solid kilnscrub-binary-mock"), make([]byte, 64)...),
		dxfData: []byte("0\nSECTION\n2\nENTITIES\n0\nENDSEC\n0\nEOF\n"),
		csvData: []byte("Category,Name,Material,Quantity,Unit,UnitCostUSD,TotalCostUSD,WeightKg\nStage 1 Quench,Quench Shell,316L SS,1,ea,450.00,450.00,32.50\n"),
		mdData:  []byte("# kilnscrub Mock BOM\n\n| Item | Material | Qty |\n|---|---|---|\n| Quench Shell | 316L SS | 1 |\n"),
		bom:     bom,
	}
}

func (m *mockExporter) ExportSTL(model.ScrubberDesign) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stlData, nil
}

func (m *mockExporter) ExportDXF(model.ScrubberDesign) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.dxfData, nil
}

func (m *mockExporter) GenerateBOM(model.ScrubberDesign) (model.BOM, error) {
	if m.err != nil {
		return model.BOM{}, m.err
	}
	return m.bom, nil
}

func (m *mockExporter) FormatBOMCSV(model.BOM) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.csvData, nil
}

func (m *mockExporter) FormatBOMMarkdown(model.BOM) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.mdData, nil
}

func TestIndexAndEmbeddedHTMLSanity(t *testing.T) {
	srv := web.NewServer(&mockDesigner{}, newMockExporter())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("GET / Content-Type = %q, want text/html", ct)
	}

	body := rec.Body.String()
	requiredSnippets := []string{
		"<!DOCTYPE html>",
		"three@0.170.0",
		"OrbitControls",
		"/api/defaults",
		"/api/size",
		"/api/export",
		"Cutaway",
		"Flow Particles",
		"viewport-3d",
		"viewport-2d",
		"chart-efficiency",
		"chart-fan",
		"chart-dp",
		"percent_of_flood",
		"hcl_removal_efficiency",
		"target_kiln_draft_in_wc",
	}
	for _, needle := range requiredSnippets {
		if !strings.Contains(body, needle) {
			t.Errorf("embedded index.html missing expected marker %q", needle)
		}
	}

	// Unknown route must return 404, not index.html.
	rec404 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec404, httptest.NewRequest(http.MethodGet, "/unknown", nil))
	if rec404.Code != http.StatusNotFound {
		t.Fatalf("GET /unknown status = %d, want 404", rec404.Code)
	}

	// Wrong method on / must return 405.
	rec405 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec405, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec405.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / status = %d, want 405", rec405.Code)
	}
}

func TestHealthAndDefaults(t *testing.T) {
	srv := web.NewServer(&mockDesigner{}, newMockExporter())

	// GET /api/health
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/health status = %d, want 200", rec.Code)
	}
	var health map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("unmarshal /api/health: %v", err)
	}
	if health["status"] != "ok" {
		t.Fatalf("health status = %q, want ok", health["status"])
	}

	// GET /api/defaults
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/defaults", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/defaults status = %d, want 200", rec.Code)
	}
	var got model.DesignInput
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal /api/defaults: %v", err)
	}
	want := model.DefaultSaltKilnInput()
	if got != want {
		t.Fatalf("GET /api/defaults = %+v, want %+v", got, want)
	}

	// Method checks
	for _, path := range []string{"/api/health", "/api/defaults"} {
		rec = httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s status = %d, want 405", path, rec.Code)
		}
	}
}

func TestSizeEndpoint(t *testing.T) {
	defaults := model.DefaultSaltKilnInput()

	t.Run("full JSON payload", func(t *testing.T) {
		designer := &mockDesigner{}
		srv := web.NewServer(designer, newMockExporter())

		full := defaults
		full.KilnVolumeCuFt = 40.0
		full.PeakExitTempF = 2300.0
		full.SaltRateLbsPerHr = 3.5
		full.FuelType = model.FuelGasSalt
		payload, _ := json.Marshal(full)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/size", bytes.NewReader(payload)))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		var out model.ScrubberDesign
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("unmarshal ScrubberDesign: %v", err)
		}
		if designer.lastInput != full {
			t.Fatalf("designer input = %+v, want %+v", designer.lastInput, full)
		}
		if out.Input.KilnVolumeCuFt != 40.0 || out.Blower.TargetKilnDraftInWC != defaults.TargetKilnDraftInWC {
			t.Fatalf("unexpected design response: %+v", out.Input)
		}
	})

	t.Run("partial JSON inherits DefaultSaltKilnInput", func(t *testing.T) {
		designer := &mockDesigner{}
		srv := web.NewServer(designer, newMockExporter())

		partial := `{"kiln_volume_cu_ft": 36, "peak_exit_temp_f": 2250, "salt_rate_lbs_per_hr": 1.5}`
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/size", strings.NewReader(partial)))
		if rec.Code != http.StatusOK {
			t.Fatalf("partial JSON status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}

		got := designer.lastInput
		if got.KilnVolumeCuFt != 36 || got.PeakExitTempF != 2250 || got.SaltRateLbsPerHr != 1.5 {
			t.Fatalf("overridden fields not applied: %+v", got)
		}
		// Verify critical omitted fields inherited valid defaults instead of zero values.
		if got.PackingFactor != defaults.PackingFactor {
			t.Errorf("PackingFactor = %v, want default %v", got.PackingFactor, defaults.PackingFactor)
		}
		if got.AmbientTempK != defaults.AmbientTempK {
			t.Errorf("AmbientTempK = %v, want default %v", got.AmbientTempK, defaults.AmbientTempK)
		}
		if got.AmbientPressurePa != defaults.AmbientPressurePa {
			t.Errorf("AmbientPressurePa = %v, want default %v", got.AmbientPressurePa, defaults.AmbientPressurePa)
		}
		if got.SprayWaterTempK != defaults.SprayWaterTempK {
			t.Errorf("SprayWaterTempK = %v, want default %v", got.SprayWaterTempK, defaults.SprayWaterTempK)
		}
		if got.TargetHClRemoval != defaults.TargetHClRemoval {
			t.Errorf("TargetHClRemoval = %v, want default %v", got.TargetHClRemoval, defaults.TargetHClRemoval)
		}
		if got.TargetKilnDraftInWC != defaults.TargetKilnDraftInWC {
			t.Errorf("TargetKilnDraftInWC = %v, want default %v", got.TargetKilnDraftInWC, defaults.TargetKilnDraftInWC)
		}
	})

	t.Run("empty object inherits all defaults", func(t *testing.T) {
		designer := &mockDesigner{}
		srv := web.NewServer(designer, newMockExporter())

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/size", strings.NewReader("{}")))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if designer.lastInput != defaults {
			t.Fatalf("input = %+v, want defaults %+v", designer.lastInput, defaults)
		}
	})

	badPayloads := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed JSON", "{bad json"},
		{"JSON null", "null"},
		{"JSON array", "[1,2,3]"},
		{"wrong field type", `{"kiln_volume_cu_ft":"twenty"}`},
		{"negative kiln volume", `{"kiln_volume_cu_ft":-10}`},
		{"positive kiln draft", `{"target_kiln_draft_in_wc":0.15}`},
		{"invalid fuel type", `{"fuel_type":"coal"}`},
		{"target hcl removal >= 1", `{"target_hcl_removal":1.0}`},
		{"packing factor <= 0", `{"packing_factor_1_m":0}`},
	}

	for _, tc := range badPayloads {
		t.Run("bad input: "+tc.name, func(t *testing.T) {
			designer := &mockDesigner{}
			srv := web.NewServer(designer, newMockExporter())

			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/size", strings.NewReader(tc.body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
			var errResp map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil || errResp["error"] == "" {
				t.Fatalf("expected JSON error body, got %q", rec.Body.String())
			}
			if designer.callCount != 0 {
				t.Fatalf("designer should not be called on invalid input, callCount=%d", designer.callCount)
			}
		})
	}

	t.Run("designer error returns 500", func(t *testing.T) {
		designer := &mockDesigner{err: errors.New("solver diverged")}
		srv := web.NewServer(designer, newMockExporter())

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/size", strings.NewReader("{}")))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})

	t.Run("GET method not allowed", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/size", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})
}

func TestExportEndpoint(t *testing.T) {
	exporter := newMockExporter()
	defaults := model.DefaultSaltKilnInput()

	cases := []struct {
		format       string
		wantCT       string
		wantFilename string
		checkBody    func(t *testing.T, body []byte)
	}{
		{
			format:       "stl",
			wantCT:       "model/stl",
			wantFilename: "kilnscrub.stl",
			checkBody: func(t *testing.T, b []byte) {
				if !bytes.Equal(b, exporter.stlData) {
					t.Fatalf("STL payload mismatch")
				}
			},
		},
		{
			format:       "dxf",
			wantCT:       "application/dxf",
			wantFilename: "kilnscrub.dxf",
			checkBody: func(t *testing.T, b []byte) {
				if !bytes.Contains(b, []byte("ENTITIES")) || !bytes.Contains(b, []byte("EOF")) {
					t.Fatalf("unexpected DXF payload: %q", string(b))
				}
			},
		},
		{
			format:       "bom_csv",
			wantCT:       "text/csv",
			wantFilename: "kilnscrub-bom.csv",
			checkBody: func(t *testing.T, b []byte) {
				if !bytes.Equal(b, exporter.csvData) {
					t.Fatalf("CSV payload mismatch")
				}
			},
		},
		{
			format:       "bom_md",
			wantCT:       "text/markdown",
			wantFilename: "kilnscrub-bom.md",
			checkBody: func(t *testing.T, b []byte) {
				if !bytes.Equal(b, exporter.mdData) {
					t.Fatalf("Markdown payload mismatch")
				}
			},
		},
		{
			format:       "bom_json",
			wantCT:       "application/json",
			wantFilename: "kilnscrub-bom.json",
			checkBody: func(t *testing.T, b []byte) {
				var got model.BOM
				if err := json.Unmarshal(b, &got); err != nil {
					t.Fatalf("unmarshal bom_json: %v", err)
				}
				if got.Title != exporter.bom.Title || got.TotalCostUSD != exporter.bom.TotalCostUSD {
					t.Fatalf("bom_json mismatch: %+v", got)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run("query format="+tc.format, func(t *testing.T) {
			designer := &mockDesigner{}
			srv := web.NewServer(designer, exporter)

			partial := `{"kiln_volume_cu_ft": 32}`
			req := httptest.NewRequest(http.MethodPost, "/api/export?format="+tc.format, strings.NewReader(partial))
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, tc.wantCT) {
				t.Errorf("Content-Type = %q, want substring %q", ct, tc.wantCT)
			}
			cd := rec.Header().Get("Content-Disposition")
			if !strings.Contains(cd, "attachment") || !strings.Contains(cd, tc.wantFilename) {
				t.Errorf("Content-Disposition = %q, want attachment with %q", cd, tc.wantFilename)
			}
			if designer.lastInput.KilnVolumeCuFt != 32 || designer.lastInput.PackingFactor != defaults.PackingFactor {
				t.Errorf("partial JSON defaults not applied: %+v", designer.lastInput)
			}
			tc.checkBody(t, rec.Body.Bytes())
		})
	}

	t.Run("format fallback in JSON body", func(t *testing.T) {
		designer := &mockDesigner{}
		srv := web.NewServer(designer, exporter)

		body := `{"format":"dxf","kiln_volume_cu_ft":28}`
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/export", strings.NewReader(body)))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if designer.lastInput.KilnVolumeCuFt != 28 {
			t.Fatalf("KilnVolumeCuFt = %v, want 28", designer.lastInput.KilnVolumeCuFt)
		}
		if !bytes.Equal(rec.Body.Bytes(), exporter.dxfData) {
			t.Fatalf("expected DXF bytes")
		}
	})

	t.Run("empty body with query format uses defaults", func(t *testing.T) {
		designer := &mockDesigner{}
		srv := web.NewServer(designer, exporter)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/export?format=stl", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if designer.lastInput != defaults {
			t.Fatalf("input = %+v, want defaults %+v", designer.lastInput, defaults)
		}
	})

	t.Run("invalid or missing format returns 400", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, exporter)
		for _, url := range []string{"/api/export", "/api/export?format=obj", "/api/export?format=pdf"} {
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, url, strings.NewReader("{}")))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("POST %s status = %d, want 400", url, rec.Code)
			}
		}
	})

	t.Run("invalid JSON or params returns 400", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, exporter)
		for _, body := range []string{"{broken", `{"kiln_volume_cu_ft": -1}`} {
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/export?format=stl", strings.NewReader(body)))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("body %q status = %d, want 400", body, rec.Code)
			}
		}
	})

	t.Run("exporter error returns 500", func(t *testing.T) {
		badExp := newMockExporter()
		badExp.err = errors.New("mesh build failed")
		srv := web.NewServer(&mockDesigner{}, badExp)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/export?format=stl", strings.NewReader("{}")))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})
}

func TestListenAndServeGracefulShutdownAndConcurrency(t *testing.T) {
	srv := web.NewServer(&mockDesigner{}, newMockExporter())

	// Exercise concurrent requests under -race via httptest.Server.
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var resp *http.Response
			var err error
			if i%2 == 0 {
				resp, err = http.Post(ts.URL+"/api/size", "application/json", strings.NewReader(`{"kiln_volume_cu_ft":24}`))
			} else {
				resp, err = http.Post(ts.URL+"/api/export?format=bom_json", "application/json", strings.NewReader(`{"kiln_volume_cu_ft":24}`))
			}
			if err != nil {
				t.Errorf("request %d failed: %v", i, err)
				return
			}
			defer resp.Body.Close()
			_, _ = io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("request %d status = %d, want 200", i, resp.StatusCode)
			}
		}(i)
	}
	wg.Wait()

	// Exercise ListenAndServe context cancellation cleanly.
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, "127.0.0.1:0")
	}()

	time.Sleep(40 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("ListenAndServe returned unexpected error on graceful shutdown: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("ListenAndServe did not shut down within timeout")
	}
}
