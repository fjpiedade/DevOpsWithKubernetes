package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

func main() {
	logFile := env("LOG_FILE", "/usr/src/app/files/log.txt")
	counterFile := env("COUNTER_FILE", "/usr/src/app/files/pingpong.txt")
	port := env("PORT", "9901")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		logLine, err := readFile(logFile)
		if err != nil {
			log.Printf("Could not read %s: %v", logFile, err)
			http.Error(w, "Log data is not available yet", http.StatusServiceUnavailable)
			return
		}

		counter, err := readFile(counterFile)
		if os.IsNotExist(err) {
			counter = "0"
		} else if err != nil {
			log.Printf("Could not read %s: %v", counterFile, err)
			http.Error(w, "Ping-pong data is not available", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintf(w, "%s.\nPing / Pongs: %s\n", logLine, counter)
	})

	log.Printf("Log reader listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
