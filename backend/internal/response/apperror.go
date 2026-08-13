package response

// AppError carries a client-facing Status plus the underlying error (for
// logging — never sent to the client). Repository/interactor layers return
// this instead of a bare error so the central error handler always knows
// the right status code and log level, without needing net/http imported
// anywhere below the controller.
type AppError struct {
	Status Status
	Err    error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Status.Message
}

// Unwrap lets errors.As/errors.Is see through to the wrapped cause.
func (e *AppError) Unwrap() error {
	return e.Err
}

func newAppError(status Status, err error) *AppError {
	return &AppError{Status: status, Err: err}
}

func NewBadRequest(err error) *AppError {
	return newAppError(BadRequest, err)
}

func NewUnauthorized(err error) *AppError {
	return newAppError(Unauthorized, err)
}

func NewProblemNotFound(err error) *AppError {
	return newAppError(ProblemNotFound, err)
}

func NewNoProblemToday(err error) *AppError {
	return newAppError(NoProblemToday, err)
}

func NewInternalError(err error) *AppError {
	return newAppError(InternalError, err)
}

func NewDatabaseError(err error) *AppError {
	return newAppError(DatabaseError, err)
}

func NewBadGateway(err error) *AppError {
	return newAppError(BadGateway, err)
}

func NewServiceUnavailable(err error) *AppError {
	return newAppError(ServiceUnavailable, err)
}
