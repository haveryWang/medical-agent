package store

import (
	"context"
	"time"

	"medical-agent/backend/internal/models"
	"medical-agent/backend/internal/security"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *MongoStore) FindUserByAccount(ctx context.Context, account string) (models.User, error) {
	var user models.User
	err := s.db.Collection("users").FindOne(ctx, bson.M{"account": account, "status": "active"}).Decode(&user)
	return user, err
}

func (s *MongoStore) FindUserByID(ctx context.Context, id primitive.ObjectID) (models.User, error) {
	var user models.User
	err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": id, "status": "active"}).Decode(&user)
	return user, err
}

func (s *MongoStore) CreateSession(ctx context.Context, userID primitive.ObjectID, token string, ttl time.Duration) (models.Session, error) {
	now := time.Now()
	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: security.TokenHash(token),
		ExpiresAt: now.Add(ttl),
		Revoked:   false,
		CreatedAt: now,
	}
	_, err := s.db.Collection("sessions").InsertOne(ctx, session)
	return session, err
}

func (s *MongoStore) FindSessionByToken(ctx context.Context, token string) (models.Session, error) {
	var session models.Session
	err := s.db.Collection("sessions").FindOne(ctx, bson.M{
		"tokenHash": security.TokenHash(token),
		"revoked":   false,
		"expiresAt": bson.M{"$gt": time.Now()},
	}).Decode(&session)
	return session, err
}

func (s *MongoStore) RevokeSession(ctx context.Context, token string) error {
	_, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"tokenHash": security.TokenHash(token)}, bson.M{"$set": bson.M{"revoked": true}})
	return err
}
