-- Add integer seconds columns and staging table for fast COPY
ALTER TABLE stop_times
    ADD COLUMN IF NOT EXISTS arrival_sec INTEGER,
    ADD COLUMN IF NOT EXISTS departure_sec INTEGER;

-- Create unlogged staging table without indexes or constraints for bulk load
CREATE UNLOGGED TABLE IF NOT EXISTS staging_stop_times (
    trip_id TEXT NOT NULL,
    arrival_sec INTEGER,
    departure_sec INTEGER,
    stop_id TEXT NOT NULL,
    stop_sequence INTEGER NOT NULL,
    stop_headsign TEXT,
    pickup_type INTEGER,
    drop_off_type INTEGER,
    shape_dist_traveled DOUBLE PRECISION,
    timepoint INTEGER
);