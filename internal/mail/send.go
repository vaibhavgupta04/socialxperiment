package mail

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendOTP sends the OTP to the given email address via SMTP.
func SendOTP(to string, otp string) error {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, pass, host)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: Your OTP Code\r\n" +
		"\r\n" +
		"Your OTP is: " + otp + "\r\n")

	addr := fmt.Sprintf("%s:%s", host, port)
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}
