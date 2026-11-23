package web_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/AaronFilson/kilnscrub/internal/web"
)

type mockStudioBackend struct {
	mu             sync.Mutex
	presetsJSON    []byte
	presetsErr     error
	evalJSON       []byte
	evalErr        error
	lastEvalBody   []byte
	packetBytes    []byte
	packetFilename string
	packetErr      error
	lastPacketBody []byte
}

func newMockStudioBackend(t *testing.T) *mockStudioBackend {
	t.Helper()

	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, err := zw.Create("README_PACKET.txt")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := f.Write([]byte("kilnscrub build packet test\n")); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	presets := []byte(`{"presets":[{"id":"wood_salt_24","name":"Reference Studio Wood / Salt Kiln","tagline":"24 cu.ft. Cone 11 studio reference","audience":"Studio Potters","badge":"Reference","project":{"preset_id":"wood_salt_24","studio_name":"Reference Studio","firing_cone":"Cone 11","firing_duration_hours":10,"salting_hours":3.0,"vapor_glaze_mode":"salt","build_path":"cots_no_weld","corrosion_tier":"standard_316l","stage_alignment":"inline_id_blower","packed_bed_layout":"single_column","venturi_layout":"single_throat","compact_footprint":false,"shared_baffled_sump":true,"control_tier":"rpi_modbus","snap_to_cots":true,"include_silencer":true,"include_plume_diluter":false,"property_line_ft":35,"ambient_winter_f":35,"available_voltage":240,"garden_hose_gpm":5}}]}`)

	evalResp := []byte(`{"project":{"preset_id":"wood_salt_24","studio_name":"Reference Studio","salting_hours":3.0,"corrosion_tier":"standard_316l","stage_alignment":"inline_id_blower","packed_bed_layout":"single_column","venturi_layout":"single_throat","compact_footprint":false,"shared_baffled_sump":true,"control_tier":"rpi_modbus"},"cots":{"snapped":true,"actual_percent_of_flood":54.2,"estimated_total_usd":3450,"bom":{"title":"COTS BOM","items":[],"total_cost_usd":3450,"total_weight_kg":112}},"urban":{"overall_status":"ok","electrical":{"status":"ok","required_voltage":240,"required_amps":14.2,"circuit_label":"240V 20A single-phase VFD","summary":"Fits 240V studio circuit","fits_available_voltage":true},"water":{"status":"ok","required_gpm":2.3,"total_firing_gallons":180,"summary":"Garden hose OK"},"noise":{"status":"ok","bare_dba_3ft":82,"silenced_dba_3ft":68,"property_line_dba":50.5,"property_line_ft":35,"urban_limit_dba":60,"silencer_recommended":true,"summary":"Below 60 dBA urban limit"},"plume":{"status":"caution","exhaust_temp_f":184,"dew_point_f":184,"visible_plume_risk":"high","min_stack_height_ft":10,"dilution_cfm_needed":1500,"dilution_reheat_kw":21.5,"cold_dilution_works":false,"bare_plume_length_ft":28,"diluted_plume_length_ft":9,"squirrel_cage_verdict":"120V squirrel-cage blower works downstream with HV cone; unheated cold air alone stays supersaturated.","summary":"Clean condensed steam mist in winter"},"drain":{"status":"ok","salting_hours":3.0,"baking_soda_lbs_per_firing":9.8,"baking_soda_15lb_bags":1,"baking_soda_50lb_sacks":1,"bulk_packaging_label":"1 x 15 lb bulk bag (~9.8 lbs over 3.0 hr salting; 1 x 50 lb sack covers ~5 firings)","target_drain_ph":"6.5-8.5","summary":"Neutralized salt brine"},"corrosion":{"status":"caution","tier":"standard_316l","material_summary":"Standard 316L SS (Light / Occasional) with alkaline-buffered spray","soda_comparison":"Post-firing buffered pH 8 flush and flange weld check keeps 316L spools serviceable.","inspection_interval":"Every 5-10 firings"}},"architecture_options":{"cost_delta_usd":0,"footprint_length_ft":9.5,"footprint_width_ft":3.0,"footprint_area_sqft":28.5,"footprint_savings_pct":0,"savings_summary":[],"warnings":[]},"recirculation":{"sump_volume_gal":45,"once_through_water_gal":1260,"recirc_makeup_water_gal":235,"water_saved_pct":81.3,"salting_hours":3.0,"hcl_neutralized_lbs":4.2,"baking_soda_lbs_per_firing":9.8,"baking_soda_15lb_bags":1,"baking_soda_50lb_sacks":1,"initial_charge_lbs":11.3,"end_of_firing_tds_g_l":26.4,"dump_tds_limit_g_l":150,"dump_specific_gravity":1.105,"firings_before_full_dump":5,"top_up_soda_lbs_between_firings":9.8,"dump_triggers":["Hydrometer SG >= 1.105","pH < 7.0","Cool <= 140F per IPC 702.5"],"schedule_steps":["Charge initial pH 8 buffer","Top up 9.8 lbs between firings"],"summary":"Saves 81% water across 5 firings before full dump."},"shop_plan":{"title":"Reference Studio Shop Plan","stage_alignment":"inline_id_blower","compact_footprint":false,"overall_length_in":114,"overall_width_in":36,"overall_height_in":122,"parts":[{"num":1,"id":"quench_tower","label":"Stage 1 Quench","stage":"Stage 1","material":"316L SS","dimensions_in":"18 x 54 in","dimensions_mm":"457 x 1372 mm","sku":"Q-316L","qty":1,"cost_usd":580,"notes":"Bolted top"}],"cut_list":[{"part_num":2,"name":"Venturi Cone","shape":"cone_frustum","material":"316L SS","flat_width_in":0,"flat_height_in":0,"outer_radius_in":18.4,"inner_radius_in":7.2,"sweep_deg":112,"qty":1,"notes":"Converging frustum"}],"clearances":{"flue_to_combustibles_in":36,"property_line_min_ft":30,"stack_above_roof_ft":6,"service_aisle_in":36}},"build_guide_md":"# Build Guide\n\n- [ ] Bolt flanges","operations_checklist_md":"# Firing Checklist\n\n- [ ] Test 210F trip"}`)

	return &mockStudioBackend{
		presetsJSON:    presets,
		evalJSON:       evalResp,
		packetBytes:    zipBuf.Bytes(),
		packetFilename: "kilnscrub-wood_salt_24-packet.zip",
	}
}

