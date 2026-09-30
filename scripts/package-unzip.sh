#!/usr/bin/env bash
# Deprecated: macOS /usr/bin/unzip packaging is unsafe for MBWS attachments
# (AMFI kills relocated Apple platform binaries). Use build-unzip.sh instead.
set -euo pipefail
echo "package-unzip.sh is deprecated — building portable Go unzip via build-unzip.sh" >&2
exec "$(cd "$(dirname "$0")" && pwd)/build-unzip.sh"
