package gtfs

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const gtfsURL = "http://gtfs.ovapi.nl/gtfs-nl.zip"

// Service handles the GTFS data processing.
type Service struct {
	pool *pgxpool.Pool
}

// NewService creates a new GTFS service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// IngestGTFSData downloads and processes GTFS data.
func (s *Service) IngestGTFSData() error {
	log.Println("Starting GTFS data ingestion...")

	zipPath, err := s.getGTFSData()
	if err != nil {
		return fmt.Errorf("failed to get GTFS data: %w", err)
	}

	extractPath, err := s.extractGTFS(zipPath)
	if err != nil {
		return fmt.Errorf("failed to extract GTFS data: %w", err)
	}
	defer os.RemoveAll(extractPath) // Clean up the extracted files.

	if err := s.processGTFS(extractPath); err != nil {
		return fmt.Errorf("failed to process GTFS data: %w", err)
	}

	log.Println("GTFS data ingestion completed successfully.")
	return nil
}

func (s *Service) getGTFSData() (string, error) {
	localPath := "/app/gtfs-data/gtfs-nl.zip"
	if _, err := os.Stat(localPath); err == nil {
		log.Printf("Using local GTFS data from %s", localPath)
		return localPath, nil
	}

	log.Printf("Downloading GTFS data from %s", gtfsURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", gtfsURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Arrivo-Transit-API/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download GTFS data: status code %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "gtfs-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}

	log.Printf("GTFS data downloaded to %s", tmpFile.Name())
	return tmpFile.Name(), nil
}

func (s *Service) processGTFS(gtfsPath string) error {
	log.Println("Processing GTFS data...")

	if err := s.processStops(gtfsPath); err != nil {
		return fmt.Errorf("failed to process stops: %w", err)
	}

	if err := s.processRoutes(gtfsPath); err != nil {
		return fmt.Errorf("failed to process routes: %w", err)
	}

	if err := s.processTrips(gtfsPath); err != nil {
		return fmt.Errorf("failed to process trips: %w", err)
	}

	if err := s.processStopTimesFast(gtfsPath); err != nil {
		return fmt.Errorf("failed to process stop_times: %w", err)
	}

	log.Println("Successfully processed GTFS data")
	return nil
}

