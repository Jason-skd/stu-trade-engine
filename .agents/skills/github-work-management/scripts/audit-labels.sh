#!/usr/bin/env bash
set -euo pipefail

# === 校验审计参数 ===
if (( $# != 2 )); then
  echo "usage: audit-labels.sh OWNER/REPO LABELS_YML" >&2
  exit 2
fi

repo=$1
labels_file=$2
command -v gh >/dev/null
command -v uv >/dev/null

expected=$(mktemp)
actual=$(mktemp)
trap 'rm -f "$expected" "$actual"' EXIT

# === 归一化预期标签 ===
uv run --with pyyaml python -c '
import json, sys, yaml
data = yaml.safe_load(open(sys.argv[1]))
print(json.dumps(sorted(
    ({"name": item["name"], "color": item["color"].upper(), "description": item["description"]} for item in data["labels"]),
    key=lambda item: item["name"],
)))
' "$labels_file" > "$expected"
gh api "repos/$repo/labels?per_page=100" > "$actual"

# === 报告标签差异 ===
uv run python -c '
import json, sys
expected = json.load(open(sys.argv[1]))
actual = [
    {"name": item["name"], "color": item["color"].upper(), "description": item.get("description")}
    for item in json.load(open(sys.argv[2]))
]
print(json.dumps({
    "missing": [item for item in expected if item not in actual],
    "unexpected": [item for item in actual if item not in expected],
}, indent=2, sort_keys=True))
' "$expected" "$actual"
