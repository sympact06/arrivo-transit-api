package services

import (
	"arrivo-transit-api/internal/cache"
	"arrivo-transit-api/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type OVAPIClient interface {
	GetDepartures(ctx context.Context, stopID string) ([]models.Departure, error)
}

type TransitService struct {
	ovapiClient OVAPIClient
	lruCache    *cache.LRUCache
	redisClient *redis.Client
	db          *pgxpool.Pool
}

func NewTransitService(ovapiClient OVAPIClient, lruCache *cache.LRUCache, redisClient *redis.Client, db *pgxpool.Pool) *TransitService {
	return &TransitService{
		ovapiClient: ovapiClient,
		lruCache:    lruCache,
		redisClient: redisClient,
		db:          db,
	}
}

func (s *TransitService) GetDepartures(ctx context.Context, stopID string) ([]models.Departure, error) {
	// Try LRU cache first
	cacheKey := fmt.Sprintf("departures:%s", stopID)
	if data, found := s.lruCache.Get(ctx, cacheKey); found {
		var departures []models.Departure
		if err := json.Unmarshal(data, &departures); err == nil {
			return departures, nil
		}
	}

	// Try Redis cache
	if s.redisClient != nil {
		data, err := s.redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var departures []models.Departure
			if err := json.Unmarshal(data, &departures); err == nil {
				// Store in LRU cache for faster access
				s.lruCache.Set(ctx, cacheKey, data)
				return departures, nil
			}
		}
	}

	// Fetch from OVAPI
	departures, err := s.ovapiClient.GetDepartures(ctx, stopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get departures from OVAPI: %w", err)
	}

	// Cache the result
	data, _ := json.Marshal(departures)
	s.lruCache.Set(ctx, cacheKey, data)
	if s.redisClient != nil {
		s.redisClient.Set(ctx, cacheKey, data, 30*time.Second)
	}

	return departures, nil
}

