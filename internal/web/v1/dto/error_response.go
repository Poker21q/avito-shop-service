package dto

type ErrorResponse struct {

	// Сообщение об ошибке, описывающее проблему.
	Errors string `json:"errors,omitempty"`
}
