package jobs

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeEmailOTP = "email:send_otp"
	TypeSMSOTP   = "sms:send_otp"
)

type OTPTaskPayload struct {
	Identifier string `json:"identifier"`
	OTP        string `json:"otp"`
}

// NewEmailTask creates a new Asynq task to send OTP via email.
func NewEmailTask(identifier, otp string) *asynq.Task {
	payload, _ := json.Marshal(OTPTaskPayload{Identifier: identifier, OTP: otp})
	return asynq.NewTask(TypeEmailOTP, payload)
}

// NewSMSTask creates a new Asynq task to send OTP via SMS.
func NewSMSTask(identifier, otp string) *asynq.Task {
	payload, _ := json.Marshal(OTPTaskPayload{Identifier: identifier, OTP: otp})
	return asynq.NewTask(TypeSMSOTP, payload)
}

// NewAsynqClient initializes and returns an Asynq client.
func NewAsynqClient() *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: "redis:6379"})
}
