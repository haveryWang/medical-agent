package store

import (
	"context"
	"time"

	"medical-agent/backend/internal/models"
	"medical-agent/backend/internal/security"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *MongoStore) Seed(ctx context.Context) error {
	count, err := s.db.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now()
	passwordHash, err := security.HashPassword("admin123")
	if err != nil {
		return err
	}
	adminID := primitive.NewObjectID()
	_, err = s.db.Collection("users").InsertOne(ctx, models.User{
		ID:           adminID,
		Account:      "admin",
		PasswordHash: passwordHash,
		DisplayName:  "张医生",
		Roles:        []string{"系统管理员", "知识库管理员"},
		Permissions:  []string{"chat:use", "knowledge:read", "knowledge:write", "system:read"},
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return err
	}
	return nil
}
