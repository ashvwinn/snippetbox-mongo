package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (string, error)
	Exists(id string) (bool, error)
	AddSnippet(userID, snippetID string) error
	IsSnippetOwner(userID, snippetID string) (bool, error)
	RemoveSnippet(userID, snippetID string) error
}

type User struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty"`
	Name           string               `bson:"name"`
	Email          string               `bson:"email"`
	HashedPassword []byte               `bson:"hashed_password"`
	Created        time.Time            `bson:"created"`
	Snippets       []primitive.ObjectID `bson:"snippets,omitempty"`
}

type UserModel struct {
	Users *mongo.Collection
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	doc := bson.M{
		"name":            name,
		"email":           email,
		"hashed_password": string(hashedPassword),
		"created":         time.Now().UTC(),
	}

	_, err = m.Users.InsertOne(context.TODO(), doc)
	if err != nil {
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, e := range writeErr.WriteErrors {
				if e.Code == 11000 {
					return ErrDuplicateEmail
				}
			}
		}
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (string, error) {
	var user struct {
		ID             primitive.ObjectID `bson:"_id"`
		HashedPassword string             `bson:"hashed_password"`
	}

	err := m.Users.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	return user.ID.Hex(), nil
}

func (m *UserModel) Exists(id string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, nil
	}

	count, err := m.Users.CountDocuments(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (m *UserModel) AddSnippet(userID, snippetID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	snippetObjID, err := primitive.ObjectIDFromHex(snippetID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": userObjID}
	update := bson.M{"$addToSet": bson.M{"snippets": snippetObjID}}

	_, err = m.Users.UpdateOne(context.TODO(), filter, update)
	return err
}

func (m *UserModel) IsSnippetOwner(userID, snippetID string) (bool, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	snippetObjID, err := primitive.ObjectIDFromHex(snippetID)
	if err != nil {
		return false, err
	}

	var user User
	err = m.Users.FindOne(context.TODO(), bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	for _, sID := range user.Snippets {
		if sID == snippetObjID {
			return true, nil
		}
	}

	return false, nil
}

func (m *UserModel) RemoveSnippet(userID, snippetID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	snippetObjID, err := primitive.ObjectIDFromHex(snippetID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": userObjID}
	update := bson.M{"$pull": bson.M{"snippets": snippetObjID}}

	_, err = m.Users.UpdateOne(context.TODO(), filter, update)
	return err
}