func (s *Service) processStops(gtfsPath string) error {
	log.Println("Processing stops.txt...")
	stopsFile, err := os.Open(filepath.Join(gtfsPath, "stops.txt"))
	if err != nil {
		return fmt.Errorf("failed to open stops.txt: %w", err)
	}
	defer stopsFile.Close()

	reader := csv.NewReader(stopsFile)
	header, err := reader.Read() // Read header
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background()) // Rollback on error

	// Use direct Exec instead of prepared statements for batch inserts

	batchSize := 1000
	valueStrings := make([]string, 0, batchSize)
	valueArgs := make([]interface{}, 0, batchSize*12)
	i := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		i++
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*12-11, i*12-10, i*12-9, i*12-8, i*12-7, i*12-6, i*12-5, i*12-4, i*12-3, i*12-2, i*12-1, i*12))

		stopData := make(map[string]string)
		for i, value := range record {
			stopData[header[i]] = value
		}

		lat, _ := strconv.ParseFloat(stopData["stop_lat"], 64)
		lon, _ := strconv.ParseFloat(stopData["stop_lon"], 64)
		locationType, _ := strconv.Atoi(stopData["location_type"])
		wheelchairBoarding, _ := strconv.Atoi(stopData["wheelchair_boarding"])

		valueArgs = append(valueArgs, stopData["stop_id"], stopData["stop_code"], stopData["stop_name"], stopData["stop_desc"], lat, lon, stopData["zone_id"], stopData["stop_url"], locationType, stopData["parent_station"], stopData["stop_timezone"], wheelchairBoarding)

		if len(valueStrings) == batchSize {
			stmt := fmt.Sprintf("INSERT INTO stops (stop_id, stop_code, stop_name, stop_desc, stop_lat, stop_lon, zone_id, stop_url, location_type, parent_station, stop_timezone, wheelchair_boarding) VALUES %s ON CONFLICT (stop_id) DO UPDATE SET stop_code = EXCLUDED.stop_code, stop_name = EXCLUDED.stop_name, stop_desc = EXCLUDED.stop_desc, stop_lat = EXCLUDED.stop_lat, stop_lon = EXCLUDED.stop_lon, zone_id = EXCLUDED.zone_id, stop_url = EXCLUDED.stop_url, location_type = EXCLUDED.location_type, parent_station = EXCLUDED.parent_station, stop_timezone = EXCLUDED.stop_timezone, wheelchair_boarding = EXCLUDED.wheelchair_boarding",
				strings.Join(valueStrings, ","))
			_, err = tx.Exec(context.Background(), stmt, valueArgs...)
			if err != nil {
				return err
			}
			valueStrings = make([]string, 0, batchSize)
			valueArgs = make([]interface{}, 0, batchSize*12)
			i = 0
		}
	}

	if len(valueStrings) > 0 {
		stmt := fmt.Sprintf("INSERT INTO stops (stop_id, stop_code, stop_name, stop_desc, stop_lat, stop_lon, zone_id, stop_url, location_type, parent_station, stop_timezone, wheelchair_boarding) VALUES %s ON CONFLICT (stop_id) DO UPDATE SET stop_code = EXCLUDED.stop_code, stop_name = EXCLUDED.stop_name, stop_desc = EXCLUDED.stop_desc, stop_lat = EXCLUDED.stop_lat, stop_lon = EXCLUDED.stop_lon, zone_id = EXCLUDED.zone_id, stop_url = EXCLUDED.stop_url, location_type = EXCLUDED.location_type, parent_station = EXCLUDED.parent_station, stop_timezone = EXCLUDED.stop_timezone, wheelchair_boarding = EXCLUDED.wheelchair_boarding",
			strings.Join(valueStrings, ","))
		_, err = tx.Exec(context.Background(), stmt, valueArgs...)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

func (s *Service) processRoutes(gtfsPath string) error {
	log.Println("Processing routes.txt...")
	routesFile, err := os.Open(filepath.Join(gtfsPath, "routes.txt"))
	if err != nil {
		return fmt.Errorf("failed to open routes.txt: %w", err)
	}
	defer routesFile.Close()

	reader := csv.NewReader(routesFile)
	header, err := reader.Read() // Read header
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background()) // Rollback on error

	batchSize := 1000
	valueStrings := make([]string, 0, batchSize)
	valueArgs := make([]interface{}, 0, batchSize*9)
	i := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		i++
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9-8, i*9-7, i*9-6, i*9-5, i*9-4, i*9-3, i*9-2, i*9-1, i*9))

		routeData := make(map[string]string)
		for i, value := range record {
			routeData[header[i]] = value
		}

		routeType, _ := strconv.Atoi(routeData["route_type"])

		valueArgs = append(valueArgs, routeData["route_id"], routeData["agency_id"], routeData["route_short_name"], routeData["route_long_name"], routeData["route_desc"], routeType, routeData["route_url"], routeData["route_color"], routeData["route_text_color"])

		if len(valueStrings) == batchSize {
			stmt := fmt.Sprintf("INSERT INTO routes (id, agency_id, route_short_name, route_long_name, route_desc, route_type, route_url, route_color, route_text_color) VALUES %s ON CONFLICT (id) DO UPDATE SET agency_id = EXCLUDED.agency_id, route_short_name = EXCLUDED.route_short_name, route_long_name = EXCLUDED.route_long_name, route_desc = EXCLUDED.route_desc, route_type = EXCLUDED.route_type, route_url = EXCLUDED.route_url, route_color = EXCLUDED.route_color, route_text_color = EXCLUDED.route_text_color",
				strings.Join(valueStrings, ","))
			_, err = tx.Exec(context.Background(), stmt, valueArgs...)
			if err != nil {
				return err
			}
			valueStrings = make([]string, 0, batchSize)
			valueArgs = make([]interface{}, 0, batchSize*9)
			i = 0
		}
	}

	if len(valueStrings) > 0 {
		stmt := fmt.Sprintf("INSERT INTO routes (id, agency_id, route_short_name, route_long_name, route_desc, route_type, route_url, route_color, route_text_color) VALUES %s ON CONFLICT (id) DO UPDATE SET agency_id = EXCLUDED.agency_id, route_short_name = EXCLUDED.route_short_name, route_long_name = EXCLUDED.route_long_name, route_desc = EXCLUDED.route_desc, route_type = EXCLUDED.route_type, route_url = EXCLUDED.route_url, route_color = EXCLUDED.route_color, route_text_color = EXCLUDED.route_text_color",
			strings.Join(valueStrings, ","))
		_, err = tx.Exec(context.Background(), stmt, valueArgs...)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Finished processing routes.txt")
	return nil
}

