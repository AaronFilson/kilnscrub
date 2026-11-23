#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = ["pillow", "matplotlib"]
# ///
"""Generate deterministic, high-craft animated engineering GIFs for kilnscrub.

Outputs (all in docs/assets/):
  1. docs/assets/scrubber-4stage-flow.gif      (960x500, 32 frames)
  2. docs/assets/failsafe-interlock-demo.gif   (960x480, 36 frames)
  3. docs/assets/stage-comparison-tradeoff.gif (960x480, 36 frames)

All engineering figures match the verified 24 cu.ft. wood/salt reference design
in pkg/scrubber/scrubber.go and docs/ENGINEERING_VALIDATION.md.

Usage:
  uv run scripts/generate_gifs.py
  # or with --preview-dir /tmp/preview to also save key PNG frames for inspection
"""

from __future__ import annotations

import argparse
import importlib.util
import math
import os
from pathlib import Path
from typing import Sequence

from PIL import Image, ImageDraw, ImageFont

# ---------------------------------------------------------------------------
# Color Palette — High-Contrast Dark-Slate Engineering Studio Theme
# ---------------------------------------------------------------------------
BG = (11, 17, 32)             # #0B1120 deep slate
BG_HEADER = (15, 23, 42)      # #0F172A header bar
GRID_LINE = (20, 31, 52)      # #141F34 blueprint grid
CARD_BG = (19, 28, 49)        # #131C31 panel surface
CARD_BG_ALT = (24, 36, 61)    # #18243D elevated surface
BORDER = (42, 58, 86)         # #2A3A56 subtle border
BORDER_HI = (71, 95, 134)     # #475F86 highlight border

TEXT_PRI = (248, 250, 252)    # #F8FAFC crisp white
TEXT_SEC = (203, 213, 225)    # #CBD5E1 light slate
TEXT_MUTED = (148, 163, 184)  # #94A3B8 muted label
TEXT_DIM = (100, 116, 139)    # #64748B dim

ORANGE = (249, 115, 22)       # #F97316 hot flue gas
ORANGE_LT = (251, 146, 60)    # #FB923C ember highlight
YELLOW = (251, 191, 36)       # #FBBF24 HCl / warning
RED = (239, 68, 68)           # #EF4444 alarm / trip
RED_DARK = (127, 29, 29)      # #7F1D1D alarm surface
CYAN = (56, 189, 248)         # #38BDF8 quench spray / water
BLUE = (14, 165, 233)         # #0EA5E9 deep water
COBALT = (96, 165, 250)       # #60A5FA blower / draft
VIOLET = (167, 139, 250)      # #A78BFA venturi / kinetic
EMERALD = (16, 185, 129)      # #10B981 safe / absorber / NaHCO3
MINT = (52, 211, 153)         # #34D399 bright safe
STEEL = (148, 163, 184)       # #94A3B8 316L SS shell
STEEL_DK = (71, 85, 105)      # #475569 dark metal
CPVC_WALL = (125, 211, 252)   # #7DD3FC CPVC/PP column highlight
WHITE = (255, 255, 255)

SS = 2  # 2x supersampling factor for razor-sharp anti-aliased rendering


# ---------------------------------------------------------------------------
# Font Discovery & Helpers
# ---------------------------------------------------------------------------
def _find_font_dir() -> Path | None:
    spec = importlib.util.find_spec("matplotlib")
    if spec and spec.origin:
        candidate = Path(spec.origin).parent / "mpl-data" / "fonts" / "ttf"
        if candidate.is_dir():
            return candidate
    for p in (
        Path("/usr/share/fonts/truetype/dejavu"),
        Path("/usr/share/fonts/TTF"),
        Path("/usr/local/share/fonts"),
    ):
        if p.is_dir():
            return p
    return None


FONT_DIR = _find_font_dir()
_FONT_CACHE: dict[tuple[str, int], ImageFont.FreeTypeFont | ImageFont.ImageFont] = {}


def get_font(style: str, size: int) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    """Return a cached TrueType font scaled by SS (supersampling factor)."""
    key = (style, size)
    if key in _FONT_CACHE:
        return _FONT_CACHE[key]
    names = {
        "regular": "DejaVuSans.ttf",
        "bold": "DejaVuSans-Bold.ttf",
        "mono": "DejaVuSansMono.ttf",
        "mono_bold": "DejaVuSansMono-Bold.ttf",
    }
    px = max(8, int(round(size * SS)))
    font = None
    if FONT_DIR is not None:
        fpath = FONT_DIR / names.get(style, "DejaVuSans.ttf")
        if fpath.is_file():
            font = ImageFont.truetype(str(fpath), px)
    if font is None:
        font = ImageFont.load_default(size=px)
    _FONT_CACHE[key] = font
    return font


def s(v: float) -> int:
    return int(round(v * SS))


def sf(v: float) -> float:
    return v * SS


def lerp(a: float, b: float, t: float) -> float:
    return a + (b - a) * t


def clamp(v: float, lo: float = 0.0, hi: float = 1.0) -> float:
    return max(lo, min(hi, v))


def ease_in_out(t: float) -> float:
    t = clamp(t, 0.0, 1.0)
    return t * t * (3.0 - 2.0 * t)


def lerp_color(c1: tuple[int, int, int], c2: tuple[int, int, int], t: float) -> tuple[int, int, int]:
    t = clamp(t)
    return (
        int(round(lerp(c1[0], c2[0], t))),
        int(round(lerp(c1[1], c2[1], t))),
        int(round(lerp(c1[2], c2[2], t))),
    )


def rgba(c: tuple[int, int, int], alpha: int) -> tuple[int, int, int, int]:
    return (c[0], c[1], c[2], clamp_byte(alpha))


def clamp_byte(v: float) -> int:
    return max(0, min(255, int(round(v))))


def new_canvas(w: int, h: int) -> tuple[Image.Image, ImageDraw.ImageDraw]:
    img = Image.new("RGBA", (w * SS, h * SS), (*BG, 255))
    draw = ImageDraw.Draw(img, "RGBA")
    # Subtle engineering grid
    step = 24
    for x in range(0, w, step):
        draw.line([(s(x), 0), (s(x), s(h))], fill=(*GRID_LINE, 180), width=1)
    for y in range(0, h, step):
        draw.line([(0, s(y)), (s(w), s(y))], fill=(*GRID_LINE, 180), width=1)
    return img, draw


def draw_rrect(
    draw: ImageDraw.ImageDraw,
    box: tuple[float, float, float, float],
    radius: float,
    fill: tuple[int, ...] | None = None,
    outline: tuple[int, ...] | None = None,
    width: float = 1.2,
) -> None:
    x0, y0, x1, y1 = box
    draw.rounded_rectangle(
        [s(x0), s(y0), s(x1), s(y1)],
        radius=s(radius),
        fill=fill,
        outline=outline,
        width=max(1, s(width)) if outline else 0,
    )


def draw_text(
    draw: ImageDraw.ImageDraw,
    xy: tuple[float, float],
    text: str,
    style: str = "regular",
    size: int = 12,
    fill: tuple[int, ...] = TEXT_PRI,
    anchor: str = "la",
) -> None:
    font = get_font(style, size)
    draw.text((s(xy[0]), s(xy[1])), text, font=font, fill=fill, anchor=anchor)


def draw_pill(
    draw: ImageDraw.ImageDraw,
    box: tuple[float, float, float, float],
    text: str,
    bg: tuple[int, ...],
    fg: tuple[int, ...] = TEXT_PRI,
    border: tuple[int, ...] | None = None,
    style: str = "bold",
    size: int = 10,
    radius: float = 6,
) -> None:
    draw_rrect(draw, box, radius=radius, fill=bg, outline=border, width=1.1)
    cx = (box[0] + box[2]) / 2.0
    cy = (box[1] + box[3]) / 2.0
    draw_text(draw, (cx, cy), text, style=style, size=size, fill=fg, anchor="mm")


def draw_arrow(
    draw: ImageDraw.ImageDraw,
    x0: float,
    y0: float,
    x1: float,
    y1: float,
    color: tuple[int, ...],
    width: float = 2.0,
    head: float = 6.0,
) -> None:
    draw.line([(s(x0), s(y0)), (s(x1), s(y1))], fill=color, width=max(1, s(width)))
    ang = math.atan2(y1 - y0, x1 - x0)
    a1 = ang + math.radians(150)
    a2 = ang - math.radians(150)
    pts = [
        (s(x1), s(y1)),
        (s(x1 + head * math.cos(a1)), s(y1 + head * math.sin(a1))),
        (s(x1 + head * math.cos(a2)), s(y1 + head * math.sin(a2))),
    ]
    draw.polygon(pts, fill=color)


def finalize_frame(img_rgba: Image.Image, target_w: int, target_h: int) -> Image.Image:
    rgb = Image.new("RGB", img_rgba.size, BG)
    rgb.paste(img_rgba, mask=img_rgba.split()[3])
    return rgb.resize((target_w, target_h), Image.Resampling.LANCZOS)


def save_gif(
    frames: Sequence[Image.Image],
    out_path: Path,
    durations_ms: int | Sequence[int],
    colors: int = 240,
) -> None:
    """Save frames using a unified global palette + lossless transparent delta compression.

    Frame 0 is stored fully opaque; subsequent frames replace unchanged pixels with
    transparent index `trans_idx` (disposal=1), slashing GIF size by 65-80% with
    zero visual quality loss and zero palette-shift flicker.
    """
    out_path.parent.mkdir(parents=True, exist_ok=True)
    # Build a representative mosaic of evenly sampled frames to derive a clean global palette
    sample_indices = sorted(
        set(int(round(i * (len(frames) - 1) / min(7, len(frames) - 1))) for i in range(min(8, len(frames))))
    )
    w, h = frames[0].size
    sw, sh = w // 2, h // 2
    cols = 4
    rows = math.ceil(len(sample_indices) / cols)
    mosaic = Image.new("RGB", (sw * cols, sh * rows), BG)
    for idx, fi in enumerate(sample_indices):
        r, c = divmod(idx, cols)
        mosaic.paste(frames[fi].resize((sw, sh), Image.Resampling.BILINEAR), (c * sw, r * sh))

    # Reserve palette index 255 for transparent delta pixels
    max_colors = min(colors, 255)
    palette_img = mosaic.quantize(colors=max_colors, method=Image.Quantize.MEDIANCUT, dither=Image.Dither.NONE)
    pal = palette_img.getpalette()
    if pal is None:
        raise RuntimeError("Failed to build GIF palette")
    if len(pal) < 768:
        pal = pal + [0] * (768 - len(pal))
    trans_idx = 255

    quantized = [f.quantize(palette=palette_img, dither=Image.Dither.NONE) for f in frames]

    # Build delta-compressed P-mode frames: frame 0 opaque, frames 1..N-1 transparent where unchanged
    out_frames: list[Image.Image] = []
    f0 = quantized[0].copy()
    f0.putpalette(pal)
    out_frames.append(f0)
    prev_bytes = f0.tobytes()

    for qf in quantized[1:]:
        cur_bytes = qf.tobytes()
        diff = bytes(trans_idx if c == p else c for c, p in zip(cur_bytes, prev_bytes))
        df = Image.frombytes("P", (w, h), diff)
        df.putpalette(pal)
        df.info["transparency"] = trans_idx
        out_frames.append(df)
        prev_bytes = cur_bytes

    out_frames[0].save(
        out_path,
        save_all=True,
        append_images=out_frames[1:],
        duration=durations_ms,
        loop=0,
        optimize=True,
        disposal=1,
    )


