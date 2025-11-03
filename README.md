# 📝 Todo List API (Go)

## 📌 Overview

This is a simple RESTful API built with **Go (Golang)** that allows users to manage a list of tasks.
Each task has a title, an activation date, and a completion status.
The data is stored **in memory** (no database) using Go’s `sync.Map`.

The project follows a clean structure with separated packages:

- `handlers/` — API endpoint logic
- `models/` — data structures
- `main.go` — app entry point

---

## ⚙️ Features

- Create, update, delete, and list tasks
- Mark tasks as done
- Filter tasks by status (`active` or `done`)
- Simple in-memory storage (no external DB required)
- Ready for containerization with Docker

---

## 🚀 Technologies

- **Language:** Go 1.22+
- **Framework:** net/http (standard library)
- **Tools:** Docker, Makefile

---

## 🧩 API Endpoints

### ➕ Create Task

**POST** `/api/todo-list/tasks`

**Request Body:**

```json
{
  "title": "Buy groceries",
  "activeAt": "2025-10-31"
}
```

**Response:**

```json
{
  "id": "a1b2c3d4-e5f6-7890-1234-56789abcde"
}
```

---

### ✏️ Update Task

**PUT** `/api/todo-list/tasks/{id}`

**Request Body:**

```json
{
  "title": "Buy milk and bread",
  "activeAt": "2025-11-01"
}
```

**Response:** `204 No Content`

---

### ✅ Mark as Done

**PUT** `/api/todo-list/tasks/{id}/done`
**Response:** `204 No Content`

---

### ❌ Delete Task

**DELETE** `/api/todo-list/tasks/{id}`
**Response:** `204 No Content`

---

### 📋 List Tasks

**GET** `/api/todo-list/tasks?status=active`
**GET** `/api/todo-list/tasks?status=done`

**Response:**

```json
[
  {
    "id": "a1b2c3",
    "title": "Buy groceries",
    "activeAt": "2025-10-31"
  }
]
```

---

## 🏗️ Project Structure

```
todo-list/
│
├── handlers/
│   ├── tasks.go        # All request handlers (create, update, delete, list)
│   └── utils.go        # Helper functions for JSON and validation
│
├── models/
│   └── task.go         # Task struct definition
│
├── main.go             # Application entry point
├── Dockerfile          # Docker build configuration
├── docker-compose.yml  # Optional Docker Compose setup
├── Makefile            # Useful shortcuts for build/run
└── README.md           # Documentation
```

---

## 🧰 Setup and Run

### ✅ Run locally

Make sure you have Go installed:

```bash
go run main.go
```

Server runs at:  
👉 [https://todo-list-la0q.onrender.com](https://todo-list-la0q.onrender.com)

---

### 🐳 Run with Docker

If you have Docker installed:

```bash
make build
make up
```

Or manually:

```bash
docker build -t todo-list .
docker run -p 8080:8080 todo-list
```

---

## 🌐 Deployment

You can deploy easily to platforms like **Render** or **Railway**, since the app already includes a `Dockerfile`.
Render will automatically detect and run it.

---
