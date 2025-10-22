package dto

type HttpResponse[T any] struct {
	Result T `json:"result" validate:"required"`
}

type HttpError struct {
	Error string `json:"error" validate:"required"`
}
