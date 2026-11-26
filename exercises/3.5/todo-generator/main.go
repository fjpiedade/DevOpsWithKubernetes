package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const wikipediaRandomURL = "https://en.wikipedia.org/wiki/Special:Random"

type Todo struct {
	Text string `json:"text"`
}

func getRandomArticleURL(client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, wikipediaRandomURL, nil)
	if err != nil {
		return "", fmt.Errorf("create Wikipedia request: %w", err)
	}

	req.Header.Set("User-Agent", "TodoGenerator/1.0 (Kubernetes CronJob)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request random Wikipedia article: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Wikipedia returned unexpected status: %s", resp.Status)
	}

	finalURL := resp.Request.URL.String()

	parsedURL, err := url.Parse(finalURL)
	if err != nil {
		return "", fmt.Errorf("parse Wikipedia article URL: %w", err)
	}

	if parsedURL.Scheme != "https" || parsedURL.Host != "en.wikipedia.org" ||
		!strings.HasPrefix(parsedURL.Path, "/wiki/") {
		return "", fmt.Errorf("unexpected Wikipedia article URL: %s", finalURL)
	}

	return finalURL, nil
}

func createTodo(client *http.Client, backendURL, articleURL string) error {
	todo := Todo{
		Text: "Read " + articleURL,
	}

	body, err := json.Marshal(todo)
	if err != nil {
		return fmt.Errorf("encode todo: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		backendURL+"/todos",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create backend request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send todo to backend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf(
			"backend returned %s: %s",
			resp.Status,
			strings.TrimSpace(string(responseBody)),
		)
	}

	var created Todo
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return fmt.Errorf("decode backend response: %w", err)
	}

	log.Printf("Created todo: %s", created.Text)
	return nil
}

func main() {
	backendURL := strings.TrimRight(os.Getenv("TODO_BACKEND_URL"), "/")
	if backendURL == "" {
		log.Fatal("TODO_BACKEND_URL environment variable is required")
	}

	parsedBackendURL, err := url.ParseRequestURI(backendURL)
	if err != nil || parsedBackendURL.Scheme != "http" ||
		parsedBackendURL.Host == "" {
		log.Fatalf("invalid TODO_BACKEND_URL: %s", backendURL)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	articleURL, err := getRandomArticleURL(client)
	if err != nil {
		log.Fatalf("could not obtain random article: %v", err)
	}

	log.Printf("Selected Wikipedia article: %s", articleURL)

	if err := createTodo(client, backendURL, articleURL); err != nil {
		log.Fatalf("could not create todo: %v", err)
	}
}
