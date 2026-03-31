## 2025-05-15 - Observability and Error Tracking
**Learning:** Connectors lacked per-exchange error metrics, making it hard to identify specific exchange failures in a shared environment.
**Action:** Implement `exchange_errors_total` metric and structured logging (slog) in all connectors for better fetch lifecycle tracking.
