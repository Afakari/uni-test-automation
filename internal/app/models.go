package app

import "time"

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateTodoRequest struct {
	Title     *string `json:"title,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
}
