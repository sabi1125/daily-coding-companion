package response

import "net/http"

type Category int

const (
	Expected Category = iota
	Operational
	Unexpected
)

type Status struct {
	Code     int
	Message  string
	Category Category
}

// NOTE: every time a status is added here it should get a matching
// constructor in apperror.go.
var (
	BadRequest         = Status{Code: http.StatusBadRequest, Message: "Bad Request", Category: Expected}
	Unauthorized       = Status{Code: http.StatusUnauthorized, Message: "Unauthorized", Category: Expected}
	ProblemNotFound    = Status{Code: http.StatusNotFound, Message: "Problem not found", Category: Expected}
	NoProblemToday     = Status{Code: http.StatusNotFound, Message: "no problem available today", Category: Expected}
	InternalError      = Status{Code: http.StatusInternalServerError, Message: "internal server error", Category: Unexpected}
	DatabaseError      = Status{Code: http.StatusInternalServerError, Message: "internal server error", Category: Operational}
	BadGateway         = Status{Code: http.StatusBadGateway, Message: "Bad gateway", Category: Operational}
	ServiceUnavailable = Status{Code: http.StatusServiceUnavailable, Message: "Service Unavailable", Category: Operational}
)
