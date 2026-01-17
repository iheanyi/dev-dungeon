-- Migration 003: Fix item template IDs
-- SKIPPED: The code fix for template IDs is already deployed.
-- New saves will have correct template IDs. Old corrupted saves
-- will fail to load items (which is acceptable for early dev).
-- Original migration was too complex and caused dirty state issues.

SELECT 1;
