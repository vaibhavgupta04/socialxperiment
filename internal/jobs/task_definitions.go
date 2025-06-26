package jobs

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"github.com/gopro/internal/mail"
	"github.com/gopro/internal/sms"
)

func HandleEmailTask(ctx context.Context, t *asynq.Task) error {
	var p OTPTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("[EMAIL] Sending OTP to %s", p.Identifier)
	return mail.SendOTP(p.Identifier, p.OTP)
}

func HandleSMSTask(ctx context.Context, t *asynq.Task) error {
	var p OTPTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("[SMS] Sending OTP to %s", p.Identifier)
	return sms.SendOTP(p.Identifier, p.OTP)
}
