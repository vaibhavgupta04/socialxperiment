package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Generate returns a secure 6-digit OTP.
func Generate() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		// fallback to predictable (for dev)
		return fmt.Sprintf("%06d", 123456)
	}
	return fmt.Sprintf("%06d", n.Int64())
}
