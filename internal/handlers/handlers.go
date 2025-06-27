package handlers

import (
	"time"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gopro/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Poll struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Options     []string  `json:"options"`
	CreatedAt   time.Time `json:"created_at"`
	PublicURL   string    `json:"public_url"`
	CreatedBy   string    `json:"created_by"`
}

func CreatePoll(rdb *redis.Client, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		userIDStr := "00000000-0000-0000-0000-000000000000"
		if userID != nil {
			if s, ok := userID.(string); ok {
				userIDStr = s
			}
		}

		var req struct {
			PollName    string   `json:"poll_name"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Options     []string `json:"options"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
		}
		if len(req.Options) < 2 {
			return fiber.NewError(fiber.StatusBadRequest, "At least 2 options required")
		}

		pollID := uuid.New()
		poll := models.Poll{
			ID:            pollID,
			WebsiteID:     req.PollName,
			Title:         req.Title,
			Description:   req.Description,
			CreatedBy:     userIDStr,
			CreatedAt:     time.Now(),
			ShareableLink: c.BaseURL() + "/poll/" + pollID.String(),
		}
		if err := db.Create(&poll).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create poll")
		}

		for _, opt := range req.Options {
			option := models.PollOption{
				ID:         uuid.New(),
				PollID:     pollID,
				OptionText: opt,
			}
			db.Create(&option)
		}

		publicURL := c.BaseURL() + "/poll/" + pollID.String()
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"poll_id":     pollID,
			"public_url":  publicURL,
			"title":       poll.Title,
			"created_at":  poll.CreatedAt,
			"description": poll.Description,
			"options":     req.Options,
			"created_by":  userIDStr,
			"poll_name":   req.PollName,
		})
	}
}

type UserVoteInfo struct {
	PhoneNumber string `json:"phone_number"`
	OptionText  string `json:"option_text"`
}

func GetPoll(rdb *redis.Client, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pollIDStr := c.Params("poll_id")
		if pollIDStr == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing poll_id")
		}

		pollID, err := uuid.Parse(pollIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid poll_id")
		}

		var poll models.Poll
		if err := db.Where("id = ?", pollID).First(&poll).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fiber.NewError(fiber.StatusNotFound, "Poll not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve poll")
		}

		publicURL := c.BaseURL() + "/poll/" + poll.ID.String()
		// Fetch options for the poll
		var options []string
		if err := db.Model(&models.PollOption{}).Where("poll_id = ?", poll.ID).Pluck("option_text", &options).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch poll options")
		}

		return c.JSON(Poll{
			ID:          poll.ID.String(),
			Title:       poll.Title,
			Description: poll.Description,
			Options:     options,
			CreatedAt:   poll.CreatedAt,
			PublicURL:   publicURL,
			CreatedBy:   poll.CreatedBy,
		})
	}
}

func GetPollData(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return fiber.ErrUnauthorized
		}
		pollID := c.Params("poll_id")
		if pollID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing poll_id")
		}

		var results []UserVoteInfo
		err := db.Table("votes").
			Select("users.identifier as phone_number, poll_options.option_text").
			Joins("JOIN users ON users.id = votes.user_id").
			Joins("JOIN poll_options ON poll_options.id = votes.option_id").
			Where("votes.poll_id = ?", pollID).
			Scan(&results).Error
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch votes")
		}

		return c.JSON(results)
	}
}

func CastPoll(rdb *redis.Client, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return fiber.ErrUnauthorized
		}
		userIDStr, ok := userID.(string)
		if !ok {
			return fiber.ErrUnauthorized
		}

		pollIDStr := c.Params("poll_id")
		if pollIDStr == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing poll_id")
		}
		pollID, err := uuid.Parse(pollIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid poll_id")
		}

		var req struct {
			OptionID string `json:"option_id"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
		}
		optionID, err := uuid.Parse(req.OptionID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid option_id")
		}

		// Ensure the option belongs to the poll
		var option models.PollOption
		if err := db.Where("id = ? AND poll_id = ?", optionID, pollID).First(&option).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Option does not belong to poll")
		}

		// Get user UUID
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		vote := models.Vote{
			ID:       uuid.New(),
			PollID:   pollID,
			OptionID: optionID,
			UserID:   userUUID,
			VotedAt:  time.Now(),
		}
		if err := db.Create(&vote).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to cast vote")
		}

		return c.JSON(fiber.Map{"message": "Vote cast successfully"})
	}
}

func RequestOTP(rdb *redis.Client, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		var req struct {
			Phone string `json:"phone"`
		}
		if err := c.BodyParser(&req); err != nil || req.Phone == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid phone number")
		}

		// Load credentials
		apiKey := os.Getenv("OTP_API_KEY")
		apiToken := os.Getenv("OTP_API_TOKEN")
		if apiKey == "" || apiToken == "" {
			return fiber.NewError(fiber.StatusInternalServerError, "OTP credentials not set")
		}

		// Build form data
		
		form := url.Values{}
		form.Set("channel", "sms")
		form.Set("phone_sms", req.Phone)
		form.Set("callback_url", "https://2fba-223-233-86-2.ngrok-free.app/auth/callback")
		form.Set("success_redirect_url", "https://2fba-223-233-86-2.ngrok-free.app/success")
		form.Set("fail_redirect_url", "https://2fba-223-233-86-2.ngrok-free.app/failure")


		form.Set("metadata", `{"phone":"`+req.Phone+`"}`)

		// Make HTTP request to OTP.dev
		reqURL := "https://otp.dev/api/verify/"
		client := &http.Client{}
		httpReq, _ := http.NewRequest("POST", reqURL, strings.NewReader(form.Encode()))
		httpReq.SetBasicAuth(apiKey, apiToken)
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		fmt.Println("Requesting OTP for phone:", req.Phone, httpReq)
		resp, err := client.Do(httpReq)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to contact OTP service")
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Invalid OTP service response")
		}

		// Return the OTP link and info
		return c.JSON(result)
	}
}

func OTPCallback(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload struct {
			OtpID      string `json:"otp_id"`
			AuthStatus string `json:"auth_status"`
			Phone      string `json:"phone_sms"`
			OtpSecret  string `json:"otp_secret"`
		}
		if err := c.BodyParser(&payload); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid callback payload")
		}
		if payload.AuthStatus != "verified" {
			return c.SendStatus(fiber.StatusOK) // Skip if not verified
		}
		// Save session
		sessionKey := fmt.Sprintf("otp:verified:%s", payload.Phone)
		err := rdb.Set(context.TODO(), sessionKey, "1", 10*time.Minute).Err()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to store session")
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func OTPSuccess(rdb *redis.Client, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "OTP verified successfully. You may proceed.",
		})
	}
}

func OTPFailure(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "failure",
			"message": "OTP verification failed. Please try again.",
		})
	}
}
