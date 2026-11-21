package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
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

func env(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}

func connectDB() (*sql.DB, error) {
	host := env("DB_HOST")
	port := env("DB_PORT")
	user := env("DB_USER")
	password := env("DB_PASSWORD")
	dbName := env("DB_NAME")

	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + dbName,
		RawQuery: "sslmode=disable",
	}

	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	for {
		if err := db.PingContext(ctx); err == nil {
			log.Println("Connected to PostgreSQL")
			break
		}

		if ctx.Err() != nil {
			db.Close()
			return nil, fmt.Errorf("could not connect to PostgreSQL: %w", ctx.Err())
		}

		log.Println("Waiting for PostgreSQL...")
		time.Sleep(2 * time.Second)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			text VARCHAR(140) NOT NULL
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("could not create todos table: %w", err)
	}

	return db, nil
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

func todosHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTodos(db, w, r)

		case http.MethodPost:
			createTodo(db, w, r)

		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func getTodos(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(r.Context(), `
		SELECT id, text
		FROM todos
		ORDER BY id
	`)
	if err != nil {
		log.Printf("Could not retrieve todos: %v", err)
		http.Error(w, "Could not retrieve todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	todos := make([]Todo, 0)

	for rows.Next() {
		var todo Todo

		if err := rows.Scan(&todo.ID, &todo.Text); err != nil {
			log.Printf("Could not scan todo: %v", err)
			http.Error(w, "Could not retrieve todos", http.StatusInternalServerError)
			return
		}

		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error reading todos: %v", err)
		http.Error(w, "Could not retrieve todos", http.StatusInternalServerError)
		return
	}

	log.Printf("GET /todos | returned=%d", len(todos))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		log.Printf("Failed to encode todos: %v", err)
	}
}

func createTodo(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	var todo Todo

	err := db.QueryRowContext(r.Context(), `
		INSERT INTO todos (text)
		VALUES ($1)
		RETURNING id, text
	`, text).Scan(&todo.ID, &todo.Text)

	if err != nil {
		log.Printf("Could not create todo: %v", err)
		http.Error(w, "Could not create todo", http.StatusInternalServerError)
		return
	}

	log.Printf("POST /todos | created todo id=%d", todo.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		log.Printf("Failed to encode created todo: %v", err)
	}
}

func main() {
	port := env("PORT")

	db, err := connectDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/todos", todosHandler(db))

	log.Printf("Todo backend listening on port %s", port)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
