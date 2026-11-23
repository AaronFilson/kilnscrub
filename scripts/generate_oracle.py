#!/usr/bin/env python3
"""Generate ground-truth thermophysical and scrubber test fixtures for Go unit tests.

Uses peer-reviewed Python engineering libraries (`chemicals`, `fluids`, `scipy`):
- Ideal-gas heat capacity Cp(T) & enthalpy via `chemicals.heat_capacity.Shomate`
  with official NIST Chemistry WebBook 7-coefficient Shomate parameters.
- Low-pressure gas viscosity mu(T) via `chemicals.viscosity.Lucas_gas`
  using critical properties and dipole moments from `chemicals`.
- Water saturation vapor pressure Psat(T) via `chemicals.iapws.iapws95_Psat`
  (IAPWS-95 formulation).
- Packed-bed pressure drop via `fluids.packed_bed.dP_packed_bed(..., Method="Ergun")`
  cross-checked against `fluids.packed_bed.Ergun`.
- Venturi scrubber cut diameter d50 via Calvert (1970, 1972) inertial impaction
  penetration model with Davies/Allen-Raabe Cunningham slip correction, solved
  with `scipy.optimize.brentq` (note: `fluids` 1.3.1 does not expose a built-in
  `Calvert_cut_diameter` function).
"""

import inspect
import json
import math
import os
from pathlib import Path

import chemicals
from chemicals import critical, dipole, heat_capacity, iapws, identifiers, viscosity
import fluids
from fluids import packed_bed
from scipy.optimize import brentq


# Official NIST Chemistry WebBook gas-phase Shomate coefficients:
# (Tmin_K, Tmax_K, A, B, C, D, E, F, G, H)
# Cp [J/(mol*K)] = A + B*t + C*t^2 + D*t^3 + E/t^2, where t = T/1000
# H(T) - H(298.15) [kJ/mol] = A*t + B*t^2/2 + C*t^3/3 + D*t^4/4 - E/t + F - H
NIST_SHOMATE = {
    "N2": [
        (100.0, 500.0, 28.98641, 1.853978, -9.647459, 16.63537, 0.000117, -8.671914, 226.4168, 0.0),
        (500.0, 2000.0, 19.50583, 19.88705, -8.598535, 1.369784, 0.527601, -4.935202, 212.3900, 0.0),
        (2000.0, 6000.0, 35.51872, 1.128728, -0.196103, 0.014662, -4.553760, -18.97091, 224.9810, 0.0),
    ],
    "CO2": [
        (298.0, 1200.0, 24.99735, 55.18696, -33.69137, 7.948387, -0.136638, -403.6075, 228.2431, -393.5224),
        (1200.0, 6000.0, 58.16639, 2.720074, -0.492289, 0.038844, -6.447293, -425.9186, 263.6125, -393.5224),
    ],
    "H2O": [
        # NIST WebBook lists 500-1700 K (calibrated to H(298.15)=0); valid down to 298 K within <0.1%
        (298.0, 1700.0, 30.09200, 6.832514, 6.793435, -2.534480, 0.082139, -250.8810, 223.3967, -241.8264),
        (1700.0, 6000.0, 41.96426, 8.622053, -1.499780, 0.098119, -11.15764, -272.1797, 219.7809, -241.8264),
    ],
    "O2": [
        (100.0, 700.0, 31.32234, -20.23531, 57.86644, -36.50624, -0.007374, -8.903471, 246.7945, 0.0),
        (700.0, 2000.0, 30.03235, 8.772972, -3.988133, 0.788313, -0.741599, -11.32468, 236.1663, 0.0),
        (2000.0, 6000.0, 20.91111, 10.72071, -2.020498, 0.146449, 9.245722, 5.337651, 237.6185, 0.0),
    ],
    "HCl": [
        (298.0, 1200.0, 32.12392, -13.45805, 19.86852, -6.853936, -0.049672, -101.6206, 228.6866, -92.31201),
        (1200.0, 6000.0, 31.91923, 3.203184, -0.541539, 0.035925, -3.438525, -108.0150, 218.2768, -92.31201),
    ],
}

SPECIES_CAS = {
    "N2": "7727-37-9",
    "CO2": "124-38-9",
    "H2O": "7732-18-5",
    "O2": "7782-44-7",
    "HCl": "7647-01-0",
}


def select_shomate_coeffs(species: str, T: float):
    ranges = NIST_SHOMATE[species]
    for row in ranges:
        tmin, tmax = row[0], row[1]
        if tmin <= T <= tmax:
            return row
    if T < ranges[0][0]:
        return ranges[0]
    return ranges[-1]