func (s *Service) processTrips(gtfsPath string) error {
	log.Println("Processing trips.txt...")
	tripsFile, err := os.Open(filepath.Join(gtfsPath, "trips.txt"))
	if err != nil {
		return fmt.Errorf("failed to open trips.txt: %w", err)
	}
	defer tripsFile.Close()

	reader := csv.NewReader(tripsFile)
	header, err := reader.Read() // Read header
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background()) // Rollback on error

	batchSize := 1000
	valueStrings := make([]string, 0, batchSize)
	valueArgs := make([]interface{}, 0, batchSize*10)
	i := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		i++
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*10-9, i*10-8, i*10-7, i*10-6, i*10-5, i*10-4, i*10-3, i*10-2, i*10-1, i*10))

		tripData := make(map[string]string)
		for i, value := range record {
			tripData[header[i]] = value
		}

		directionID, _ := strconv.Atoi(tripData["direction_id"])
		wheelchairAccessible, _ := strconv.Atoi(tripData["wheelchair_accessible"])
		bikesAllowed, _ := strconv.Atoi(tripData["bikes_allowed"])

		valueArgs = append(valueArgs, tripData["route_id"], tripData["service_id"], tripData["trip_id"], tripData["trip_headsign"], tripData["trip_short_name"], directionID, tripData["block_id"], tripData["shape_id"], wheelchairAccessible, bikesAllowed)

		if len(valueStrings) == batchSize {
			stmt := fmt.Sprintf(`
				INSERT INTO trips (route_id, service_id, id, trip_headsign, trip_short_name, direction_id, block_id, shape_id, wheelchair_accessible, bikes_allowed)
				VALUES %s
				ON CONFLICT (id) DO UPDATE SET
					route_id = EXCLUDED.route_id,
					service_id = EXCLUDED.service_id,
					trip_headsign = EXCLUDED.trip_headsign,
					trip_short_name = EXCLUDED.trip_short_name,
					direction_id = EXCLUDED.direction_id,
					block_id = EXCLUDED.block_id,
					shape_id = EXCLUDED.shape_id,
					wheelchair_accessible = EXCLUDED.wheelchair_accessible,
					bikes_allowed = EXCLUDED.bikes_allowed
			`, strings.Join(valueStrings, ","))
			_, err = tx.Exec(context.Background(), stmt, valueArgs...)
			if err != nil {
				return err
			}
			valueStrings = make([]string, 0, batchSize)
			valueArgs = make([]interface{}, 0, batchSize*10)
			i = 0
		}
	}

	if len(valueStrings) > 0 {
		stmt := fmt.Sprintf(`
			INSERT INTO trips (route_id, service_id, id, trip_headsign, trip_short_name, direction_id, block_id, shape_id, wheelchair_accessible, bikes_allowed)
			VALUES %s
			ON CONFLICT (id) DO UPDATE SET
				route_id = EXCLUDED.route_id,
				service_id = EXCLUDED.service_id,
				trip_headsign = EXCLUDED.trip_headsign,
				trip_short_name = EXCLUDED.trip_short_name,
				direction_id = EXCLUDED.direction_id,
				block_id = EXCLUDED.block_id,
				shape_id = EXCLUDED.shape_id,
				wheelchair_accessible = EXCLUDED.wheelchair_accessible,
				bikes_allowed = EXCLUDED.bikes_allowed
		`, strings.Join(valueStrings, ","))
		_, err = tx.Exec(context.Background(), stmt, valueArgs...)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Finished processing trips.txt")
	return nil
}

