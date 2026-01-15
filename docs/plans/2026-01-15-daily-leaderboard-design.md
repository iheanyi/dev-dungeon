# Daily Leaderboard & Seed Security Design

**Date:** 2026-01-15
**Status:** Approved

## Overview

Two improvements to the daily run system:
1. **Unpredictable daily seeds** - Prevent users from pre-calculating future seeds
2. **Navigable daily leaderboard** - Browse past 7 days of daily challenges

## Design Decisions

### Daily Seeds: Crypto-Random Generation

**Problem:** Current implementation uses `time.Now().UTC().Truncate(24*time.Hour).UnixNano()` which is completely predictable. Users can calculate future seeds and practice ahead of time.

**Solution:** Generate seeds using `crypto/rand` when the day starts.

```go
// Instead of: seed = today.UnixNano()
bytes := make([]byte, 8)
crypto_rand.Read(bytes)
seed = int64(binary.BigEndian.Uint64(bytes))
```

Seeds are stored in `daily_seeds` table and only become knowable once generated.

### Daily Runs: SSH-Only

**Decision:** Remove daily runs from local CLI entirely.

**Rationale:**
- Daily runs are a competitive feature (same seed + leaderboard)
- Local CLI can't submit scores or view leaderboards
- Keeps local experience simple, competitive features on SSH

**Flow:**
- `./devdungeon` locally → Standard runs, seeded runs (fully offline)
- `ssh player@devdungeon.io` → Standard, seeded, AND daily runs with leaderboards

### Leaderboard Navigation

**Horizontal (days):** Arrow keys (`←`/`→`) to navigate last 7 days

**Vertical (entries):** Fixed "Top N + Your Rank" view
- Show top 10 entries
- If player not in top 10, show separator then their position
- Highlight player's entry wherever it appears

```
═══ Daily Leaderboard (Jan 15) ═══  ← →
 #1  rootkiller     12,450  Floor 8
 #2  daemon_slayer  11,200  Floor 7
 #3  fork_master     9,800  Floor 7
...
 #8  you_here        5,200  Floor 4
...
#47  process_zero    2,100  Floor 2  ← You
```

**Extensibility:** Data layer supports full pagination (offset/limit) for future scrollable UI.

## Data Model Changes

### Repository Interface Additions

```go
// Daily seed operations
GetDailySeed(ctx context.Context, date time.Time) (*DailySeed, error)
GetOrCreateDailySeed(ctx context.Context) (int64, error)  // existing, update impl

// Leaderboard cursor for stable pagination
type LeaderboardCursor struct {
    Score int
    ID    int  // tiebreaker for equal scores
}

// Leaderboard operations (new)
GetDailyLeaderboard(ctx context.Context, date time.Time, limit int, cursor *LeaderboardCursor) (entries []LeaderboardEntry, nextCursor *LeaderboardCursor, error)
GetPlayerDailyRank(ctx context.Context, date time.Time, userID int) (rank int, entry *LeaderboardEntry, error)
```

### Why Cursor-Based Pagination

Offset/limit has issues for leaderboards:
- New scores can be inserted at any rank while browsing
- Offset would shift results between pages (duplicates or missed entries)
- Cursor on `(score, id)` is stable even with concurrent submissions

### Query: GetDailyLeaderboard

```sql
-- First page (no cursor)
SELECT id, nanoid, user_id, username, run_type, seed, score,
       floors_cleared, time_seconds, class, created_at
FROM leaderboard
WHERE run_type = 'daily'
  AND seed = (SELECT seed FROM daily_seeds WHERE date = $1)
ORDER BY score DESC, id ASC
LIMIT $2

-- Subsequent pages (with cursor)
SELECT id, nanoid, user_id, username, run_type, seed, score,
       floors_cleared, time_seconds, class, created_at
FROM leaderboard
WHERE run_type = 'daily'
  AND seed = (SELECT seed FROM daily_seeds WHERE date = $1)
  AND ((score < $cursor_score) OR (score = $cursor_score AND id > $cursor_id))
ORDER BY score DESC, id ASC
LIMIT $2
```

### Query: GetPlayerDailyRank

```sql
WITH ranked AS (
  SELECT *, RANK() OVER (ORDER BY score DESC) as rank
  FROM leaderboard
  WHERE run_type = 'daily'
    AND seed = (SELECT seed FROM daily_seeds WHERE date = $1)
)
SELECT rank, id, nanoid, user_id, username, score, floors_cleared,
       time_seconds, class, created_at
FROM ranked
WHERE user_id = $2
```

## Implementation Tasks

1. **Update daily seed generation** (`internal/db/postgres.go`)
   - Change from `UnixNano()` to `crypto/rand`
   - Update in-memory implementation too

2. **Add repository methods**
   - `GetDailySeed(date)` - fetch seed for specific date
   - `GetDailyLeaderboard(date, limit, offset)` - paginated daily entries
   - `GetPlayerDailyRank(date, userID)` - player's rank for a day

3. **Add leaderboard UI component** (`internal/ui/`)
   - New view type for daily leaderboard
   - Arrow key navigation for days (last 7)
   - Top N + your rank display
   - Highlight current player

4. **Remove daily runs from local CLI**
   - Update main menu to hide daily option when not in SSH session
   - Or gate the option behind a "connected" check

5. **Tests**
   - Seed randomness (not predictable from date)
   - Leaderboard pagination
   - Rank calculation
   - Day navigation bounds (can't go beyond 7 days or into future)

## Out of Scope

- Calendar date picker (arrows sufficient for 7 days)
- Infinite scroll (top N + rank is enough for now)
- Daily run rewards/streaks (future feature)
