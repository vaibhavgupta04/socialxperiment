package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/gopro/internal/config"
	"github.com/gopro/internal/handlers"
	"github.com/gopro/internal/redis"
	"github.com/gopro/internal/token"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gopro/internal/middleware"
	"github.com/gopro/internal/db"
)

func main() {
	cfg := config.LoadEnv()
	token.Init(cfg)
	rdb := redis.InitRedis(cfg)
	pgdb := db.InitPostgres(cfg)

	app := fiber.New(fiber.Config{
		Prefork:       true,
		ServerHeader:  "GoPro",
		BodyLimit:     1 * 1024 * 1024,
	})

	app.Use(middleware.CORSMiddleware())
	app.Use(limiter.New(limiter.Config{
		Max:        5000,
		Expiration: 1 * 60 * 1000000000,
	}))
	app.Post("/send-otp", handlers.SendOTP(rdb))
	app.Post("/verify-otp", handlers.VerifyOTP(rdb))

	// secure := app.Group("/secure", middleware.JWTMiddleware())
	app.Post("/create-poll", handlers.CreatePoll(pgdb))
	app.Post("/polls/:poll_id/vote", handlers.CastVote(pgdb))
	

	app.Get(("polls/:poll_id"), handlers.GetPollData(pgdb))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})



	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}