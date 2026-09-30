#!/usr/bin/env bash
# Cross-build ziplist binaries and pack per-platform archives for MBWS attachment ingest.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/dist/ziplist"
VERSION="${ZIPLIST_VERSION:-1.0.0}"
mkdir -p "$OUT"
cd "${ROOT}/ziplist"

build_one() {
  local goos="$1" goarch="$2" platform="$3" ext="${4:-}"
  local bin="ziplist${ext}"
  local dir="${OUT}/${platform}"
  rm -rf "$dir"
  mkdir -p "$dir"
  echo "building ${platform}..."
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "${dir}/${bin}" .
  if [[ "$goos" == "windows" ]]; then
    (cd "$dir" && zip -q "${OUT}/ziplist-${VERSION}-${platform}.zip" "$bin")
  else
    tar -C "$dir" -czf "${OUT}/ziplist-${VERSION}-${platform}.tar.gz" "$bin"
  fi
  echo "  -> ${OUT}/ziplist-${VERSION}-${platform}.$([[ $goos == windows ]] && echo zip || echo tar.gz)"
}

build_one darwin arm64 darwin-arm64
build_one darwin amd64 darwin-x64
build_one windows amd64 win-x64 .exe
build_one linux amd64 linux-x64

echo "ziplist ${VERSION} packages ready under ${OUT}"
ls -la "$OUT"/*.{tar.gz,zip} 2>/dev/null || ls -la "$OUT"
