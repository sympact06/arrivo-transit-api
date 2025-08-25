-- Revert stop_times optimization changes
ALTER TABLE stop_times
    DROP COLUMN IF EXISTS arrival_sec,
    DROP COLUMN IF EXISTS departure_sec;

DROP TABLE IF EXISTS staging_stop_times;