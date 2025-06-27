package models

import (
	"time"
	"github.com/google/uuid"
)

// OTPRequest represents the request payload to send OTP to an identifier (email or phone).
type OTPRequest struct {
	Identifier string `json:"identifier" validate:"required"` // Can be email or phone number
}

// VerifyRequest represents the payload to verify OTP for a given identifier.
type VerifyRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	OTP        string `json:"otp" validate:"required,len=6"`
}

// OTPResponse represents a generic response after sending OTP.
type OTPResponse struct {
	Message string `json:"message"`
}

// AuthResponse represents a response with a JWT token after successful OTP verification.
type AuthResponse struct {
	AccessToken string `json:"access_token"`
}


type Poll struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
    WebsiteID   string
    Title       string
    Description string
    CreatedBy   string
    CreatedAt   time.Time
    Options     []PollOption `gorm:"foreignKey:PollID"`
    ShareableLink string `gorm:"type:varchar(255);unique"`
}

type PollOption struct {
    ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
    PollID  uuid.UUID
    OptionText string
}

type User struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
    Identifier string    `gorm:"unique"`
}

type Vote struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    PollID    uuid.UUID
    OptionID  uuid.UUID
    UserID    uuid.UUID
    VotedAt   time.Time
}