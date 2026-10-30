package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func generateUUID() (string, error) {
	bufferOfBytes := make([]byte, 16)

	if _, err := rand.Read(bufferOfBytes); err != nil {
		return "", err
	}

	bufferOfBytes[6] = (bufferOfBytes[6] & 0x0f) | 0x40
	bufferOfBytes[8] = (bufferOfBytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bufferOfBytes[0:4],
		bufferOfBytes[4:6],
		bufferOfBytes[6:8],
		bufferOfBytes[8:10],
		bufferOfBytes[10:16],
	), nil
}

func main() {
	randomString, err := generateUUID()
	if err != nil {
		log.Fatalf("failed to generate UUID: %v", err)
	}

	go func() {
		for {
			fmt.Printf("%s: %s\n",
				time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				randomString,
			)
			time.Sleep(5 * time.Second)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintf(w, "%s: %s\n",
			time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			randomString,
		)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "9901"
	}

	log.Printf("Server started on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
