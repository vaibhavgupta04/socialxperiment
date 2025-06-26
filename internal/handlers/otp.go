package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/gopro/internal/jobs"
	"github.com/gopro/internal/models"
	"github.com/gopro/internal/token"
	"gorm.io/gorm"
)

type OTPRequest struct {
	Identifier string `json:"identifier"`
	PollID     string `json:"poll_id"`
}

type VerifyRequest struct {
	Identifier string `json:"identifier"`
	OTP        string `json:"otp"`
}

func SendOTP(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req OTPRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		otp := generateOTP()
		key := fmt.Sprintf("otp:%s", req.Identifier)
		ctx := context.Background()
		rdb.Set(ctx, key, otp, 5*time.Minute)

		client := jobs.NewAsynqClient()
		var task *asynq.Task
		if strings.Contains(req.Identifier, "@") {
			task = jobs.NewEmailTask(req.Identifier, otp)
		} else {
			task = jobs.NewSMSTask(req.Identifier, otp)
		}
		client.Enqueue(task)

		return c.JSON(fiber.Map{"message": "OTP sent successfully", "otp": otp, "identifier": req.Identifier})
	}
}

func VerifyOTP(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req VerifyRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		key := fmt.Sprintf("otp:%s", req.Identifier)
		ctx := context.Background()
		storedOTP, err := rdb.Get(ctx, key).Result()
		if err != nil || storedOTP != req.OTP {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid OTP"})
		}

		rdb.Del(ctx, key)
		tokenStr, err := token.GenerateJWT(req.Identifier)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate token"})
		}

		return c.JSON(fiber.Map{"access_token": tokenStr})
	}
}

// RegisterUser handles user registration and sends OTP
func RegisterUser(rdb *redis.Client, pgdb interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req OTPRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		// Save user to DB using pgdb (GORM)
		db, ok := pgdb.(*gorm.DB)
		if !ok {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "invalid db instance"})
		}
		user := models.User{
			ID:         uuid.New(),
			Identifier: req.Identifier,
		}
		if err := db.FirstOrCreate(&user, models.User{Identifier: req.Identifier}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not save user"})
		}
		// Send OTP (reuse SendOTP logic)
		otp := generateOTP()
		key := fmt.Sprintf("otp:%s", req.Identifier)
		ctx := context.Background()
		rdb.Set(ctx, key, otp, 5*time.Minute)
		client := jobs.NewAsynqClient()
		var task *asynq.Task
		if strings.Contains(req.Identifier, "@") {
			task = jobs.NewEmailTask(req.Identifier, otp)
		} else {
			task = jobs.NewSMSTask(req.Identifier, otp)
		}
		client.Enqueue(task)
		return c.JSON(fiber.Map{"message": "OTP sent for registration", "identifier": req.Identifier})
	}
}

// LoginUser handles login and sends OTP
func LoginUser(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req OTPRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		// Send OTP (reuse SendOTP logic)
		otp := generateOTP()
		key := fmt.Sprintf("otp:%s", req.Identifier)
		ctx := context.Background()
		rdb.Set(ctx, key, otp, 5*time.Minute)
		client := jobs.NewAsynqClient()
		var task *asynq.Task
		if strings.Contains(req.Identifier, "@") {
			task = jobs.NewEmailTask(req.Identifier, otp)
		} else {
			task = jobs.NewSMSTask(req.Identifier, otp)
		}
		client.Enqueue(task)
		// Note: In a real application, you would also check if the user exists in the database
		// and handle any necessary authentication logic here.
		// For simplicity, we are just sending the OTP.
		return c.JSON(fiber.Map{"message": "OTP sent for login", "identifier": req.Identifier})
	}
}

// LogoutUser is a dummy handler for logout (JWT is stateless)
func LogoutUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Invalidate JWT on client side (e.g., remove from storage/cookie)
		return c.JSON(fiber.Map{"message": "Logged out. Please delete JWT token on client side."})
	}
}

func generateOTP() string {
	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
}
