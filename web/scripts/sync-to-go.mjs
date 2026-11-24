import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const outDir = path.resolve(__dirname, "../out");
const targetDir = path.resolve(__dirname, "../../internal/web/static");

if (!fs.existsSync(outDir) || !fs.existsSync(path.join(outDir, "index.html"))) {
  console.error(`ERROR: Next.js static export not found at ${outDir}. Run 'npm run build' first.`);
  process.exit(1);
}

// Preserve clean replacement of internal/web/static without stale hashed chunks.
fs.rmSync(targetDir, { recursive: true, force: true });
fs.mkdirSync(targetDir, { recursive: true });

function copyFiltered(src, dst) {
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const srcPath = path.join(src, entry.name);
    const dstPath = path.join(dst, entry.name);
    if (entry.isDirectory()) {
      fs.mkdirSync(dstPath, { recursive: true });
      copyFiltered(srcPath, dstPath);
    } else {
      // Skip Next.js RSC .txt debug payloads for non-routed single-page bundle to keep Go binary compact.
      if (entry.name.endsWith(".txt") && path.dirname(srcPath) === outDir) {
        continue;
      }
      fs.copyFileSync(srcPath, dstPath);
    }
  }
}

copyFiltered(outDir, targetDir);

let fileCount = 0;
let totalBytes = 0;
function countStats(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      countStats(p);
    } else {
      fileCount++;
      totalBytes += fs.statSync(p).size;
    }
  }
}
countStats(targetDir);
console.log(
  `Synced Next.js static export -> internal/web/static (${fileCount} files, ${(totalBytes / 1024).toFixed(1)} KiB)`
);
