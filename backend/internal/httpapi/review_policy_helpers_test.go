package httpapi

import (
	"context"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func requestWithUser(ctx context.Context, user models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func testUser() models.User {
	return models.User{
		ID:          primitive.NewObjectID(),
		Account:     "admin",
		DisplayName: "管理员",
		Permissions: []string{"review_notes:write", "policy:write"},
		Status:      "active",
	}
}
