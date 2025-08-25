package models

// Route represents a transit route/line
type Route struct {
	ID              string  `json:"id"`                // Route identifier
	AgencyID        *string `json:"agency_id,omitempty"`  // Agency operating this route
	ShortName       *string `json:"short_name,omitempty"` // Short name (e.g., "1", "A")
	LongName        *string `json:"long_name,omitempty"`  // Full descriptive name
	Description     *string `json:"description,omitempty"` // Route description
	Type            int     `json:"type"`               // Route type (0=tram, 1=subway, 2=rail, 3=bus, etc.)
	URL             *string `json:"url,omitempty"`      // Route URL
	Color           *string `json:"color,omitempty"`    // Route color (hex)
	TextColor       *string `json:"text_color,omitempty"` // Text color (hex)
	Distance        *float64 `json:"distance,omitempty"` // Distance for location-based searches
}

// RouteType constants for different transit modes
const (
	RouteTypeTram      = 0
	RouteTypeSubway    = 1
	RouteTypeRail      = 2
	RouteTypeBus       = 3
	RouteTypeFerry     = 4
	RouteTypeCableTram = 5
	RouteTypeAerialLift = 6
	RouteTypeFunicular = 7
	RouteTypeTrolleybus = 11
	RouteTypeMonorail  = 12
)

// GetRouteTypeName returns a human-readable name for the route type
func (r *Route) GetRouteTypeName() string {
	switch r.Type {
	case RouteTypeTram:
		return "Tram"
	case RouteTypeSubway:
		return "Subway"
	case RouteTypeRail:
		return "Rail"
	case RouteTypeBus:
		return "Bus"
	case RouteTypeFerry:
		return "Ferry"
	case RouteTypeCableTram:
		return "Cable Tram"
	case RouteTypeAerialLift:
		return "Aerial Lift"
	case RouteTypeFunicular:
		return "Funicular"
	case RouteTypeTrolleybus:
		return "Trolleybus"
	case RouteTypeMonorail:
		return "Monorail"
	default:
		return "Unknown"
	}
}