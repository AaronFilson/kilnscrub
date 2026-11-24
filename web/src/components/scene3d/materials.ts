import * as THREE from "three";

export interface MaterialBundle {
  // Industrial Metals & Polymers
  ss: THREE.MeshStandardMaterial;
  ssSolid: THREE.MeshStandardMaterial;
  flange: THREE.MeshStandardMaterial;
  bolt: THREE.MeshStandardMaterial;
  gasket: THREE.MeshStandardMaterial;
  hot: THREE.MeshStandardMaterial;
  hotSolid: THREE.MeshStandardMaterial;
  cpvc: THREE.MeshStandardMaterial;
  cpvcSolid: THREE.MeshStandardMaterial;
  cpvcFlange: THREE.MeshStandardMaterial;
  cutGlass: THREE.MeshStandardMaterial;
  packingPP: THREE.MeshStandardMaterial;
  packingCeramic: THREE.MeshStandardMaterial;
  packingBio: THREE.MeshStandardMaterial;
  demister: THREE.MeshStandardMaterial;
  sumpWater: THREE.MeshStandardMaterial;
  blower: THREE.MeshStandardMaterial;
  blowerFlange: THREE.MeshStandardMaterial;
  blowerDark: THREE.MeshStandardMaterial;
  blade: THREE.MeshStandardMaterial;
  shroud: THREE.MeshStandardMaterial;
  darkSteel: THREE.MeshStandardMaterial;
  motor: THREE.MeshStandardMaterial;
  motorFin: THREE.MeshStandardMaterial;
  sprayMist: THREE.MeshStandardMaterial;
  diluter: THREE.MeshStandardMaterial;

  // Phase 9 Systems / Deluxe
  skidSteel: THREE.MeshStandardMaterial;
  braceSteel: THREE.MeshStandardMaterial;
  hoseSupply: THREE.MeshStandardMaterial;
  hoseReturn: THREE.MeshStandardMaterial;
  conduit: THREE.MeshStandardMaterial;
  panelBox: THREE.MeshStandardMaterial;
  safetyRed: THREE.MeshStandardMaterial;
  valveGreen: THREE.MeshStandardMaterial;
  acousticBox: THREE.MeshStandardMaterial;
  acousticTrim: THREE.MeshStandardMaterial;
  silencerBody: THREE.MeshStandardMaterial;
  necZone: THREE.MeshStandardMaterial;
  sightGlass: THREE.MeshStandardMaterial;
  rubberSeal: THREE.MeshStandardMaterial;
  collectorPlate: THREE.MeshStandardMaterial;

  // Phase 10 / 10B / 10C Local Ceramic & Steam-Jet
  brick: THREE.MeshStandardMaterial;
  brickSolid: THREE.MeshStandardMaterial;
  mortar: THREE.MeshStandardMaterial;
  terracotta: THREE.MeshStandardMaterial;
  stoneware: THREE.MeshStandardMaterial;
  hotCore: THREE.MeshStandardMaterial;
  kentank: THREE.MeshStandardMaterial;
  kentankRib: THREE.MeshStandardMaterial;
  timber: THREE.MeshStandardMaterial;
  pipeCyan: THREE.MeshStandardMaterial;
  brass: THREE.MeshStandardMaterial;
  iron: THREE.MeshStandardMaterial;
  plinth: THREE.MeshStandardMaterial;
  boilerSteel: THREE.MeshStandardMaterial;
  steamerSS: THREE.MeshStandardMaterial;
  steelTrim: THREE.MeshStandardMaterial;
  dialWhite: THREE.MeshStandardMaterial;
  standpipePVC: THREE.MeshStandardMaterial;
  steamHeader: THREE.MeshStandardMaterial;
  steamPlume: THREE.MeshStandardMaterial;

  textures: THREE.Texture[];
  dispose: () => void;
}

function makeCanvas(w: number, h: number): [HTMLCanvasElement, CanvasRenderingContext2D] | null {
  if (typeof document === "undefined") return null;
  const c = document.createElement("canvas");
  c.width = w;
  c.height = h;
  const ctx = c.getContext("2d");
  if (!ctx) return null;
  return [c, ctx];
}

