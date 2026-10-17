package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	capUSD, err := strconv.ParseFloat(os.Getenv("HARD_CAP_USD"), 64)
	if err != nil || capUSD <= 0 {
		log.Fatal("HARD_CAP_USD must be positive")
	}
	account := accountClient{key: key, http: &http.Client{Timeout: 20 * time.Second}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := account.setBudget(ctx, capUSD, "monthly"); err != nil {
		log.Fatal(err)
	}
	// One credential and one base URL cover both the budget control and the inference.
	ai := openai.NewClient(option.WithAPIKey(key), option.WithBaseURL(baseURL))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /requests", func(w http.ResponseWriter, r *http.Request) {
		var input PropertyRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		prompt, err := input.prompt()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := ai.Chat.Completions.New(r.Context(), openai.ChatCompletionNewParams{
			Model: "auto", Messages: []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)},
		})
		if err != nil {
			var apiErr *openai.Error
			if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
				http.Error(w, apiErr.Error(), apiErr.StatusCode)
				return
			}
			http.Error(w, "inference failed", http.StatusBadGateway)
			return
		}
		if len(result.Choices) == 0 {
			http.Error(w, "empty response", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"kind": input.Kind, "property": input.Property, "note": result.Choices[0].Message.Content})
	})
	mux.HandleFunc("GET /usage", func(w http.ResponseWriter, r *http.Request) {
		data, err := account.usage(r.Context())
		if err != nil {
			var apiErr *apiError
			if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
				http.Error(w, apiErr.Error(), apiErr.Status)
				return
			}
			http.Error(w, "usage request failed", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
}
