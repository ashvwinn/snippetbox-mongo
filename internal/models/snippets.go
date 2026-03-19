package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SnippetModelInterface interface {
	Insert(title, content string, expires int, userId string) (string, error)
	Get(id string) (Snippet, error)
	GetByUserID(userID string) ([]Snippet, error)
	Latest() ([]Snippet, error)
	Delete(id string) error
	Update(id, title, content string, expires int) error
}

type Snippet struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Title   string             `bson:"title"`
	Content string             `bson:"content"`
	Created time.Time          `bson:"created"`
	Expires time.Time          `bson:"expires"`
	UserID  primitive.ObjectID `bson:"user_id,omitempty"`
}

type SnippetModel struct {
	Snippets *mongo.Collection
}

func (m *SnippetModel) Insert(title string, content string, expires int, userId string) (string, error) {
	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()

	doc := bson.M{
		"title":   title,
		"content": content,
		"created": now,
		"expires": now.AddDate(0, 0, expires),
		"user_id": userObjID,
	}

	res, err := m.Snippets.InsertOne(context.TODO(), doc)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m *SnippetModel) Get(id string) (Snippet, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Snippet{}, ErrNoRecord
	}

	var s Snippet

	filter := bson.M{
		"_id":     objID,
		"expires": bson.M{"$gt": time.Now().UTC()},
	}

	err = m.Snippets.FindOne(context.TODO(), filter).Decode(&s)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Snippet{}, ErrNoRecord
		}
		return Snippet{}, err
	}

	return s, nil
}

func (m *SnippetModel) GetByUserID(userID string) ([]Snippet, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"user_id": userObjID,
	}

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}})

	cursor, err := m.Snippets.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var snippets []Snippet

	for cursor.Next(context.TODO()) {
		var s Snippet
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	filter := bson.M{
		"expires": bson.M{"$gt": time.Now().UTC()},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(10)

	cursor, err := m.Snippets.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var snippets []Snippet

	for cursor.Next(context.TODO()) {
		var s Snippet
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}

func (m *SnippetModel) Delete(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNoRecord
	}

	res, err := m.Snippets.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return ErrNoRecord
	}

	return nil
}

func (m *SnippetModel) Update(id, title, content string, expires int) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNoRecord
	}

	update := bson.M{
		"$set": bson.M{
			"title":   title,
			"content": content,
			"expires": time.Now().UTC().AddDate(0, 0, expires),
		},
	}

	res, err := m.Snippets.UpdateOne(context.TODO(), bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return ErrNoRecord
	}

	return nil
}
