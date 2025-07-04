package models

import (
	"time"
)

type Memo struct {
	ID           string     `firestore:"id" json:"id"`
	Title        string     `firestore:"title" json:"title"`
	Description  string     `firestore:"description" json:"description"`
	Tags         []string   `firestore:"tags" json:"tags"`
	LinkedTodos  []string   `firestore:"linked_todos" json:"linked_todos"`
	CreatedAt    time.Time  `firestore:"created_at" json:"created_at"`
	LastModified time.Time  `firestore:"last_modified" json:"last_modified"`
	ClosedAt     *time.Time `firestore:"closed_at,omitempty" json:"closed_at,omitempty"`
}
