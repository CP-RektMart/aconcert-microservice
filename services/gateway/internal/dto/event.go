package dto

type ListEventsRequest struct {
	Query  string `query:"query"`
	SortBy string `query:"sortBy"`
	Order  string `query:"order"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

type GetEventRequest struct {
	ID string `params:"id"`
}

type CreateEventRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	LocationID  string   `json:"locationId" validate:"required"`
	Artist      []string `json:"artist" validate:"required"`
	EventDate   string   `json:"eventDate" validate:"required"`
	Thumbnail   string   `json:"thumbnail" validate:"required"`
	Images      []string `json:"images" validate:"required"`
}

type CreateEventResponse struct {
	ID string `json:"id" validate:"required"`
}

type UpdateEventRequest struct {
	ID          string   `params:"id" swaggerignore:"true"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	LocationID  string   `json:"locationId"`
	Artist      []string `json:"artist"`
	EventDate   string   `json:"eventDate"`
	Thumbnail   string   `json:"thumbnail"`
	Images      []string `json:"images"`
}

type UpdateEventResponse struct {
	ID string `json:"id" validate:"required"`
}

type DeleteEventRequest struct {
	ID string `params:"id" swaggerignore:"true"`
}

type EventResponse struct {
	ID          string   `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	LocationID  string   `json:"locationId" validate:"required"`
	Artist      []string `json:"artist" validate:"required"`
	EventDate   string   `json:"eventDate" validate:"required"`
	Thumbnail   string   `json:"thumbnail" validate:"required"`
	Images      []string `json:"images" validate:"required"`
	CreatedAt   string   `json:"createdAt" validate:"required"`
	UpdatedAt   string   `json:"updatedAt" validate:"required"`
}

type EventListResponse struct {
	List []EventResponse `json:"list" validate:"required"`
}

type EventZoneResponse struct {
	ID          string  `json:"id" validate:"required"`
	EventID     string  `json:"eventId" validate:"required"`
	LocationID  string  `json:"locationId" validate:"required"`
	ZoneNumber  int     `json:"zoneNumber" validate:"required"`
	Price       float64 `json:"price" validate:"required"`
	Color       string  `json:"color" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	IsSoldOut   bool    `json:"isSoldOut" validate:"required"`
}

type GetEventZoneByEventIDRequest struct {
	EventID string `params:"id" swaggerignore:"true"`
}

type EventZoneListResponse struct {
	List []EventZoneResponse `json:"list" validate:"required"`
}

type CreateEventZoneRequest struct {
	EventID     string  `params:"id" validate:"required"`
	LocationID  string  `json:"locationId" validate:"required"`
	ZoneNumber  int     `json:"zoneNumber" validate:"required"`
	Price       float64 `json:"price" validate:"required"`
	Color       string  `json:"color" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
}

type CreateEventZoneResponse struct {
	ID string `json:"id" validate:"required"`
}

type UpdateEventZoneRequest struct {
	ID          string  `params:"id" swaggerignore:"true"`
	EventID     string  `json:"eventId"`
	LocationID  string  `json:"locationId"`
	ZoneNumber  int     `json:"zoneNumber"`
	Price       float64 `json:"price"`
	Color       string  `json:"color"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	IsSoldOut   bool    `json:"isSoldOut"`
}

type UpdateEventZoneResponse struct {
	ID string `json:"id" validate:"required"`
}

type DeleteEventZoneRequest struct {
	ID string `params:"id" swaggerignore:"true"`
}
