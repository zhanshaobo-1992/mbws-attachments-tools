#!/usr/bin/env bash
# Package Info-ZIP-compatible unzip binaries for MBWS attachment ingest.
# - darwin-*: uses macOS /usr/bin/unzip (often universal; same binary for arm64/x64 packs)
# - win-x64: repacks GnuWin32 unzip.exe to a flat zip (entry = unzip.exe)
# - linux-x64 (optional): if UNZIP_LINUX_BIN is set to a static/linux unzip path
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/dist/unzip"
VERSION="${UNZIP_VERSION:-6.0.0}"
mkdir -p "$OUT"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

pack_darwin() {
  local platform="$1"
  local dir="${OUT}/${platform}"
  rm -rf "$dir"
  mkdir -p "$dir"
  if [[ ! -x /usr/bin/unzip ]]; then
    echo "missing /usr/bin/unzip" >&2
    exit 1
  fi
  cp /usr/bin/unzip "${dir}/unzip"
  chmod +x "${dir}/unzip"
  tar -C "$dir" -czf "${OUT}/unzip-${VERSION}-${platform}.tar.gz" unzip
  echo "packed ${platform}"
}

pack_win() {
  local zipurl="${UNZIP_WIN_URL:-https://downloads.sourceforge.net/project/gnuwin32/unzip/5.51-1/unzip-5.51-1-bin.zip}"
  local raw="${TMP}/win.zip"
  echo "downloading Windows unzip from GnuWin32..."
  curl -sL --fail --max-time 120 -o "$raw" "$zipurl"
  local extract="${TMP}/win"
  mkdir -p "$extract"
  unzip -q -o "$raw" -d "$extract"
  local src
  src="$(find "$extract" -name 'unzip.exe' | head -1)"
  if [[ -z "$src" ]]; then
    echo "unzip.exe not found in GnuWin32 package" >&2
    exit 1
  fi
  local dir="${OUT}/win-x64"
  rm -rf "$dir"
  mkdir -p "$dir"
  cp "$src" "${dir}/unzip.exe"
  (cd "$dir" && zip -q "${OUT}/unzip-${VERSION}-win-x64.zip" unzip.exe)
  echo "packed win-x64 (note: upstream GnuWin32 is historically 5.51; packaged as ${VERSION} for MBWS semver slot)"
}

pack_linux() {
  if [[ -z "${UNZIP_LINUX_BIN:-}" ]]; then
    echo "skip linux-x64 (set UNZIP_LINUX_BIN=/path/to/unzip to include)"
    return 0
  fi
  local dir="${OUT}/linux-x64"
  rm -rf "$dir"
  mkdir -p "$dir"
  cp "$UNZIP_LINUX_BIN" "${dir}/unzip"
  chmod +x "${dir}/unzip"
  tar -C "$dir" -czf "${OUT}/unzip-${VERSION}-linux-x64.tar.gz" unzip
  echo "packed linux-x64"
}

pack_darwin darwin-arm64
pack_darwin darwin-x64
pack_win
pack_linux

echo "unzip ${VERSION} packages ready under ${OUT}"
ls -la "$OUT"/*.{tar.gz,zip} 2>/dev/null || ls -la "$OUT"
