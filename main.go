package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/akii1ira/todo-list/handlers"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "🚀 Todo List API running. Use /api/todo-list/tasks")
	})

	http.HandleFunc("/api/todo-list/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateTaskHandler(w, r)
		case http.MethodGet:
			handlers.ListTasksHandler(w, r)
		default:
			handlers.JSONWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Todo server started on port", port)
	http.ListenAndServe(":"+port, nil)
}
