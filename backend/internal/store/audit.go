package store

import (
	"context"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *MongoStore) Audit(ctx context.Context, actorID primitive.ObjectID, action string, target string, result string, requestID string) {
	_, _ = s.db.Collection("audit_logs").InsertOne(ctx, models.AuditLog{
		ID:        primitive.NewObjectID(),
		ActorID:   actorID,
		Action:    action,
		Target:    target,
		Result:    result,
		RequestID: requestID,
		CreatedAt: time.Now(),
	})
}
