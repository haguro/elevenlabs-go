package elevenlabs

import (
	"fmt"
	"strings"
)

// APIError represents an error response from the API.
//
// At this stage, any error that is not a ValidationError is returned in this format.
type APIError struct {
	Detail APIErrorDetail `json:"detail"`
}

// APIErrorDetail contains detailed information about an APIError.
type APIErrorDetail struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	AdditionalInfo string `json:"additional_info,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error - %s", e.Detail.Message)
}

// ValidationError represents a request validation error response from the API.
type ValidationError struct {
	Detail *[]ValidationErrorDetailItem `json:"detail"`
}

type ValidationErrorDetailItem struct {
	Loc  []ValidationErrorDetailLocItem `json:"loc"`
	Msg  string                         `json:"msg"`
	Type string                         `json:"type"`
}

type ValidationErrorDetailLocItem string

func (i *ValidationErrorDetailLocItem) UnmarshalJSON(b []byte) error {
	*i = ValidationErrorDetailLocItem(strings.Trim(string(b), "\""))
	return nil
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", (*e.Detail)[0].Msg)
}
