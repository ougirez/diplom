package domain

type BasicResponse struct{}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type AuthTokenResponse struct {
	AuthToken string `json:"auth_token"`
}
