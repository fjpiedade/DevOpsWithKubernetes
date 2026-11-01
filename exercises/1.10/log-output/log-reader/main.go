package main

import (
	"log"
	"net/http"
	"os"
)

func getLogFilePath() string {
	if path := os.Getenv("LOG_FILE"); path != "" {
		return path
	}

	return "./files/log.txt"
}

func main() {
	filePath := getLogFilePath()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "Log file not found", http.StatusNotFound)
				return
			}

			http.Error(w, "Failed to read log file", http.StatusInternalServerError)
			log.Printf("Failed to read log file: %v", err)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(content)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "9901"
	}

	log.Printf("Log Reader started on port %s", port)
	log.Printf("Reading from: %s", filePath)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
