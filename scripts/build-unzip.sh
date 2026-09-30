#!/usr/bin/env bash
# Cross-build portable Go unzip binaries (NOT macOS /usr/bin/unzip).
# Apple's system unzip is AMFI-restricted and dies (exit 137/-1) when relocated
# into the MBWS attachment cache — hence a static Go reimplementation.
#
# CLI aligned with Info-ZIP subset used by MBWS plugins:
#   unzip -Z1 <zip>
#   unzip -o <zip> -d <dir>
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/dist/unzip"
VERSION="${UNZIP_VERSION:-6.0.0}"
mkdir -p "$OUT"
cd "${ROOT}/unzip"

build_one() {
  local goos="$1" goarch="$2" platform="$3" ext="${4:-}"
  local bin="unzip${ext}"
  local dir="${OUT}/${platform}"
  rm -rf "$dir"
  mkdir -p "$dir"
  echo "building ${platform}..."
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "${dir}/${bin}" .
  if [[ "$goos" == "windows" ]]; then
    (cd "$dir" && zip -q "${OUT}/unzip-${VERSION}-${platform}.zip" "$bin")
  else
    tar -C "$dir" -czf "${OUT}/unzip-${VERSION}-${platform}.tar.gz" "$bin"
  fi
  echo "  -> ${OUT}/unzip-${VERSION}-${platform}.$([[ $goos == windows ]] && echo zip || echo tar.gz)"
}

build_one darwin arm64 darwin-arm64
build_one darwin amd64 darwin-x64
build_one windows amd64 win-x64 .exe
build_one linux amd64 linux-x64

echo "unzip ${VERSION} packages ready under ${OUT}"
ls -la "$OUT"/*.{tar.gz,zip} 2>/dev/null || ls -la "$OUT"
