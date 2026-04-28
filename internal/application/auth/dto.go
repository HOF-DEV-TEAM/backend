// Package auth provides the authentication application service and related DTOs.
package auth

// LoginRequest is the payload for standard email/password login.
type LoginRequest struct {
	Email            string `json:"email"             validate:"required,email"`
	Password         string `json:"password"          validate:"required"`
	DeviceIdentifier string `json:"device_identifier"`
}

// AdminLoginRequest is the admin-only login payload (no device required).
type AdminLoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthenticateRequest carries both tokens for a session refresh.
type AuthenticateRequest struct {
	Token        string `json:"token"         validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// SessionResponse is returned on successful authentication.
type SessionResponse struct {
	User             UserDTO            `json:"user"`
	Subscription     SubscriptionDTO    `json:"subscription"`
	GlobalParameters GlobalParamsDTO    `json:"global_parameters"`
	Token            string             `json:"token"`
	RefreshToken     string             `json:"refresh_token"`
}

// GlobalParamsDTO carries app-wide feature flags for the client.
type GlobalParamsDTO struct {
	ActivateSubscription bool `json:"activate_subscription"`
}

// UserDTO is the safe user representation embedded in session responses.
type UserDTO struct {
	ID         string   `json:"id"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Email      string   `json:"email"`
	IsVerified uint8    `json:"is_verified"`
	Roles      []string `json:"roles"`
}

// SubscriptionDTO carries minimal subscription state for the client.
type SubscriptionDTO struct {
	Status          int    `json:"status"`
	NextPaymentDate string `json:"next_payment_date,omitempty"`
	PlanName        string `json:"plan_name,omitempty"`
}
