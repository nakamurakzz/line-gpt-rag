package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	Body string `json:"body"`
}

type Body struct {
	URL string `json:"url"`
}

type Response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func HandleRequest(ctx context.Context, event Event) (Response, error) {
	var body Body
	err := json.Unmarshal([]byte(event.Body), &body)
	if err != nil {
		return Response{StatusCode: 400, Body: `{"msg": "error ready body, Invalid JSON"}`}, err
	}

	log.Printf("request: %v", body)

	// read from env
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	if openaiApiKey == "" {
		return Response{
			StatusCode: 500,
			Body:       `{"msg": "OPENAI_API_KEY is not set"}`,
		}, fmt.Errorf("OPENAI_API_KEY is not set")
	}
	log.Printf("OPENAI_API_KEY: %v", openaiApiKey[:3])

	return Response{
		StatusCode: 200,
		Body:       `{"msg": "ok"}`,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
