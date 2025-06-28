package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/gopro/internal/config"
	"github.com/gopro/internal/handlers"
	"github.com/gopro/internal/redis"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gopro/internal/middleware"
	"github.com/gopro/internal/db"
	"github.com/gofiber/jwt/v3"
	"os"
)

func main() {
	cfg := config.LoadEnv()
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

	app.Post("/auth/request", handlers.RequestOTP(rdb, pgdb))
	app.Post("/auth/callback", handlers.OTPCallback())

	
	secure := app.Group("/", jwtware.New(jwtware.Config{
		SigningKey:   []byte(os.Getenv("JWT_SECRET")),
		ErrorHandler: fiber.DefaultErrorHandler, // Custom error handler for unauthorized access
	}))
	secure.Post("/create", handlers.CreatePoll(rdb, pgdb))
	secure.Post("/poll/:poll_id", handlers.GetPoll(rdb, pgdb))
	secure.Post("/vote/:poll_id", handlers.CastPoll(rdb, pgdb))
	

	app.Get(("polldata/:poll_id"), handlers.GetPollData(pgdb))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})



	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}