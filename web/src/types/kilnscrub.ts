export interface DesignInput {
  kiln_volume_cu_ft: number;
  peak_exit_temp_f: number;
  salt_rate_lbs_per_hr: number;
  fuel_type:
    | "wood_salt"
    | "wood"
    | "gas_salt"
    | "gas"
    | "coal_salt"
    | "coal"
    | "peat_salt"
    | "peat"
    | "biomass_ag_salt"
    | "biomass_ag"
    | string;
  fuel_sulfur_pct?: number;
  excess_air_fraction: number;
  ambient_temp_k?: number;
  ambient_pressure_pa?: number;
  venturi_throat_velocity_m_s: number;
  venturi_lg_ratio: number;
  target_hcl_removal: number;
  packing_name?: string;
  packing_diameter_m: number;
  packing_voidage: number;
  sump_ph_target?: number;
  target_kiln_draft_in_wc: number;
  duct_length_m?: number;
}

export interface StudioProject {
  preset_id: string;
  studio_name: string;
  firing_cone: string;
  firing_duration_hours: number;
  salting_hours: number;
  vapor_glaze_mode: "none" | "soda" | "salt" | string;
  build_path:
    | "cots_no_weld"
    | "welded_dxf"
    | "custom_fab_deluxe"
    | "local_ceramic_masonry"
    | string;
  zero_electricity: boolean;
  local_region:
    | "nairobi"
    | "mumbai"
    | "kathmandu"
    | "lagos_accra"
    | "tonala_puebla"
    | "cusco_andes"
    | "khurja_dhaka"
    | "cairo_fez"
    | "highland_peat"
    | "general_local"
    | string;
  local_alkali:
    | "soda_ash"
    | "slaked_lime"
    | "trona_kanwa"
    | "wood_ash_lye"
    | "baking_soda"
    | string;
  local_draft_mode:
    | "natural_chimney"
    | "steam_ejector_subsonic"
    | "steam_ejector_lp"
    | "steam_ejector_hp"
    | string;
  local_boiler_source?:
    | "dedicated_hearth"
    | "kiln_exhaust_coil"
    | "stage1_feedwater_preheater"
    | "stage1_water_jacket"
    | string;
  local_pump_mode?:
    | "manual_treadle"
    | "steam_ppcp"
    | "steam_jet_syphon"
    | "sump_pool_no_pump"
    | string;
  local_afterburner?: boolean;
  deluxe_condensing_reclaimer?: boolean;
  deluxe_variable_throat_vfd?: boolean;
  deluxe_stage1_heat_recovery?: boolean;
  corrosion_tier: "standard_316l" | "heavy_c276_frp" | string;
  stage_alignment: "inline_id_blower" | "eductor_dry_fan" | string;
  packed_bed_layout: "single_column" | "parallel_manifold" | string;
  parallel_pipe_size_in: number;
  packing_media_option:
    | "pp_pall_rings"
    | "aquaculture_bioballs"
    | "diy_ceramic_rings"
    | string;
  blower_drive_option:
    | "industrial_radial_vfd"
    | "regen_side_channel"
    | "dual_bypass_blowers"
    | "compressor_air_atomize"
    | "compressor_eductor_only"
    | string;
  shop_compressor_scfm: number;
  venturi_layout: "single_throat" | "parallel_mini_venturis" | string;
  compact_footprint: boolean;
  shared_baffled_sump: boolean;
  control_tier: "rpi_modbus" | "electromechanical_only" | string;
  snap_to_cots: boolean;
  include_silencer: boolean;
  include_plume_diluter: boolean;
  property_line_ft: number;
  ambient_winter_f: number;
  available_voltage: number;
  garden_hose_gpm: number;
  input: DesignInput;
}

export interface StudioPreset {
  id: string;
  name: string;
  badge: string;
  audience: string;
  tagline: string;
  project: StudioProject;
}

export interface EfficiencyPoint {
  particle_diameter_m: number;
  particle_diameter_um: number;
  wood_ash_efficiency: number;
  nacl_fume_efficiency: number;
}

export interface CurvePoint {
  flow_m3_s: number;
  flow_acfm: number;
  pressure_pa: number;
  pressure_in_wc: number;
}

