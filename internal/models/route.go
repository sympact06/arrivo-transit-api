package models

// Route represents a transit route/line
type Route struct {
	ID        string `json:"id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
	Type      int    `json:"type"`
}

// Route type constants based on GTFS specification
const (
	RouteTypeTram       = 0  // Tram, Streetcar, Light rail
	RouteTypeSubway     = 1  // Subway, Metro
	RouteTypeRail       = 2  // Rail
	RouteTypeBus        = 3  // Bus
	RouteTypeFerry      = 4  // Ferry
	RouteTypeCableTram  = 5  // Cable tram
	RouteTypeAerialLift = 6  // Aerial lift, suspended cable car
	RouteTypeFunicular  = 7  // Funicular
	RouteTypeTrolleybus = 11 // Trolleybus
	RouteTypeMonorail   = 12 // Monorail
)