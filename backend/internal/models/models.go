package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Account      string             `bson:"account" json:"account"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	DisplayName  string             `bson:"displayName" json:"displayName"`
	Roles        []string           `bson:"roles" json:"roles"`
	Permissions  []string           `bson:"permissions" json:"permissions"`
	Status       string             `bson:"status" json:"status"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Session struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	TokenHash string             `bson:"tokenHash" json:"-"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
	Revoked   bool               `bson:"revoked" json:"revoked"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type KnowledgeBase struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Description     string             `bson:"description" json:"description"`
	Scenario        string             `bson:"scenario" json:"scenario"`
	Tags            []string           `bson:"tags" json:"tags"`
	Department      string             `bson:"department" json:"department"`
	Status          string             `bson:"status" json:"status"`
	BuildStatus     string             `bson:"buildStatus" json:"buildStatus"`
	DocumentCount   int                `bson:"documentCount" json:"documentCount"`
	ChunkCount      int                `bson:"chunkCount" json:"chunkCount"`
	RetrievalTopK   int                `bson:"retrievalTopK" json:"retrievalTopK"`
	SimilarityFloor float64            `bson:"similarityFloor" json:"similarityFloor"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Document struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	KnowledgeBaseID primitive.ObjectID `bson:"knowledgeBaseId" json:"knowledgeBaseId"`
	FileName        string             `bson:"fileName" json:"fileName"`
	FileType        string             `bson:"fileType" json:"fileType"`
	SizeBytes       int64              `bson:"sizeBytes" json:"sizeBytes"`
	StoragePath     string             `bson:"storagePath" json:"storagePath"`
	Status          string             `bson:"status" json:"status"`
	FailureReason   string             `bson:"failureReason,omitempty" json:"failureReason,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Chunk struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	KnowledgeBaseID primitive.ObjectID `bson:"knowledgeBaseId" json:"knowledgeBaseId"`
	DocumentID      primitive.ObjectID `bson:"documentId" json:"documentId"`
	Text            string             `bson:"text" json:"text"`
	Section         string             `bson:"section" json:"section"`
	ChunkIndex      int                `bson:"chunkIndex" json:"chunkIndex"`
	VectorID        string             `bson:"vectorId" json:"vectorId"`
	Checksum        string             `bson:"checksum" json:"checksum"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}

type IngestionJob struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	KnowledgeBaseID primitive.ObjectID `bson:"knowledgeBaseId" json:"knowledgeBaseId"`
	DocumentID      primitive.ObjectID `bson:"documentId" json:"documentId"`
	Status          string             `bson:"status" json:"status"`
	Step            string             `bson:"step" json:"step"`
	Attempts        int                `bson:"attempts" json:"attempts"`
	Error           string             `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Conversation struct {
	ID               primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID   `bson:"userId" json:"userId"`
	Title            string               `bson:"title" json:"title"`
	Status           string               `bson:"status" json:"status"`
	KnowledgeBaseIDs []primitive.ObjectID `bson:"knowledgeBaseIds" json:"knowledgeBaseIds"`
	CreatedAt        time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time            `bson:"updatedAt" json:"updatedAt"`
}

type Citation struct {
	ChunkID         primitive.ObjectID `bson:"chunkId" json:"chunkId"`
	DocumentID      primitive.ObjectID `bson:"documentId" json:"documentId"`
	KnowledgeBaseID primitive.ObjectID `bson:"knowledgeBaseId" json:"knowledgeBaseId"`
	Title           string             `bson:"title" json:"title"`
	Snippet         string             `bson:"snippet" json:"snippet"`
	Score           float64            `bson:"score" json:"score"`
}

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID `bson:"conversationId" json:"conversationId"`
	Role           string             `bson:"role" json:"role"`
	Content        string             `bson:"content" json:"content"`
	Status         string             `bson:"status" json:"status"`
	Citations      []Citation         `bson:"citations" json:"citations"`
	ModelName      string             `bson:"modelName" json:"modelName"`
	PromptContext  string             `bson:"promptContext" json:"promptContext"`
	DurationMS     int64              `bson:"durationMs" json:"durationMs"`
	TokenUsage     int                `bson:"tokenUsage" json:"tokenUsage"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ActorID   primitive.ObjectID `bson:"actorId" json:"actorId"`
	Action    string             `bson:"action" json:"action"`
	Target    string             `bson:"target" json:"target"`
	Result    string             `bson:"result" json:"result"`
	RequestID string             `bson:"requestId" json:"requestId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type ModelConfig struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DeepSeekBaseURL        string             `bson:"deepSeekBaseUrl" json:"deepSeekBaseUrl"`
	DeepSeekAPIKey         string             `bson:"deepSeekAPIKey,omitempty" json:"deepSeekAPIKey,omitempty"`
	DeepSeekChatModel      string             `bson:"deepSeekChatModel" json:"deepSeekChatModel"`
	QwenEmbeddingBaseURL   string             `bson:"qwenEmbeddingBaseUrl" json:"qwenEmbeddingBaseUrl"`
	QwenEmbeddingAPIKey    string             `bson:"qwenEmbeddingAPIKey,omitempty" json:"qwenEmbeddingAPIKey,omitempty"`
	QwenEmbeddingModel     string             `bson:"qwenEmbeddingModel" json:"qwenEmbeddingModel"`
	QwenEmbeddingDimension int                `bson:"qwenEmbeddingDimension" json:"qwenEmbeddingDimension"`
	UpdatedAt              time.Time          `bson:"updatedAt" json:"updatedAt"`
}
