package mocks

import (
	"time"

	"github.com/ASH-WIN-10/snippetbox/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mockSnippetID = primitive.NewObjectID()

var mockSnippet = models.Snippet{
	ID:      mockSnippetID,
	Title:   "An old silent pond",
	Content: "An old silent pond...",
	Created: time.Now(),
	Expires: time.Now(),
}

type SnippetModel struct{}

func (m *SnippetModel) Insert(title, content string, expires int) (string, error) {
	return primitive.NewObjectID().Hex(), nil
}

func (m *SnippetModel) Get(id string) (models.Snippet, error) {
	if id == mockSnippetID.Hex() {
		return mockSnippet, nil
	}
	return models.Snippet{}, models.ErrNoRecord
}

func (m *SnippetModel) Latest() ([]models.Snippet, error) {
	return []models.Snippet{mockSnippet}, nil
}
