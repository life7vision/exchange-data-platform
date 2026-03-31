## 2025-05-14 - Path Traversal & Missing Authentication in RunOnce Endpoint
**Vulnerability:** The `/run-once` endpoint allowed arbitrary strings in `Datasets` and `Markets` fields, which were used directly in file system paths by the storage layer, enabling potential path traversal. Additionally, the endpoint had no authentication.
**Learning:** Even internal-facing endpoints should have basic authentication and strict input validation, especially when inputs are used to construct file paths.
**Prevention:** Use allowlists or strict regex validation for any user-provided strings used in path construction. Implement opt-in authentication for sensitive operations.
