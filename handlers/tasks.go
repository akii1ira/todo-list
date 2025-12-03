package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/akii1ira/todo-list/models"
	"github.com/google/uuid"
)

var store sync.Map
var mu sync.Mutex

type createUpdateRequest struct {
	Title    string `json:"title"`
	ActiveAt string `json:"activeAt"`
}

// -------------------- Вспомогательные функции --------------------

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

// -------------------- Создать задачу --------------------

func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	if len(req.Title) > 200 {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "title too long"})
		return
	}

	if _, err := ParseDate(req.ActiveAt, models.DateLayout); err != nil {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "invalid date"})
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

// -------------------- Обработчик путей с ID --------------------

func TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/todo-list/tasks/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}

	switch r.Method {

	// GET /tasks/{id}
	case http.MethodGet:
		if len(parts) == 1 {
			GetTaskHandler(w, r, id)
			return
		}

	// PUT /tasks/{id}
	case http.MethodPut:
		if len(parts) == 1 {
			UpdateTaskHandler(w, r, id)
			return
		}

		// PUT /tasks/{id}/done
		if len(parts) == 2 && parts[1] == "done" {
			MarkDoneHandler(w, r, id)
			return
		}

	// DELETE /tasks/{id}
	case http.MethodDelete:
		if len(parts) == 1 {
			DeleteTaskHandler(w, r, id)
			return
		}
	}

	JSONWrite(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

// -------------------- Получить задачу по ID --------------------

func GetTaskHandler(w http.ResponseWriter, r *http.Request, id string) {
	v, ok := store.Load(id)
	if !ok {
		JSONWrite(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	JSONWrite(w, http.StatusOK, v.(*models.Task))
}

// -------------------- Обновить задачу --------------------

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, id string) {
	v, ok := store.Load(id)
	if !ok {
		JSONWrite(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	var req createUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if req.Title == "" || req.ActiveAt == "" {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "title and activeAt are required"})
		return
	}

	if len(req.Title) > 200 {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "title too long"})
		return
	}

	if _, err := ParseDate(req.ActiveAt, models.DateLayout); err != nil {
		JSONWrite(w, http.StatusBadRequest, map[string]string{"error": "invalid date"})
		return
	}

	if !checkUniqueness(req.Title, req.ActiveAt, id) {
		JSONWrite(w, http.StatusNotFound, map[string]string{"error": "duplicate task"})
		return
	}

	t := v.(*models.Task)
	t.Title = req.Title
	t.ActiveAt = req.ActiveAt

	store.Store(id, t)
	w.WriteHeader(http.StatusNoContent)

}

// -------------------- Пометить выполненной --------------------

func MarkDoneHandler(w http.ResponseWriter, r *http.Request, id string) {
	v, ok := store.Load(id)
	if !ok {
		JSONWrite(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	t := v.(*models.Task)
	t.Done = true
	store.Store(id, t)

	w.WriteHeader(http.StatusNoContent)

}

// -------------------- Удалить задачу --------------------

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, id string) {
	_, ok := store.Load(id)
	if !ok {
		JSONWrite(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	store.Delete(id)
	w.WriteHeader(http.StatusNoContent)

}

// -------------------- Список задач --------------------
func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}

	now := DateOnly(time.Now())
	tasks := []*models.Task{}

	store.Range(func(_, v interface{}) bool {
		t := v.(*models.Task)
		d, err := ParseDate(t.ActiveAt, models.DateLayout)
		if err != nil {
			return true
		}

		if status == "done" && t.Done {
			tasks = append(tasks, t)
		} else if status == "active" && !t.Done && (d.Before(now) || d.Equal(now)) {
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
