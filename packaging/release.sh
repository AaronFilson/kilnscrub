#!/usr/bin/env bash
# Reproducible cross-platform release archive builder for kilnscrub.
# Produces platform archives (.tar.gz for Linux/macOS, .zip for Windows)
# and SHA256SUMS.txt matching .github/workflows/release.yml.
#
# Usage:
#   ./packaging/release.sh [VERSION] [OUT_DIR]
#
# Examples:
#   ./packaging/release.sh v1.0.0 dist/release
#   ./packaging/release.sh

set -euo pipefail

export PATH="${PATH}:/usr/local/go/bin"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

VERSION="${1:-${VERSION:-$(git -C "${REPO_ROOT}" describe --tags --exact-match 2>/dev/null || git -C "${REPO_ROOT}" rev-parse --short HEAD 2>/dev/null || echo dev)}}"
OUT_DIR="${2:-${REPO_ROOT}/dist/release}"
GO_BIN="${GO:-go}"
PKG="./cmd/kilnscrub"
LDFLAGS="-s -w -X main.version=${VERSION}"

# Normalize OUT_DIR to an absolute path before changing directories.
mkdir -p "${OUT_DIR}"
OUT_DIR="$(cd "${OUT_DIR}" && pwd)"

STAGE_ROOT="$(mktemp -d)"
trap 'rm -rf "${STAGE_ROOT}"' EXIT

TARGETS=(
  "linux amd64 tar.gz"
  "linux arm64 tar.gz"
  "darwin amd64 tar.gz"
  "darwin arm64 tar.gz"
  "windows amd64 zip"
)

# Verify required bundled files exist before building.
for req in \
  "README.md" \
  "LICENSE" \
  "configs/kilnscrub.example.json" \
  "packaging/systemd/kilnscrub-daemon.service"; do
  if [ ! -f "${REPO_ROOT}/${req}" ]; then
    echo "ERROR: required release file missing: ${req}" >&2
    exit 1
  fi
done

# Deterministic mtime for reproducible archives (git commit timestamp if available).
SOURCE_DATE_EPOCH="${SOURCE_DATE_EPOCH:-$(git -C "${REPO_ROOT}" log -1 --format=%ct 2>/dev/null || echo 0)}"
export SOURCE_DATE_EPOCH

create_tar_gz() {
  local stage_parent="$1"
  local dir_name="$2"
  local archive_path="$3"

  if tar --version 2>/dev/null | grep -qi "GNU tar"; then
    tar \
      --sort=name \
      --mtime="@${SOURCE_DATE_EPOCH}" \
      --owner=0 \
      --group=0 \
      --numeric-owner \
      -C "${stage_parent}" \
      -cf - \
      "${dir_name}" | gzip -n > "${archive_path}"
  else
    tar -C "${stage_parent}" -cf - "${dir_name}" | gzip -n > "${archive_path}"
  fi
}

create_zip() {
  local stage_parent="$1"
  local dir_name="$2"
  local archive_path="$3"

  rm -f "${archive_path}"
  if command -v python3 >/dev/null 2>&1; then
    python3 - "${stage_parent}" "${dir_name}" "${archive_path}" "${SOURCE_DATE_EPOCH}" <<'PY'
import datetime
import os
import sys
import zipfile

stage_parent, dir_name, archive_path, epoch_str = sys.argv[1:5]
epoch = int(epoch_str)
dt = datetime.datetime.fromtimestamp(max(epoch, 315532800), tz=datetime.timezone.utc)
zip_dt = (dt.year, dt.month, dt.day, dt.hour, dt.minute, dt.second)

root_path = os.path.join(stage_parent, dir_name)
entries = []
for current_root, dirs, files in os.walk(root_path):
    dirs.sort()
    files.sort()
    for fname in files:
        full_path = os.path.join(current_root, fname)
        rel_path = os.path.relpath(full_path, stage_parent)
        entries.append((full_path, rel_path))

with zipfile.ZipFile(archive_path, "w", compression=zipfile.ZIP_DEFLATED) as zf:
    for full_path, rel_path in entries:
        zi = zipfile.ZipInfo(filename=rel_path.replace(os.sep, "/"), date_time=zip_dt)
        mode = os.stat(full_path).st_mode & 0o777
        zi.external_attr = (0o100000 | mode) << 16
        with open(full_path, "rb") as src:
            zf.writestr(zi, src.read(), compress_type=zipfile.ZIP_DEFLATED)
PY
  elif command -v zip >/dev/null 2>&1; then
    (cd "${stage_parent}" && zip -r -q -X "${archive_path}" "${dir_name}")
  else
    echo "ERROR: neither 'python3' nor 'zip' is available to create ${archive_path}" >&2
    exit 1
  fi
}

compute_sha256sums() {
  local dir="$1"
  shift
  (
    cd "${dir}"
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum "$@" > SHA256SUMS.txt
      sha256sum -c SHA256SUMS.txt
    elif command -v shasum >/dev/null 2>&1; then
      shasum -a 256 "$@" > SHA256SUMS.txt
      shasum -a 256 -c SHA256SUMS.txt
    else
      echo "ERROR: neither 'sha256sum' nor 'shasum' is available" >&2
      exit 1
    fi
  )
}

cd "${REPO_ROOT}"

ARCHIVES=()

for target in "${TARGETS[@]}"; do
  read -r goos goarch fmt <<<"${target}"
  ext=""
  if [ "${goos}" = "windows" ]; then
    ext=".exe"
  fi

  bundle_name="kilnscrub-${VERSION}-${goos}-${goarch}"
  bundle_dir="${STAGE_ROOT}/${bundle_name}"
  mkdir -p "${bundle_dir}/configs" "${bundle_dir}/packaging/systemd"

  echo "Building ${goos}/${goarch} (${VERSION})..."
  CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
    "${GO_BIN}" build -trimpath -ldflags="${LDFLAGS}" \
    -o "${bundle_dir}/kilnscrub${ext}" \
    "${PKG}"

  cp "${REPO_ROOT}/README.md" "${bundle_dir}/README.md"
  cp "${REPO_ROOT}/LICENSE" "${bundle_dir}/LICENSE"
  cp "${REPO_ROOT}/configs/kilnscrub.example.json" "${bundle_dir}/configs/kilnscrub.example.json"
  cp "${REPO_ROOT}/packaging/systemd/kilnscrub-daemon.service" "${bundle_dir}/packaging/systemd/kilnscrub-daemon.service"

  archive_name="${bundle_name}.${fmt}"
  archive_path="${OUT_DIR}/${archive_name}"
  if [ "${fmt}" = "tar.gz" ]; then
    create_tar_gz "${STAGE_ROOT}" "${bundle_name}" "${archive_path}"
  else
    create_zip "${STAGE_ROOT}" "${bundle_name}" "${archive_path}"
  fi

  ARCHIVES+=("${archive_name}")
done

SORTED_ARCHIVES=()
while IFS= read -r archive_line; do
  [ -n "${archive_line}" ] && SORTED_ARCHIVES+=("${archive_line}")
done < <(printf '%s\n' "${ARCHIVES[@]}" | LC_ALL=C sort)
compute_sha256sums "${OUT_DIR}" "${SORTED_ARCHIVES[@]}"

echo "Release archives created in ${OUT_DIR}:"
ls -lh "${OUT_DIR}"
