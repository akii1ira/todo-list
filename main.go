package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ActiveAt  string `json:"activeAt"`
	Done      bool   `json:"done"`
	CreatedAt int64  `json:"-"`
}

type createUpdateRequest struct {
	Title    string `json:"title"`
	ActiveAt string `json:"activeAt"`
}

var store sync.Map
var mu sync.Mutex

const dateLayout = "2006-01-02"

// --- Вспомогательная функция для ответа ---
func jsonWrite(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// --- POST /api/todo-list/tasks ---
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req createUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonWrite(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.ActiveAt = strings.TrimSpace(req.ActiveAt)

	if req.Title == "" || req.ActiveAt == "" {
		jsonWrite(w, http.StatusBadRequest, map[string]string{"error": "title and activeAt are required"})
		return
	}

	id := uuid.New().String()
	now := time.Now().Unix()
	task := &Task{
		ID:        id,
		Title:     req.Title,
		ActiveAt:  req.ActiveAt,
		Done:      false,
		CreatedAt: now,
	}

	store.Store(id, task)
	jsonWrite(w, http.StatusCreated, map[string]string{"id": id})
}

// --- PUT /api/todo-list/tasks/{id} ---
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/todo-list/tasks/")
	if id == "" {
		jsonWrite(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}

	var req createUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonWrite(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.ActiveAt = strings.TrimSpace(req.ActiveAt)

	if req.Title == "" || req.ActiveAt == "" {
		jsonWrite(w, http.StatusBadRequest, map[string]string{"error": "title and activeAt are required"})
		return
	}

	v, ok := store.Load(id)
	if !ok {
		jsonWrite(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	task := v.(*Task)
	task.Title = req.Title
	task.ActiveAt = req.ActiveAt
	store.Store(id, task)

	w.WriteHeader(http.StatusNoContent)
}

// --- DELETE /api/todo-list/tasks/{id} ---
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		jsonWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/todo-list/tasks/")
	if id == "" {
		jsonWrite(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}

	_, ok := store.Load(id)
	if !ok {
		jsonWrite(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	store.Delete(id)
	w.WriteHeader(http.StatusNoContent)
}

// --- PUT /api/todo-list/tasks/{id}/done ---
func markDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/todo-list/tasks/")
	id = strings.TrimSuffix(id, "/done")
	if id == "" {
		jsonWrite(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}

	v, ok := store.Load(id)
	if !ok {
		jsonWrite(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	task := v.(*Task)
	task.Done = true
	store.Store(id, task)
	w.WriteHeader(http.StatusNoContent)
}

// --- GET /api/todo-list/tasks?status=active|done ---
func listTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	status := r.URL.Query().Get("status")
	tasks := make([]*Task, 0)

	store.Range(func(_, v interface{}) bool {
		t := v.(*Task)
		if status == "done" && t.Done {
			tasks = append(tasks, t)
		} else if status == "" || status == "active" {
			if !t.Done {
				tasks = append(tasks, t)
			}
		}
		return true
	})

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt < tasks[j].CreatedAt
	})

	jsonWrite(w, http.StatusOK, tasks)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Todo List API running. Use /api/todo-list/tasks")
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/todo-list/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createTaskHandler(w, r)
		case http.MethodGet:
			listTasksHandler(w, r)
		default:
			jsonWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	http.HandleFunc("/api/todo-list/tasks/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/done") {
			markDoneHandler(w, r)
			return
		}

		switch r.Method {
		case http.MethodPut:
			updateTaskHandler(w, r)
		case http.MethodDelete:
			deleteTaskHandler(w, r)
		default:
			jsonWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server started on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("server error:", err)
	}
}