export interface ScrubberDesign {
  input: DesignInput;
  inlet_gas: {
    temperature_k: number;
    mass_flow_kg_s: number;
    volumetric_flow_m3_s: number;
    volumetric_flow_acfm: number;
  };
  quenched_gas: {
    temperature_k: number;
    mass_flow_kg_s: number;
    volumetric_flow_m3_s: number;
    volumetric_flow_acfm: number;
  };
  outlet_gas: {
    temperature_k: number;
    volumetric_flow_acfm: number;
  };
  quench: {
    inlet_temp_k: number;
    adiabatic_sat_temp_k: number;
    outlet_temp_k: number;
    dew_point_k: number;
    evaporated_water_gpm: number;
    total_spray_water_gpm: number;
    overspray_factor: number;
    tower_diameter_m: number;
    tower_height_m: number;
    nozzle_count: number;
    pressure_drop_pa: number;
    pressure_drop_in_wc: number;
    material: string;
  };
  venturi: {
    inlet_diameter_m: number;
    throat_diameter_m: number;
    outlet_diameter_m: number;
    converging_length_m: number;
    throat_length_m: number;
    diverging_length_m: number;
    throat_velocity_m_s: number;
    lg_ratio: number;
    liquid_flow_gpm: number;
    droplet_diameter_m: number;
    cut_diameter_d50_um: number;
    wood_ash_cut_diameter_d50_um: number;
    efficiency_curve: EfficiencyPoint[];
    overall_wood_ash_efficiency: number;
    overall_nacl_fume_efficiency: number;
    pressure_drop_pa: number;
    pressure_drop_in_wc: number;
    material: string;
  };
  packed_bed: {
    column_diameter_m: number;
    column_height_m: number;
    flooding_velocity_m_s: number;
    superficial_gas_velocity_m_s: number;
    percent_of_flood: number;
    packing_height_m: number;
    packing_volume_m3: number;
    ntu: number;
    htu_m: number;
    packing_spec: string;
    hcl_inlet_ppm: number;
    hcl_outlet_ppm: number;
    hcl_removal_efficiency: number;
    so2_inlet_ppm?: number;
    so2_outlet_ppm?: number;
    so2_removal_efficiency?: number;
    sulfur_acid_dew_point_f?: number;
    nahco3_consumption_kg_hr: number;
    recirculation_flow_gpm: number;
    sump_ph: number;
    sump_volume_m3: number;
    bed_pressure_drop_pa: number;
    demister_pressure_drop_pa: number;
    demister_height_m: number;
    total_pressure_drop_pa: number;
    total_pressure_drop_in_wc: number;
    material: string;
  };
  blower: {
    breakdown: {
      duct_in_wc: number;
      quench_in_wc: number;
      venturi_in_wc: number;
      packed_bed_in_wc: number;
      demister_in_wc: number;
      kiln_draft_in_wc: number;
    };
    target_kiln_draft_in_wc: number;
    total_static_pressure_pa: number;
    total_static_pressure_in_wc: number;
    flow_acfm: number;
    efficiency: number;
    shaft_power_hp: number;
    recommended_motor_hp: number;
    wheel_diameter_m: number;
    operating_rpm: number;
    operating_point: CurvePoint;
    fan_curve: CurvePoint[];
    system_curve: CurvePoint[];
    material: string;
  };
  envelope: {
    length_m: number;
    width_m: number;
    height_m: number;
    duct_diameter_m: number;
    footprint_area_m2: number;
  };
  warnings: string[];
}

export interface COTSPart {
  sku: string;
  name: string;
  nominal_in: number;
  material: string;
  source: string;
  unit_cost_usd: number;
  weight_kg: number;
  notes: string;
}

export interface BOMItem {
  category: string;
  name: string;
  material: string;
  quantity: number;
  unit: string;
  spec: string;
  unit_cost_usd: number;
  total_cost_usd: number;
  weight_kg: number;
}

export interface COTSSummary {
  snapped: boolean;
  quench_vessel: COTSPart;
  packed_column: COTSPart;
  sump_tank: COTSPart;
  spray_nozzle: COTSPart;
  packing_media: COTSPart;
  blower_vfd: COTSPart;
  silencer?: COTSPart | null;
  plume_diluter?: COTSPart | null;
  sensor_kit: COTSPart;
  actual_percent_of_flood: number;
  estimated_total_usd: number;
  bom: {
    title: string;
    items: BOMItem[];
    total_cost_usd: number;
    total_weight_kg: number;
    notes: string[];
  };
}

