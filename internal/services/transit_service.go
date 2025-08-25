package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"arrivo-transit-api/internal/cache"
	"arrivo-transit-api/internal/models"
	"arrivo-transit-api/internal/ovapi"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	redisCacheDuration = 5 * time.Minute
)

type TransitService struct {
	ovapiClient *ovapi.Client
	lruCache    *cache.LRUCache
	redisClient *redis.Client
	db          *pgxpool.Pool
}

func NewTransitService(ovapiClient *ovapi.Client, lruCache *cache.LRUCache, redisClient *redis.Client, db *pgxpool.Pool) *TransitService {
	return &TransitService{
		ovapiClient: ovapiClient,
		lruCache:    lruCache,
		redisClient: redisClient,
		db:          db,
	}
}

func (s *TransitService) GetDepartures(ctx context.Context, stopID string) ([]models.Departure, error) {
	cacheKey := fmt.Sprintf("departures:%s", stopID)

	// 1. Try in-memory LRU cache
	if cachedData, ok := s.lruCache.Get(ctx, cacheKey); ok {
		var departures []models.Departure
		if err := json.Unmarshal(cachedData, &departures); err == nil {
			log.Printf("CACHE HIT (LRU): %s", cacheKey)
			return departures, nil
		}
	}

	// 2. Try Redis cache
	if cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		var departures []models.Departure
		if err := json.Unmarshal(cachedData, &departures); err == nil {
			log.Printf("CACHE HIT (Redis): %s", cacheKey)
			// Store in LRU for future requests
			s.lruCache.Set(ctx, cacheKey, cachedData)
			return departures, nil
		}
	}

	log.Printf("CACHE MISS: %s", cacheKey)
	
	// TODO: Replace with real OVapi integration when service is available
	// For now, return mock data to test the API structure
	departures := []models.Departure{
		{
			Line:        "1",
			Destination: "Centraal Station",
			Departure:   time.Now().Add(3 * time.Minute),
		},
		{
			Line:        "5",
			Destination: "Amstelveen Centrum",
			Departure:   time.Now().Add(7 * time.Minute),
		},
		{
			Line:        "12",
			Destination: "Station Sloterdijk",
			Departure:   time.Now().Add(12 * time.Minute),
		},
	}

	// 4. Store in caches
	if marshaledData, err := json.Marshal(departures); err == nil {
		s.redisClient.Set(ctx, cacheKey, marshaledData, redisCacheDuration)
		s.lruCache.Set(ctx, cacheKey, marshaledData)
	}

	return departures, nil
}

func (s *TransitService) SearchStops(ctx context.Context, query string, lat, lon *float64) ([]models.Stop, error) {
	var rows pgx.Rows
	var err error

	baseQuery := `
		SELECT stop_id, stop_name, stop_lat, stop_lon %s
		FROM stops 
		WHERE stop_name ILIKE $1 
		%s 
		LIMIT 20;`

	if lat != nil && lon != nil {
		// Query with distance calculation and ordering
		distanceField := ", (6371000 * acos(cos(radians($2)) * cos(radians(stop_lat)) * cos(radians(stop_lon) - radians($3)) + sin(radians($2)) * sin(radians(stop_lat)))) AS distance"
		orderBy := "ORDER BY distance"
		searchQuery := fmt.Sprintf(baseQuery, distanceField, orderBy)
		rows, err = s.db.Query(ctx, searchQuery, "%"+query+"%", *lat, *lon)
	} else {
		// Original query without location
		searchQuery := fmt.Sprintf(baseQuery, "", "")
		rows, err = s.db.Query(ctx, searchQuery, "%"+query+"%")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query stops: %w", err)
	}
	defer rows.Close()

	var stops []models.Stop
	for rows.Next() {
		var stop models.Stop
		var err error
		if lat != nil && lon != nil {
			err = rows.Scan(&stop.ID, &stop.Name, &stop.Lat, &stop.Lon, &stop.Distance)
		} else {
			err = rows.Scan(&stop.ID, &stop.Name, &stop.Lat, &stop.Lon)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan stop: %w", err)
		}
		stops = append(stops, stop)
	}

	return stops, nil
}

