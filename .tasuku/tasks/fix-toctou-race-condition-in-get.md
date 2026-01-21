---
status: done
priority: 0
tags: [bug, data-integrity, security]
created_at: 2026-01-20T18:07:46.14539Z
updated_at: 2026-01-21T01:45:28.055098Z
---

# Fix TOCTOU race condition in GetOrCreateDailySeed (postgres.go:362-389) - use...

Fix TOCTOU race condition in GetOrCreateDailySeed (postgres.go:362-389) - use INSERT ON CONFLICT or advisory lock