def eval_polyline(pts: Sequence[tuple[float, float]], u: float) -> tuple[float, float]:
    """Evaluate position along a piecewise-linear path for u in [0, 1]."""
    segs = []
    total = 0.0
    for i in range(len(pts) - 1):
        dx = pts[i + 1][0] - pts[i][0]
        dy = pts[i + 1][1] - pts[i][1]
        d = math.hypot(dx, dy)
        segs.append((pts[i], pts[i + 1], d))
        total += d
    target = clamp(u) * total
    acc = 0.0
    for p0, p1, d in segs:
        if acc + d >= target or d == 0:
            frac = 0.0 if d == 0 else (target - acc) / d
            return (lerp(p0[0], p1[0], frac), lerp(p0[1], p1[1], frac))
        acc += d
    return pts[-1]


# ===========================================================================
# GIF 1: 4-Stage Wet Scrubber Cutaway Flow (960x500, 32 frames)
# ===========================================================================
def render_scrubber_flow_frame(frame_idx: int, total_frames: int) -> Image.Image:
    W, H = 960, 500
    t = frame_idx / total_frames  # [0, 1) seamless periodic phase
    img, draw = new_canvas(W, H)

    # --- Header Bar ---
    draw.rectangle([0, 0, s(W), s(48)], fill=(*BG_HEADER, 255))
    draw.line([(0, s(48)), (s(W), s(48))], fill=(*BORDER, 255), width=s(1.2))
    draw_text(draw, (18, 15), "kilnscrub — 4-Stage Wet Scrubber Process Flow", "bold", 16, TEXT_PRI)
    draw_text(draw, (18, 34), "Verified First-Principles Cutaway · 24 cu.ft. Wood/Salt Studio Kiln", "regular", 10, TEXT_MUTED)
    draw_pill(
        draw,
        (600, 11, 944, 37),
        "CONE 11 · 2400°F · 2.0 lbs/hr NaCl · 256 ACFM",
        bg=(25, 39, 68, 255),
        fg=CYAN,
        border=(56, 189, 248, 120),
        style="mono_bold",
        size=10,
        radius=12,
    )

    # --- Top Stage Header Pills (y = 58..80) ---
    draw_pill(draw, (16, 58, 134, 80), "KILN FIREBOX", bg=(67, 24, 14, 230), fg=ORANGE_LT, border=(249, 115, 22, 160), size=9.5)
    draw_pill(draw, (160, 58, 304, 80), "1 · ADIABATIC QUENCH", bg=(14, 46, 74, 230), fg=CYAN, border=(56, 189, 248, 160), size=9.5)
    draw_pill(draw, (330, 58, 492, 80), "2 · WET VENTURI", bg=(43, 28, 74, 230), fg=VIOLET, border=(167, 139, 250, 160), size=9.5)
    draw_pill(draw, (518, 58, 690, 80), "3 · PACKED BED + DEMISTER", bg=(12, 56, 44, 230), fg=MINT, border=(16, 185, 129, 160), size=9.5)
    draw_pill(draw, (718, 58, 944, 80), "4 · ID BLOWER + PLUME DILUTER", bg=(20, 44, 82, 230), fg=COBALT, border=(96, 165, 250, 160), size=9.5)

    # --- Connecting Ducts (drawn behind equipment) ---
    duct_fill = (24, 34, 54, 255)
    # Kiln -> Quench top duct
    draw.rectangle([s(110), s(142), s(190), s(166)], fill=(65, 28, 18, 255), outline=(*ORANGE, 180), width=s(1.5))
    # Quench bottom -> Venturi inlet
    draw.rectangle([s(280), s(248), s(336), s(274)], fill=duct_fill, outline=(*STEEL, 190), width=s(1.5))
    # Venturi exit -> Packed bed bottom plenum
    draw.rectangle([s(486), s(248), s(540), s(274)], fill=duct_fill, outline=(*STEEL, 190), width=s(1.5))
    # Packed bed top -> Blower suction
    draw.polygon(
        [(s(668), s(114)), (s(754), s(114)), (s(754), s(212)), (s(732), s(212)), (s(732), s(136)), (s(668), s(136))],
        fill=(18, 39, 52, 255),
        outline=(*CPVC_WALL, 180),
        width=s(1.4),
    )

    # --- 0. Studio Kiln Firebox (x = 18..126, y = 134..324) ---
    # Outer brick body + arched crown
    draw.pieslice([s(20), s(134), s(124), s(206)], 180, 360, fill=(49, 35, 38, 255), outline=(140, 82, 65, 255), width=s(2))
    draw.rectangle([s(20), s(170), s(124), s(324)], fill=(49, 35, 38, 255), outline=(140, 82, 65, 255), width=s(2))
    # Subtle brick courses
    for by in range(186, 320, 18):
        draw.line([(s(22), s(by)), (s(122), s(by))], fill=(75, 50, 48, 160), width=s(1))
    # Inner glowing fire chamber
    glow_pulse = 0.5 + 0.5 * math.sin(2 * math.pi * t)
    fire_outer = lerp_color((220, 62, 18), (249, 115, 22), glow_pulse)
    fire_inner = lerp_color((251, 146, 60), (254, 240, 138), glow_pulse)
    draw.pieslice([s(32), s(160), s(112), s(224)], 180, 360, fill=(*fire_outer, 235))
    draw.rectangle([s(32), s(192), s(112), s(310)], fill=(*fire_outer, 235))
    # Animated flame tongues inside kiln
    for fi in range(4):
        fx = 44 + fi * 18
        fh = 38 + 14 * math.sin(2 * math.pi * (t * 2 + fi * 0.27))
        pts = [
            (s(fx - 8), s(310)),
            (s(fx - 3), s(310 - fh * 0.65)),
            (s(fx + 2 * math.sin(2 * math.pi * (t + fi * 0.3))), s(310 - fh)),
            (s(fx + 8), s(310)),
        ]
        draw.polygon(pts, fill=(*fire_inner, 230))
    # Kiln temperature & draft label inside/below
    draw_pill(draw, (30, 212, 114, 236), "2400°F", bg=(24, 14, 14, 215), fg=(254, 215, 170), border=(*ORANGE, 200), style="mono_bold", size=12)
    draw_text(draw, (72, 338), "579 ACFM · -0.05\"", "mono_bold", 9.5, ORANGE_LT, "mm")
    draw_text(draw, (72, 353), "Wood + Salt Flue", "regular", 9.5, TEXT_MUTED, "mm")

    # --- 1. Stage 1: Adiabatic Quench Tower (x = 180..284, y = 104..324) ---
    # 316L SS shell with warm-to-cool vertical gradient interior
    qx0, qy0, qx1, qy1 = 180, 104, 284, 324
    # Vertical temperature gradient bands inside tower
    bands = 16
    for bi in range(bands):
        frac = bi / (bands - 1)
        by0 = lerp(qy0 + 4, qy1 - 24, bi / bands)
        by1 = lerp(qy0 + 4, qy1 - 24, (bi + 1) / bands)
        c = lerp_color((92, 34, 20), (14, 52, 82), ease_in_out(frac))
        draw.rectangle([s(qx0 + 3), s(by0), s(qx1 - 3), s(by1 + 1)], fill=(*c, 225))
    # Bottom liquid pool in quench sump
    draw.rectangle([s(qx0 + 3), s(qy1 - 22), s(qx1 - 3), s(qy1 - 3)], fill=(14, 90, 138, 220))
    # Shell wall outline (316L SS)
    draw_rrect(draw, (qx0, qy0, qx1, qy1), radius=8, fill=None, outline=(*STEEL, 240), width=2.4)
    # Top spray manifold & 2 full-cone nozzles
    draw.line([(s(232), s(90)), (s(232), s(118))], fill=(*CYAN, 255), width=s(2.4))
    draw.line([(s(206), s(118)), (s(258), s(118))], fill=(*CYAN, 255), width=s(2.4))
    for nx in (208, 256):
        draw.polygon([(s(nx - 5), s(118)), (s(nx + 5), s(118)), (s(nx + 2), s(126)), (s(nx - 2), s(126))], fill=(*CYAN, 255))
        # Translucent spray cone mist
        draw.polygon(
            [(s(nx), s(126)), (s(nx - 24), s(225)), (s(nx + 24), s(225))],
            fill=(56, 189, 248, 34),
        )
        # Animated falling spray droplets inside cone
        for di in range(8):
            du = (t * 2.0 + di / 8.0 + (0.06 if nx > 230 else 0.0)) % 1.0
            spread = (di % 3 - 1) * 14 * du
            dx = nx + spread
            dy = 128 + du * 145
            alpha = int(235 * (1.0 - 0.55 * du))
            draw.ellipse([s(dx - 2), s(dy - 3), s(dx + 2), s(dy + 3)], fill=(56, 189, 248, alpha))
    # Quench callouts
    draw_pill(draw, (191, 290, 273, 312), "183.9°F", bg=(10, 32, 54, 230), fg=CYAN, border=(*CYAN, 180), style="mono_bold", size=11)
    draw_text(draw, (232, 338), "1.16 GPM Spray", "mono_bold", 9.5, CYAN, "mm")
    draw_text(draw, (232, 353), "14\" 316L SS · 0.41\"", "regular", 9.5, TEXT_MUTED, "mm")

    # --- 2. Stage 2: Wet Venturi Scrubber (x = 334..488, y = 212..310) ---
    # Converging-Throat-Diverging cutaway polygon
    vy_mid = 261
    venturi_poly = [
        (s(334), s(vy_mid - 26)),
        (s(386), s(vy_mid - 11)),
        (s(424), s(vy_mid - 11)),
        (s(488), s(vy_mid - 28)),
        (s(488), s(vy_mid + 28)),
        (s(424), s(vy_mid + 11)),
        (s(386), s(vy_mid + 11)),
        (s(334), s(vy_mid + 26)),
    ]
    draw.polygon(venturi_poly, fill=(28, 25, 56, 240), outline=(*VIOLET, 220), width=s(2.2))
    # Glowing high-velocity throat zone
    draw.rectangle([s(386), s(vy_mid - 9), s(424), s(vy_mid + 9)], fill=(139, 92, 246, 90))
    # High-velocity streamlines in throat
    for si in range(3):
        sy = vy_mid - 5 + si * 5
        sx = 384 + ((t * 4.0 + si * 0.33) % 1.0) * 32
        draw.line([(s(sx), s(sy)), (s(sx + 10), s(sy))], fill=(221, 214, 254, 210), width=s(1.6))
    # Top & bottom opposing liquid injection jets at throat entry (x=392)
    draw.line([(s(392), s(vy_mid - 32)), (s(392), s(vy_mid - 12))], fill=(*CYAN, 255), width=s(2.4))
    draw.line([(s(392), s(vy_mid + 32)), (s(392), s(vy_mid + 12))], fill=(*CYAN, 255), width=s(2.4))
    # Sheared mist cone in diverging diffuser
    draw.polygon(
        [(s(394), s(vy_mid - 8)), (s(482), s(vy_mid - 22)), (s(482), s(vy_mid + 22)), (s(394), s(vy_mid + 8))],
        fill=(56, 189, 248, 38),
    )
    # Info card neatly positioned above Venturi (y = 104..198)
    draw_rrect(draw, (332, 104, 490, 198), radius=8, fill=(24, 22, 46, 235), outline=(*VIOLET, 130), width=1.2)
    draw_text(draw, (411, 120), "HIGH-VELOCITY THROAT", "bold", 9.5, VIOLET, "mm")
    draw_text(draw, (411, 139), "70 m/s  ·  18.88 inWC", "mono_bold", 10.5, TEXT_PRI, "mm")
    draw_text(draw, (411, 158), "d50 = 0.21 µm (NaCl)", "mono_bold", 10, CYAN, "mm")
    draw_text(draw, (411, 177), "84.2% NaCl · 96.5% Ash", "regular", 9.5, TEXT_SEC, "mm")
    draw.line([(s(405), s(198)), (s(405), s(244))], fill=(*VIOLET, 130), width=s(1.2))
    draw_text(draw, (411, 338), "2.30 GPM Sheared Spray", "mono_bold", 9.5, VIOLET, "mm")
    draw_text(draw, (411, 353), "1.85\" Throat · 316L SS", "regular", 9.5, TEXT_MUTED, "mm")

    # --- 3. Stage 3: Packed-Bed HCl Absorber + Chevron Demister (x = 536..672, y = 94..324) ---
    px0, py0, px1, py1 = 536, 94, 672, 324
    draw_rrect(draw, (px0, py0, px1, py1), radius=8, fill=(14, 34, 42, 240), outline=(*CPVC_WALL, 220), width=2.4)
    # Bottom NaHCO3 sump liquid (y = 294..321)
    draw.rectangle([s(px0 + 3), s(294), s(px1 - 3), s(py1 - 3)], fill=(16, 135, 100, 170))
    draw_text(draw, (604, 308), "NaHCO₃ Sump pH 8.0", "mono_bold", 8.5, (209, 250, 229), "mm")

    # Packing bed zone (y = 178..258) — 1" PP Pall Rings texture
    bed_y0, bed_y1 = 178, 258
    draw.rectangle([s(px0 + 5), s(bed_y0), s(px1 - 5), s(bed_y1)], fill=(15, 62, 54, 210), outline=(*EMERALD, 170), width=s(1.2))
    # Cross-hatch / Pall ring circles inside bed
    for ry in range(bed_y0 + 10, bed_y1 - 4, 14):
        offset = 7 if ((ry // 14) % 2) else 0
        for rx in range(px0 + 14 + offset, px1 - 10, 16):
            draw.ellipse([s(rx - 5), s(ry - 5), s(rx + 5), s(ry + 5)], outline=(52, 211, 153, 135), width=s(1.2))
    # Support grate & hold-down grid
    draw.line([(s(px0 + 4), s(bed_y1)), (s(px1 - 4), s(bed_y1))], fill=(*MINT, 240), width=s(2.4))
    draw.line([(s(px0 + 4), s(bed_y0)), (s(px1 - 4), s(bed_y0))], fill=(*MINT, 200), width=s(1.6))

    # Counter-current NaHCO3 distributor spray above bed (y = 152)
    draw.line([(s(px0 + 18), s(154)), (s(px1 - 18), s(154))], fill=(*MINT, 240), width=s(2.4))
    for di in range(10):
        du = (t * 2.0 + di / 10.0) % 1.0
        dx = px0 + 22 + (di % 5) * 23 + math.sin(di * 1.7) * 3
        dy = 156 + du * 132
        draw.ellipse([s(dx - 1.8), s(dy - 2.6), s(dx + 1.8), s(dy + 2.6)], fill=(52, 211, 153, 200))

    # Chevron Mist Eliminator (Demister) at top (y = 114..136)
    draw.rectangle([s(px0 + 6), s(114), s(px1 - 6), s(136)], fill=(20, 50, 68, 220), outline=(*CYAN, 160), width=s(1))
    for cx in range(px0 + 14, px1 - 12, 16):
        pts = [(s(cx), s(134)), (s(cx + 6), s(125)), (s(cx), s(116))]
        draw.line(pts, fill=(*CYAN, 230), width=s(1.8))
    draw_text(draw, (604, 104), "CHEVRON DEMISTER", "bold", 8, CYAN, "mm")

    # Bed metric pill centered on column
    draw_pill(draw, (548, 206, 660, 230), "99.04% HCl", bg=(7, 36, 30, 235), fg=MINT, border=(*EMERALD, 200), style="mono_bold", size=11)
    draw_text(draw, (604, 338), "55.5% Flood · 1.70\"", "mono_bold", 9.5, MINT, "mm")
    draw_text(draw, (604, 353), "12\" CPVC · 1\" Pall Rings", "regular", 9.5, TEXT_MUTED, "mm")

    # --- 4. Stage 4: ID Blower + Optional HV Squirrel-Cage Plume Diluter ---
    # Centrifugal blower volute housing centered at (760, 232)
    bcx, bcy, br = 760, 232, 38
    # Discharge vertical stack rising from right side of volute (x = 782..814, y = 122..232)
    draw.rectangle([s(782), s(122), s(814), s(232)], fill=(20, 38, 64, 240), outline=(*COBALT, 220), width=s(2.0))
    # Volute snail circle
    draw.ellipse([s(bcx - br), s(bcy - br), s(bcx + br), s(bcy + br)], fill=(20, 38, 64, 250), outline=(*COBALT, 240), width=s(2.5))
    # Rotating radial blades (8 blades spinning smoothly 1 full turn per loop -> 45 deg symmetry)
    rot = t * 2 * math.pi  # 1 full rev per loop
    for bi in range(8):
        ang = rot + bi * (math.pi / 4.0)
        x1 = bcx + (br - 7) * math.cos(ang)
        y1 = bcy + (br - 7) * math.sin(ang)
        x0 = bcx + 9 * math.cos(ang)
        y0 = bcy + 9 * math.sin(ang)
        draw.line([(s(x0), s(y0)), (s(x1), s(y1))], fill=(*STEEL, 240), width=s(2.6))
    # Hub center
    draw.ellipse([s(bcx - 10), s(bcy - 10), s(bcx + 10), s(bcy + 10)], fill=(*COBALT, 255), outline=(*BG, 255), width=s(2))

    # Optional HV Squirrel-Cage Plume Diluter unit injecting side ambient air into stack at y = 160
    draw_rrect(draw, (828, 140, 940, 180), radius=7, fill=(18, 42, 72, 240), outline=(*CYAN, 180), width=1.4)
    draw_text(draw, (884, 154), "HV DILUTER (OPT)", "bold", 8.5, CYAN, "mm")
    draw_text(draw, (884, 168), "120V · 1500 CFM", "mono_bold", 8.5, TEXT_PRI, "mm")
    draw_arrow(draw, 828, 160, 814, 160, CYAN, width=2.2, head=5.5)

    # High-velocity exit cone on top of stack (y = 104..122)
    draw.polygon([(s(782), s(122)), (s(788), s(104)), (s(808), s(104)), (s(814), s(122))], fill=(30, 58, 96, 255), outline=(*CYAN, 230), width=s(1.8))

    # Plume Comparison Visualization (top-right, y = 92..130, cleanly below header pill!)
    # Dashed ghost outline showing bare 41 ft plume reach
    for gx in range(818, 940, 10):
        draw.line([(s(gx), s(95)), (s(min(gx + 5, 940)), s(95))], fill=(148, 163, 184, 115), width=s(1.2))
        draw.line([(s(gx), s(123)), (s(min(gx + 5, 940)), s(123))], fill=(148, 163, 184, 115), width=s(1.2))
    draw_text(draw, (910, 109), "Bare 41 ft", "mono", 8.3, TEXT_DIM, "mm")
    # Crisp animated compact diluted plume (12 ft) shooting up & right
    for pi in range(6):
        pu = (t * 2.0 + pi / 6.0) % 1.0
        px = 798 + pu * 66
        py = 108 - 5 * math.sin(pu * math.pi) + 3.5 * math.sin(2 * math.pi * (t + pi * 0.2))
        pr = 4.5 + pu * 8.0
        pa = int(170 * (1.0 - pu) ** 0.8)
        draw.ellipse([s(px - pr), s(py - pr * 0.7), s(px + pr), s(py + pr * 0.7)], fill=(224, 242, 254, pa))

    # Plume shortening callout card (x = 824..944, y = 190..292)
    draw_rrect(draw, (824, 190, 944, 292), radius=8, fill=(16, 30, 52, 235), outline=(*COBALT, 140), width=1.2)
    draw_text(draw, (884, 206), "WINTER PLUME", "bold", 8.5, COBALT, "mm")
    draw_text(draw, (884, 226), "41 ft → 12 ft", "mono_bold", 11.5, CYAN, "mm")
    draw_text(draw, (884, 245), "256 ACFM", "mono_bold", 10, TEXT_PRI, "mm")
    draw_text(draw, (884, 261), "21.21 inWC", "mono_bold", 10, ORANGE_LT, "mm")
    draw_text(draw, (884, 278), "2.0 HP · 240V", "regular", 9, TEXT_SEC, "mm")

    draw_text(draw, (822, 338), "Draft: -0.05 inWC PID", "mono_bold", 9.5, COBALT, "mm")
    draw_text(draw, (822, 353), "FRP/PP Housing · VFD", "regular", 9.5, TEXT_MUTED, "mm")

    # --- Animated Gas, HCl, and NaCl/Ash Particles Flowing Through Entire Train ---
    main_path = [
        (82, 168),
        (134, 154),
        (232, 154),   # Enter Quench top
        (232, 261),   # Down through Quench
        (334, 261),   # Enter Venturi
        (405, 261),   # Throat (70 m/s)
        (488, 261),   # Exit Venturi
        (604, 266),   # Enter Packed Bed bottom
        (604, 125),   # Rise through bed & demister
        (743, 125),   # Overhead duct to blower
        (743, 232),   # Blower suction
        (798, 232),   # Blower discharge
        (798, 108),   # Clean stack exit
    ]

    # 1) Carrier gas particles (28 particles evenly spaced, color transitions hot->cool->clean)
    n_gas = 28
    for i in range(n_gas):
        u = (t + i / n_gas) % 1.0
        px, py = eval_polyline(main_path, u)
        # Slight transverse swirl
        wobble = 4.5 * math.sin(2 * math.pi * (u * 7 + i * 0.37))
        if u < 0.24:
            col = lerp_color(ORANGE, CYAN, clamp((u - 0.12) / 0.12))
            rad = 3.2
        elif u < 0.50:
            col = lerp_color(CYAN, VIOLET, clamp((u - 0.30) / 0.15))
            rad = 2.8
        elif u < 0.72:
            col = lerp_color(VIOLET, MINT, clamp((u - 0.52) / 0.16))
            rad = 2.6
        else:
            col = (186, 230, 253)
            rad = 2.6
        # Keep tight in venturi throat (u ~ 0.34..0.44)
        if 0.33 < u < 0.45:
            wobble *= 0.25
        draw.ellipse(
            [s(px - rad), s(py + wobble - rad), s(px + rad), s(py + wobble + rad)],
            fill=(*col, 220),
        )

    # 2) HCl acid gas molecules (yellow diamonds) — survive Quench & Venturi, REMOVED inside Stage 3 Packed Bed (u ~ 0.56..0.67)!
    n_hcl = 14
    for i in range(n_hcl):
        u = ((t + i / n_hcl) % 1.0) * 0.66
        px, py = eval_polyline(main_path, u)
        wobble = 7.0 * math.cos(2 * math.pi * (u * 6 + i * 0.41))
        if 0.33 < u < 0.45:
            wobble *= 0.25
        alpha = 235 if u < 0.54 else int(235 * clamp((0.66 - u) / 0.12))
        if alpha > 15:
            r = 3.2
            pts = [(s(px), s(py + wobble - r)), (s(px + r), s(py + wobble)), (s(px), s(py + wobble + r)), (s(px - r), s(py + wobble))]
            draw.polygon(pts, fill=(*YELLOW, alpha))

    # 3) Ash & NaCl fume specks (bright white/gray squares) — CAPTURED in Stage 2 Venturi (u ~ 0.37..0.47)!
    n_pm = 12
    for i in range(n_pm):
        u = ((t + i / n_pm) % 1.0) * 0.46
        px, py = eval_polyline(main_path, u)
        wobble = 6.5 * math.sin(2 * math.pi * (u * 5 + i * 0.53))
        if 0.33 < u < 0.45:
            wobble *= 0.2
        alpha = 230 if u < 0.37 else int(230 * clamp((0.46 - u) / 0.09))
        if alpha > 15:
            r = 2.4
            draw.rectangle([s(px - r), s(py + wobble - r), s(px + r), s(py + wobble + r)], fill=(241, 245, 249, alpha))

    # --- Bottom 4 Engineering Spec Cards (y = 374..456) ---
    specs = [
        (
            (16, 374, 240, 456),
            CYAN,
            "STAGE 1 · ADIABATIC QUENCH",
            "2400°F → 183.9°F  |  1.16 GPM",
            "ΔP 0.41 inWC · PID 188°F (210°F trip)",
        ),
        (
            (250, 374, 474, 456),
            VIOLET,
            "STAGE 2 · WET VENTURI",
            "70 m/s Throat  |  d50 = 0.21 µm",
            "ΔP 18.88 inWC · 84% NaCl / 96.5% Ash",
        ),
        (
            (484, 374, 708, 456),
            EMERALD,
            "STAGE 3 · PACKED BED + DEM",
            "99.04% HCl  |  55.5% of Flood",
            "ΔP 1.44 bed + 0.26 dem = 1.70 inWC",
        ),
        (
            (718, 374, 944, 456),
            COBALT,
            "STAGE 4 · BLOWER + DILUTER",
            "256 ACFM @ 21.21 inWC  |  2.0 HP",
            "Draft -0.05\" · Plume 41 ft → 12 ft",
        ),
    ]
    for box, accent, title, line1, line2 in specs:
        draw_rrect(draw, box, radius=7, fill=(*CARD_BG, 245), outline=(*BORDER, 255), width=1.2)
        draw.rectangle([s(box[0] + 1), s(box[1] + 1), s(box[2] - 1), s(box[1] + 4)], fill=(*accent, 230))
        draw_text(draw, (box[0] + 10, box[1] + 19), title, "bold", 9.5, accent, "lm")
        draw_text(draw, (box[0] + 10, box[1] + 42), line1, "mono_bold", 9.8, TEXT_PRI, "lm")
        draw_text(draw, (box[0] + 10, box[1] + 64), line2, "regular", 9.0, TEXT_SEC, "lm")

    # --- Footer Legend Bar (y = 466..492) ---
    draw_rrect(draw, (16, 466, 944, 492), radius=6, fill=(*BG_HEADER, 240), outline=(*BORDER, 180), width=1)
    legend = [
        (32, ORANGE, "circle", "Hot Flue (2400°F)"),
        (178, YELLOW, "diamond", "HCl Acid Gas (99% scrubbed)"),
        (378, (241, 245, 249), "square", "Ash / 0.21 µm NaCl"),
        (536, CYAN, "circle", "Water Spray (1.16 GPM)"),
        (706, MINT, "circle", "NaHCO₃ Liquor (pH 8)"),
        (856, (186, 230, 253), "circle", "Clean Plume"),
    ]
    for lx, lcol, shape, label in legend:
        ly = 479
        if shape == "circle":
            draw.ellipse([s(lx - 4), s(ly - 4), s(lx + 4), s(ly + 4)], fill=(*lcol, 255))
        elif shape == "diamond":
            draw.polygon([(s(lx), s(ly - 4.5)), (s(lx + 4.5), s(ly)), (s(lx), s(ly + 4.5)), (s(lx - 4.5), s(ly))], fill=(*lcol, 255))
        else:
            draw.rectangle([s(lx - 3.5), s(ly - 3.5), s(lx + 3.5), s(ly + 3.5)], fill=(*lcol, 255))
        draw_text(draw, (lx + 9, ly), label, "regular", 9.2, TEXT_SEC, "lm")

    return finalize_frame(img, W, H)


# ===========================================================================
# GIF 2: Fail-Safe Interlock & Bypass Demo (960x480, 36 frames)
# ===========================================================================
def _interlock_profile(n_frames: int = 36) -> list[dict]:
    """Build exact deterministic 3-phase safety telemetry profile across 36 frames."""
    prof = []
    for i in range(n_frames):
        if i < 12:
            # Phase 1: Normal Firing (188°F PID setpoint, 1.16 GPM, Closed damper)
            phase = 1
            temp = 188.0 + 0.6 * math.sin(i * 0.9)
            gpm = 1.16
            tripped = False
            damper_deg = 0.0
            vfd = 82.0
            draft = -0.05
            reset_state = "idle"
        elif i < 26:
            # Phase 2: Water Loss -> Thermal Spike -> >=210°F Trip!
            phase = 2
            k = i - 12  # 0..13
            gpm = 0.0
            if k == 0:
                temp, tripped, damper_deg, vfd, draft = 191.5, False, 0.0, 82.0, -0.05
            elif k == 1:
                temp, tripped, damper_deg, vfd, draft = 201.0, False, 0.0, 82.0, -0.05
            elif k == 2:
                temp, tripped, damper_deg, vfd, draft = 211.8, True, 35.0, 38.0, -0.02
            elif k == 3:
                temp, tripped, damper_deg, vfd, draft = 220.5, True, 75.0, 10.0, -0.01
            elif k <= 8:
                temp = 224.0 - (k - 4) * 1.4
                tripped, damper_deg, vfd, draft = True, 90.0, 0.0, 0.00
            else:
                temp = 217.0 - (k - 8) * 1.6
                tripped, damper_deg, vfd, draft = True, 90.0, 0.0, 0.00
            reset_state = "blocked" if tripped else "idle"
        else:
            # Phase 3: Water Restored -> Cool <190°F -> Manual Reset -> Normal Resumed
            phase = 3
            k = i - 26  # 0..9
            gpm = 1.16
            if k == 0:
                temp, tripped, damper_deg, vfd, draft, reset_state = 204.0, True, 90.0, 0.0, 0.00, "cooling"
            elif k == 1:
                temp, tripped, damper_deg, vfd, draft, reset_state = 195.5, True, 90.0, 0.0, 0.00, "cooling"
            elif k == 2:
                temp, tripped, damper_deg, vfd, draft, reset_state = 188.5, True, 90.0, 0.0, 0.00, "ready"
            elif k == 3:
                temp, tripped, damper_deg, vfd, draft, reset_state = 187.8, True, 90.0, 0.0, 0.00, "ready"
            elif k == 4:
                temp, tripped, damper_deg, vfd, draft, reset_state = 187.9, True, 90.0, 0.0, 0.00, "clicking"
            elif k == 5:
                temp, tripped, damper_deg, vfd, draft, reset_state = 188.0, False, 55.0, 35.0, -0.02, "clicked"
            elif k == 6:
                temp, tripped, damper_deg, vfd, draft, reset_state = 188.1, False, 20.0, 68.0, -0.04, "normal"
            else:
                temp, tripped, damper_deg, vfd, draft, reset_state = 188.0, False, 0.0, 82.0, -0.05, "normal"
        prof.append(
            {
                "phase": phase,
                "temp": temp,
                "gpm": gpm,
                "tripped": tripped,
                "damper_deg": damper_deg,
                "vfd": vfd,
                "draft": draft,
                "reset_state": reset_state,
            }
        )
    return prof


INTERLOCK_PROFILE = _interlock_profile(36)


def render_interlock_frame(frame_idx: int, total_frames: int) -> Image.Image:
    W, H = 960, 480
    st = INTERLOCK_PROFILE[frame_idx]
    phase = st["phase"]
    temp = st["temp"]
    gpm = st["gpm"]
    tripped = st["tripped"]
    damper_deg = st["damper_deg"]
    vfd = st["vfd"]
    draft = st["draft"]
    reset_state = st["reset_state"]
    t_anim = frame_idx / total_frames

    img, draw = new_canvas(W, H)

    # --- Header Bar (y = 0..46) ---
    draw.rectangle([0, 0, s(W), s(46)], fill=(*BG_HEADER, 255))
    draw.line([(0, s(46)), (s(W), s(46))], fill=(*BORDER, 255), width=s(1.2))
    draw_text(draw, (18, 15), "Fail-Safe Thermal Interlock & Bypass Sequence", "bold", 15, TEXT_PRI)
    draw_text(draw, (18, 33), "Protecting CPVC/PP Column (210°F Limit) · Spring-Return Fail-Open Damper", "regular", 9.5, TEXT_MUTED)

    # 3-step Phase Stepper (top-right)
    steps = [
        (1, "1 · Normal Firing", (536, 11, 666, 35), EMERALD),
        (2, "2 · Loss & Trip", (674, 11, 804, 35), RED),
        (3, "3 · Clear & Reset", (812, 11, 944, 35), CYAN),
    ]
    for pnum, label, box, col in steps:
        active = phase == pnum
        bg = (*lerp_color(BG_HEADER, col, 0.28), 255) if active else (20, 30, 50, 220)
        fg = TEXT_PRI if active else TEXT_DIM
        bdr = (*col, 230) if active else (*BORDER, 160)
        draw_pill(draw, box, label, bg=bg, fg=fg, border=bdr, style="bold" if active else "regular", size=9.2, radius=11)

    # --- Dynamic Safety Alarm Banner (y = 54..86) ---
    if tripped and reset_state in ("blocked", "cooling"):
        pulse = 0.5 + 0.5 * math.sin(frame_idx * 1.8)
        b_bg = lerp_color((110, 20, 24), (165, 28, 32), pulse)
        if reset_state == "blocked":
            b_txt = "INTERLOCK TRIPPED (LATCHED) — QUENCH ≥ 210°F / LOW FLOW · BYPASS OPEN · VFD 0%"
        else:
            b_txt = "INTERLOCK LATCHED — WATER RESTORED (1.16 GPM) · COOLING BELOW 190°F BEFORE RESET"
        draw_pill(draw, (16, 54, 944, 86), b_txt, bg=(*b_bg, 250), fg=WHITE, border=(*RED, 255), style="bold", size=10.2, radius=7)
    elif tripped and reset_state in ("ready", "clicking"):
        draw_pill(
            draw,
            (16, 54, 944, 86),
            "FAULT CLEARED (< 190°F & 1.16 GPM) — INTERLOCK LATCHED · OPERATOR CLICK [RESET INTERLOCK]",
            bg=(100, 68, 12, 245),
            fg=(254, 243, 199),
            border=(*YELLOW, 235),
            style="bold",
            size=10.0,
            radius=7,
        )
    elif phase == 2 and not tripped:
        draw_pill(
            draw,
            (16, 54, 944, 86),
            "WARNING — QUENCH WATER LOST (0.00 GPM) · OUTLET TEMPERATURE RISING RAPIDLY TOWARD 210°F!",
            bg=(115, 52, 12, 245),
            fg=(255, 237, 213),
            border=(*ORANGE, 240),
            style="bold",
            size=10.0,
            radius=7,
        )
    else:
        msg = (
            "MANUAL RESET COMPLETE — BYPASS CLOSING · NORMAL PID DRAFT & QUENCH CONTROL RESUMING"
            if phase == 3
            else "NORMAL FIRING — QUENCH PID 188°F · SPRAY 1.16 GPM · BYPASS CLOSED · KILN DRAFT -0.05 inWC"
        )
        draw_pill(draw, (16, 54, 944, 86), msg, bg=(12, 66, 48, 240), fg=(209, 250, 229), border=(*EMERALD, 220), style="bold", size=10.0, radius=7)

    # --- Left Panel: Live Piping, Quench, Spring-Return Bypass & CPVC Column (x = 16..480, y = 96..360) ---
    draw_rrect(draw, (16, 96, 480, 360), radius=9, fill=(*CARD_BG, 245), outline=(*BORDER, 255), width=1.2)
    draw_text(draw, (28, 111), "FAIL-SAFE DUCT & BYPASS SCHEMATIC", "bold", 9.8, TEXT_MUTED, "lm")

    # Bypass Relief Stack going UP from Tee (x = 148..178, y = 114..220)
    byp_active = damper_deg > 15.0
    byp_fill = (84, 30, 18, 240) if byp_active else (22, 30, 48, 230)
    byp_bdr = ORANGE if byp_active else STEEL_DK
    draw.rectangle([s(148), s(126), s(178), s(228)], fill=byp_fill, outline=(*byp_bdr, 230), width=s(1.8))
    draw_text(draw, (238, 134), "SAFE BYPASS RELIEF", "bold", 9.0, ORANGE_LT if byp_active else TEXT_DIM, "lm")
    draw_text(draw, (238, 148), "(316L SS Emergency Flue)", "regular", 8.5, TEXT_MUTED, "lm")
    if byp_active:
        draw_arrow(draw, 163, 174, 163, 132, ORANGE_LT, width=2.4, head=6)

    # Main horizontal hot flue from Kiln (x = 30..216, y = 214..242)
    draw.rectangle([s(74), s(216), s(218), s(242)], fill=(78, 30, 18, 240), outline=(*ORANGE, 210), width=s(1.8))
    # Kiln box
    draw_rrect(draw, (26, 196, 80, 264), radius=6, fill=(76, 30, 22, 250), outline=(*ORANGE, 220), width=1.8)
    draw_text(draw, (53, 222), "KILN", "bold", 10, TEXT_PRI, "mm")
    draw_text(draw, (53, 239), "2400°F", "mono_bold", 9.5, ORANGE_LT, "mm")

    # Tee Junction + Animated Spring-Return Bypass Damper Blade at (163, 228)
    draw_rrect(draw, (144, 212, 182, 246), radius=4, fill=(32, 42, 66, 255), outline=(*STEEL, 230), width=1.6)
    # Pivot at (148, 216): 0° = horizontal sealing UPWARD bypass port (CLOSED), 90° = vertical blocking right path / opening bypass (OPEN)
    piv_x, piv_y = 148, 216
    rad_ang = math.radians(damper_deg)
    bx1 = piv_x + 30 * math.cos(rad_ang)
    by1 = piv_y + 30 * math.sin(rad_ang)
    blade_col = RED if byp_active else EMERALD
    draw.line([(s(piv_x), s(piv_y)), (s(bx1), s(by1))], fill=(*blade_col, 255), width=s(4.2))
    draw.ellipse([s(piv_x - 4), s(piv_y - 4), s(piv_x + 4), s(piv_y + 4)], fill=(*WHITE, 255))
    # Spring-return actuator badge below tee (cleanly separated from Quench tower!)
    draw_pill(
        draw,
        (94, 260, 208, 282),
        f"BYPASS: {'OPEN' if byp_active else 'CLOSED'} ({int(round(damper_deg))}°)",
        bg=(110, 24, 28, 240) if byp_active else (14, 62, 46, 240),
        fg=WHITE if byp_active else MINT,
        border=(*blade_col, 230),
        style="mono_bold",
        size=8.4,
        radius=5,
    )
    draw_text(draw, (151, 294), "Spring-Return Fail-Open", "regular", 8.2, TEXT_MUTED, "mm")

    # Stage 1 Quench Vessel (x = 218..288, y = 186..278)
    q_hot = clamp((temp - 188.0) / 30.0)
    q_fill = lerp_color((16, 52, 80), (112, 34, 22), q_hot)
    draw_rrect(draw, (218, 188, 288, 276), radius=7, fill=(*q_fill, 240), outline=(*STEEL, 230), width=1.8)
    draw_text(draw, (253, 177), "QUENCH", "bold", 8.8, CYAN if gpm > 0 else RED, "mm")
    # Spray water inlet & droplets
    draw.line([(s(253), s(188)), (s(253), s(198))], fill=(*(CYAN if gpm > 0 else RED), 255), width=s(2.4))
    if gpm > 0.1:
        for di in range(6):
            du = ((frame_idx * 0.22 + di / 6.0) % 1.0)
            dx = 253 + (di % 3 - 1) * 14 * du
            dy = 200 + du * 56
            draw.ellipse([s(dx - 2), s(dy - 2.6), s(dx + 2), s(dy + 2.6)], fill=(*CYAN, 225))
    else:
        draw_text(draw, (253, 222), "NO SPRAY", "bold", 8.5, (254, 202, 202), "mm")
    # Live RTD probe tag at Quench outlet
    rtd_col = RED if temp >= 210.0 else (ORANGE_LT if temp >= 195.0 else CYAN)
    draw_pill(draw, (224, 246, 282, 268), f"{temp:.1f}°F", bg=(10, 18, 34, 235), fg=rtd_col, border=(*rtd_col, 220), style="mono_bold", size=9.2)

    # Duct Quench -> CPVC Column
    draw.rectangle([s(288), s(218), s(324), s(240)], fill=(20, 38, 56, 240), outline=(*STEEL, 200), width=s(1.6))

    # Stage 3 CPVC/PP Column (x = 324..394, y = 166..278)
    cpvc_col = (14, 58, 50) if not byp_active else (22, 34, 52)
    draw_rrect(draw, (324, 166, 394, 278), radius=7, fill=(*cpvc_col, 240), outline=(*CPVC_WALL, 220), width=1.8)
    draw_text(draw, (359, 181), "CPVC / PP", "bold", 8.8, CPVC_WALL, "mm")
    draw_text(draw, (359, 194), "COLUMN", "bold", 8.8, CPVC_WALL, "mm")
    draw_pill(
        draw,
        (330, 212, 388, 234),
        "ISOLATED" if byp_active else "PROTECTED",
        bg=(16, 48, 38, 240),
        fg=MINT,
        border=(*EMERALD, 190),
        style="bold",
        size=7.6,
        radius=4,
    )
    draw_text(draw, (359, 256), "Max 210°F", "mono", 8.0, TEXT_MUTED, "mm")

    # Blower circle (x = 434, y = 228)
    draw.line([(s(394), s(228)), (s(414), s(228))], fill=(*CPVC_WALL, 210), width=s(2.4))
    bl_col = COBALT if vfd > 5 else STEEL_DK
    draw.ellipse([s(412), s(206), s(456), s(250)], fill=(20, 36, 62, 245), outline=(*bl_col, 230), width=s(2))
    spin = (frame_idx * 0.65) if vfd > 5 else 0.0
    for bi in range(6):
        ang = spin + bi * math.pi / 3.0
        draw.line(
            [(s(434 + 5 * math.cos(ang)), s(228 + 5 * math.sin(ang))), (s(434 + 16 * math.cos(ang)), s(228 + 16 * math.sin(ang)))],
            fill=(*bl_col, 235),
            width=s(2),
        )
    draw_text(draw, (434, 264), f"VFD {int(round(vfd))}%", "mono_bold", 8.8, COBALT if vfd > 5 else RED, "mm")

    # Animated particles along active path
    if byp_active:
        path = [(78, 229), (163, 229), (163, 128)]
        pcol = ORANGE_LT
    else:
        path = [(78, 229), (163, 229), (253, 229), (359, 229), (434, 229)]
        pcol = CYAN
    for pi in range(9):
        pu = (t_anim * 4.0 + pi / 9.0) % 1.0
        px, py = eval_polyline(path, pu)
        c = ORANGE_LT if (byp_active or pu < 0.42) else (ORANGE if gpm == 0 else pcol)
        draw.ellipse([s(px - 3), s(py - 3), s(px + 3), s(py + 3)], fill=(*c, 225))

    # Interactive Manual Reset Interlock Button (bottom of left panel, y = 314..348)
    btn_box = (276, 314, 466, 348)
    if reset_state == "clicking":
        draw_pill(draw, btn_box, "CLICKING: RESET INTERLOCK", bg=(16, 150, 105, 255), fg=WHITE, border=(*WHITE, 255), style="bold", size=9.2, radius=7)
        # Cursor pointer icon clicking the button
        cx, cy = 446, 340
        draw.polygon([(s(cx), s(cy)), (s(cx), s(cy + 14)), (s(cx + 4), s(cy + 10)), (s(cx + 10), s(cy + 12))], fill=(*WHITE, 255), outline=(*BG, 255))
    elif reset_state == "ready":
        draw_pill(draw, btn_box, "RESET INTERLOCK (READY)", bg=(180, 120, 16, 255), fg=WHITE, border=(*YELLOW, 255), style="bold", size=9.2, radius=7)
    elif reset_state in ("blocked", "cooling"):
        draw_pill(draw, btn_box, "RESET LOCKED (TEMP/FLOW)", bg=(45, 24, 30, 230), fg=(248, 113, 113), border=(*RED, 160), style="bold", size=8.8, radius=7)
    else:
        draw_pill(draw, btn_box, "INTERLOCK ARMED · OK", bg=(16, 52, 42, 220), fg=MINT, border=(*EMERALD, 150), style="bold", size=9.0, radius=7)

    draw_text(draw, (28, 326), "Trips: ≥210°F | <0.25 GPM | Power Loss", "mono", 8.3, TEXT_MUTED, "lm")
    draw_text(draw, (28, 342), "Latches until <190°F + manual operator reset", "regular", 8.3, TEXT_DIM, "lm")

    # --- Right Panel: Live Trend Chart (x = 494..944, y = 96..360) ---
    draw_rrect(draw, (494, 96, 944, 360), radius=9, fill=(*CARD_BG, 245), outline=(*BORDER, 255), width=1.2)
    draw_text(draw, (508, 111), "LIVE QUENCH TEMP (°F) & SPRAY FLOW (GPM)", "bold", 9.8, TEXT_MUTED, "lm")

    # Upper subplot: Temperature (y = 128..264, range 170..235°F)
    gx0, gy0, gx1, gy1 = 548, 128, 926, 264
    draw.rectangle([s(gx0), s(gy0), s(gx1), s(gy1)], fill=(12, 19, 36, 245), outline=(*BORDER, 200), width=s(1))

    def y_temp(val: float) -> float:
        return lerp(gy1, gy0, clamp((val - 170.0) / 65.0))

    def x_fr(idx: int) -> float:
        return lerp(gx0 + 4, gx1 - 4, idx / (total_frames - 1))

    # Danger zone red tint above 210°F
    y210 = y_temp(210.0)
    draw.rectangle([s(gx0 + 1), s(gy0 + 1), s(gx1 - 1), s(y210)], fill=(239, 68, 68, 30))

    # Horizontal grid & labeled limits
    for tv, lbl, col in [(230, "230°", TEXT_DIM), (210, "210°", RED), (188, "188°", EMERALD), (175, "175°", TEXT_DIM)]:
        yy = y_temp(tv)
        if tv in (210, 188):
            for dx in range(int(gx0), int(gx1), 10):
                draw.line([(s(dx), s(yy)), (s(min(dx + 6, gx1)), s(yy))], fill=(*col, 190), width=s(1.3))
        else:
            draw.line([(s(gx0), s(yy)), (s(gx1), s(yy))], fill=(*GRID_LINE, 220), width=s(1))
        draw_text(draw, (gx0 - 6, yy), lbl, "mono_bold" if tv in (210, 188) else "mono", 8.5, col, "rm")

    draw_text(draw, (gx1 - 6, y210 - 9), "210°F CPVC TRIP LIMIT", "bold", 8.2, RED, "rm")
    draw_text(draw, (gx1 - 6, y_temp(188.0) + 10), "188°F PID Setpoint (Tsat 183.9°F)", "regular", 8.0, MINT, "rm")

    # Phase vertical dividers on chart
    for div_i, lbl in [(12, "FAULT"), (26, "RESET")]:
        xx = x_fr(div_i)
        draw.line([(s(xx), s(gy0)), (s(xx), s(gy1))], fill=(*BORDER_HI, 120), width=s(1))
        draw_text(draw, (xx + 4, gy0 + 8), lbl, "mono_bold", 7.8, TEXT_DIM, "lm")

    # Draw temperature trace up to current frame_idx
    if frame_idx > 0:
        for i in range(frame_idx):
            p0 = (s(x_fr(i)), s(y_temp(INTERLOCK_PROFILE[i]["temp"])))
            p1 = (s(x_fr(i + 1)), s(y_temp(INTERLOCK_PROFILE[i + 1]["temp"])))
            t_hi = max(INTERLOCK_PROFILE[i]["temp"], INTERLOCK_PROFILE[i + 1]["temp"])
            seg_c = RED if t_hi >= 210.0 else (ORANGE_LT if t_hi >= 195.0 else EMERALD)
            draw.line([p0, p1], fill=(*seg_c, 255), width=s(2.4))
    # Highlight trip marker if reached (frame 14)
    if frame_idx >= 14:
        tx, ty = x_fr(14), y_temp(INTERLOCK_PROFILE[14]["temp"])
        draw.ellipse([s(tx - 5), s(ty - 5), s(tx + 5), s(ty + 5)], fill=(*RED, 255), outline=(*WHITE, 255), width=s(1.2))
    # Current cursor dot
    cx, cy = x_fr(frame_idx), y_temp(temp)
    cur_c = RED if temp >= 210.0 else (ORANGE_LT if temp >= 195.0 else EMERALD)
    draw.ellipse([s(cx - 4.5), s(cy - 4.5), s(cx + 4.5), s(cy + 4.5)], fill=(*cur_c, 255), outline=(*WHITE, 255), width=s(1.4))

    # Lower subplot: Water Flow GPM (y = 290..346, range 0..1.5 GPM) — label sits cleanly above box!
    draw_text(draw, (gx0, 279), "SPRAY WATER FLOW (GPM)", "bold", 8.2, TEXT_MUTED, "lm")
    draw_text(draw, (gx1, 279), "Low-Flow Trip < 0.25 GPM", "mono", 8.0, RED, "rm")
    wy0, wy1 = 290, 346
    draw.rectangle([s(gx0), s(wy0), s(gx1), s(wy1)], fill=(12, 19, 36, 245), outline=(*BORDER, 200), width=s(1))

    def y_gpm(val: float) -> float:
        return lerp(wy1 - 7, wy0 + 9, clamp(val / 1.35))

    y_min_flow = y_gpm(0.25)
    for dx in range(int(gx0), int(gx1), 10):
        draw.line([(s(dx), s(y_min_flow)), (s(min(dx + 5, gx1)), s(y_min_flow))], fill=(*RED, 160), width=s(1))
    draw_text(draw, (gx0 - 6, y_gpm(1.16)), "1.16", "mono", 8.2, CYAN, "rm")
    draw_text(draw, (gx0 - 6, y_gpm(0.0)), "0.00", "mono", 8.2, RED, "rm")

    if frame_idx > 0:
        for i in range(frame_idx):
            g0 = INTERLOCK_PROFILE[i]["gpm"]
            g1 = INTERLOCK_PROFILE[i + 1]["gpm"]
            p0 = (s(x_fr(i)), s(y_gpm(g0)))
            p1 = (s(x_fr(i + 1)), s(y_gpm(g1)))
            draw.line([p0, p1], fill=(*(RED if g1 < 0.25 else CYAN), 255), width=s(2.2))
    draw.ellipse(
        [s(x_fr(frame_idx) - 4), s(y_gpm(gpm) - 4), s(x_fr(frame_idx) + 4), s(y_gpm(gpm) + 4)],
        fill=(*(RED if gpm < 0.25 else CYAN), 255),
        outline=(*WHITE, 255),
        width=s(1.2),
    )

    # --- Bottom 5 Live Telemetry Readout Cards (y = 372..464) ---
    tiles = [
        (
            "QUENCH OUTLET",
            f"{temp:.1f}°F",
            "CPVC Limit: 210°F",
            RED if temp >= 210.0 else (ORANGE_LT if temp >= 195.0 else EMERALD),
        ),
        (
            "SPRAY WATER",
            f"{gpm:.2f} GPM",
            "Trip Floor: < 0.25",
            RED if gpm < 0.25 else CYAN,
        ),
        (
            "BYPASS DAMPER",
            "OPEN" if byp_active else "CLOSED",
            f"Spring-Return ({int(round(damper_deg))}°)",
            RED if byp_active else EMERALD,
        ),
        (
            "BLOWER VFD",
            f"{int(round(vfd))}%",
            "256 ACFM" if vfd > 70 else ("STOPPED" if vfd == 0 else "RAMPING"),
            RED if vfd == 0 else COBALT,
        ),
        (
            "KILN DRAFT",
            f"{draft:+.2f} inWC" if draft != 0 else "0.00 inWC",
            "Setpoint: -0.05 inWC",
            COBALT if abs(draft + 0.05) < 0.01 else YELLOW,
        ),
    ]
    tw = 178
    gap = 9
    for idx, (label, val, sub, col) in enumerate(tiles):
        tx0 = 16 + idx * (tw + gap)
        box = (tx0, 372, tx0 + tw, 464)
        draw_rrect(draw, box, radius=8, fill=(*CARD_BG, 245), outline=(*col, 155), width=1.4)
        draw.rectangle([s(tx0 + 1), s(373), s(tx0 + tw - 1), s(376)], fill=(*col, 230))
        draw_text(draw, (tx0 + tw / 2, 391), label, "bold", 9.0, TEXT_MUTED, "mm")
        draw_text(draw, (tx0 + tw / 2, 421), val, "mono_bold", 17.0, col, "mm")
        draw_text(draw, (tx0 + tw / 2, 448), sub, "regular", 8.8, TEXT_SEC, "mm")

    return finalize_frame(img, W, H)


# ===========================================================================
# GIF 3: Urban Studio Architecture Tradeoff Comparison (960x480, 36 frames)
# ===========================================================================
ARCHITECTURES = [
    {
        "short": "2-Stage Gas / Soda",
        "tag": "120V · ZERO HCl · LOW ΔP",
        "accent": CYAN,
        "subtitle": "Adiabatic Quench + Packed Bed · Backyard Gas / Soda (Na₂CO₃) Kilns",
        "stages": [
            ("1 · Quench", True, CYAN),
            ("2 · Venturi", False, VIOLET),
            ("3 · Packed Bed", True, EMERALD),
            ("4 · 120V Fan", True, COBALT),
            ("5 · Diluter/WESP", False, VIOLET),
        ],
        "bars": [
            ("PM Capture", 0.65, "~65% PM", "Coarse ash; no venturi", YELLOW),
            ("HCl Neutralization", 1.00, "0 ppm (N/A)", "Zero HCl (Gas / Soda)", EMERALD),
            ("Static Pressure ΔP", 0.19, "~4.0 inWC", "Low ΔP (no throat loss)", CYAN),
            ("Power & Voltage", 0.25, "0.5 HP · 120V", "Standard 120V 20A GFCI", EMERALD),
            ("Noise @ 40 ft", 0.36, "38 dBA", "Whisper-quiet (<60 lim)", EMERALD),
            ("Est. DIY / COTS", 0.35, "~$2,800", "No-weld CPVC + 316L", CYAN),
        ],
        "plume_ft": 31,
        "plume_txt": "31 ft bare vapor (0 ppm acid)",
        "best_for": [
            "Backyard gas, raku, or soda kilns",
            "Studios with only 120V 20A power",
            "Zero baking-soda reagent upkeep",
        ],
        "tradeoff": "Skips high-ΔP venturi: lowest cost & noise, but limited submicron (<1 µm) fume capture.",
        "badges": [("120V PLUG-IN", EMERALD), ("ZERO HCl", CYAN), ("LOWEST COST", MINT)],
    },
    {
        "short": "4-Stage Wood / Salt",
        "tag": "VERIFIED 24 CU.FT. BASELINE",
        "accent": EMERALD,
        "subtitle": "Quench + 70 m/s Venturi + Packed Bed + 2 HP ID Blower · Full Salt Control",
        "stages": [
            ("1 · Quench", True, CYAN),
            ("2 · Venturi", True, VIOLET),
            ("3 · Packed Bed", True, EMERALD),
            ("4 · 2 HP ID", True, COBALT),
            ("5 · Diluter/WESP", False, VIOLET),
        ],
        "bars": [
            ("PM Capture", 0.89, "84% NaCl / 97% Ash", "d50 = 0.21 µm @ 70 m/s", EMERALD),
            ("HCl Neutralization", 0.99, "99.04% HCl", "1048 → 10 ppm (NaHCO₃)", EMERALD),
            ("Static Pressure ΔP", 0.85, "21.21 inWC", "18.88\" venturi + 1.70\" bed", ORANGE_LT),
            ("Power & Voltage", 0.72, "2.0 HP · 240V", "240V 20A VFD (~10.6 A)", ORANGE_LT),
            ("Noise @ 40 ft", 0.56, "43.2 dBA", "Silenced (<60 dBA lim)", EMERALD),
            ("Est. DIY / COTS", 0.68, "~$5,550", "No-weld 316L + CPVC kit", YELLOW),
        ],
        "plume_ft": 41,
        "plume_txt": "41 ft bare winter steam reach",
        "best_for": [
            "Urban Cone 10–11 wood & salt kilns",
            "Submicron NaCl fume + 99% HCl",
            "Validated kilnscrub 24 cu.ft. kit",
        ],
        "tradeoff": "Venturi throat needs 18.88 inWC (240V 2 HP blower), delivering proven d50 = 0.21 µm capture.",
        "badges": [("240V 20A VFD", YELLOW), ("99.04% HCl", EMERALD), ("d50 0.21 µm", VIOLET)],
    },
    {
        "short": "5-Stage Plume / WESP",
        "tag": "12 FT PLUME · OR LOW-ΔP WESP",
        "accent": VIOLET,
        "subtitle": "+ 120V HV Squirrel-Cage Plume Diluter (or Wet ESP Submicron Alternative)",
        "stages": [
            ("1 · Quench", True, CYAN),
            ("2 · Vent/WESP", True, VIOLET),
            ("3 · Packed Bed", True, EMERALD),
            ("4 · ID Blower", True, COBALT),
            ("5 · HV Diluter", True, CYAN),
        ],
        "bars": [
            ("PM Capture", 0.97, "95–99% PM", "0.21 µm or >98% WESP", MINT),
            ("HCl Neutralization", 0.99, "99.04% HCl", "Full NaHCO₃ scrub pH 8", EMERALD),
            ("Static Pressure ΔP", 0.85, "21.2\" (3.5\" WESP)", "0\" added ΔP; WESP cuts ΔP", VIOLET),
            ("Power & Voltage", 0.86, "240V + 120V", "ID VFD + 1500 CFM HV fan", VIOLET),
            ("Noise @ 40 ft", 0.62, "45.5 dBA", "Plume 41→12 ft (<60 lim)", EMERALD),
            ("Est. DIY / COTS", 0.82, "~$6.3k ($9k WESP)", "+$750 HV diluter add-on", VIOLET),
        ],
        "plume_ft": 12,
        "plume_txt": "12 ft HV diluted plume (41→12 ft!)",
        "best_for": [
            "Tight urban yards (<40 ft fence)",
            "Winter steam visibility control",
            "WESP option cuts blower ΔP to ~4\"",
        ],
        "tradeoff": "+$750 120V squirrel-cage slashes winter plume 41→12 ft; WESP trades capex for low-ΔP PM.",
        "badges": [("12 FT PLUME", CYAN), ("45.5 dBA OK", EMERALD), ("WESP OPTION", VIOLET)],
    },
]


def render_tradeoff_frame(frame_idx: int, total_frames: int) -> Image.Image:
    W, H = 960, 480
    per_arch = total_frames // 3  # 12 frames per architecture
    arch_idx = frame_idx // per_arch
    local_f = frame_idx % per_arch
    prev_idx = (arch_idx - 1) % 3

    # Smooth 4-frame cubic transition into each architecture, then 8-frame readable hold
    trans_frames = 4
    blend = ease_in_out((local_f + 1) / (trans_frames + 1)) if local_f < trans_frames else 1.0
    hold_progress = (local_f + 1) / per_arch

    cur = ARCHITECTURES[arch_idx]
    prv = ARCHITECTURES[prev_idx]
    accent = lerp_color(prv["accent"], cur["accent"], blend)

    img, draw = new_canvas(W, H)

    # --- Header Bar (y = 0..46) ---
    draw.rectangle([0, 0, s(W), s(46)], fill=(*BG_HEADER, 255))
    draw.line([(0, s(46)), (s(W), s(46))], fill=(*BORDER, 255), width=s(1.2))
    draw_text(draw, (18, 15), "Urban Studio Scrubber Architecture Tradeoffs", "bold", 15, TEXT_PRI)
    draw_text(draw, (18, 33), "Comparing 2-Stage Gas/Soda · 4-Stage Wood/Salt Baseline · 5-Stage Plume/WESP", "regular", 9.5, TEXT_MUTED)

    # Cycle progress dots (top-right)
    for di in range(3):
        dx = 884 + di * 22
        is_cur = di == arch_idx
        r = 6 if is_cur else 4
        col = ARCHITECTURES[di]["accent"] if is_cur else BORDER_HI
        draw.ellipse([s(dx - r), s(23 - r), s(dx + r), s(23 + r)], fill=(*col, 255))

    # --- 3 Top Architecture Selector Cards (y = 54..102) ---
    card_w = 300
    card_gap = 14
    for idx, arch in enumerate(ARCHITECTURES):
        cx0 = 16 + idx * (card_w + card_gap)
        box = (cx0, 54, cx0 + card_w, 102)
        active = idx == arch_idx
        a_col = arch["accent"]
        bg = (*lerp_color(CARD_BG, a_col, 0.16), 255) if active else (*CARD_BG, 210)
        bdr = (*a_col, 240) if active else (*BORDER, 180)
        draw_rrect(draw, box, radius=8, fill=bg, outline=bdr, width=2.0 if active else 1.0)
        if active:
            draw.rectangle([s(cx0 + 1), s(55), s(cx0 + card_w - 1), s(58)], fill=(*a_col, 255))
            # Live dwell-progress bar along bottom of active architecture card
            prog_x1 = cx0 + 4 + (card_w - 8) * hold_progress
            draw.rectangle([s(cx0 + 4), s(98), s(prog_x1), s(100)], fill=(*a_col, 225))
        draw_text(draw, (cx0 + 12, 71), arch["short"], "bold", 11.5, TEXT_PRI if active else TEXT_MUTED, "lm")
        draw_text(draw, (cx0 + 12, 88), arch["tag"], "mono_bold", 8.6, a_col if active else TEXT_DIM, "lm")

    # --- Active Flow Train Strip (y = 112..164) ---
    draw_rrect(draw, (16, 112, 944, 164), radius=8, fill=(*CARD_BG, 245), outline=(*accent, 140), width=1.2)
    draw_text(draw, (28, 126), cur["subtitle"], "bold", 9.6, TEXT_PRI, "lm")
    sx = 28
    for si, (st_lbl, enabled, scol) in enumerate(cur["stages"]):
        sw = 138 if si < 4 else 148
        sbox = (sx, 136, sx + sw, 158)
        if enabled:
            draw_pill(draw, sbox, st_lbl, bg=(*lerp_color(BG, scol, 0.25), 255), fg=TEXT_PRI, border=(*scol, 220), style="bold", size=8.6, radius=5)
        else:
            draw_pill(draw, sbox, f"{st_lbl} (Skip)", bg=(18, 25, 42, 180), fg=TEXT_DIM, border=(*BORDER, 120), style="regular", size=8.2, radius=5)
        if si < 4:
            draw_arrow(draw, sx + sw + 3, 147, sx + sw + 20, 147, accent if enabled else BORDER, width=1.6, head=4.5)
        sx += sw + 24

    # --- Left Panel: 6 Live Animated Bar Gauges (x = 16..644, y = 174..438) ---
    draw_rrect(draw, (16, 174, 644, 438), radius=9, fill=(*CARD_BG, 245), outline=(*BORDER, 255), width=1.2)
    draw_text(draw, (30, 190), "PERFORMANCE, HYDRAULICS & STUDIO RESOURCE GAUGES", "bold", 9.6, TEXT_MUTED, "lm")

    row_y0 = 208
    row_h = 37
    bar_x0, bar_x1 = 174, 434
    bar_w = bar_x1 - bar_x0

    for ri in range(6):
        ry = row_y0 + ri * row_h
        lbl, cur_val, cur_txt, cur_sub, cur_col = cur["bars"][ri]
        _, prv_val, _, _, prv_col = prv["bars"][ri]
        val = lerp(prv_val, cur_val, blend)
        bcol = lerp_color(prv_col, cur_col, blend)

        # Metric label (left)
        draw_text(draw, (30, ry + 13), lbl, "bold", 10.0, TEXT_SEC, "lm")

        # Gauge track
        track_box = (bar_x0, ry + 4, bar_x1, ry + 22)
        draw_rrect(draw, track_box, radius=6, fill=(13, 21, 38, 255), outline=(*BORDER, 190), width=1.0)

        # Subtle ghost ticks showing the other two architectures' levels for immediate relative comparison
        for oi, oarch in enumerate(ARCHITECTURES):
            if oi != arch_idx:
                ox = bar_x0 + bar_w * oarch["bars"][ri][1]
                draw.line([(s(ox), s(ry + 5)), (s(ox), s(ry + 21))], fill=(*oarch["accent"], 95), width=s(1.5))

        # Animated filled bar
        fill_x1 = bar_x0 + max(12.0, bar_w * val)
        draw_rrect(draw, (bar_x0 + 1, ry + 5, fill_x1, ry + 21), radius=5, fill=(*bcol, 230))
        # Subtle top specular highlight on bar
        draw.line([(s(bar_x0 + 6), s(ry + 8)), (s(max(bar_x0 + 8, fill_x1 - 5)), s(ry + 8))], fill=(255, 255, 255, 75), width=s(1.4))

        # Exact value + concise engineering sub-label (right)
        draw_text(draw, (446, ry + 8), cur_txt, "mono_bold", 10.2, bcol, "lm")
        draw_text(draw, (446, ry + 22), cur_sub, "regular", 8.4, TEXT_MUTED, "lm")

    # --- Right Panel: Best Fit, Plume Bar & Verdict (x = 656..944, y = 174..438) ---
    draw_rrect(draw, (656, 174, 944, 438), radius=9, fill=(*CARD_BG, 245), outline=(*accent, 170), width=1.4)
    draw.rectangle([s(657), s(175), s(943), s(178)], fill=(*accent, 235))
    draw_text(draw, (670, 195), "BEST STUDIO FIT", "bold", 9.6, accent, "lm")

    for bi, bullet in enumerate(cur["best_for"]):
        by = 216 + bi * 21
        draw.ellipse([s(671), s(by - 3), s(677), s(by + 3)], fill=(*accent, 230))
        draw_text(draw, (685, by), bullet, "regular", 9.4, TEXT_PRI, "lm")

    # Animated Winter Steam Plume Meter inside Right Card (y = 286..336)
    draw_rrect(draw, (668, 282, 932, 336), radius=6, fill=(13, 21, 38, 245), outline=(*BORDER, 180), width=1)
    plume_val = lerp(prv["plume_ft"], cur["plume_ft"], blend)
    draw_text(draw, (678, 295), "WINTER STEAM PLUME REACH", "bold", 8.4, TEXT_MUTED, "lm")
    p_x0, p_x1 = 678, 922
    draw_rrect(draw, (p_x0, 306, p_x1, 318), radius=4, fill=(22, 33, 54, 255))
    p_col = MINT if plume_val <= 15 else (YELLOW if plume_val <= 34 else ORANGE_LT)
    p_w = max(14.0, (p_x1 - p_x0) * clamp(plume_val / 45.0))
    draw_rrect(draw, (p_x0, 306, p_x0 + p_w, 318), radius=4, fill=(*p_col, 235))
    # 40 ft property line marker
    fence_x = p_x0 + (p_x1 - p_x0) * (40.0 / 45.0)
    draw.line([(s(fence_x), s(304)), (s(fence_x), s(320))], fill=(*RED, 220), width=s(1.6))
    draw_text(draw, (fence_x, 295), "40ft fence", "mono", 7.6, RED, "mm")
    draw_text(draw, (678, 327), cur["plume_txt"], "mono_bold", 8.4, p_col, "lm")

    # Key Engineering Tradeoff Callout (wrapped cleanly)
    draw_text(draw, (670, 351), "KEY TRADEOFF", "bold", 8.8, TEXT_MUTED, "lm")
    words = cur["tradeoff"].split()
    lines: list[str] = []
    line = ""
    for w_word in words:
        cand = f"{line} {w_word}".strip()
        if len(cand) <= 36:
            line = cand
        else:
            lines.append(line)
            line = w_word
    if line:
        lines.append(line)
    for li, ltxt in enumerate(lines[:3]):
        draw_text(draw, (670, 367 + li * 14), ltxt, "regular", 9.0, TEXT_SEC, "lm")

    # 3 Compact Verdict Pills at bottom of right card (y = 408..428)
    px = 668
    pw = 84
    for btxt, bcol in cur["badges"]:
        draw_pill(draw, (px, 408, px + pw, 428), btxt, bg=(*lerp_color(BG, bcol, 0.25), 245), fg=bcol, border=(*bcol, 190), style="mono_bold", size=7.6, radius=5)
        px += pw + 6

    # --- Footer Bar (y = 448..472) ---
    draw_rrect(draw, (16, 448, 944, 472), radius=6, fill=(*BG_HEADER, 240), outline=(*BORDER, 180), width=1)
    draw_text(
        draw,
        (28, 460),
        "Verified kilnscrub 24 cu.ft. metrics · Ghost bar ticks show alternate architectures · Property line limit: 60 dBA & 40 ft setback",
        "regular",
        8.8,
        TEXT_MUTED,
        "lm",
    )

    return finalize_frame(img, W, H)


# ===========================================================================
# Main Entrypoint
# ===========================================================================
def generate_all(out_dir: Path, preview_dir: Path | None = None) -> None:
    out_dir.mkdir(parents=True, exist_ok=True)
    if preview_dir is not None:
        preview_dir.mkdir(parents=True, exist_ok=True)

    # 1. 4-Stage Process Flow (960x500, 32 frames, 85ms)
    n1 = 32
    f1 = [render_scrubber_flow_frame(i, n1) for i in range(n1)]
    p1 = out_dir / "scrubber-4stage-flow.gif"
    save_gif(f1, p1, durations_ms=85, colors=240)
    if preview_dir is not None:
        for idx in (0, 16):
            f1[idx].save(preview_dir / f"flow_frame_{idx:02d}.png")

    # 2. Fail-Safe Interlock Demo (960x480, 36 frames, variable pacing for clarity)
    n2 = 36
    f2 = [render_interlock_frame(i, n2) for i in range(n2)]
    d2 = []
    for i in range(n2):
        if i in (14, 15, 29, 30):
            d2.append(280)  # subtle emphasis pause on trip & manual reset click
        elif i == 11 or i == 25 or i == 35:
            d2.append(300)  # brief beat between story phases
        else:
            d2.append(135)
    p2 = out_dir / "failsafe-interlock-demo.gif"
    save_gif(f2, p2, durations_ms=d2, colors=240)
    if preview_dir is not None:
        for idx in (4, 16, 30):
            f2[idx].save(preview_dir / f"interlock_frame_{idx:02d}.png")

    # 3. Architecture Tradeoff Comparison (960x480, 36 frames, smooth ease + readable hold)
    n3 = 36
    f3 = [render_tradeoff_frame(i, n3) for i in range(n3)]
    d3 = []
    for i in range(n3):
        local = i % 12
        if local < 4:
            d3.append(80)   # snappy smooth ease-in-out animation
        elif local == 11:
            d3.append(450)  # comfortable hold before cycling to next architecture
        else:
            d3.append(170)
    p3 = out_dir / "stage-comparison-tradeoff.gif"
    save_gif(f3, p3, durations_ms=d3, colors=240)
    if preview_dir is not None:
        for idx in (6, 18, 30):
            f3[idx].save(preview_dir / f"tradeoff_frame_{idx:02d}.png")

    for path, n_fr in ((p1, n1), (p2, n2), (p3, n3)):
        size_kb = path.stat().st_size / 1024.0
        with Image.open(path) as im:
            w, h = im.size
        print(f"Generated {path}: {w}x{h}, {n_fr} frames, {size_kb:.1f} KB")


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate kilnscrub animated engineering GIFs.")
    parser.add_argument(
        "--out-dir",
        type=Path,
        default=Path(__file__).resolve().parent.parent / "docs" / "assets",
        help="Output directory for GIFs (default: docs/assets)",
    )
    parser.add_argument(
        "--preview-dir",
        type=Path,
        default=None,
        help="Optional scratch directory to dump key PNG frames for visual QA",
    )
    args = parser.parse_args()
    generate_all(args.out_dir, args.preview_dir)


if __name__ == "__main__":
    main()
