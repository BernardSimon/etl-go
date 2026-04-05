#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_NAME="${APP_NAME:-etl-go}"
MAIN_PACKAGE="${MAIN_PACKAGE:-.}"
VERSION="${VERSION:-${1:-}}"

if [[ -z "${VERSION}" ]]; then
  if git -C "${ROOT_DIR}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    VERSION="$(git -C "${ROOT_DIR}" describe --tags --always --dirty)"
  else
    VERSION="v0.0.0"
  fi
fi

VERSION="${VERSION#refs/tags/}"
VERSION="${VERSION#/}"

TARGETS=(
  "windows amd64"
  "windows arm64"
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
)

OUTPUT_ROOT="${ROOT_DIR}/bin/${VERSION}"
LDFLAGS="${LDFLAGS:--s -w}"

echo "==> release version: ${VERSION}"
echo "==> output dir: ${OUTPUT_ROOT}"

rm -rf "${OUTPUT_ROOT}"
mkdir -p "${OUTPUT_ROOT}"

for target in "${TARGETS[@]}"; do
  read -r goos goarch <<<"${target}"

  ext=""
  if [[ "${goos}" == "windows" ]]; then
    ext=".exe"
  fi

  target_dir="${OUTPUT_ROOT}/${goos}/${goarch}"
  binary_name="${APP_NAME}_${VERSION}_${goos}_${goarch}${ext}"
  output_file="${target_dir}/${binary_name}"

  mkdir -p "${target_dir}"

  echo "==> building ${goos}/${goarch}"
  CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
    go build -trimpath -ldflags "${LDFLAGS}" -o "${output_file}" "${MAIN_PACKAGE}"
done

echo
echo "release artifacts:"
find "${OUTPUT_ROOT}" -type f | sort
