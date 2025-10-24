#!/usr/bin/env bash

set -euo pipefail

TEST_DIR="${1:-cmd/pda/cmdtest}"

command -v fzf >/dev/null 2>&1 || { echo "err: fzf is required" >&2; exit 1; }
[[ -d "$TEST_DIR" ]] || { echo "err: dir not found: $TEST_DIR" >&2; exit 1; }

mapfile -t files < <(find "$TEST_DIR" -maxdepth 1 -type f -printf '%f\n' | sort)
[[ ${#files[@]} -gt 0 ]] || { echo "err: no files in $TEST_DIR" >&2; exit 1; }

sel="$(printf '%s\n' "${files[@]}" | fzf --prompt="pick base test > " --height=60% --border)"
[[ -n "${sel:-}" ]] || { echo "cancelled" >&2; exit 1; }

ext="ct"
read -rp "new test name: " raw_name
[[ -n "${raw_name// }" ]] || { echo "err: name is empty"; exit 1; }

sanitised="$(printf '%s' "$raw_name" \
  | tr '[:space:]' '_' \
  | sed 's/[^A-Za-z0-9._-]/_/g' \
  | sed 's/_\+/_/g; s/^_//; s/_$//')"
[[ -n "$sanitised" ]] || { echo "err: name empty after sanitising"; exit 1; }

declare -A used=()
while IFS= read -r f; do
    if [[ "$f" =~ ^([0-9]{3})_ ]]; then
        used["${BASH_REMATCH[1]}"]=1
    fi
done < <(printf '%s\n' "${files[@]}")

num=1
while :; do
    prefix=$(printf '%03d' "$num")
    if [[ -z "${used[$prefix]:-}" && ! -e "$TEST_DIR/$prefix"_* ]]; then
        break
    fi
    ((num++))
done

filename="${prefix}_${sanitised}.${ext}"
src="$TEST_DIR/$sel"
dst="$TEST_DIR/$filename"


if [[ -e "$dst" ]]; then
    echo "err: dst already exists: $dst" >&2
    exit 1
fi

cp -- "$src" "$dst"

echo "created: $dst"
echo "based on: $src"

if [[ -n "${EDITOR:-}" ]]; then
    read -rp "open in $EDITOR?" yn
    if [[ "$yn" =~ ^[Yy]$ ]]; then
        "$EDITOR" "$dst"
    fi
fi