def shomate_cp_and_enthalpy(species: str, T: float):
    """Compute Cp [J/mol/K] via chemicals.heat_capacity.Shomate and H(T)-H(298.15) [J/mol]."""
    _, _, A, B, C, D, E, F, _G, H = select_shomate_coeffs(species, T)
    # Call chemicals.heat_capacity.Shomate directly; verify scaling convention (T in K vs t=T/1000).
    t = T / 1000.0
    cp_nist = A + B * t + C * (t ** 2) + D * (t ** 3) + E / (t ** 2)
    cp_chem = heat_capacity.Shomate(T, A, B, C, D, E)
    if abs(cp_chem - cp_nist) / abs(cp_nist) > 1e-9:
        # Some versions of chemicals expect unscaled coefficients or t=T/1000
        cp_chem_t = heat_capacity.Shomate(t, A, B, C, D, E)
        if abs(cp_chem_t - cp_nist) / abs(cp_nist) < 1e-9:
            cp = cp_chem_t
        else:
            raise ValueError(
                f"chemicals.Shomate mismatch for {species} at {T}K: "
                f"Shomate(T)={cp_chem}, Shomate(t)={cp_chem_t}, expected={cp_nist}"
            )
    else:
        cp = cp_chem

    # Sensible enthalpy H(T) - H(298.15 K) in J/mol (NIST formula yields kJ/mol)
    h_kj_mol = A * t + B * (t ** 2) / 2.0 + C * (t ** 3) / 3.0 + D * (t ** 4) / 4.0 - E / t + F - H
    h_j_mol = h_kj_mol * 1000.0
    return cp, h_j_mol


def cunningham_slip(d_p: float, mean_free_path: float = 6.6e-8) -> float:
    """Davies (1945) / Allen-Raabe Cunningham slip correction factor."""
    kn = 2.0 * mean_free_path / d_p
    return 1.0 + kn * (1.257 + 0.4 * math.exp(-1.1 / kn))


def calvert_penetration(
    d_p: float,
    mu_g: float,
    d_droplet: float,
    v_throat: float,
    rho_p: float,
    L_G: float,
    rho_L: float = 1000.0,
    f_calvert: float = 0.5,
    mean_free_path: float = 6.6e-8,
) -> float:
    """Calvert (1970, 1972) venturi scrubber fractional particle penetration Pt(d_p).

    Pt = exp( (2 * L_G * rho_L * v_throat * d_droplet / (55 * mu_g)) * F(Kp, f) )
    where:
      Kp = Cc * rho_p * d_p^2 * v_throat / (9 * mu_g * d_droplet)
      F(Kp, f) = (1/Kp) * (-0.7 - Kp*f + 1.4*ln((Kp*f + 0.7)/0.7) + 0.49/(0.7 + Kp*f))
    """
    Cc = cunningham_slip(d_p, mean_free_path)
    Kp = (Cc * rho_p * (d_p ** 2) * v_throat) / (9.0 * mu_g * d_droplet)
    x = Kp * f_calvert
    F = (-0.7 - x + 1.4 * math.log((0.7 + x) / 0.7) + 0.49 / (0.7 + x)) / Kp
    exponent = (2.0 * L_G * rho_L * v_throat * d_droplet / (55.0 * mu_g)) * F
    return math.exp(exponent)


def calvert_cut_diameter(
    mu_g: float,
    d_droplet: float,
    v_throat: float,
    rho_p: float,
    L_G: float,
    rho_L: float = 1000.0,
    f_calvert: float = 0.5,
    mean_free_path: float = 6.6e-8,
) -> float:
    """Solve for particle cut diameter d50 [m] where Calvert penetration Pt(d50) = 0.5."""
    def residual(d_p: float) -> float:
        return calvert_penetration(
            d_p=d_p,
            mu_g=mu_g,
            d_droplet=d_droplet,
            v_throat=v_throat,
            rho_p=rho_p,
            L_G=L_G,
            rho_L=rho_L,
            f_calvert=f_calvert,
            mean_free_path=mean_free_path,
        ) - 0.5

    d50 = brentq(residual, 1e-9, 1e-3, xtol=1e-20, rtol=1e-14)
    assert abs(residual(d50)) < 1e-11, f"Calvert root check failed: residual={residual(d50)}"
    return d50


def ergun_dp(dp: float, voidage: float, vs: float, rho: float, mu: float, L: float) -> float:
    """Compute total packed-bed pressure drop [Pa] via fluids.packed_bed."""
    dp_total = packed_bed.dP_packed_bed(
        dp=dp, voidage=voidage, vs=vs, rho=rho, mu=mu, L=L, Method="Ergun"
    )
    # Cross-check against direct fluids.packed_bed.Ergun
    sig = inspect.signature(packed_bed.Ergun)
    if "L" in sig.parameters:
        dp_direct = packed_bed.Ergun(dp=dp, voidage=voidage, vs=vs, rho=rho, mu=mu, L=L)
    else:
        dp_direct = packed_bed.Ergun(dp=dp, voidage=voidage, vs=vs, rho=rho, mu=mu) * L
    assert abs(dp_total - dp_direct) / abs(dp_total) < 1e-12, (
        f"Ergun cross-check mismatch: dP_packed_bed={dp_total}, direct={dp_direct}"
    )
    return dp_total


def assert_finite(obj, path="root"):
    if obj is None:
        raise ValueError(f"Null value at {path}")
    if isinstance(obj, float):
        if math.isnan(obj) or math.isinf(obj):
            raise ValueError(f"Non-finite float at {path}: {obj}")
    elif isinstance(obj, dict):
        for k, v in obj.items():
            assert_finite(v, f"{path}.{k}")
    elif isinstance(obj, list):
        for i, v in enumerate(obj):
            assert_finite(v, f"{path}[{i}]")