// parseSeconds converts HH:MM:SS (allowing >24h) into seconds since midnight. Missing or malformed returns -1.
func parseSeconds(t string) int {
    if t == "" {
        return -1
    }
    parts := strings.Split(t, ":")
    if len(parts) != 3 {
        return -1
    }
    h, _ := strconv.Atoi(parts[0])
    m, _ := strconv.Atoi(parts[1])
    s, _ := strconv.Atoi(parts[2])
    return h*3600 + m*60 + s
}


	

func (s *Service) processStopTimes(gtfsPath string) error {
	log.Println("Processing stop_times.txt with COPY...")

	ctx := context.Background()
	dsn := os.Getenv("POSTGRES_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Truncate staging table first
	if _, err := pool.Exec(ctx, "TRUNCATE staging_stop_times"); err != nil {
		return err
	}

	file, err := os.Open(filepath.Join(gtfsPath, "stop_times.txt"))
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := func(name string) int {
		for i, h := range header {
			if h == name {
				return i
			}
		}
		return -1
	}
	iTrip := idx("trip_id")
	iArr := idx("arrival_time")
	iDep := idx("departure_time")
	iStop := idx("stop_id")
	iSeq := idx("stop_sequence")
	iHead := idx("stop_headsign")
	iPick := idx("pickup_type")
	iDrop := idx("drop_off_type")
	iDist := idx("shape_dist_traveled")
	iTP := idx("timepoint")

	cols := []string{"trip_id", "arrival_sec", "departure_sec", "stop_id", "stop_sequence", "stop_headsign", "pickup_type", "drop_off_type", "shape_dist_traveled", "timepoint"}

	buffer := make([][]interface{}, 0, 50000)
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		seq, _ := strconv.Atoi(rec[iSeq])
		pick, _ := strconv.Atoi(rec[iPick])
		drop, _ := strconv.Atoi(rec[iDrop])
		dist, _ := strconv.ParseFloat(rec[iDist], 64)
		tp, _ := strconv.Atoi(rec[iTP])

		row := []interface{}{rec[iTrip], parseSeconds(rec[iArr]), parseSeconds(rec[iDep]), rec[iStop], seq, rec[iHead], pick, drop, dist, tp}
		buffer = append(buffer, row)

		if len(buffer) == cap(buffer) {
			if _, err := pool.CopyFrom(ctx, pgx.Identifier{"staging_stop_times"}, cols, pgx.CopyFromRows(buffer)); err != nil {
				return err
			}
			buffer = buffer[:0]
		}
	}
	if len(buffer) > 0 {
		if _, err := pool.CopyFrom(ctx, pgx.Identifier{"staging_stop_times"}, cols, pgx.CopyFromRows(buffer)); err != nil {
			return err
		}
	}

	// Merge into final table
	mergeSQL := `INSERT INTO stop_times (trip_id, arrival_sec, departure_sec, stop_id, stop_sequence, stop_headsign, pickup_type, drop_off_type, shape_dist_traveled, timepoint)
	SELECT trip_id, arrival_sec, departure_sec, stop_id, stop_sequence, stop_headsign, pickup_type, drop_off_type, shape_dist_traveled, timepoint FROM staging_stop_times
	ON CONFLICT (trip_id, stop_sequence) DO UPDATE SET
		arrival_sec = EXCLUDED.arrival_sec,
		departure_sec = EXCLUDED.departure_sec,
		stop_id = EXCLUDED.stop_id,
		stop_headsign = EXCLUDED.stop_headsign,
		pickup_type = EXCLUDED.pickup_type,
		drop_off_type = EXCLUDED.drop_off_type,
		shape_dist_traveled = EXCLUDED.shape_dist_traveled,
		timepoint = EXCLUDED.timepoint`

	if _, err := pool.Exec(ctx, mergeSQL); err != nil {
		return err
	}

	log.Println("Creating post-load indexes on stop_times...")
	_, err = pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS stop_times_stop_departure_idx ON stop_times (stop_id, departure_sec)")
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	_, err = pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS stop_times_trip_sequence_idx ON stop_times (trip_id, stop_sequence)")
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	_, err = pool.Exec(ctx, "ANALYZE stop_times")
	if err != nil {
		return fmt.Errorf("failed to analyze table: %w", err)
	}

	log.Println("Finished processing stop_times.txt via COPY and indexing")
	return nil
}