export interface UrbanAssessment {
  overall_status: "ok" | "caution" | "action_needed" | string;
  electrical: {
    status: string;
    required_voltage: number;
    required_amps: number;
    circuit_label: string;
    matches_available: boolean;
    summary: string;
  };
  water: {
    status: string;
    required_gpm: number;
    available_gpm: number;
    total_firing_gallons: number;
    summary: string;
  };
  noise: {
    status: string;
    bare_dba_3ft: number;
    silenced_dba_3ft: number;
    property_line_dba: number;
    property_line_ft: number;
    urban_limit_dba: number;
    silencer_recommended: boolean;
    summary: string;
  };
  plume: {
    status: string;
    exhaust_temp_f: number;
    dew_point_f: number;
    visible_plume_risk: string;
    min_stack_height_ft: number;
    dilution_cfm_needed: number;
    dilution_reheat_kw: number;
    cold_dilution_works: boolean;
    bare_plume_length_ft: number;
    diluted_plume_length_ft: number;
    squirrel_cage_verdict: string;
    summary: string;
  };
  drain: {
    status: string;
    salting_hours: number;
    baking_soda_lbs_per_firing: number;
    baking_soda_15lb_bags: number;
    baking_soda_50lb_sacks: number;
    bulk_packaging_label: string;
    target_drain_ph: string;
    summary: string;
  };
  corrosion: {
    status: string;
    tier: string;
    tier_label: string;
    material_summary: string;
    soda_comparison: string;
    inspection_interval: string;
  };
}

export interface ArchitectureOptionsResult {
  stage_alignment: string;
  packed_bed_layout: string;
  venturi_layout: string;
  compact_footprint: boolean;
  shared_baffled_sump: boolean;
  control_tier: string;
  baseline_cost_usd: number;
  selected_cost_usd: number;
  cost_delta_usd: number;
  footprint_length_ft: number;
  footprint_width_ft: number;
  footprint_area_sq_ft: number;
  footprint_savings_pct: number;
  eductor: {
    enabled: boolean;
    motive_air_cfm: number;
    quenched_acfm: number;
    mixed_acfm: number;
    mixed_temp_f: number;
    thermal_margin_gain_f: number;
    positive_pressure_inwc: number;
    dry_fan_static_inwc: number;
    dry_fan_motor_hp: number;
    cyclonic_sweep_elbow_in: number;
    cost_delta_usd: number;
    feasible: boolean;
    summary: string;
    warnings: string[];
  };
  parallel_bed: {
    enabled: boolean;
    pipe_nominal_in: number;
    pipe_actual_id_in: number;
    pipe_count: number;
    pipe_material: string;
    wall_ratio_d_to_dp: number;
    actual_flood_pct: number;
    bed_dp_inwc: number;
    cost_delta_usd: number;
    summary: string;
  };
  parallel_venturi: {
    enabled: boolean;
    throat_count: number;
    throat_diameter_in: number;
    inlet_fitting_in: number;
    throat_dp_inwc: number;
    plugging_risk: string;
    turndown_benefit: string;
    cost_delta_usd: number;
    summary: string;
  };
  blower_drive: {
    option: string;
    label: string;
    fan_cost_usd: number;
    baseline_fan_cost_usd: number;
    cost_delta_usd: number;
    required_static_inwc: number;
    throat_velocity_m_s: number;
    droplet_um: number;
    motor_hp: number;
    voltage: number;
    feasible: boolean;
    summary: string;
    caveats: string[];
  };
  compressor: {
    available_scfm: number;
    eductor_scfm_needed: number;
    eductor_hp_needed: number;
    eductor_feasible: boolean;
    eductor_verdict: string;
    atomize_scfm_needed: number;
    atomize_feasible: boolean;
    atomize_droplet_um: number;
    atomize_dp_saved_inwc: number;
    atomize_cost_saved_usd: number;
    atomize_verdict: string;
  };
  packing_media: {
    option: string;
    label: string;
    diameter_in: number;
    voidage: number;
    bed_dp_inwc: number;
    media_cost_usd: number;
    baseline_media_cost_usd: number;
    cost_delta_usd: number;
    summary: string;
  };
  savings_summary: string[];
  warnings: string[];
}

