package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/gopro/internal/config"
	"github.com/gopro/internal/jobs"
)

func main() {
	cfg := config.LoadEnv()
	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"default":  6,
			"critical": 3,
			"low":      1,
		},
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc("email:send_otp", jobs.HandleEmailTask)
	mux.HandleFunc("sms:send_otp", jobs.HandleSMSTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run worker server: %v", err)
	}
}