package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type Payload struct {
	Body string `json:"body"`
}

type Request struct {
	URL string `json:"url"`
}

func HandleRequest(ctx context.Context, payload Payload) (bool, error) {
	var req Request
	err := json.NewDecoder(strings.NewReader(payload.Body)).Decode(&req)
	if err != nil {
		return false, fmt.Errorf("cannot decode payload: %v: %v", err, payload.Body)
	}

	log.Printf("request: %v", req)

	return true, nil
}

func main() {
	lambda.Start(HandleRequest)
}
