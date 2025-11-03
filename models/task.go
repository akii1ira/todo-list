package models

type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	ActiveAt  string `json:"activeAt"`
	Done      bool   `json:"-"`
	CreatedAt int64  `json:"-"`
}

const DateLayout = "2006-01-02"
