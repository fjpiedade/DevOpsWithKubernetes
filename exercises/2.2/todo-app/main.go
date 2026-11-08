package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		log.Printf(
			"%s %s | status=%d | duration=%s",
			r.Method,
			r.URL.Path,
			recorder.status,
			time.Since(start),
		)
	})
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

func todosProxy(w http.ResponseWriter, r *http.Request) {
	backendURL := os.Getenv("TODO_BACKEND_URL")

	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	target, err := url.Parse(backendURL)
	if err != nil {
		http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}

//go:embed static/*
var frontend embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9901"
	}

	cache := newImageCache()

	// Load frontend files from the embedded filesystem.
	staticFiles, err := fs.Sub(frontend, "static")
	if err != nil {
		log.Fatal(err)
	}

	indexHTML, err := fs.ReadFile(staticFiles, "index.html")
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTP router.
	mux := http.NewServeMux()

	// Proxy requests to the todo backend.
	mux.HandleFunc("/todos", todosProxy)

	// Serve the cached image.
	mux.HandleFunc("/image", cache.serveImage)

	// Serve frontend assets.
	mux.Handle("/static/", http.StripPrefix(
		"/static/",
		http.FileServer(http.FS(staticFiles)),
	))

	// Serve application homepage.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if r.Method == http.MethodHead {
			return
		}

		if _, err := w.Write(indexHTML); err != nil {
			log.Printf("Error serving homepage: %v", err)
		}
	})

	log.Printf("Server started on port %s", port)

	if err := http.ListenAndServe(":"+port, logRequests(mux)); err != nil {
		log.Fatal(err)
	}
}
