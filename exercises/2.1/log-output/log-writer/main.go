package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const uuidLength = 36

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func newUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}

func uuidFromExistingFile(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	line := strings.TrimSpace(string(data))
	separator := strings.LastIndex(line, ": ")
	if separator < 0 {
		return "", false
	}
	id := strings.TrimSpace(line[separator+2:])
	if len(id) != uuidLength {
		return "", false
	}
	return id, true
}

func writeLog(path, id string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	line := time.Now().UTC().Format(time.RFC3339Nano) + ": " + id + "\n"
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(line), 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, path)
}

func main() {
	logFile := env("LOG_FILE", "/usr/src/app/files/log.txt")
	interval := 5 * time.Second

	id, found := uuidFromExistingFile(logFile)
	if !found {
		var err error
		id, err = newUUID()
		if err != nil {
			log.Fatalf("Could not generate UUID: %v", err)
		}
		log.Printf("Generated a new UUID")
	} else {
		log.Printf("Recovered the existing UUID from persistent storage")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := writeLog(logFile, id); err != nil {
			log.Printf("Could not update log file: %v", err)
		}
		<-ticker.C
	}
}
