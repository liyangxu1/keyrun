#!/usr/bin/env bash
set -euo pipefail

release='https://github.com/liyangxu1/keyrun/releases/download/typora-desktop'
case "$(uname -m)" in
    x86_64) asset=typora-linux-amd64 ;;
    aarch64|arm64) asset=typora-linux-arm64 ;;
    *) printf '%s\n' '不支持当前 Linux 架构。' >&2; exit 1 ;;
esac

work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
curl -fsSL --retry 2 "$release/$asset" -o "$work/$asset"
curl -fsSL --retry 2 "$release/SHA256SUMS" -o "$work/SHA256SUMS"
expected=$(awk -v file="$asset" '$2 == file && length($1) == 64 && $1 ~ /^[a-fA-F0-9]+$/ {print $1}' "$work/SHA256SUMS")
actual=$(sha256sum "$work/$asset")
if [[ "${actual%% *}" != "$expected" ]]; then
    printf '%s\n' '下载文件的 SHA-256 校验失败。' >&2
    exit 1
fi
chmod u+x "$work/$asset"
"$work/$asset" "$@"
