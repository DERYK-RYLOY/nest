#!/usr/bin/env bash
# Build the NEST Windows installer. Run `wails build` first.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ ! -f "$here/../../bin/nest.exe" ]; then
  echo "build/bin/nest.exe is missing - run 'wails build' first." >&2
  exit 1
fi

version="${1:-}"
for candidate in \
  "makensis" \
  "/c/Program Files/NSIS/makensis.exe" \
  "/c/Program Files (x86)/NSIS/makensis.exe"; do
  if [ "$candidate" = "makensis" ]; then
    command -v makensis >/dev/null 2>&1 || continue
  elif [ ! -x "$candidate" ]; then
    continue
  fi

  if [ -n "$version" ]; then
    "$candidate" -V2 -DAPP_VERSION="$version" "$here/installer.nsi"
  else
    "$candidate" -V2 "$here/installer.nsi"
  fi
  echo "Installer written to build/bin/"
  exit 0
done

echo "NSIS not found. Install it with 'winget install NSIS.NSIS' or from https://nsis.sourceforge.io/." >&2
exit 1
