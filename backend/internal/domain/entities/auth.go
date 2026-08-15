package entities

type GoogleCallbackParameters struct {
	Code  string `query:"code" validate:"required"`
	State string `query:"state" validate:"required"`
}

type GoogleIdentity struct {
	Sub        string `json:"-"`
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}