function createBrickTextures(): { map?: THREE.CanvasTexture; bump?: THREE.CanvasTexture } {
  const pair = makeCanvas(256, 256);
  const bumpPair = makeCanvas(256, 256);
  if (!pair || !bumpPair) return {};
  const [c, ctx] = pair;
  const [bc, bctx] = bumpPair;

  // Mortar background
  ctx.fillStyle = "#7c6e62";
  ctx.fillRect(0, 0, 256, 256);
  bctx.fillStyle = "#3a3a3a";
  bctx.fillRect(0, 0, 256, 256);

  const rows = 8;
  const cols = 4;
  const rowH = 256 / rows;
  const colW = 256 / cols;
  const gap = 3;

  const palette = ["#b4533c", "#a54732", "#bc5c43", "#9b402c", "#b8583e"];
  let seed = 17;
  const rand = () => {
    seed = (seed * 16807) % 2147483647;
    return (seed - 1) / 2147483646;
  };

  for (let r = 0; r < rows; r++) {
    const offset = (r % 2) * (colW * 0.5);
    for (let col = -1; col <= cols; col++) {
      const x = col * colW + offset + gap * 0.5;
      const y = r * rowH + gap * 0.5;
      const w = colW - gap;
      const h = rowH - gap;
      const baseColor = palette[Math.floor(rand() * palette.length)];
      ctx.fillStyle = baseColor;
      ctx.fillRect(x, y, w, h);

      // Subtle firebrick surface variation
      ctx.fillStyle = `rgba(255, 220, 180, ${(rand() * 0.08).toFixed(3)})`;
      ctx.fillRect(x + 2, y + 2, w - 4, h * 0.45);
      ctx.fillStyle = `rgba(20, 10, 5, ${(rand() * 0.14).toFixed(3)})`;
      ctx.fillRect(x, y + h - 4, w, 4);

      const bumpVal = Math.floor(185 + rand() * 55);
      bctx.fillStyle = `rgb(${bumpVal},${bumpVal},${bumpVal})`;
      bctx.fillRect(x, y, w, h);
    }
  }

  const map = new THREE.CanvasTexture(c);
  map.wrapS = map.wrapT = THREE.RepeatWrapping;
  map.repeat.set(3, 3);
  map.colorSpace = THREE.SRGBColorSpace;

  const bump = new THREE.CanvasTexture(bc);
  bump.wrapS = bump.wrapT = THREE.RepeatWrapping;
  bump.repeat.set(3, 3);

  return { map, bump };
}

function createBrushedMetalBump(): THREE.CanvasTexture | undefined {
  const pair = makeCanvas(256, 128);
  if (!pair) return undefined;
  const [c, ctx] = pair;
  ctx.fillStyle = "#808080";
  ctx.fillRect(0, 0, 256, 128);

  let seed = 43;
  const rand = () => {
    seed = (seed * 48271) % 2147483647;
    return (seed - 1) / 2147483646;
  };
  for (let y = 0; y < 128; y++) {
    const v = Math.floor(110 + rand() * 38);
    ctx.fillStyle = `rgb(${v},${v},${v})`;
    ctx.fillRect(0, y, 256, 1);
  }
  for (let i = 0; i < 90; i++) {
    const v = Math.floor(95 + rand() * 70);
    ctx.fillStyle = `rgba(${v},${v},${v},0.35)`;
    ctx.fillRect(rand() * 256, rand() * 128, 30 + rand() * 90, 1);
  }

  const tex = new THREE.CanvasTexture(c);
  tex.wrapS = tex.wrapT = THREE.RepeatWrapping;
  tex.repeat.set(2, 4);
  return tex;
}

