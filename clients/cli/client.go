package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/hibiken/asynq"
	"github.com/rampenke/zosma-llama-server/tasks"
)

type Config struct {
	RedisAddr     string `envconfig:"REDIS_ADDR" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
}

var cfg Config

func waitForResult(ctx context.Context, i *asynq.Inspector, queue, taskID string) (*asynq.TaskInfo, error) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			taskInfo, err := i.GetTaskInfo(queue, taskID)
			if err != nil {
				return nil, err
			}
			if taskInfo.CompletedAt.IsZero() {
				continue
			}
			return taskInfo, nil
		case <-ctx.Done():
			return nil, fmt.Errorf("context closed")
		}
	}
}

func main() {
	_ = godotenv.Overload()
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err.Error())
	}
	conn := asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPassword}
	client := asynq.NewClient(conn)
	defer client.Close()
	inspector := asynq.NewInspector(conn)
	request := &tasks.Txt2txtRequest{
		{
			{
				Role:    "user",
				Content: "I want to know about Ayurveda",
			},
		},
	}
	task, err := tasks.NewTxt2txtTask(request)
	if err != nil {
		log.Fatalf("could not create task: %v", err)
	}
	info, err := client.Enqueue(task, asynq.Queue(tasks.PromptQueue), asynq.MaxRetry(10), asynq.Timeout(3*time.Minute), asynq.Retention(2*time.Hour))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	res, err := waitForResult(ctx, inspector, tasks.PromptQueue, info.ID)
	if err != nil {
		log.Fatalf("unable to wait for resilt: %v", err)
	}
	var respStruct = &tasks.Txt2txtResponse{}
	err = json.Unmarshal(res.Result, respStruct)
	if err != nil {
		log.Fatalf("Unexpected API response: %v", err)
	}
	fmt.Printf("%v", respStruct)
}
