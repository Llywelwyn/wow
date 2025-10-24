#!/usr/bin/env bash
# Run all .tape files except the EXCLUDE list.

set -euo pipefail

TMP_HOME="pda-tmp-home"

for file in *.tape; do
  if [[ "$file" == _* ]]; then
    echo "skipping $file (starts with _)"
    continue
  fi
  echo "running vhs on $file"
  vhs "$file"
done

echo "starting clean-up"
echo "removing $TMP_HOME"
rm -rf -- "$TMP_HOME"
echo "removing any leftover txt files"
rm -f -- ./*.txt || true
echo "Finished."