function createPackingTexture(
  bgHex: string,
  ringHex: string,
  shadowHex: string
): { map?: THREE.CanvasTexture; bump?: THREE.CanvasTexture } {
  const pair = makeCanvas(256, 256);
  const bumpPair = makeCanvas(256, 256);
  if (!pair || !bumpPair) return {};
  const [c, ctx] = pair;
  const [bc, bctx] = bumpPair;

  ctx.fillStyle = bgHex;
  ctx.fillRect(0, 0, 256, 256);
  bctx.fillStyle = "#404040";
  bctx.fillRect(0, 0, 256, 256);

  let seed = 91;
  const rand = () => {
    seed = (seed * 48271) % 2147483647;
    return (seed - 1) / 2147483646;
  };

  for (let i = 0; i < 180; i++) {
    const x = rand() * 256;
    const y = rand() * 256;
    const rx = 6 + rand() * 6;
    const ry = 4 + rand() * 5;
    const rot = rand() * Math.PI;

    ctx.save();
    ctx.translate(x, y);
    ctx.rotate(rot);
    ctx.fillStyle = shadowHex;
    ctx.beginPath();
    ctx.ellipse(1.5, 1.5, rx, ry, 0, 0, Math.PI * 2);
    ctx.fill();

    ctx.strokeStyle = ringHex;
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.ellipse(0, 0, rx, ry, 0, 0, Math.PI * 2);
    ctx.stroke();
    ctx.restore();

    bctx.save();
    bctx.translate(x, y);
    bctx.rotate(rot);
    bctx.strokeStyle = "#e0e0e0";
    bctx.lineWidth = 3.2;
    bctx.beginPath();
    bctx.ellipse(0, 0, rx, ry, 0, 0, Math.PI * 2);
    bctx.stroke();
    bctx.restore();
  }

  const map = new THREE.CanvasTexture(c);
  map.wrapS = map.wrapT = THREE.RepeatWrapping;
  map.repeat.set(2, 2);
  map.colorSpace = THREE.SRGBColorSpace;

  const bump = new THREE.CanvasTexture(bc);
  bump.wrapS = bump.wrapT = THREE.RepeatWrapping;
  bump.repeat.set(2, 2);
  return { map, bump };
}

export function createParticleSpriteTexture(): THREE.CanvasTexture | null {
  const pair = makeCanvas(64, 64);
  if (!pair) return null;
  const [c, ctx] = pair;
  const g = ctx.createRadialGradient(32, 32, 2, 32, 32, 30);
  g.addColorStop(0, "rgba(255,255,255,1)");
  g.addColorStop(0.35, "rgba(255,255,255,0.88)");
  g.addColorStop(0.7, "rgba(255,255,255,0.28)");
  g.addColorStop(1, "rgba(255,255,255,0)");
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, 64, 64);

  const tex = new THREE.CanvasTexture(c);
  tex.colorSpace = THREE.SRGBColorSpace;
  return tex;
}

export function createFloorShadowTexture(): THREE.CanvasTexture | null {
  const pair = makeCanvas(256, 256);
  if (!pair) return null;
  const [c, ctx] = pair;
  const g = ctx.createRadialGradient(128, 128, 12, 128, 128, 124);
  g.addColorStop(0, "rgba(2, 6, 14, 0.72)");
  g.addColorStop(0.55, "rgba(2, 6, 14, 0.36)");
  g.addColorStop(1, "rgba(2, 6, 14, 0.0)");
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, 256, 256);
  return new THREE.CanvasTexture(c);
}

/**
 * Generates a lightweight procedural studio environment map using PMREMGenerator
 * so brushed stainless steel, brass, copper, and CPVC have realistic specular reflections
 * with zero external HDR assets.
 */
export function createStudioEnvironment(
  renderer: THREE.WebGLRenderer
): THREE.Texture | null {
  try {
    const pmrem = new THREE.PMREMGenerator(renderer);
    const envScene = new THREE.Scene();
    envScene.background = new THREE.Color(0x0c1726);

    const panels: Array<{
      pos: [number, number, number];
      scale: [number, number, number];
      color: number;
    }> = [
      { pos: [0, 8, 0], scale: [14, 0.2, 14], color: 0xe2f2ff },
      { pos: [8, 4, 6], scale: [0.2, 8, 10], color: 0xffffff },
      { pos: [-8, 3.5, 4], scale: [0.2, 7, 9], color: 0x7dd3fc },
      { pos: [0, 3, -9], scale: [12, 6, 0.2], color: 0x93c5fd },
      { pos: [0, -5, 0], scale: [16, 0.2, 16], color: 0x1e293b },
    ];
    for (const p of panels) {
      const mesh = new THREE.Mesh(
        new THREE.BoxGeometry(1, 1, 1),
        new THREE.MeshBasicMaterial({ color: p.color })
      );
      mesh.position.set(...p.pos);
      mesh.scale.set(...p.scale);
      envScene.add(mesh);
    }

    const rt = pmrem.fromScene(envScene, 0.04);
    pmrem.dispose();
    envScene.traverse((obj) => {
      if (obj instanceof THREE.Mesh) {
        obj.geometry.dispose();
        (obj.material as THREE.Material).dispose();
      }
    });
    return rt.texture;
  } catch {
    return null;
  }
}