export interface RecirculationPlan {
  sump_volume_gal: number;
  once_through_water_gal: number;
  recirc_makeup_water_gal: number;
  water_saved_pct: number;
  salting_hours: number;
  hcl_neutralized_lbs: number;
  baking_soda_lbs_per_firing: number;
  baking_soda_15lb_bags: number;
  baking_soda_50lb_sacks: number;
  initial_charge_lbs: number;
  end_of_firing_tds_g_l: number;
  dump_tds_limit_g_l: number;
  dump_specific_gravity: number;
  firings_before_full_dump: number;
  dump_triggers: string[];
  schedule_steps: string[];
  summary: string;
}

export interface FrugalAssessment {
  enabled: boolean;
  zero_electricity: boolean;
  region: string;
  region_label: string;
  power_watts: number;
  chimney_height_m: number;
  chimney_height_ft: number;
  stack_reheat_temp_f: number;
  available_draft_inwc: number;
  required_draft_inwc: number;
  draft_margin_inwc: number;
  draft_feasible: boolean;
  draft_mechanism: string;
  bed_diameter_in: number;
  bed_superficial_vel_m_s: number;
  bed_dp_inwc: number;
  packing_material: string;
  gravity_tank_liters: number;
  gravity_tank_gal: number;
  tank_elevation_ft: number;
  gravity_head_psi: number;
  spray_flow_lpm: number;
  spray_flow_gpm: number;
  manual_pump_type: string;
  jerrycans_20l_per_hr: number;
  treadle_minutes_per_hr: number;
  alkali_name: string;
  alkali_formula: string;
  alkali_kg_per_firing: number;
  alkali_lbs_per_firing: number;
  alkali_cost_usd: number;
  alkali_local_note: string;
  scaling_warning: string;
  hcl_removal_pct: number;
  coarse_ash_capture_pct: number;
  submicron_nacl_capture_pct: number;
  submicron_honest_note: string;
  thermal_immunity_note: string;
  estimated_local_cost_usd: number;
  local_currency_estimate: string;
  draft_mode: string;
  draft_mode_label: string;
  steam_enabled: boolean;
  steam_pressure_psig: number;
  steam_pressure_bar: number;
  steam_nozzle_vel_m_s: number;
  steam_nozzle_mach: number;
  steam_nozzle_throat_mm: number;
  steam_flow_kg_hr: number;
  steam_water_l_hr: number;
  steam_thermal_kw: number;
  kiln_waste_heat_kw: number;
  waste_heat_sufficient: boolean;
  ejector_draft_inwc: number;
  ejector_entrainment_ratio: number;
  condensation_growth_note: string;
  steam_source_type: string;
  steam_safety_requirements: string[];
  manual_steps: string[];
  warnings: string[];
  summary: string;
  boiler_source?: string;
  boiler_source_label?: string;
  boiler_water_saved_pct?: number;
  boiler_fuel_saved_note?: string;
  boiler_corrosion_note?: string;
  pump_mode?: string;
  pump_mode_label?: string;
  pump_steam_kg_hr?: number;
  has_overhead_tank?: boolean;
  sump_pool_submergence_in?: number;
  sump_pool_dp_inwc?: number;
  pump_engineering_note?: string;
  ambient_pressure_kpa?: number;
  elevation_note?: string;
  fuel_type?: string;
  so2_inlet_ppm?: number;
  so2_removal_pct?: number;
  sulfur_acid_dew_point_f?: number;
  afterburner_enabled?: boolean;
  afterburner_note?: string;
  tar_mitigation_note?: string;
  regional_machinery_note?: string;
}

