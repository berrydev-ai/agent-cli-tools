#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 tool|version" >&2
  exit 2
}

mode="${1:-}"
case "${mode}" in
  tool)
    if [[ -z "${TOOL:-}" ]]; then
      echo "TOOL is required, e.g. TOOL=slack-post" >&2
      exit 2
    fi
    if [[ ! "${TOOL}" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
      echo "TOOL must contain only lowercase letters, digits, and hyphens" >&2
      exit 2
    fi
    if [[ ! -d "cmd/${TOOL}" ]]; then
      echo "unknown TOOL; run 'make list' for available binaries" >&2
      exit 2
    fi
    ;;
  version)
    if [[ -z "${VERSION:-}" ]]; then
      echo "VERSION is required, e.g. VERSION=v0.1.0" >&2
      exit 2
    fi
    semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?(\+([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?$'
    if [[ ! "${VERSION}" =~ ${semver} ]]; then
      echo "VERSION must be valid SemVer with a leading v, e.g. v0.1.0" >&2
      exit 2
    fi
    ;;
  *)
    usage
    ;;
esac
