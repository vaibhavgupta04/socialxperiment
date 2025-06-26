package token

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gopro/internal/config"
)

var cfg *config.Config

// Init sets the config for JWT usage
func Init(c *config.Config) {
	log.Println("Initializing JWT")
	cfg = c
}

// GenerateJWT creates a new JWT token for a given identifier (email or phone).
func GenerateJWT(identifier string) (string, error) {
	if cfg == nil {
		return "", jwt.ErrInvalidKeyType // fallback error
	}

	claims := jwt.MapClaims{
		"sub": identifier,
		"exp": time.Now().Add(time.Duration(cfg.JWTTTL) * time.Second).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ValidateAndExtract validates the JWT and returns the user identifier (sub claim) and claims map.
func ValidateAndExtract(tokenString string) (string, map[string]interface{}, error) {
	if cfg == nil {
		return "", nil, jwt.ErrInvalidKeyType
	}

	// Remove "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil, jwt.ErrInvalidKey
	}

	identifier, ok := claims["sub"].(string)
	if !ok {
		return "", nil, jwt.ErrInvalidKey
	}

	// Optionally return all claims as map[string]interface{}
	return identifier, claims, nil
}
