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
	ID string `json:"id" validate:"required"`
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
