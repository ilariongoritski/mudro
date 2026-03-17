# Security Checklist

- [ ] Input validation on all entry points
- [ ] SQL injections: parameterized queries used
- [ ] XSS: output data escaped
- [ ] Authorization: checked on every endpoint
- [ ] Authentication: tokens/sessions have TTL
- [ ] Secrets: not stored in code, vault/env used
- [ ] Logging: sensitive data does not leak into logs
- [ ] Rate limiting: configured on public endpoints
- [ ] CORS: configured correctly
- [ ] Dependencies: no known CVEs (checked via `audit`)
- [ ] Errors: do not expose internal structure in API responses
- [ ] File uploads: type and size validation
