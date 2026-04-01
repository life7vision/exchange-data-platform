## 2025-05-15 - Enterprise Connector Security Hardening
**Vulnerability:** Lack of input validation and authentication on administrative endpoints (/run-once).
**Learning:** Open endpoints without validation are susceptible to path traversal via market/dataset names and unauthorized job triggering.
**Prevention:** Implement strict alphanumeric regex validation for input parameters and Bearer token authentication using `WORKER_API_TOKEN`.
