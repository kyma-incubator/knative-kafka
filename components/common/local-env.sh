#!/usr/bin/env bash

# Ensure The Script Is Being Sourced
[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "\nWARNING: 'Source' this script for the environment setup to apply to the current session!\n"

# Absolute Path Of Directory Containing Script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Set The $GOPATH (Up 6 Levels To GO)
export GOPATH=$(dirname $(dirname $(dirname $(dirname $(dirname $(dirname ${SCRIPT_DIR}))))))

# Log The Exported Environment Variables
echo ""
echo "Exported Env Vars..."
echo "    GOPATH=${GOPATH}"
echo ""
