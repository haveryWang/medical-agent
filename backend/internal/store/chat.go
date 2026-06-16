package store

import (
	"context"
	"strings"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoStore) CreateConversation(ctx context.Context, userID primitive.ObjectID, title string, kbIDs []primitive.ObjectID) (models.Conversation, error) {
	now := time.Now()
	if strings.TrimSpace(title) == "" {
		title = "新的问答会话"
	}
	conversation := models.Conversation{
		ID:               primitive.NewObjectID(),
		UserID:           userID,
		Title:            title,
		Status:           "active",
		KnowledgeBaseIDs: kbIDs,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	_, err := s.db.Collection("conversations").InsertOne(ctx, conversation)
	return conversation, err
}

func (s *MongoStore) ListConversations(ctx context.Context, userID primitive.ObjectID, keyword string) ([]models.Conversation, error) {
	query := bson.M{"userId": userID, "status": "active"}
	if keyword != "" {
		query["title"] = bson.M{"$regex": keyword, "$options": "i"}
	}
	cursor, err := s.db.Collection("conversations").Find(ctx, query, options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetLimit(50))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var conversations []models.Conversation
	if err := cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}
	return conversations, nil
}

func (s *MongoStore) GetConversation(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (models.Conversation, error) {
	var conversation models.Conversation
	err := s.db.Collection("conversations").FindOne(ctx, bson.M{"_id": id, "userId": userID}).Decode(&conversation)
	return conversation, err
}

func (s *MongoStore) CreateMessage(ctx context.Context, message models.Message) (models.Message, error) {
	now := time.Now()
	message.ID = primitive.NewObjectID()
	message.CreatedAt = now
	message.UpdatedAt = now
	_, err := s.db.Collection("messages").InsertOne(ctx, message)
	if err != nil {
		return message, err
	}
	_, _ = s.db.Collection("conversations").UpdateOne(ctx, bson.M{"_id": message.ConversationID}, bson.M{"$set": bson.M{"updatedAt": now}})
	return message, nil
}

func (s *MongoStore) UpdateMessage(ctx context.Context, id primitive.ObjectID, patch bson.M) error {
	patch["updatedAt"] = time.Now()
	_, err := s.db.Collection("messages").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": patch})
	return err
}

func (s *MongoStore) ListMessages(ctx context.Context, conversationID primitive.ObjectID) ([]models.Message, error) {
	cursor, err := s.db.Collection("messages").Find(ctx, bson.M{"conversationId": conversationID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MongoStore) GetMessage(ctx context.Context, id primitive.ObjectID) (models.Message, error) {
	var message models.Message
	err := s.db.Collection("messages").FindOne(ctx, bson.M{"_id": id}).Decode(&message)
	return message, err
}