func (m *mockStudioBackend) PresetsJSON() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.presetsErr != nil {
		return nil, m.presetsErr
	}
	return append([]byte(nil), m.presetsJSON...), nil
}

func (m *mockStudioBackend) EvaluateJSON(reqBody []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastEvalBody = append([]byte(nil), reqBody...)
	if m.evalErr != nil {
		return nil, m.evalErr
	}
	return append([]byte(nil), m.evalJSON...), nil
}

func (m *mockStudioBackend) BuildPacketZip(reqBody []byte) ([]byte, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastPacketBody = append([]byte(nil), reqBody...)
	if m.packetErr != nil {
		return nil, "", m.packetErr
	}
	return append([]byte(nil), m.packetBytes...), m.packetFilename, nil
}

func TestStudioPresetsEndpoint(t *testing.T) {
	t.Run("disabled when studio backend is nil", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/studio/presets", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/studio/presets status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		var got struct {
			Enabled bool  `json:"enabled"`
			Presets []any `json:"presets"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal disabled presets: %v", err)
		}
		if got.Enabled {
			t.Fatalf("enabled = true, want false")
		}
		if got.Presets == nil || len(got.Presets) != 0 {
			t.Fatalf("presets = %#v, want empty non-nil slice", got.Presets)
		}

		// HEAD returns 200 with empty body.
		recHead := httptest.NewRecorder()
		srv.Handler().ServeHTTP(recHead, httptest.NewRequest(http.MethodHead, "/api/studio/presets", nil))
		if recHead.Code != http.StatusOK || recHead.Body.Len() != 0 {
			t.Fatalf("HEAD /api/studio/presets = %d len=%d, want 200 len=0", recHead.Code, recHead.Body.Len())
		}
	})

	t.Run("enabled returns presets JSON", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		srv.SetStudioBackend(newMockStudioBackend(t))

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/studio/presets", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		var got struct {
			Presets []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"presets"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal presets: %v", err)
		}
		if len(got.Presets) != 1 || got.Presets[0].ID != "wood_salt_24" {
			t.Fatalf("unexpected presets: %+v", got.Presets)
		}
	})

	t.Run("backend error or invalid JSON returns 500", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		b := newMockStudioBackend(t)
		b.presetsErr = errors.New("catalog unavailable")
		srv.SetStudioBackend(b)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/studio/presets", nil))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}

		b.presetsErr = nil
		b.presetsJSON = []byte("{not-json")
		rec2 := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/api/studio/presets", nil))
		if rec2.Code != http.StatusInternalServerError {
			t.Fatalf("invalid backend JSON status = %d, want 500", rec2.Code)
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/presets", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST /api/studio/presets status = %d, want 405", rec.Code)
		}
	})
}

func TestStudioEvaluateEndpoint(t *testing.T) {
	t.Run("disabled returns 503", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/evaluate", strings.NewReader("{}")))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		srv.SetStudioBackend(newMockStudioBackend(t))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/studio/evaluate", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("valid project JSON returns 200 StudioEvaluation", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		b := newMockStudioBackend(t)
		srv.SetStudioBackend(b)

		payload := `{"preset_id":"wood_salt_24","studio_name":"Cedar Creek","include_plume_diluter":true}`
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/evaluate", strings.NewReader(payload)))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if string(b.lastEvalBody) != payload {
			t.Errorf("backend received %q, want %q", string(b.lastEvalBody), payload)
		}
		var out map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("unmarshal evaluation: %v", err)
		}
		for _, key := range []string{"project", "cots", "urban", "architecture_options", "recirculation", "shop_plan", "build_guide_md", "operations_checklist_md"} {
			if _, ok := out[key]; !ok {
				t.Errorf("StudioEvaluation missing key %q", key)
			}
		}
	})

	t.Run("invalid JSON or validation error returns 400", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		b := newMockStudioBackend(t)
		srv.SetStudioBackend(b)

		for _, bad := range []string{"", "{bad", "[]", "null"} {
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/evaluate", strings.NewReader(bad)))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("body %q status = %d, want 400", bad, rec.Code)
			}
		}

		b.evalErr = errors.New("invalid vapor_glaze_mode")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/evaluate", strings.NewReader(`{"vapor_glaze_mode":"lava"}`)))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("validation error status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "invalid vapor_glaze_mode") {
			t.Fatalf("expected error detail in body, got %q", rec.Body.String())
		}
	})
}

func TestStudioPacketEndpoint(t *testing.T) {
	t.Run("disabled returns 503", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/packet", strings.NewReader("{}")))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		srv.SetStudioBackend(newMockStudioBackend(t))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/studio/packet", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("valid request returns 200 application/zip attachment", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		b := newMockStudioBackend(t)
		srv.SetStudioBackend(b)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/packet", strings.NewReader(`{"preset_id":"wood_salt_24"}`)))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
			t.Errorf("Content-Type = %q, want application/zip", ct)
		}
		cd := rec.Header().Get("Content-Disposition")
		if cd != `attachment; filename="kilnscrub-wood_salt_24-packet.zip"` {
			t.Errorf("Content-Disposition = %q, want attachment with kilnscrub-wood_salt_24-packet.zip", cd)
		}

		zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
		if err != nil {
			t.Fatalf("response is not a valid zip archive: %v", err)
		}
		if len(zr.File) != 1 || zr.File[0].Name != "README_PACKET.txt" {
			t.Fatalf("unexpected zip contents: %+v", zr.File)
		}
	})

	t.Run("sanitizes unsafe download filename and defaults when empty", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		b := newMockStudioBackend(t)
		b.packetFilename = "../evil\"\r\npacket.zip"
		srv.SetStudioBackend(b)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/packet", strings.NewReader("{}")))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		cd := rec.Header().Get("Content-Disposition")
		if strings.ContainsAny(cd, "\r\n/") || !strings.HasPrefix(cd, `attachment; filename="`) {
			t.Fatalf("unsafe Content-Disposition header: %q", cd)
		}

		b.packetFilename = "   "
		rec2 := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec2, httptest.NewRequest(http.MethodPost, "/api/studio/packet", strings.NewReader("{}")))
		if cd2 := rec2.Header().Get("Content-Disposition"); cd2 != `attachment; filename="kilnscrub-build-packet.zip"` {
			t.Fatalf("fallback Content-Disposition = %q", cd2)
		}
	})

	t.Run("invalid JSON or backend error returns 400", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		b := newMockStudioBackend(t)
		srv.SetStudioBackend(b)

		for _, bad := range []string{"", "{broken", "[1,2]"} {
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/packet", strings.NewReader(bad)))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("body %q status = %d, want 400", bad, rec.Code)
			}
		}

		b.packetErr = errors.New("invalid project")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/packet", strings.NewReader("{}")))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("backend error status = %d, want 400", rec.Code)
		}
	})
}

