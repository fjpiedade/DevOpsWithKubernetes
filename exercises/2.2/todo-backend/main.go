package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Todo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type CreateTodoRequest struct {
	Text string `json:"text"`
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

var (
	todos  = []Todo{}
	nextID = 1
	mu     sync.Mutex
)

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

func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTodos(w, r)

	case http.MethodPost:
		createTodo(w, r)

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	log.Printf("GET /todos | returned=%d", len(todos))

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(todos); err != nil {
		http.Error(w, "Failed to encode todos", http.StatusInternalServerError)
		return
	}
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var request CreateTodoRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	text := strings.TrimSpace(request.Text)

	if text == "" {
		http.Error(w, "Todo text cannot be empty", http.StatusBadRequest)
		return
	}

	if len([]rune(text)) > 140 {
		http.Error(w, "Todo text cannot exceed 140 characters", http.StatusBadRequest)
		return
	}

	mu.Lock()

	todo := Todo{
		ID:   nextID,
		Text: text,
	}

	nextID++
	todos = append(todos, todo)

	mu.Unlock()

	log.Printf("POST /todos | created todo id=%d", todo.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		log.Printf("Failed to encode created todo: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/todos", todosHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Todo backend listening on port %s", port)

	if err := http.ListenAndServe(":"+port, logRequests(mux)); err != nil {
		log.Fatal(err)
	}
}
