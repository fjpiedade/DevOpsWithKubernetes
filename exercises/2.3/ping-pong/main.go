package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	counterFile = env("COUNTER_FILE", "/usr/src/app/files/pingpong.txt")
	mu          sync.Mutex
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func readCounter() (int, error) {
	data, err := os.ReadFile(counterFile)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		return 0, nil
	}

	return strconv.Atoi(value)
}

func writeCounter(value int) error {
	if err := os.MkdirAll(filepath.Dir(counterFile), 0755); err != nil {
		return err
	}

	tmpFile := counterFile + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(strconv.Itoa(value)+"\n"), 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, counterFile)
}

func main() {
	port := env("PORT", "9901")

	// Initialize the counter file before accepting HTTP requests.
	initialCounter, err := readCounter()
	if err != nil {
		log.Fatalf("Could not initialize counter: %v", err)
	}

	if err := writeCounter(initialCounter); err != nil {
		log.Fatalf("Could not create counter file: %v", err)
	}

	http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		counter, err := readCounter()
		if err != nil {
			log.Printf("Could not read counter: %v", err)
			http.Error(w, "Could not read counter", http.StatusInternalServerError)
			return
		}

		counter++

		if err := writeCounter(counter); err != nil {
			log.Printf("Could not save counter: %v", err)
			http.Error(w, "Could not save counter", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "pong %d\n", counter)
	})

	http.HandleFunc("/pings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		counter, err := readCounter()
		if err != nil {
			log.Printf("Could not read counter: %v", err)
			http.Error(w, "Could not read counter", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, counter)
	})

	log.Printf("Ping-pong server listening on :%s; counter file: %s", port, counterFile)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
