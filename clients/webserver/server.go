package main

import (
	"net/http"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/rampenke/zosma-llama-server/tasks"
)

type Config struct {
	RedisAddr     string `envconfig:"REDIS_ADDR" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
}

var cfg Config

type ServerContext struct {
	echo.Context
}

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

func ProcessQuery(c echo.Context, client *asynq.Client, inspector *asynq.Inspector) (*tasks.Txt2txtResponse, error) {

	request := &tasks.Txt2txtRequest{}
	err := c.Bind(&request)
	if err != nil {
		return nil, err
	}

	task, err := tasks.NewTxt2txtTask(request)
	if err != nil {
		log.Printf("could not create task: %v", err)
		return nil, err
	}
	info, err := client.Enqueue(task, asynq.Queue(tasks.PromptQueue), asynq.MaxRetry(10), asynq.Timeout(3*time.Minute), asynq.Retention(2*time.Hour))
	if err != nil {
		log.Printf("could not enqueue task: %v", err)
		return nil, err
	}
	log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	res, err := waitForResult(ctx, inspector, tasks.PromptQueue, info.ID)
	if err != nil {
		log.Printf("unable to wait for resilt: %v", err)
		return nil, err
	}
	var respStruct = &tasks.Txt2txtResponse{}
	err = json.Unmarshal(res.Result, respStruct)
	if err != nil {
		log.Printf("Unexpected API response: %v", err)
		return nil, err
	}
	fmt.Printf("%v", respStruct)
	return respStruct, nil
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

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/api/query", func(c echo.Context) error {
		err, res := ProcessQuery(c, client, inspector)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, res)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
