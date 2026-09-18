#!/usr/bin/env bash
set -euo pipefail

# === 校验同步参数 ===
if (( $# < 2 || $# > 3 )); then
  echo "usage: sync.sh SOURCE TARGET [--apply]" >&2
  exit 2
fi

source_root=$1
target_root=$2
mode=${3:-report}
[[ "$mode" == report || "$mode" == --apply ]] || exit 2
test -d "$target_root"

# === 比较并同步托管文件 ===
changed=0
while IFS= read -r path; do
  [[ -z "$path" || "$path" == "#"* ]] && continue
  test -f "$source_root/$path"
  if [[ ! -e "$target_root/$path" ]]; then
    echo "missing $path"
    changed=1
  elif ! cmp -s "$source_root/$path" "$target_root/$path"; then
    echo "different $path"
    changed=1
  else
    echo "current $path"
  fi

  if [[ "$mode" == --apply ]] && ! cmp -s "$source_root/$path" "$target_root/$path"; then
    mkdir -p "$(dirname "$target_root/$path")"
    cp "$source_root/$path" "$target_root/$path"
  fi
done < "$source_root/managed-files.txt"

# === 报告同步结果 ===
if [[ "$mode" == --apply ]]; then
  echo "applied managed baseline"
elif (( changed )); then
  echo "managed baseline differs"
else
  echo "managed baseline is current"
fi
