# Todo List API (Go)

## Overview

RESTful microservice for managing todo tasks. Implemented in Go, stores data in memory (sync.Map). Provides endpoints to create, update, delete, mark done and list tasks.

## Requirements

- Go
- Docker
- Docker Compose
- (Optional) Render for deployment

## Endpoints

- **POST** `/api/todo-list/tasks`  
  Body:
  ```json
  { "title": "Buy book", "activeAt": "2023-08-04" }
  ```
