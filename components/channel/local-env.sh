#!/usr/bin/env bash

# Ensure The Script Is Being Sourced
[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "\nWARNING: 'Source' this script for the environment setup to apply to the current session!\n"

# Absolute Path Of Directory Containing Script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Set Local Testing Environment Variables
export METRICS_PORT="8081"
export KAFKA_BROKERS="todo"
export KAFKA_USERNAME="todo"
export KAFKA_PASSWORD="todo"

# Log Environment Variables
echo ""
echo "Exported Env Vars..."
echo ""
echo "METRICS_PORT=${METRICS_PORT}"
echo "KAFKA_BROKERS=${KAFKA_BROKERS}"
echo "KAFKA_USERNAME=${KAFKA_USERNAME}"
echo "KAFKA_PASSWORD=${KAFKA_PASSWORD}"
echo ""