// SearchStops searches for stops by name, with optional proximity search
func (s *TransitService) SearchStops(ctx context.Context, query string, lat, lon *float64) ([]models.Stop, error) {
	var stops []models.Stop
	var err error

	if lat != nil && lon != nil {
		// Search with proximity
		query := `
			SELECT stop_id, stop_name, stop_lat, stop_lon,
				   ST_Distance(ST_Point(stop_lon, stop_lat)::geography, ST_Point($3, $2)::geography) as distance
			FROM stops 
			WHERE stop_name ILIKE $1
			ORDER BY distance
			LIMIT 20
		`
		rows, err := s.db.Query(ctx, query, "%"+query+"%", *lat, *lon)
		if err != nil {
			return nil, fmt.Errorf("failed to search stops with proximity: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var stop models.Stop
			var distance float64
			err := rows.Scan(&stop.ID, &stop.Name, &stop.Lat, &stop.Lon, &distance)
			if err != nil {
				return nil, fmt.Errorf("failed to scan stop row: %w", err)
			}
			stop.Distance = distance
			stops = append(stops, stop)
		}
	} else {
		// Regular search without proximity
		query := `
			SELECT stop_id, stop_name, stop_lat, stop_lon
			FROM stops 
			WHERE stop_name ILIKE $1
			ORDER BY stop_name
			LIMIT 20
		`
		rows, err := s.db.Query(ctx, query, "%"+query+"%")
		if err != nil {
			return nil, fmt.Errorf("failed to search stops: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var stop models.Stop
			err := rows.Scan(&stop.ID, &stop.Name, &stop.Lat, &stop.Lon)
			if err != nil {
				return nil, fmt.Errorf("failed to scan stop row: %w", err)
			}
			stops = append(stops, stop)
		}
	}

	return stops, nil
}

// GetVehiclesByRoute fetches all vehicles currently serving a specific route
func (s *TransitService) GetVehiclesByRoute(ctx context.Context, routeID string) (*models.RouteVehicles, error) {
	// Try LRU cache first
	cacheKey := fmt.Sprintf("route_vehicles:%s", routeID)
	if data, found := s.lruCache.Get(ctx, cacheKey); found {
		var routeVehicles models.RouteVehicles
		if err := json.Unmarshal(data, &routeVehicles); err == nil {
			return &routeVehicles, nil
		}
	}

	// Try Redis cache
	if s.redisClient != nil {
		data, err := s.redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var routeVehicles models.RouteVehicles
			if err := json.Unmarshal(data, &routeVehicles); err == nil {
				// Store in LRU cache for faster access
				s.lruCache.Set(ctx, cacheKey, data)
				return &routeVehicles, nil
			}
		}
	}

	// For now, return mock data since we don't have real-time vehicle data yet
	routeVehicles := s.generateMockVehicleData(routeID)

	// Cache the result
	data, _ := json.Marshal(routeVehicles)
	s.lruCache.Set(ctx, cacheKey, data)
	if s.redisClient != nil {
		s.redisClient.Set(ctx, cacheKey, data, 15*time.Second) // Shorter cache for real-time data
	}

	return routeVehicles, nil
}

// generateMockVehicleData creates mock vehicle data for testing
func (s *TransitService) generateMockVehicleData(routeID string) *models.RouteVehicles {
	numVehicles := rand.Intn(5) + 1 // 1-5 vehicles
	vehicles := make([]models.Vehicle, numVehicles)

	for i := 0; i < numVehicles; i++ {
		// Generate random coordinates around Amsterdam area
		lat := 52.3676 + (rand.Float64()-0.5)*0.1
		lon := 4.9041 + (rand.Float64()-0.5)*0.1
		bearing := rand.Float64() * 360
		speed := rand.Float64() * 50
		delay := rand.Intn(600) - 300 // -5 to +5 minutes

		vehicles[i] = models.Vehicle{
			ID:        fmt.Sprintf("vehicle_%s_%d", routeID, i+1),
			RouteID:   routeID,
			TripID:    fmt.Sprintf("trip_%s_%d", routeID, i+1),
			Lat:       lat,
			Lon:       lon,
			Bearing:   &bearing,
			Speed:     &speed,
			Timestamp: time.Now(),
			Delay:     &delay,
			Status:    "IN_TRANSIT",
		}
	}

	return &models.RouteVehicles{
		RouteID:     routeID,
		RouteName:   fmt.Sprintf("Route %s", routeID),
		Vehicles:    vehicles,
		LastUpdated: time.Now(),
	}
}

// GetAllActiveVehicles fetches all currently active vehicles across all routes
func (s *TransitService) GetAllActiveVehicles(ctx context.Context) ([]models.Vehicle, error) {
	// Try cache first
	cacheKey := "all_active_vehicles"
	if data, found := s.lruCache.Get(ctx, cacheKey); found {
		var vehicles []models.Vehicle
		if err := json.Unmarshal(data, &vehicles); err == nil {
			return vehicles, nil
		}
	}

	// For now, generate mock data for multiple routes
	mockRoutes := []string{"1", "2", "5", "13", "17"}
	var allVehicles []models.Vehicle

	for _, routeID := range mockRoutes {
		routeVehicles := s.generateMockVehicleData(routeID)
		allVehicles = append(allVehicles, routeVehicles.Vehicles...)
	}

	// Cache the result
	data, _ := json.Marshal(allVehicles)
	s.lruCache.Set(ctx, cacheKey, data)
	if s.redisClient != nil {
		s.redisClient.Set(ctx, cacheKey, data, 10*time.Second) // Very short cache for all vehicles
	}

	return allVehicles, nil
}

// SearchRoutes searches for routes/lines by name or number
func (s *TransitService) SearchRoutes(ctx context.Context, query string) ([]models.Route, error) {
	var routes []models.Route

	// Search in database
	sqlQuery := `
		SELECT route_id, route_short_name, route_long_name, route_type
		FROM routes 
		WHERE route_short_name ILIKE $1 OR route_long_name ILIKE $1
		ORDER BY route_short_name
		LIMIT 20
	`
	rows, err := s.db.Query(ctx, sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search routes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var route models.Route
		var routeType sql.NullInt32
		err := rows.Scan(&route.ID, &route.ShortName, &route.LongName, &routeType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan route row: %w", err)
		}
		if routeType.Valid {
			route.Type = int(routeType.Int32)
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// GetNearbyStops finds stops within a given radius of a location
func (s *TransitService) GetNearbyStops(ctx context.Context, lat, lon, radius float64) ([]models.Stop, error) {
	var stops []models.Stop

	// Use PostGIS to find nearby stops
	query := `
		SELECT stop_id, stop_name, stop_lat, stop_lon,
			   ST_Distance(ST_Point(stop_lon, stop_lat)::geography, ST_Point($2, $1)::geography) as distance
		FROM stops 
		WHERE ST_DWithin(ST_Point(stop_lon, stop_lat)::geography, ST_Point($2, $1)::geography, $3)
		ORDER BY distance
		LIMIT 50
	`
	rows, err := s.db.Query(ctx, query, lat, lon, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby stops: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stop models.Stop
		var distance float64
		err := rows.Scan(&stop.ID, &stop.Name, &stop.Lat, &stop.Lon, &distance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stop row: %w", err)
		}
		stop.Distance = math.Round(distance*100) / 100 // Round to 2 decimal places
		stops = append(stops, stop)
	}

	return stops, nil
}