def generate_fixtures():
    # Gather critical properties & dipole moments from `chemicals` for Lucas_gas
    species_constants = {}
    for name, cas in SPECIES_CAS.items():
        tc = float(critical.Tc(cas))
        pc = float(critical.Pc(cas))
        zc = float(critical.Zc(cas))
        mw = float(identifiers.MW(cas))
        dip = dipole.dipole_moment(cas)
        dip = 0.0 if dip is None else float(dip)
        species_constants[name] = {
            "CASRN": cas,
            "Tc_K": tc,
            "Pc_Pa": pc,
            "Zc": zc,
            "MW_g_mol": mw,
            "dipole_D": dip,
        }

    fixtures = {
        "metadata": {
            "chemicals_version": chemicals.__version__,
            "fluids_version": fluids.__version__,
            "cp_method": "chemicals.heat_capacity.Shomate (NIST Chemistry WebBook coefficients)",
            "viscosity_method": "chemicals.viscosity.Lucas_gas",
            "vapor_pressure_method": "chemicals.iapws.iapws95_Psat (IAPWS-95)",
            "ergun_method": "fluids.packed_bed.dP_packed_bed(Method='Ergun')",
            "calvert_method": "Calvert (1970, 1972) penetration integral with Davies/Allen-Raabe Cunningham slip (scipy.optimize.brentq)",
        },
        "species_constants": species_constants,
        "gas_properties": [],
        "water_vapor_pressure": [],
        "ergun_packed_bed": [],
        "calvert_venturi": [],
    }

    # 1. Flue gas species Cp (J/mol/K), sensible enthalpy H(T)-H(298.15) (J/mol), & viscosity (Pa*s)
    for T in [350.0, 500.0, 900.0, 1300.0, 1588.15]:
        entry = {"T_K": T, "species": {}}
        for name, cas in SPECIES_CAS.items():
            cp, h_j_mol = shomate_cp_and_enthalpy(name, T)
            c = species_constants[name]
            mu = float(
                viscosity.Lucas_gas(
                    T=T,
                    Tc=c["Tc_K"],
                    Pc=c["Pc_Pa"],
                    Zc=c["Zc"],
                    MW=c["MW_g_mol"],
                    dipole=c["dipole_D"],
                    CASRN=cas,
                )
            )
            entry["species"][name] = {
                "Cp_J_mol_K": float(cp),
                "H_minus_H298_J_mol": float(h_j_mol),
                "mu_Pa_s": mu,
            }
        fixtures["gas_properties"].append(entry)

    # 2. Water saturation vapor pressure (IAPWS-95) across 300-373.15 K
    for T in [300.0, 310.0, 320.0, 330.0, 340.0, 347.15, 350.0, 360.0, 370.0, 373.15]:
        psat = float(iapws.iapws95_Psat(T))
        fixtures["water_vapor_pressure"].append({
            "T_K": T,
            "Psat_Pa": psat,
        })

    # 3. Packed-bed pressure drop (Ergun equation for 1" / 0.0254 m Pall rings, voidage=0.90)
    for vs in [0.5, 1.0, 1.5, 2.0]:
        dp = float(ergun_dp(dp=0.0254, voidage=0.90, vs=vs, rho=1.05, mu=2.05e-5, L=1.2))
        fixtures["ergun_packed_bed"].append({
            "dp_m": 0.0254,
            "voidage": 0.90,
            "vs_m_s": vs,
            "rho_kg_m3": 1.05,
            "mu_Pa_s": 2.05e-5,
            "L_m": 1.2,
            "dP_Pa": dp,
        })

    # 4. Venturi particulate cut diameter (Calvert 1970/1972 model, NaCl rho_p=2160 kg/m3)
    for v_throat in [50.0, 70.0, 90.0]:
        for L_G in [0.0008, 0.0015]:
            d50 = float(
                calvert_cut_diameter(
                    mu_g=2.1e-5,
                    d_droplet=50e-6,
                    v_throat=v_throat,
                    rho_p=2160.0,
                    L_G=L_G,
                    rho_L=1000.0,
                    f_calvert=0.5,
                    mean_free_path=6.6e-8,
                )
            )
            fixtures["calvert_venturi"].append({
                "mu_g_Pa_s": 2.1e-5,
                "d_droplet_m": 50e-6,
                "v_throat_m_s": v_throat,
                "rho_p_kg_m3": 2160.0,
                "L_G": L_G,
                "rho_L_kg_m3": 1000.0,
                "f_calvert": 0.5,
                "mean_free_path_m": 6.6e-8,
                "d50_m": d50,
            })

    assert_finite(fixtures)

    repo_root = Path(__file__).resolve().parent.parent
    out_path = repo_root / "testdata" / "oracle_fixtures.json"
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with open(out_path, "w", encoding="utf-8", newline="\n") as f:
        json.dump(fixtures, f, indent=2)
        f.write("\n")

    print(f"Wrote {out_path} successfully.")


if __name__ == "__main__":
    generate_fixtures()
