package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/lambda"
	openai "github.com/sashabaranov/go-openai"
)

type Event struct {
	Body string `json:"body"`
}

type Message struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	QuoteToken string `json:"quoteToken"`
	Text       string `json:"text"`
}

type DeliveryContext struct {
	IsRedelivery bool `json:"isRedelivery"`
}

type Source struct {
	Type   string `json:"type"`
	UserID string `json:"userId"`
}

type Events struct {
	Type            string          `json:"type"`
	Message         Message         `json:"message"`
	WebhookEventID  string          `json:"webhookEventId"`
	DeliveryContext DeliveryContext `json:"deliveryContext"`
	Timestamp       int64           `json:"timestamp"`
	Source          Source          `json:"source"`
	ReplyToken      string          `json:"replyToken"`
	Mode            string          `json:"mode"`
}

type Body struct {
	URL    string   `json:"url"`
	Events []Events `json:"events"`
}

type Response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

type ResponseBody struct {
	ReplyToken string `json:"replyToken"`
	Messages   []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"messages"`
}

func HandleRequest(ctx context.Context, event Event) (Response, error) {
	log.Printf("event: %v", event)

	var body Body
	err := json.Unmarshal([]byte(event.Body), &body)
	if err != nil {
		return Response{StatusCode: 400, Body: `{"msg": "error ready body, Invalid JSON"}`}, err
	}

	if len(body.Events) == 0 {
		return Response{StatusCode: 200, Body: `{"message": "success"}`}, nil
	}

	// read from env
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	if openaiApiKey == "" {
		return Response{
			StatusCode: 500,
			Body:       `{"msg": "OPENAI_API_KEY is not set"}`,
		}, fmt.Errorf("OPENAI_API_KEY is not set")
	}
	lineAccessToken := os.Getenv("LINE_ACCESS_TOKEN")
	if lineAccessToken == "" {
		return Response{
			StatusCode: 500,
		}, fmt.Errorf("LINE_ACCESS_TOKEN is not set")
	}

	// call openai api
	openAiClient := openai.NewClient(openaiApiKey)
	openaiResp, err := openAiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "if your answer will be over 1500 characters, it must be submitted in multiple messages." + body.Events[0].Message.Text,
				},
			},
		},
	)
	if err != nil {
		log.Printf("failed to call openai api: %v", err)
		return Response{
			StatusCode: 500,
			Body:       `{"msg": "failed to call openai api"}`,
		}, fmt.Errorf("failed to call openai api: %v", err)
	}
	c := openaiResp.Choices[0].Message.Content

	var messages []string
	// message は 300文字ずつに分割する
	l := utf8.RuneCountInString(c)
	for i := 0; i < l; i += 300 {
		end := i + 300
		if end > l {
			end = l
		}
		messages = append(messages, c[i:end])
	}

	var rb ResponseBody
	rb.ReplyToken = body.Events[0].ReplyToken
	for _, m := range messages {
		rb.Messages = append(rb.Messages, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			Type: "text",
			Text: m,
		})
	}

	rBody, err := json.Marshal(rb)
	if err != nil {
		log.Printf("failed to marshal response body: %v", err)
		return Response{
			StatusCode: 500,
			Body:       `{"msg": "failed to marshal response body"}`,
		}, fmt.Errorf("failed to marshal response body: %v", err)
	}
	b := bytes.NewBuffer(rBody)

	req, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/reply", b)
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return Response{
			StatusCode: 500,
			Body:       `{"msg": "failed to create request"}`,
		}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", lineAccessToken))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to request: %v", err)
		return Response{
			StatusCode: 500,
			Body:       `{"msg": "failed to request"}`,
		}, fmt.Errorf("failed to request: %v", err)
	}
	defer resp.Body.Close()

	return Response{
		StatusCode: 200,
		Body:       `{"msg": "success"}`,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
