# Loop Cycle Template

## Date: <date>
## Branch: loop/<date>-<slug>

### SCAN
- Services status:
- Healthz checks:
- go vet:
- Log scan:
- Recent changes:

### PLAN
- Improvement 1: [description] (risk: low/medium/high)
- Improvement 2: [description] (risk: low/medium/high)
- Improvement 3: [description] (risk: low/medium/high)

### BUILD
- Files changed:
- Lines diff:
- Commit:

### TEST
- go test ./...: [PASS/FAIL]
- docker compose build: [PASS/FAIL]
- Notes:

### DEPLOY
- Services deployed:
- healthz after 30s:
- Rollback plan:

### MONITOR
- 5min watch result:
- Errors detected:
- Final status: [STABLE/ROLLED_BACK]

## Telegram notification sent: YES/NO
