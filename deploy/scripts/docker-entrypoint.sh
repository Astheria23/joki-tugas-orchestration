#!/bin/sh
set -e

echo "Starting Joki Orchestrator Service..."

# Execute the binary as PID 1 to handle signals gracefully
exec /app/orchestrator
