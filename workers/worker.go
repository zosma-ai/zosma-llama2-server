package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rampenke/zosma-llama-server/tasks"
)

type Config struct {
	RedisAddr     string `envconfig:"REDIS_ADDR" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
	LlamaApiHost  string `envconfig:"LLAMA_API_HOST" required:"true"`
}

var cfg Config

func main() {
	_ = godotenv.Overload()
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err.Error())
	}
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPassword},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				tasks.PromptQueue: 3,
			},
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Handle(tasks.TypeTxt2txt, tasks.NewTxt2txtProcessor(cfg.LlamaApiHost))
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