export function createMaterialBundle(
  cutaway: boolean,
  isHeavyAlloy: boolean,
  isSubsonicSteam: boolean
): MaterialBundle {
  const side = cutaway ? THREE.DoubleSide : THREE.FrontSide;
  const textures: THREE.Texture[] = [];

  const brickTex = createBrickTextures();
  if (brickTex.map) textures.push(brickTex.map);
  if (brickTex.bump) textures.push(brickTex.bump);

  const metalBump = createBrushedMetalBump();
  if (metalBump) textures.push(metalBump);

  const ppTex = createPackingTexture("#0c5e54", "#2dd4bf", "#073630");
  if (ppTex.map) textures.push(ppTex.map);
  if (ppTex.bump) textures.push(ppTex.bump);

  const cerTex = createPackingTexture("#7c4a2b", "#dfa06e", "#4e2a14");
  if (cerTex.map) textures.push(cerTex.map);
  if (cerTex.bump) textures.push(cerTex.bump);

  const bioTex = createPackingTexture("#132742", "#3b82f6", "#0b1729");
  if (bioTex.map) textures.push(bioTex.map);
  if (bioTex.bump) textures.push(bioTex.bump);

  const alloyTint = isHeavyAlloy ? 0xaab9c8 : 0xc6d3e1;
  const cpvcTint = isHeavyAlloy ? 0x1e7874 : 0x2d83a8;

  const bundle: Omit<MaterialBundle, "textures" | "dispose"> = {
    ss: new THREE.MeshStandardMaterial({
      color: alloyTint,
      metalness: 0.78,
      roughness: 0.26,
      bumpMap: metalBump,
      bumpScale: 0.004,
      side,
    }),
    ssSolid: new THREE.MeshStandardMaterial({
      color: alloyTint,
      metalness: 0.78,
      roughness: 0.26,
      bumpMap: metalBump,
      bumpScale: 0.004,
      side: THREE.DoubleSide,
    }),
    flange: new THREE.MeshStandardMaterial({
      color: 0xa8b8ca,
      metalness: 0.80,
      roughness: 0.28,
      side: THREE.DoubleSide,
    }),
    bolt: new THREE.MeshStandardMaterial({
      color: 0xe2e8f0,
      metalness: 0.88,
      roughness: 0.20,
    }),
    gasket: new THREE.MeshStandardMaterial({
      color: 0x181c24,
      metalness: 0.05,
      roughness: 0.88,
    }),
    hot: new THREE.MeshStandardMaterial({
      color: 0xf97316,
      metalness: 0.45,
      roughness: 0.36,
      emissive: 0x451506,
      side,
    }),
    hotSolid: new THREE.MeshStandardMaterial({
      color: 0xf97316,
      metalness: 0.45,
      roughness: 0.36,
      emissive: 0x451506,
      side: THREE.DoubleSide,
    }),
    cpvc: new THREE.MeshStandardMaterial({
      color: cpvcTint,
      metalness: 0.16,
      roughness: 0.36,
      side,
    }),
    cpvcSolid: new THREE.MeshStandardMaterial({
      color: cpvcTint,
      metalness: 0.18,
      roughness: 0.34,
      side: THREE.DoubleSide,
    }),
    cpvcFlange: new THREE.MeshStandardMaterial({
      color: isHeavyAlloy ? 0x2ca29c : 0x449ec4,
      metalness: 0.22,
      roughness: 0.32,
      side: THREE.DoubleSide,
    }),
    cutGlass: new THREE.MeshStandardMaterial({
      color: 0x67e8f9,
      metalness: 0.15,
      roughness: 0.14,
      transparent: true,
      opacity: 0.26,
      depthWrite: false,
      side: THREE.DoubleSide,
    }),
    packingPP: new THREE.MeshStandardMaterial({
      color: 0xffffff,
      map: ppTex.map,
      bumpMap: ppTex.bump,
      bumpScale: 0.022,
      metalness: 0.08,
      roughness: 0.72,
    }),
    packingCeramic: new THREE.MeshStandardMaterial({
      color: 0xffffff,
      map: cerTex.map,
      bumpMap: cerTex.bump,
      bumpScale: 0.028,
      metalness: 0.05,
      roughness: 0.85,
    }),
    packingBio: new THREE.MeshStandardMaterial({
      color: 0xffffff,
      map: bioTex.map,
      bumpMap: bioTex.bump,
      bumpScale: 0.024,
      metalness: 0.12,
      roughness: 0.65,
    }),
    demister: new THREE.MeshStandardMaterial({
      color: 0x93c5fd,
      metalness: 0.45,
      roughness: 0.32,
      transparent: true,
      opacity: 0.78,
    }),
    sumpWater: new THREE.MeshStandardMaterial({
      color: 0x0ea5e9,
      metalness: 0.12,
      roughness: 0.14,
      transparent: true,
      opacity: 0.60,
    }),
    blower: new THREE.MeshStandardMaterial({
      color: 0x10b981,
      metalness: 0.42,
      roughness: 0.32,
      side: THREE.DoubleSide,
    }),
    blowerFlange: new THREE.MeshStandardMaterial({
      color: 0x34d399,
      metalness: 0.50,
      roughness: 0.28,
      side: THREE.DoubleSide,
    }),
    blowerDark: new THREE.MeshStandardMaterial({
      color: 0x36495f,
      metalness: 0.38,
      roughness: 0.42,
    }),
    blade: new THREE.MeshStandardMaterial({
      color: 0xe6f0fa,
      metalness: 0.85,
      roughness: 0.18,
    }),
    shroud: new THREE.MeshStandardMaterial({
      color: 0x94a3b8,
      metalness: 0.72,
      roughness: 0.26,
    }),
    darkSteel: new THREE.MeshStandardMaterial({
      color: 0x425166,
      metalness: 0.48,
      roughness: 0.42,
    }),
    motor: new THREE.MeshStandardMaterial({
      color: 0x4b6078,
      metalness: 0.42,
      roughness: 0.38,
    }),
    motorFin: new THREE.MeshStandardMaterial({
      color: 0x64748b,
      metalness: 0.48,
      roughness: 0.34,
    }),
    sprayMist: new THREE.MeshStandardMaterial({
      color: 0x38bdf8,
      transparent: true,
      opacity: 0.22,
      side: THREE.DoubleSide,
      depthWrite: false,
    }),
    diluter: new THREE.MeshStandardMaterial({
      color: 0x38bdf8,
      metalness: 0.45,
      roughness: 0.32,
      side: THREE.DoubleSide,
    }),

    // Phase 9 Systems / Deluxe
    skidSteel: new THREE.MeshStandardMaterial({
      color: 0x2f3e52,
      metalness: 0.64,
      roughness: 0.36,
    }),
    braceSteel: new THREE.MeshStandardMaterial({
      color: 0x6366f1,
      metalness: 0.48,
      roughness: 0.38,
    }),
    hoseSupply: new THREE.MeshStandardMaterial({
      color: 0x06b6d4,
      metalness: 0.22,
      roughness: 0.38,
    }),
    hoseReturn: new THREE.MeshStandardMaterial({
      color: 0x38bdf8,
      metalness: 0.20,
      roughness: 0.44,
    }),
    conduit: new THREE.MeshStandardMaterial({
      color: 0xf59e0b,
      metalness: 0.52,
      roughness: 0.34,
    }),
    panelBox: new THREE.MeshStandardMaterial({
      color: 0xcbd5e1,
      metalness: 0.50,
      roughness: 0.28,
    }),
    safetyRed: new THREE.MeshStandardMaterial({
      color: 0xef4444,
      metalness: 0.22,
      roughness: 0.32,
      emissive: 0x450a0a,
    }),
    valveGreen: new THREE.MeshStandardMaterial({
      color: 0x10b981,
      metalness: 0.22,
      roughness: 0.34,
    }),
    acousticBox: new THREE.MeshStandardMaterial({
      color: 0xa855f7,
      metalness: 0.14,
      roughness: 0.52,
      transparent: true,
      opacity: 0.20,
      depthWrite: false,
      side: THREE.DoubleSide,
    }),
    acousticTrim: new THREE.MeshStandardMaterial({
      color: 0x9333ea,
      metalness: 0.40,
      roughness: 0.40,
    }),
    silencerBody: new THREE.MeshStandardMaterial({
      color: 0x7e22ce,
      metalness: 0.44,
      roughness: 0.36,
      side: THREE.DoubleSide,
    }),
    necZone: new THREE.MeshStandardMaterial({
      color: 0x22c55e,
      metalness: 0.05,
      roughness: 0.7,
      transparent: true,
      opacity: 0.18,
      depthWrite: false,
      side: THREE.DoubleSide,
    }),
    sightGlass: new THREE.MeshStandardMaterial({
      color: 0xbae6fd,
      metalness: 0.28,
      roughness: 0.08,
      transparent: true,
      opacity: 0.74,
    }),
    rubberSeal: new THREE.MeshStandardMaterial({
      color: 0x161b22,
      metalness: 0.06,
      roughness: 0.86,
    }),
    collectorPlate: new THREE.MeshStandardMaterial({
      color: 0x334e68,
      metalness: 0.28,
      roughness: 0.40,
      side: THREE.DoubleSide,
    }),

    // Phase 10 / 10B / 10C Local Ceramic & Steam-Jet
    brick: new THREE.MeshStandardMaterial({
      color: 0xffffff,
      map: brickTex.map,
      bumpMap: brickTex.bump,
      bumpScale: 0.018,
      metalness: 0.05,
      roughness: 0.88,
      side,
    }),
    brickSolid: new THREE.MeshStandardMaterial({
      color: 0xf4e8e0,
      map: brickTex.map,
      bumpMap: brickTex.bump,
      bumpScale: 0.018,
      metalness: 0.05,
      roughness: 0.88,
      side: THREE.DoubleSide,
    }),
    mortar: new THREE.MeshStandardMaterial({
      color: 0x8d7d6e,
      metalness: 0.04,
      roughness: 0.94,
    }),
    terracotta: new THREE.MeshStandardMaterial({
      color: 0xc2633a,
      metalness: 0.08,
      roughness: 0.62,
      side: THREE.DoubleSide,
    }),
    stoneware: new THREE.MeshStandardMaterial({
      color: 0xd49d6a,
      metalness: 0.06,
      roughness: 0.78,
    }),
    hotCore: new THREE.MeshStandardMaterial({
      color: 0xf97316,
      metalness: 0.25,
      roughness: 0.40,
      emissive: 0x7c2d12,
      emissiveIntensity: 0.9,
      side: THREE.DoubleSide,
    }),
    kentank: new THREE.MeshStandardMaterial({
      color: 0x1d6fb8,
      metalness: 0.14,
      roughness: 0.32,
    }),
    kentankRib: new THREE.MeshStandardMaterial({
      color: 0x155794,
      metalness: 0.16,
      roughness: 0.30,
    }),
    timber: new THREE.MeshStandardMaterial({
      color: 0x855834,
      metalness: 0.04,
      roughness: 0.88,
    }),
    pipeCyan: new THREE.MeshStandardMaterial({
      color: 0x06b6d4,
      metalness: 0.24,
      roughness: 0.36,
    }),
    brass: new THREE.MeshStandardMaterial({
      color: 0xd9a83c,
      metalness: 0.82,
      roughness: 0.24,
    }),
    iron: new THREE.MeshStandardMaterial({
      color: 0x374151,
      metalness: 0.60,
      roughness: 0.44,
    }),
    plinth: new THREE.MeshStandardMaterial({
      color: 0x4b5563,
      metalness: 0.08,
      roughness: 0.90,
    }),
    boilerSteel: new THREE.MeshStandardMaterial({
      color: 0x24303e,
      metalness: 0.62,
      roughness: 0.46,
      bumpMap: metalBump,
      bumpScale: 0.003,
    }),
    steamerSS: new THREE.MeshStandardMaterial({
      color: 0xdbe5f0,
      metalness: 0.86,
      roughness: 0.20,
      bumpMap: metalBump,
      bumpScale: 0.003,
    }),
    steelTrim: new THREE.MeshStandardMaterial({
      color: 0x94a3b8,
      metalness: 0.76,
      roughness: 0.28,
    }),
    dialWhite: new THREE.MeshStandardMaterial({
      color: 0xf8fafc,
      metalness: 0.08,
      roughness: 0.36,
    }),
    standpipePVC: new THREE.MeshStandardMaterial({
      color: 0xe2e8f0,
      metalness: 0.12,
      roughness: 0.20,
      transparent: true,
      opacity: 0.42,
      depthWrite: false,
      side: THREE.DoubleSide,
    }),
    steamHeader: new THREE.MeshStandardMaterial({
      color: 0xc8844c,
      metalness: 0.78,
      roughness: 0.28,
    }),
    steamPlume: new THREE.MeshStandardMaterial({
      color: 0xe0f2fe,
      emissive: 0x38bdf8,
      emissiveIntensity: isSubsonicSteam ? 0.52 : 0.88,
      transparent: true,
      opacity: isSubsonicSteam ? 0.38 : 0.50,
      side: THREE.DoubleSide,
      depthWrite: false,
    }),
  };

  return {
    ...bundle,
    textures,
    dispose: () => {
      Object.values(bundle).forEach((m) => m.dispose());
      textures.forEach((t) => t.dispose());
    },
  };
}
