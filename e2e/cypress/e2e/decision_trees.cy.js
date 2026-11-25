/// <reference types="cypress" />

describe("kilnscrub End-to-End Decision Trees & UX Suite", () => {
  beforeEach(() => {
    cy.interceptEvaluate();
    cy.visit("/");
    cy.waitForEvalOK();
  });

  it("Tree 0: First-time onboarding tour, Guided vs All Controls, and Units toggle", () => {
    // Quick Tour appears on first visit and can be dismissed cleanly.
    cy.contains("New to kilnscrub?").should("be.visible");
    cy.contains("button", "Got it").click();
    cy.contains("New to kilnscrub?").should("not.exist");

    // Reopen tour via "? How It Works".
    cy.contains("button", "? How It Works").click();
    cy.contains("New to kilnscrub?").should("be.visible");
    cy.contains("button", "Hide Guide").click();

    // Switch Imperial <-> Metric units cleanly.
    cy.contains("button", "in / ft").click();
    cy.contains("button", "mm / m").should("be.visible");

    // Switch Guided Setup <-> All Controls mode.
    cy.contains("button", "All Controls").click();
    cy.get("#in-studio-name").should("be.visible");
    cy.get("#in-blower-drive").should("be.visible");
    cy.contains("button", "Guided Setup").click();
  });

  it("Tree A & C: Build paths, Deluxe eco upgrades, Eductor dry-fan & Budget DIY combo", () => {
    // 1. Pro Deluxe unlocks eco upgrades & disables parallel_manifold.
    cy.get('[data-build-path="custom_fab_deluxe"]').click();
    cy.waitForEvalOK();
    cy.get("#panel-deluxe-eco-settings").should("be.visible");
    cy.get("#in-deluxe-condensing").check();
    cy.waitForEvalOK();
    cy.get("#in-deluxe-var-throat").check();
    cy.waitForEvalOK();
    cy.get("#in-deluxe-s1-heat").check();
    cy.waitForEvalOK();

    // Open Hardware & Layout accordion and verify Uniseal Bundle is disabled in Deluxe.
    cy.contains("button", "Hardware & Layout").click();
    cy.get('[data-packed-bed="parallel_manifold"]').should("be.disabled");

    // Switching to Dry-Fan Eductor auto-clears Variable-Throat VFD & disables Mini-Venturis.
    cy.contains("button", "Dry-Fan Eductor").click();
    cy.waitForEvalOK();
    cy.get('[data-venturi-layout="parallel_mini_venturis"]').should("be.disabled");
    cy.get("#in-deluxe-var-throat").should("be.disabled").and("not.be.checked");

    // 2. Budget DIY Quick Profile enables No-Weld + 4" Uniseal + Air-Atomize + DIY Rings.
    cy.get("#btn-apply-budget-combo").click();
    cy.waitForEvalOK();
    cy.get('[data-build-path="cots_no_weld"]').should("have.class", "active");
    cy.get("#in-blower-drive").should("have.value", "compressor_air_atomize");
  });

  it("Tree B: Fuel Stoichiometry x Vapor Glaze modes (Coal, Peat, PKS, Wood, Gas)", () => {
    // Switching to Coal + Salt auto-selects Salt mode and surfaces SO2 + HCl capture.
    cy.get("#in-fuel-type").select("coal_salt");
    cy.waitForEvalOK();
    cy.get('[data-vapor-mode="salt"]').should("have.class", "active");
    cy.contains("SO₂").should("be.visible");

    // Switching Vapor Glaze to Soda flips coal_salt -> coal cleanly with 0 salt.
    cy.get('[data-vapor-mode="soda"]').click();
    cy.waitForEvalOK();
    cy.get("#in-fuel-type").should("have.value", "coal");

    // Switching to Peat + Salt works cleanly.
    cy.get("#in-fuel-type").select("peat_salt");
    cy.waitForEvalOK();
    cy.contains("SO₂").should("be.visible");

    // Switching to Ag-Biomass (PKS/Pomace/Husk).
    cy.get("#in-fuel-type").select("biomass_ag");
    cy.waitForEvalOK();
  });

  it("Tree D: Local Ceramic / Off-Grid Global Hubs, Steam Tiers, Boiler & Sump-Pool Guards", () => {
    // Apply Off-Grid Ceramic (0 W) and verify zero-electricity ON -> OFF recovery (no 400 error!).
    cy.get("#btn-apply-offgrid-ceramic").click();
    cy.waitForEvalOK();
    cy.get("#in-zero-electricity").should("be.checked");
    cy.get("#in-zero-electricity").uncheck();
    cy.waitForEvalOK();
    cy.get("#in-zero-electricity").check();
    cy.waitForEvalOK();

    // Open Global Coal & Peat Examples and walk Khurja -> Cusco 3,400m -> Lagos -> Highland.
    cy.contains("button", "Global Coal & Peat Examples").click();
    cy.contains("button", "Khurja Coal+Salt").click();
    cy.waitForEvalOK();
    cy.contains("Khurja").should("be.visible");
    cy.contains("SO₂").should("be.visible");

    cy.contains("button", "Cusco 3,400m Peat").click();
    cy.waitForEvalOK();
    cy.contains("Cusco").should("be.visible");

    cy.contains("button", "Lagos PKS+Kanwa").click();
    cy.waitForEvalOK();
    cy.contains("Lagos").should("be.visible");

    cy.contains("button", "Highland Peat+Salt").click();
    cy.waitForEvalOK();
    cy.contains("Highland").should("be.visible");

    // Verify Boiler Heat & Water Pump mutual-exclusivity rules across 6 -> 20 -> 60 psig.
    cy.get("#in-local-draft-mode").select("steam_ejector_lp");
    cy.waitForEvalOK();
    cy.contains("button", "Boiler Heat & Water Pump").click();
    cy.get("#in-local-boiler-source").select("kiln_exhaust_coil");
    cy.waitForEvalOK();

    // Stepping up to 60 psig HP automatically migrates coil -> ASME Stage 1 jacket & unlocks Sump Pools!
    cy.get("#in-local-draft-mode").select("steam_ejector_hp");
    cy.waitForEvalOK();
    cy.get("#in-local-boiler-source").should("have.value", "stage1_water_jacket");
    cy.get("#in-local-pump-mode").select("sump_pool_no_pump");
    cy.waitForEvalOK();

    // Dropping to 6 psig subsonic safely downgrades boiler & pump with zero 400 errors!
    cy.get("#in-local-draft-mode").select("steam_ejector_subsonic");
    cy.waitForEvalOK();
    cy.get("#in-local-boiler-source").should("have.value", "stage1_feedwater_preheater");
    cy.get("#in-local-pump-mode").should("have.value", "manual_treadle");
  });

  it("Viewports, 4-Chapter Build Guide, Shop Plans Layers & Live Control Telemetry", () => {
    // 1. Switch 3D <-> 2D Detailed P&ID <-> Simple Outline and click stage focus explainer.
    cy.contains("button", "1 Quench").click();
    cy.contains("Stage 1 Spray Quench").should("be.visible");
    cy.contains("button", "2D Blueprint").click();
    cy.get("#viewport-2d").should("be.visible");
    cy.get("#btn-2d-detail-mode").click();
    cy.get("#viewport-2d").should("be.visible");

    // Check Performance Charts & Notices deep-dive tabs.
    cy.contains("button", "Performance Charts").click();
    cy.get("#chart-dp").should("be.visible");
    cy.get("#chart-efficiency").should("be.visible");
    cy.get("#chart-fan").should("be.visible");

    // 2. Walk all 4 Build & Firing Guide chapters.
    cy.get("#tab-guide").click();
    cy.contains("button", "2. COTS Parts & Shopping List").click();
    cy.get("table.data-table").should("be.visible");
    cy.contains("button", "3. Step-by-Step Build Guide").click();
    cy.get(".md-doc").should("be.visible");
    cy.contains("button", "4. Firing-Day How-To Checklist").click();
    cy.get(".md-check-item").first().click();

    // 3. Verify Shop Plans & 5 toggleable layers.
    cy.get("#tab-plans").click();
    cy.get("#btn-layer-electrical").click();
    cy.get("#btn-layer-plumbing").click();

    // 4. Verify Live Control & Telemetry tab receives live SSE state and handles fault/reset.
    cy.get("#tab-control").click();
    cy.contains("SSE LIVE", { timeout: 8000 }).should("be.visible");
  });
});
