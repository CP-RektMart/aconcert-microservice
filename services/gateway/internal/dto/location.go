package dto

type CreateLocationZoneRequest struct {
	ZoneNumber int      `json:"zoneNumber" validate:"required"`
	ZoneName   string   `json:"zoneName" validate:"required"`
	Capacity   int      `json:"capacity" validate:"required"`
	Reserved   bool     `json:"reserved" validate:"required"`
	Price      float64  `json:"price" validate:"required"`
	Exclusive  []string `json:"exclusive" validate:"required"`
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
	ZoneNumber int      `json:"zoneNumber" validate:"required"`
	ZoneName   string   `json:"zoneName" validate:"required"`
	Capacity   int      `json:"capacity" validate:"required"`
	Reserved   bool     `json:"reserved" validate:"required"`
	Price      float64  `json:"price" validate:"required"`
	Exclusive  []string `json:"exclusive" validate:"required"`
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
	ID            string         `json:"id"`
	VenueName     string         `json:"venueName"`
	City          string         `json:"city"`
	StateProvince string         `json:"stateProvince"`
	Country       string         `json:"country"`
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
	Zones         []ZoneResponse `json:"zones"`
}

type ZoneResponse struct {
	ZoneNumber int      `json:"zoneNumber"`
	ZoneName   string   `json:"zoneName"`
	Capacity   int      `json:"capacity"`
	Reserved   bool     `json:"reserved"`
	Price      float64  `json:"price"`
	Exclusive  []string `json:"exclusive"`
}
