package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const cacheDuration = 10 * time.Minute

type ImageCache struct {
	mu            sync.Mutex
	imageFile     string
	timestampFile string
	graceFile     string
	client        *http.Client
}

func newImageCache() *ImageCache {
	imageFile := os.Getenv("IMAGE_FILE")
	if imageFile == "" {
		imageFile = "/data/image.jpg"
	}

	return &ImageCache{
		imageFile:     imageFile,
		timestampFile: imageFile + ".timestamp",
		graceFile:     imageFile + ".grace",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ImageCache) getImage() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := os.Stat(c.imageFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		timestamp, timestampErr := os.ReadFile(c.timestampFile)
		if timestampErr == nil {
			savedTime, parseErr := time.Parse(time.RFC3339Nano, string(timestamp))
			if parseErr == nil && time.Since(savedTime) < cacheDuration {
				return nil
			}

			// Serve the old image once after expiration.
			grace, _ := os.ReadFile(c.graceFile)
			if string(grace) != "used" {
				return os.WriteFile(c.graceFile, []byte("used"), 0644)
			}
		}
	}

	return c.downloadImage()
}

func (c *ImageCache) downloadImage() error {
	resp, err := c.client.Get("https://picsum.photos/1200")
	if err != nil {
		return fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Lorem Picsum returned status %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(c.imageFile), 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(c.imageFile), "image-*.tmp")
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, c.imageFile); err != nil {
		return err
	}

	if err := os.WriteFile(c.timestampFile, []byte(time.Now().UTC().Format(time.RFC3339Nano)), 0644); err != nil {
		return err
	}

	if err := os.Remove(c.graceFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (c *ImageCache) serveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := c.getImage(); err != nil {
		log.Printf("Image cache error: %v", err)

		// If a previous image exists, keep serving it when Picsum is unavailable.
		if _, statErr := os.Stat(c.imageFile); statErr != nil {
			http.Error(w, "Image unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	http.ServeFile(w, r, c.imageFile)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9901"
	}

	cache := newImageCache()

	http.HandleFunc("/image", cache.serveImage)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="pt">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>To-do Application</title>

    <style>
        body {
            margin: 0;
            padding: 30px 20px;
            font-family: Arial, sans-serif;
            text-align: center;
            background-color: #f4f4f4;
        }

        h1 {
            text-align: center;
            margin-bottom: 30px;
            color: #333;
        }

        .image-container {
            display: flex;
            justify-content: center;
            align-items: center;
            width: 100%;
        }

        .image-container img {
            width: 700px;
            height: 450px;
            object-fit: cover;
            border-radius: 12px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
			margin-bottom: 20px;
        }

        @media (max-width: 650px) {
            .image-container img {
                width: 100%;
                height: 250px;
            }
        }
    </style>
</head>

<body>
    <h1>To-do Application with Go</h1>

    <div class="image-container">
        <img src="/image" alt="Random image from Lorem Picsum">
    </div>
	<span>DevOps with Kubernetes 2026</span>
</body>
</html>`)
	})

	log.Printf("Server started on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