func TestStudioConcurrency(t *testing.T) {
	srv := web.NewServer(&mockDesigner{}, newMockExporter())
	b := newMockStudioBackend(t)
	srv.SetStudioBackend(b)

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 4 {
			case 0:
				rec := httptest.NewRecorder()
				srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/studio/presets", nil))
				if rec.Code != http.StatusOK {
					t.Errorf("presets status = %d", rec.Code)
				}
			case 1:
				rec := httptest.NewRecorder()
				srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/evaluate", strings.NewReader("{}")))
				if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
					t.Errorf("evaluate status = %d", rec.Code)
				}
			case 2:
				rec := httptest.NewRecorder()
				srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/studio/packet", strings.NewReader("{}")))
				if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
					t.Errorf("packet status = %d", rec.Code)
				}
			case 3:
				srv.SetStudioBackend(b)
			}
		}(i)
	}
	wg.Wait()
}

func TestEmbeddedHTMLStudioUXAndJSIntegrity(t *testing.T) {
	srv := web.NewServer(&mockDesigner{}, newMockExporter())
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	// 1. Verify all 4 tabs, potter-friendly Studio UX, Phase 7 selectable options, bulk soda, recirculation,
	//    complete Shop Plans & Cut Sheets tab, plume diluter, tooltips, unit literals, and 3D builders.
	requiredMarkers := []string{
		"Design &amp; 3D Studio",
		"Build &amp; Firing Guide",
		"Live Control &amp; Telemetry",
		"Shop Plans &amp; Cut Sheets",
		`id="tab-plans"`,
		`id="view-plans"`,
		`id="plan-svg"`,
		`id="plan-cutsheet-svg"`,
		`id="plan-parts-tbody"`,
		`id="btn-plan-download-svg"`,
		`id="btn-plan-print"`,
		`id="btn-plan-unit-in"`,
		`id="btn-plan-unit-mm"`,
		"/api/studio/presets",
		"/api/studio/evaluate",
		"/api/studio/packet",
		`id="preset-picker-list"`,
		`id="readiness-bar"`,
		`id="panel-stage-options"`,
		"Stage Design &amp; Cost-Saving Options",
		`data-stage-alignment="inline_id_blower"`,
		`data-stage-alignment="eductor_dry_fan"`,
		`data-packed-bed-layout="single_column"`,
		`data-packed-bed-layout="parallel_manifold"`,
		`data-venturi-layout="single_throat"`,
		`data-venturi-layout="parallel_mini_venturis"`,
		`data-control-tier="rpi_modbus"`,
		`data-control-tier="electromechanical_only"`,
		`id="in-compact-footprint"`,
		`id="in-shared-baffled-sump"`,
		`id="in-salting-hours"`,
		"Active Salting Window",
		"Standard 316L SS (Light / Occasional)",
		`data-corrosion="standard_316l"`,
		"15 lb",
		"50 lb",
		"baking_soda_15lb_bags",
		"baking_soda_50lb_sacks",
		"salting_hours",
		"bulk_packaging_label",
		"architecture_options",
		"recirculation",
		"shop_plan",
		"Sump Recirculation &amp; Dump Schedule",
		"Architecture Cost &amp; Trade-Off Comparison",
		"1.105",
		"NFPA 211",
		`id="advanced-engineering-drawer"`,
		"Advanced Engineering Parameters",
		"Download Complete Build Packet (.zip)",
		"Step-by-Step Build Guide",
		"Firing-Day How-To Checklist",
		"COTS Parts &amp; Shopping List",
		"Save project.json",
		"Load project.json",
		"Cone 06",
		"Cone 6",
		"Cone 10",
		"Cone 11",
		"No-Weld COTS",
		"Welded Sheet Metal",
		"include_plume_diluter",
		"squirrel_cage_verdict",
		"dilution_cfm_needed",
		"dilution_reheat_kw",
		"bare_plume_length_ft",
		"diluted_plume_length_ft",
		".sidebar > details.adv-drawer",
		"flex-shrink: 0;",
		"CPVC 210°F Cutoff",
		`<span class="unit-literal">Na₂CO₃</span>`,
		`<span class="unit-literal">NaCl</span>`,
		`<span class="unit-literal">HCl</span>`,
		`<span class="unit-literal">µm</span>`,
		`<span class="unit-literal">d₅₀</span>`,
		`<span class="unit-literal">pH</span>`,
		`<span class="unit-literal">inWC</span>`,
		`<span class="unit-literal">GPM</span>`,
		"TubeGeometry",
		"CurvePath",
		"ExtrudeGeometry",
		"computeVoluteProfiles",
		"buildRectToRoundGeometry",
		"buildImpellerGroup",
		"addBoltedFlange",
	}
	for _, needle := range requiredMarkers {
		if !strings.Contains(body, needle) {
			t.Errorf("embedded index.html missing studio marker %q", needle)
		}
	}

	// Verify removed terminology & nag strings do NOT appear anywhere in embedded HTML.
	forbiddenMarkers := []string{
		"Hobby 316L",
		"4-lb",
		"4lb",
		"Grocery 4-lb",
		"Switching from Salt (NaCl) to Soda",
		"Switching to Soda",
	}
	for _, bad := range forbiddenMarkers {
		if strings.Contains(body, bad) {
			t.Errorf("embedded index.html must NOT contain deprecated string %q", bad)
		}
	}

	// Verify Advanced Engineering drawer is collapsed by default (no `open` attribute).
	if strings.Contains(body, `<details id="advanced-engineering-drawer" open`) ||
		strings.Contains(body, `<details id="advanced-engineering-drawer" class="adv-drawer" open`) {
		t.Errorf("advanced-engineering-drawer should be collapsed by default")
	}

	// 2. Verify every static el("...") and document.getElementById("...") references an existing id="..." in HTML.
	idDefRe := regexp.MustCompile(`\bid="([a-zA-Z0-9_-]+)"`)
	definedIDs := make(map[string]bool)
	for _, m := range idDefRe.FindAllStringSubmatch(body, -1) {
		definedIDs[m[1]] = true
	}

	scriptRe := regexp.MustCompile(`(?s)<script(?:\s+type="module")?>(.*?)</script>`)
	scripts := scriptRe.FindAllStringSubmatch(body, -1)
	if len(scripts) < 2 {
		t.Fatalf("expected at least 2 inline <script> blocks, found %d", len(scripts))
	}

	refRe := regexp.MustCompile(`(?:document\.getElementById|\bel)\("([a-zA-Z0-9_-]+)"\)`)
	for idx, match := range scripts {
		jsCode := match[1]
		for _, ref := range refRe.FindAllStringSubmatch(jsCode, -1) {
			id := ref[1]
			if !definedIDs[id] {
				t.Errorf("script[%d] references missing DOM id %q", idx, id)
			}
		}
	}

	// 3. Syntax-check embedded scripts with `node --check` if node is available.
	nodeBin, err := exec.LookPath("node")
	if err != nil {
		t.Log("node not found in PATH; skipping node --check syntax verification")
		return
	}

	tmpDir := t.TempDir()
	for idx, match := range scripts {
		ext := ".js"
		if strings.Contains(match[0], `type="module"`) {
			ext = ".mjs"
		}
		scriptPath := filepath.Join(tmpDir, "embedded_"+string(rune('0'+idx))+ext)
		if err := os.WriteFile(scriptPath, []byte(match[1]), 0o600); err != nil {
			t.Fatalf("write temp script: %v", err)
		}
		cmd := exec.Command(nodeBin, "--check", scriptPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("node --check failed for script[%d]: %v\n%s", idx, err, string(out))
		}
	}
}
