package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeTxt2txt = "text:text"
)

const (
	PromptQueue = "prompts"
)

type JsonTxt2txtResponse struct {
	Responses []string
}

type Txt2txtPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Txt2txtRequest [][]Txt2txtPrompt

type Txt2txtResponse []string

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

func NewTxt2txtTask(req *Txt2txtRequest) (*asynq.Task, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	// task options can be passed to NewTask, which can be overridden at enqueue time.
	return asynq.NewTask(TypeTxt2txt, payload, asynq.MaxRetry(5), asynq.Timeout(20*time.Minute)), nil
}

//---------------------------------------------------------------
// Write a function HandleXXXTask to handle the input task.
// Note that it satisfies the asynq.HandlerFunc interface.
//
// Handler doesn't need to be a function. You can define a type
// that satisfies asynq.Handler interface. See examples below.
//---------------------------------------------------------------

// ImageProcessor implements asynq.Handler interface.
type Txt2txtProcessor struct {
	// ... fields for struct
	host string
}

func (api *Txt2txtProcessor) TextToText(req *Txt2txtRequest) (*Txt2txtResponse, error) {
	if req == nil {
		return nil, errors.New("missing request")
	}

	postURL := api.host + "/api/query"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		log.Printf("API URL: %s", postURL)
		log.Printf("Error with API Request: %s", string(jsonData))

		return nil, err
	}

	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	respStruct := &Txt2txtResponse{}

	err = json.Unmarshal(body, respStruct)
	if err != nil {
		log.Printf("API URL: %s", postURL)
		log.Printf("Unexpected API response: %s", string(body))

		return nil, err
	}

	return respStruct, nil
}

func (processor *Txt2txtProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p Txt2txtRequest
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Generating text for prompt: %v", p)
	//res := []byte("Welcome")
	res, err := processor.TextToText(&p)
	if err != nil {
		log.Printf("TextToText Failed: %v", err)
		return err
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		return err
	}

	if _, err := t.ResultWriter().Write(jsonRes); err != nil {
		log.Printf("Failed to write task result: %v", err)
		return err
	}
	// Image resizing code ...
	return nil
}

func NewTxt2txtProcessor(host string) *Txt2txtProcessor {
	return &Txt2txtProcessor{host: host}
}