// GetVehiclesByRoute fetches all vehicles currently serving a specific route
func (s *TransitService) GetVehiclesByRoute(ctx context.Context, routeID string) (*models.RouteVehicles, error) {
	cacheKey := fmt.Sprintf("vehicles:route:%s", routeID)

	// Try LRU cache first
	if cachedData, found := s.lruCache.Get(ctx, cacheKey); found {
		log.Printf("CACHE HIT (LRU): %s", cacheKey)
		var routeVehicles models.RouteVehicles
		if err := json.Unmarshal(cachedData, &routeVehicles); err == nil {
			return &routeVehicles, nil
		}
	}

	// Try Redis cache
	if cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		log.Printf("CACHE HIT (Redis): %s", cacheKey)
		var routeVehicles models.RouteVehicles
		if err := json.Unmarshal(cachedData, &routeVehicles); err == nil {
			// Store in LRU cache for faster access
			s.lruCache.Set(ctx, cacheKey, cachedData)
			return &routeVehicles, nil
		}
	}

	// For now, return mock data since OVapi vehicle tracking is not implemented
	// In production, this would fetch real-time vehicle positions from OVapi
	routeVehicles := &models.RouteVehicles{
		RouteID:   routeID,
		RouteName: fmt.Sprintf("Route %s", routeID),
		Vehicles: []models.Vehicle{
			{
				ID:        fmt.Sprintf("VEH_%s_001", routeID),
				RouteID:   routeID,
				TripID:    fmt.Sprintf("TRIP_%s_001", routeID),
				Lat:       52.3676 + (float64(len(routeID)%10) * 0.001), // Mock coordinates around Amsterdam
				Lon:       4.9041 + (float64(len(routeID)%10) * 0.001),
				Bearing:   func() *float64 { b := 45.0; return &b }(),
				Speed:     func() *float64 { s := 25.5; return &s }(),
				Timestamp: time.Now(),
				Delay:     func() *int { d := 120; return &d }(), // 2 minutes late
				Status:    "IN_TRANSIT",
				Occupancy: func() *string { o := "MANY_SEATS_AVAILABLE"; return &o }(),
			},
			{
				ID:        fmt.Sprintf("VEH_%s_002", routeID),
				RouteID:   routeID,
				TripID:    fmt.Sprintf("TRIP_%s_002", routeID),
				Lat:       52.3700 + (float64(len(routeID)%7) * 0.001),
				Lon:       4.9000 + (float64(len(routeID)%7) * 0.001),
				Bearing:   func() *float64 { b := 225.0; return &b }(),
				Speed:     func() *float64 { s := 18.2; return &s }(),
				Timestamp: time.Now(),
				Delay:     func() *int { d := -30; return &d }(), // 30 seconds early
				Status:    "STOPPED_AT",
				StopID:    func() *string { s := "2992167"; return &s }(),
				Occupancy: func() *string { o := "FEW_SEATS_AVAILABLE"; return &o }(),
			},
		},
		LastUpdated: time.Now(),
	}

	// Cache the result
	if marshaledData, err := json.Marshal(routeVehicles); err == nil {
		s.redisClient.Set(ctx, cacheKey, marshaledData, redisCacheDuration)
		s.lruCache.Set(ctx, cacheKey, marshaledData)
	}

	return routeVehicles, nil
}

// GetAllActiveVehicles fetches all currently active vehicles across all routes
func (s *TransitService) GetAllActiveVehicles(ctx context.Context) ([]models.Vehicle, error) {
	cacheKey := "vehicles:all:active"

	// Try LRU cache first
	if cachedData, found := s.lruCache.Get(ctx, cacheKey); found {
		log.Printf("CACHE HIT (LRU): %s", cacheKey)
		var vehicles []models.Vehicle
		if err := json.Unmarshal(cachedData, &vehicles); err == nil {
			return vehicles, nil
		}
	}

	// Try Redis cache
	if cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		log.Printf("CACHE HIT (Redis): %s", cacheKey)
		var vehicles []models.Vehicle
		if err := json.Unmarshal(cachedData, &vehicles); err == nil {
			// Store in LRU cache for faster access
			s.lruCache.Set(ctx, cacheKey, cachedData)
			return vehicles, nil
		}
	}

	// For now, return mock data since OVapi vehicle tracking is not implemented
	// In production, this would fetch all active vehicles from OVapi
	vehicles := []models.Vehicle{
		{
			ID:        "VEH_001",
			RouteID:   "1",
			TripID:    "TRIP_001",
			Lat:       52.3676,
			Lon:       4.9041,
			Bearing:   func() *float64 { b := 90.0; return &b }(),
			Speed:     func() *float64 { s := 30.0; return &s }(),
			Timestamp: time.Now(),
			Status:    "IN_TRANSIT",
			Occupancy: func() *string { o := "MANY_SEATS_AVAILABLE"; return &o }(),
		},
		{
			ID:        "VEH_002",
			RouteID:   "2",
			TripID:    "TRIP_002",
			Lat:       52.3700,
			Lon:       4.9000,
			Bearing:   func() *float64 { b := 180.0; return &b }(),
			Speed:     func() *float64 { s := 22.5; return &s }(),
			Timestamp: time.Now(),
			Delay:     func() *int { d := 60; return &d }(),
			Status:    "STOPPED_AT",
			StopID:    func() *string { s := "2992167"; return &s }(),
			Occupancy: func() *string { o := "FEW_SEATS_AVAILABLE"; return &o }(),
		},
		{
			ID:        "VEH_003",
			RouteID:   "5",
			TripID:    "TRIP_003",
			Lat:       52.3650,
			Lon:       4.9100,
			Bearing:   func() *float64 { b := 270.0; return &b }(),
			Speed:     func() *float64 { s := 15.8; return &s }(),
			Timestamp: time.Now(),
			Status:    "IN_TRANSIT",
			Occupancy: func() *string { o := "STANDING_ROOM_ONLY"; return &o }(),
		},
	}

	// Cache the result
	if marshaledData, err := json.Marshal(vehicles); err == nil {
		s.redisClient.Set(ctx, cacheKey, marshaledData, redisCacheDuration)
		s.lruCache.Set(ctx, cacheKey, marshaledData)
	}

	return vehicles, nil
}

