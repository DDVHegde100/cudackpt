#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="$(tr -d ' \n' < VERSION)"
PKG="cudackpt_${VERSION}_amd64"
STAGE="$ROOT/build/deb/$PKG"
rm -rf "$STAGE"

make install DESTDIR="$STAGE"
mkdir -p "$STAGE/DEBIAN"
cat > "$STAGE/DEBIAN/control" <<EOF
Package: cudackpt
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Dhruv Hegde <ddvhegde100@gmail.com>
Description: CUDA process checkpoint and restore with CRIU
 Single-GPU checkpoint/restore control plane and LD_PRELOAD shim.
EOF

install -d "$STAGE/lib/systemd/system"
install -m 644 deploy/cudackpt-run.service "$STAGE/lib/systemd/system/"
install -m 644 deploy/cudackpt.socket "$STAGE/lib/systemd/system/"
install -m 644 deploy/cudackpt@.service "$STAGE/lib/systemd/system/"
install -m 644 deploy/cudackpt-agent.service "$STAGE/lib/systemd/system/"

OUT="$ROOT/build/${PKG}.deb"
dpkg-deb --build "$STAGE" "$OUT"
echo "built $OUT"
