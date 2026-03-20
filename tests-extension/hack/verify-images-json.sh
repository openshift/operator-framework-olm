#!/usr/bin/env bash

# verify-images-json.sh
#
# Verifies that the 'images' subcommand outputs valid JSON without log pollution.
#
# Usage:
#   ./hack/verify-images-json.sh [path-to-binary]
#
# Example:
#   ./hack/verify-images-json.sh ./bin/olmv0-tests-ext

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Default binary path
BINARY="${1:-${PROJECT_ROOT}/bin/olmv0-tests-ext}"

echo "Verifying olmv0-tests-ext images output is valid JSON..."

# Check if binary exists
if [[ ! -f "${BINARY}" ]]; then
    echo -e "ERROR: Binary not found at: ${BINARY}"
    echo "Please run 'make build' first."
    exit 1
fi

# Run the images command and capture all output (stdout + stderr)
output=$("${BINARY}" images 2>&1)

# Create temporary directory for Go validation program
tmpdir=$(mktemp -d)
trap 'rm -rf "${tmpdir}"' EXIT

# Write Go validation program to temporary file
cat > "${tmpdir}/validate.go" <<'GO_CODE'
package main

import (
	"encoding/json"
	"io"
	"os"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	if !json.Valid(data) {
		os.Exit(1)
	}
}
GO_CODE

# Validate JSON using Go's json.Valid function
# This matches exactly what binary.go:548 does (json.Unmarshal)
if ! echo "${output}" | go run "${tmpdir}/validate.go" 2>/dev/null; then
    echo -e "ERROR: 'olmv0-tests-ext images' output is not valid JSON!"
    echo "This usually means log statements are polluting the JSON output."
    echo ""
    echo "Output was:"
    echo "----------------------------------------"
    echo "${output}"
    echo "----------------------------------------"
    echo ""
    exit 1
fi

echo -e "olmv0-tests-ext images output is valid JSON"
exit 0
