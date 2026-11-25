package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func getPingPongCount(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ping-pong returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func main() {
	logFile := env("LOG_FILE", "/usr/src/app/files/log.txt")
	configFile := env("CONFIG_FILE", "/usr/src/app/config/information.txt")
	message := env("MESSAGE", "")

	pingPongURL := env(
		"PINGPONG_URL",
		"http://ping-pong-svc:80/pings",
	)
	port := env("PORT", "9901")

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fileContent, err := readFile(configFile)
		if err != nil {
			log.Printf("Could not read configuration file: %v", err)
			http.Error(w, "Configuration file is not available", http.StatusServiceUnavailable)
			return
		}

		logLine, err := readFile(logFile)
		if err != nil {
			log.Printf("Could not read log file: %v", err)
			http.Error(w, "Log data is not available yet", http.StatusServiceUnavailable)
			return
		}

		counter, err := getPingPongCount(client, pingPongURL)
		if err != nil {
			log.Printf("Could not contact Ping-pong: %v", err)
			http.Error(w, "Ping-pong service is unavailable", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintf(
			w,
			"file content: %s\nenv variable: MESSAGE=%s\n%s.\nPing / Pongs: %s\n",
			fileContent,
			message,
			logLine,
			counter,
		)
	})

	log.Printf("Log-reader listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
