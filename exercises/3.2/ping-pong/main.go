package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func connectDB() (*sql.DB, error) {
	host := env("DB_HOST", "postgres")
	port := env("DB_PORT", "5432")
	user := env("DB_USER", "pingpong")
	password := env("DB_PASSWORD", "pingpong123")
	dbName := env("DB_NAME", "pingpong")

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
		CREATE TABLE IF NOT EXISTS pingpong_counter (
			id INTEGER PRIMARY KEY,
			count BIGINT NOT NULL
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("could not create counter table: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO pingpong_counter (id, count)
		VALUES (1, 0)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("could not initialize counter: %w", err)
	}

	return db, nil
}

func main() {
	port := env("PORT", "9901")

	db, err := connectDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// GKE Ingress health check: this endpoint must return HTTP 200.
	// It does not increment the Ping-pong counter.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var counter int64

		err := db.QueryRow(`
			UPDATE pingpong_counter
			SET count = count + 1
			WHERE id = 1
			RETURNING count
		`).Scan(&counter)

		if err != nil {
			log.Printf("Could not increment counter: %v", err)
			http.Error(w, "Could not increment counter", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "pong %d\n", counter)
	})

	http.HandleFunc("/pings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var counter int64

		err := db.QueryRow(`
			SELECT count
			FROM pingpong_counter
			WHERE id = 1
		`).Scan(&counter)

		if err != nil {
			log.Printf("Could not read counter: %v", err)
			http.Error(w, "Could not read counter", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, counter)
	})

	log.Printf("Ping-pong server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
