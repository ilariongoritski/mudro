#!/bin/bash
curl -s -X POST http://127.0.0.1:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"login": "admin", "password": "admin123"}'

echo ""

curl -s -X POST http://127.0.0.1:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"login": "user", "password": "user123"}'

echo ""
