#!/usr/bin/env bash

PROJECT_ROOT=$(git rev-parse --show-toplevel)

# Load Token
source ${PROJECT_ROOT}/.discord_token

# Run the bot
go run ${PROJECT_ROOT}/main.go