// extractGTFS extracts the GTFS zip file to a temporary directory.
func (s *Service) extractGTFS(zipPath string) (string, error) {
	log.Printf("Extracting GTFS data from %s", zipPath)

	extractPath, err := os.MkdirTemp("", "gtfs-extract-*")
	if err != nil {
		return "", err
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(extractPath, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file handles explicitly.
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", err
		}
	}

	log.Printf("GTFS data extracted to %s", extractPath)
	return extractPath, nil
}

// processStopTimesFast uses pgx.CopyFrom into staging table then merges.
func (s *Service) processStopTimesFast(gtfsPath string) error {
	ctx := context.Background()
	log.Println("Processing stop_times.txt via COPY ...")

	if _, err := s.pool.Exec(ctx, "TRUNCATE staging_stop_times"); err != nil {
		return err
	}

	file, err := os.Open(filepath.Join(gtfsPath, "stop_times.txt"))
	if err != nil {
		return fmt.Errorf("failed to open stop_times.txt: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := func(name string) int {
		for i, h := range header {
			if h == name {
				return i
			}
		}
		return -1
	}
	iTrip := idx("trip_id")
	iArr := idx("arrival_time")
	iDep := idx("departure_time")
	iStop := idx("stop_id")
	iSeq := idx("stop_sequence")
	iHead := idx("stop_headsign")
	iPick := idx("pickup_type")
	iDrop := idx("drop_off_type")
	iDist := idx("shape_dist_traveled")
	iTP := idx("timepoint")

	copyCols := []string{"trip_id", "arrival_sec", "departure_sec", "stop_id", "stop_sequence", "stop_headsign", "pickup_type", "drop_off_type", "shape_dist_traveled", "timepoint"}
	buffer := make([][]interface{}, 0, 50000)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		seq, _ := strconv.Atoi(rec[iSeq])
		pick, _ := strconv.Atoi(rec[iPick])
		drop, _ := strconv.Atoi(rec[iDrop])
		dist, _ := strconv.ParseFloat(rec[iDist], 64)
		tp, _ := strconv.Atoi(rec[iTP])

		buffer = append(buffer, []interface{}{
			rec[iTrip],
			parseSeconds(rec[iArr]),
			parseSeconds(rec[iDep]),
			rec[iStop],
			seq,
			rec[iHead],
			pick,
			drop,
			dist,
			tp,
		})

		if len(buffer) == 50000 {
			if _, err = s.pool.CopyFrom(ctx, pgx.Identifier{"staging_stop_times"}, copyCols, pgx.CopyFromRows(buffer)); err != nil {
				return err
			}
			buffer = buffer[:0]
		}
	}
	if len(buffer) > 0 {
		if _, err = s.pool.CopyFrom(ctx, pgx.Identifier{"staging_stop_times"}, copyCols, pgx.CopyFromRows(buffer)); err != nil {
			return err
		}
	}

	mergeSQL := `INSERT INTO stop_times AS t (
  trip_id, arrival_sec, departure_sec, stop_id, stop_sequence, stop_headsign, pickup_type, drop_off_type, shape_dist_traveled, timepoint)
  SELECT trip_id, arrival_sec, departure_sec, stop_id, stop_sequence, stop_headsign, pickup_type, drop_off_type, shape_dist_traveled, timepoint FROM staging_stop_times
  ON CONFLICT (trip_id, stop_sequence) DO UPDATE SET
    arrival_sec = EXCLUDED.arrival_sec,
    departure_sec = EXCLUDED.departure_sec,
    stop_id = EXCLUDED.stop_id,
    stop_headsign = EXCLUDED.stop_headsign,
    pickup_type = EXCLUDED.pickup_type,
    drop_off_type = EXCLUDED.drop_off_type,
    shape_dist_traveled = EXCLUDED.shape_dist_traveled,
    timepoint = EXCLUDED.timepoint`

	if _, err := s.pool.Exec(ctx, mergeSQL); err != nil {
		return err
	}

	log.Println("Finished processing stop_times.txt via COPY")
	return nil
}