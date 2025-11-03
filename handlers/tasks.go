package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/akii1ira/todo-list/models"
)

var store sync.Map
var mu sync.Mutex

type createUpdateRequest struct {
	Title    string `json:"title"`
	ActiveAt string `json:"activeAt"`
}

func checkUniqueness(title, activeAt, skipID string) bool {
	unique := true
	store.Range(func(_, v interface{}) bool {
		t := v.(*models.Task)
		if t.ID == skipID {
			return true
		}
		if strings.EqualFold(t.Title, title) && t.ActiveAt == activeAt {
			unique = false
			return false
		}
		return true
	})
	return unique
}

func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req createUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.ActiveAt = strings.TrimSpace(req.ActiveAt)

	if req.Title == "" || req.ActiveAt == "" {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "title and activeAt are required"})
		return
	}

	if _, err := ParseDate(req.ActiveAt, models.DateLayout); err != nil {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "activeAt must be valid date YYYY-MM-DD"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if !checkUniqueness(req.Title, req.ActiveAt, "") {
		JSONWrite(w, http.StatusNotFound, map[string]string{"error": "task already exists"})
		return
	}

	task := &models.Task{
		ID:        uuid.New().String(),
		Title:     req.Title,
		ActiveAt:  req.ActiveAt,
		Done:      false,
		CreatedAt: time.Now().Unix(),
	}

	store.Store(task.ID, task)
	JSONWrite(w, http.StatusCreated, map[string]string{"id": task.ID})
}

func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONWrite(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}

	now := time.Now()
	tasks := []*models.Task{}

	store.Range(func(_, v interface{}) bool {
		t := v.(*models.Task)
		d, err := ParseDate(t.ActiveAt, models.DateLayout)
		if err != nil {
			return true
		}

		if status == "done" && t.Done {
			tasks = append(tasks, t)
		} else if status == "active" && !t.Done && (d.Before(DateOnly(now)) || d.Equal(DateOnly(now))) {
			tasks = append(tasks, t)
		}
		return true
	})

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt < tasks[j].CreatedAt
	})

	out := []map[string]string{}
	for _, t := range tasks {
		out = append(out, map[string]string{
			"id":       t.ID,
			"title":    t.Title,
			"activeAt": t.ActiveAt,
		})
	}

	JSONWrite(w, http.StatusOK, out)
}
