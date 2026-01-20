# Integration Test Schema Out of Sync with Migrations

## Problem

Integration tests in `internal/db/postgres_integration_test.go` fail with:

```
ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification (SQLSTATE 42P10)
```

This occurs when `AddLeaderboardEntry` uses `ON CONFLICT (user_id, seed)` but the test database lacks the required unique constraint.

## Root Cause

The integration test uses a **hardcoded test schema** (`const testSchema`) instead of running actual migrations. When migration 004 added the `idx_leaderboard_user_seed` unique index to production, the test schema was not updated.

Additionally, test data was creating entries without unique seeds, causing UPSERT behavior to merge entries instead of creating separate ones.

## Solution

### 1. Add missing constraint to test schema

In `internal/db/postgres_integration_test.go`, add the unique index after the `leaderboard_entries` table:

```sql
-- Unique constraint for UPSERT on (user_id, seed)
CREATE UNIQUE INDEX IF NOT EXISTS idx_leaderboard_user_seed
ON leaderboard_entries(user_id, seed);
```

### 2. Update test data to use unique seeds

When testing multiple leaderboard entries, ensure each has a unique `(user_id, seed)` combination:

```go
entries := []*LeaderboardEntry{
    {UserID: user1.ID, RunType: "standard", Seed: 1001, Score: 1000, ...},
    {UserID: user2.ID, RunType: "standard", Seed: 1002, Score: 2000, ...},
    {UserID: user1.ID, RunType: "standard", Seed: 1004, Score: 3000, ...}, // Different seed!
}
```

## Prevention

When adding database constraints or indexes in migrations:

1. **Update the test schema** in `postgres_integration_test.go` to match
2. **Update test data** to respect new constraints
3. **Run integration tests locally** before pushing: `go test -tags=integration ./internal/db/...`

## Related Files

- `internal/db/postgres_integration_test.go` - Test schema and test cases
- `internal/db/migrate/migrations/004_leaderboard_unique_seed.up.sql` - Production migration
- `internal/db/postgres.go:AddLeaderboardEntry` - Uses ON CONFLICT

## Commits

- `c3c08a4` - Add missing unique index to test schema
- `7974ec3` - Update test data to use unique seeds
