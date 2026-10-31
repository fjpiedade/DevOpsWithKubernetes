package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	counter int
	mutex   sync.Mutex
)

func main() {
	http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		fmt.Fprintf(w, "pong %d\n", counter)
		counter++
	})

	port := "9901"
	log.Printf("Ping-pong server started on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
