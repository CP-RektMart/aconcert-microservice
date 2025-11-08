package dto

type CreateLocationZoneRequest struct {
	ZoneNumber   int    `json:"zoneNumber" validate:"required"`
	ZoneName     string `json:"zoneName" validate:"required"`
	Capacity     int    `json:"capacity" validate:"required"`
	SeatsPerRow  int    `json:"seatsPerRow" validate:"required"`
	NumberOfRows int    `json:"numberOfRows" validate:"required"`
}

type CreateLocationRequest struct {
	VenueName     string                      `json:"venueName" validate:"required"`
	City          string                      `json:"city" validate:"required"`
	StateProvince string                      `json:"stateProvince" validate:"required"`
	Country       string                      `json:"country" validate:"required"`
	Latitude      float64                     `json:"latitude" validate:"required"`
	Longitude     float64                     `json:"longitude" validate:"required"`
	Zones         []CreateLocationZoneRequest `json:"zones" validate:"required"`
}

type CreateLocationResponse struct {
	ID string `json:"id"`
}

type GetLocationRequest struct {
	ID string `params:"id" swaggerignore:"true"`
}

type ListLocationsRequest struct{}

type ListLocationsResponse struct {
	List []LocationResponse `json:"list"`
}

type UpdateLocationZoneRequest struct {
	ZoneNumber   int    `json:"zoneNumber" validate:"required"`
	ZoneName     string `json:"zoneName" validate:"required"`
	Capacity     int    `json:"capacity" validate:"required"`
	SeatsPerRow  int    `json:"seatsPerRow" validate:"required"`
	NumberOfRows int    `json:"numberOfRows" validate:"required"`
}

type UpdateLocationRequest struct {
	ID            string                      `params:"id" swaggerignore:"true"`
	VenueName     string                      `json:"venueName" validate:"required"`
	City          string                      `json:"city" validate:"required"`
	StateProvince string                      `json:"stateProvince" validate:"required"`
	Country       string                      `json:"country" validate:"required"`
	Latitude      float64                     `json:"latitude" validate:"required"`
	Longitude     float64                     `json:"longitude" validate:"required"`
	Zones         []UpdateLocationZoneRequest `json:"zones" validate:"required"`
}

type UpdateLocationResponse struct {
	ID string `json:"id"`
}

type DeleteLocationRequest struct {
	ID string `params:"id" swaggerignore:"true"`
}

type LocationResponse struct {
	ID            string         `json:"id" validate:"required"`
	VenueName     string         `json:"venueName" validate:"required"`
	City          string         `json:"city" validate:"required"`
	StateProvince string         `json:"stateProvince" validate:"required"`
	Country       string         `json:"country" validate:"required"`
	Latitude      float64        `json:"latitude" validate:"required"`
	Longitude     float64        `json:"longitude" validate:"required"`
	Zones         []ZoneResponse `json:"zones" validate:"required"`
}

type ZoneResponse struct {
	ZoneNumber   int    `json:"zoneNumber" validate:"required"`
	ZoneName     string `json:"zoneName" validate:"required"`
	Capacity     int    `json:"capacity" validate:"required"`
	SeatsPerRow  int    `json:"seatsPerRow" validate:"required"`
	NumberOfRows int    `json:"numberOfRows" validate:"required"`
}