// SearchRoutes searches for routes/lines by name or short name
func (s *TransitService) SearchRoutes(ctx context.Context, query string) ([]models.Route, error) {
	cacheKey := fmt.Sprintf("routes:search:%s", query)

	// Try LRU cache first
	if cachedData, found := s.lruCache.Get(ctx, cacheKey); found {
		log.Printf("CACHE HIT (LRU): %s", cacheKey)
		var routes []models.Route
		if err := json.Unmarshal(cachedData, &routes); err == nil {
			return routes, nil
		}
	}

	// Try Redis cache
	if cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		log.Printf("CACHE HIT (Redis): %s", cacheKey)
		var routes []models.Route
		if err := json.Unmarshal(cachedData, &routes); err == nil {
			// Store in LRU cache for faster access
			s.lruCache.Set(ctx, cacheKey, cachedData)
			return routes, nil
		}
	}

	// Query database
	searchQuery := `
		SELECT id, agency_id, route_short_name, route_long_name, route_desc, route_type, route_url, route_color, route_text_color
		FROM routes 
		WHERE route_short_name ILIKE $1 OR route_long_name ILIKE $1
		ORDER BY 
			CASE 
				WHEN route_short_name ILIKE $1 THEN 1
				WHEN route_long_name ILIKE $1 THEN 2
				ELSE 3
			END,
			route_short_name, route_long_name
		LIMIT 20`

	rows, err := s.db.Query(ctx, searchQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query routes: %w", err)
	}
	defer rows.Close()

	var routes []models.Route
	for rows.Next() {
		var route models.Route
		var agencyID, shortName, longName, description, url, color, textColor sql.NullString
		
		err := rows.Scan(
			&route.ID,
			&agencyID,
			&shortName,
			&longName,
			&description,
			&route.Type,
			&url,
			&color,
			&textColor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}

		// Handle nullable fields
		if agencyID.Valid {
			route.AgencyID = &agencyID.String
		}
		if shortName.Valid {
			route.ShortName = &shortName.String
		}
		if longName.Valid {
			route.LongName = &longName.String
		}
		if description.Valid {
			route.Description = &description.String
		}
		if url.Valid {
			route.URL = &url.String
		}
		if color.Valid {
			route.Color = &color.String
		}
		if textColor.Valid {
			route.TextColor = &textColor.String
		}

		routes = append(routes, route)
	}

	// Cache the result
	if marshaledData, err := json.Marshal(routes); err == nil {
		s.redisClient.Set(ctx, cacheKey, marshaledData, redisCacheDuration)
		s.lruCache.Set(ctx, cacheKey, marshaledData)
	}

	return routes, nil
}

func (s *TransitService) GetNearbyStops(ctx context.Context, lat, lon, radius float64) ([]models.Stop, error) {
	query := `
		SELECT stop_id, stop_name, stop_lat, stop_lon, ( 
			6371000 * acos( 
				cos( radians($1) ) 
				* cos( radians( stop_lat ) ) 
				* cos( radians( stop_lon ) - radians($2) ) 
				+ sin( radians($1) ) 
				* sin( radians( stop_lat ) ) 
			) 
		) AS distance 
		FROM stops 
		WHERE ( 
			6371000 * acos( 
				cos( radians($1) ) 
				* cos( radians( stop_lat ) ) 
				* cos( radians( stop_lon ) - radians($2) ) 
				+ sin( radians($1) ) 
				* sin( radians( stop_lat ) ) 
			) 
		) < $3 
		ORDER BY distance
		LIMIT 50;`

	rows, err := s.db.Query(ctx, query, lat, lon, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to query nearby stops: %w", err)
	}
	defer rows.Close()

	var stops []models.Stop
	for rows.Next() {
		var stop models.Stop
		if err := rows.Scan(&stop.ID, &stop.Name, &stop.Lat, &stop.Lon, &stop.Distance); err != nil {
			return nil, fmt.Errorf("failed to scan stop: %w", err)
		}
		stops = append(stops, stop)
	}

	return stops, nil
}