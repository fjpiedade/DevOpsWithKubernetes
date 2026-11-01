package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	filePath := os.Getenv("LOG_FILE")
	if filePath == "" {
		filePath = "./files/log.txt"
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	log.Printf("Log Writer started")
	log.Printf("Generated UUID: %s", randomString)
	log.Printf("Writing to: %s", filePath)

	for {
		line := fmt.Sprintf("%s: %s\n",
			time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			randomString,
		)

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("failed to open log file: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		_, writeErr := file.WriteString(line)
		closeErr := file.Close()

		if writeErr != nil {
			log.Printf("failed to write log file: %v", writeErr)
		} else if closeErr != nil {
			log.Printf("failed to close log file: %v", closeErr)
		} else {
			log.Print("Wrote new log entry")
		}

		time.Sleep(5 * time.Second)
	}
}
