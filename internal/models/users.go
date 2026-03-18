package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
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
		// Handle duplicate email
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, e := range writeErr.WriteErrors {
				if e.Code == 11000 { // duplicate key error
					return ErrDuplicateEmail
				}
			}
		}
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var user struct {
		ID             int    `bson:"id"`
		HashedPassword string `bson:"hashed_password"`
	}

	filter := bson.M{
		"email": email,
	}

	err := m.Users.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	return user.ID, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	count, err := m.Users.CountDocuments(context.TODO(), bson.M{"id": id})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
