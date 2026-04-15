#!/usr/bin/env bash
set -euo pipefail

echo "Release demo flow:"
echo "1) make demo-up"
echo "2) cd frontend && npm.cmd run dev"
echo "3) make demo-check"

make demo-up

echo
echo "Start the frontend in another terminal: cd frontend && npm.cmd run dev"
echo "Then run: make demo-check"
