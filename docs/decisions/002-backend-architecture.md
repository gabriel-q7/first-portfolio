# ADR 002: Single-process backend with SQLite

Status: accepted.

For the one-vCPU, 1 GB VPS, the backend remains a layered modular monolith.
SQLite implements the existing repository interface, while a bounded local
cache replaces network caching.

This removes network database/cache daemons, connection pools, background
workers, and metrics collectors. The trade-off is a single-writer database and
one backend replica, which is acceptable for the expected portfolio workload.
If write concurrency or horizontal scaling becomes necessary, persistence must
be reconsidered as a separate architecture change.
