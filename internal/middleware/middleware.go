package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gopro/internal/token"
    "github.com/redis/go-redis/v9"
    "context"
    "fmt"
)

// CORS middleware: allows all origins, methods, and headers (customize as needed)
func CORSMiddleware() fiber.Handler {
    return cors.New(cors.Config{
        AllowOrigins: "*",
        AllowHeaders: "Origin, Content-Type, Accept, Authorization",
        AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
    })
}

// JWT middleware: checks for Authorization header and validates JWT
func JWTMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        tokenString := c.Get("Authorization")
        if tokenString == "" {
            return fiber.ErrUnauthorized
        }
        userID, userAttr, err := token.ValidateAndExtract(tokenString)
        if err != nil {
            return fiber.ErrUnauthorized
        }
        c.Locals("user_id", userID)
        c.Locals("user_attr", userAttr)
        return c.Next()
    }
}

// Example: Placeholder for another middleware
func ExampleMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Do something before
        err := c.Next()
        // Do something after
        return err
    }
}




func RequireOTP(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		phone := c.Get("X-Phone") // e.g., passed in header
		if phone == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Phone number required")
		}
		key := fmt.Sprintf("otp:verified:%s", phone)
		exists, err := rdb.Exists(context.TODO(), key).Result()
		if err != nil || exists == 0 {
			return fiber.NewError(fiber.StatusUnauthorized, "OTP verification required")
		}
		return c.Next()
	}
}