export interface ShopPlan {
  title: string;
  build_path: string;
  is_deluxe: boolean;
  local_draft_mode?: string;
  steam_ejector_enabled?: boolean;
  stage_alignment: string;
  packed_bed_layout: string;
  venturi_layout: string;
  compact_footprint: boolean;
  overall_length_in: number;
  overall_width_in: number;
  overall_height_in: number;
  parts: Array<{
    num: number;
    id: string;
    label: string;
    stage: string;
    material: string;
    dimensions_in: string;
    dimensions_mm: string;
    sku: string;
    qty: number;
    cost_usd: number;
    notes: string;
  }>;
  cut_list: Array<{
    part_num: number;
    name: string;
    shape: string;
    material: string;
    flat_width_in: number;
    flat_height_in: number;
    outer_radius_in?: number;
    inner_radius_in?: number;
    sweep_deg?: number;
    qty: number;
    notes: string;
  }>;
  clearances: {
    flue_to_combustibles_in: number;
    property_line_min_ft: number;
    stack_above_roof_ft: number;
    service_aisle_in: number;
    nec_panel_depth_in: number;
    nec_panel_width_in: number;
    pump_service_in: number;
    cleanout_swing_in: number;
    hose_bend_radius_in: number;
    wiring_access_note: string;
  };
  electrical: Array<{
    id: string;
    label: string;
    from: string;
    to: string;
    voltage: number;
    amps: number;
    breaker_amps: number;
    wire_gauge: string;
    conduit: string;
    nec_ref: string;
    is_safety_loop: boolean;
    notes: string;
  }>;
  plumbing: Array<{
    id: string;
    label: string;
    fluid: string;
    from: string;
    to: string;
    size_in: number;
    material: string;
    flow_gpm: number;
    pressure_psi: number;
    valves_fittings: string;
    notes: string;
  }>;
  pump_station: {
    location: string;
    pump_type: string;
    flow_gpm: number;
    head_ft: number;
    hp: number;
    voltage: number;
    suction_size_in: number;
    discharge_size_in: number;
    strainer: string;
    backflow_preventer: string;
    drain_spec: string;
  };
  bracing: Array<{
    id: string;
    component: string;
    member_spec: string;
    mounting: string;
    load_lbs: number;
    anchor_spec: string;
    notes: string;
  }>;
  acoustics: Array<{
    id: string;
    location: string;
    treatment: string;
    material: string;
    thickness_in: number;
    attenuation_dba: number;
    notes: string;
  }>;
  acoustic_summary: {
    bare_dba_3ft: number;
    treated_dba_3ft: number;
    property_line_dba: number;
    property_line_ft: number;
    total_reduction_dba: number;
  };
  deluxe_spec?: {
    enabled: boolean;
    vessel_material: string;
    flange_standard: string;
    sight_glasses: string;
    cleanout_hatches: string;
    skid_frame: string;
    panel_standard: string;
    pump_upgrade: string;
    cost_delta_usd: number;
    summary: string;
    highlights: string[];
    condensing_reclaimer?: boolean;
    variable_throat_vfd?: boolean;
    stage1_heat_recovery?: boolean;
    eco_cost_delta_usd?: number;
    water_saved_pct?: number;
    water_saved_gal_per_firing?: number;
    energy_saved_pct?: number;
    energy_saved_kwh_per_firing?: number;
    recovered_heat_kw?: number;
    eco_summary?: string;
    eco_highlights?: string[];
  } | null;
}

export interface StudioEvaluation {
  project: StudioProject;
  design: ScrubberDesign;
  cots: COTSSummary;
  urban: UrbanAssessment;
  architecture_options: ArchitectureOptionsResult;
  recirculation: RecirculationPlan;
  shop_plan: ShopPlan;
  frugal?: FrugalAssessment | null;
  build_guide_md: string;
  operations_checklist_md: string;
}

export interface ControlStatus {
  state: string;
  filtered_draft_pa: number;
  filtered_draft_in_wc: number;
  interlock_active: boolean;
  updated_at: string;
  readings: {
    draft_pa: number;
    draft_in_wc: number;
    quench_inlet_temp_k: number;
    quench_outlet_temp_k: number;
    water_flow_m3_s: number;
    water_flow_gpm: number;
    sump_ph: number;
    blower_speed_fraction: number;
    timestamp: string;
  };
  commands: {
    blower_vfd: number;
    quench_spray_valve: number;
    bypass_damper_open: boolean;
    alarm_latched: boolean;
    trip_reason: string;
  };
}

export interface ControlConfig {
  target_draft_in_wc: number;
  target_draft_pa: number;
  quench_setpoint_k: number;
  safety: {
    max_quench_outlet_temp_k: number;
    min_water_flow_m3_s: number;
    min_sump_ph: number;
    max_sump_ph: number;
    max_draft_magnitude_pa: number;
  };
}

export interface ControlStatusResponse {
  enabled: boolean;
  hal_mode?: string;
  supports_fault_injection?: boolean;
  status?: ControlStatus;
  config?: ControlConfig;
}
