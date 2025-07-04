package models

import (
	"time"
)

type TodoStatus string

const (
	StatusBacklog    TodoStatus = "backlog"
	StatusTodo       TodoStatus = "todo"
	StatusInProgress TodoStatus = "in_progress"
	StatusDone       TodoStatus = "done"
)

type TodoPriority string

const (
	PriorityHigh   TodoPriority = "high"
	PriorityNormal TodoPriority = "normal"
)

type Todo struct {
	ID           string       `firestore:"id" json:"id"`
	Title        string       `firestore:"title" json:"title"`
	Description  string       `firestore:"description" json:"description"`
	Status       TodoStatus   `firestore:"status" json:"status"`
	Priority     TodoPriority `firestore:"priority" json:"priority"`
	Tags         []string     `firestore:"tags" json:"tags"`
	ParentID     string       `firestore:"parent_id,omitempty" json:"parent_id,omitempty"`
	CreatedAt    time.Time    `firestore:"created_at" json:"created_at"`
	LastModified time.Time    `firestore:"last_modified" json:"last_modified"`
	ClosedAt     *time.Time   `firestore:"closed_at,omitempty" json:"closed_at,omitempty"`
}
