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

	kbs := []models.KnowledgeBase{
		{
			ID:              primitive.NewObjectID(),
			Name:            "糖尿病诊疗知识库",
			Description:     "包含各类疾病的诊疗规范和指南",
			Scenario:        "临床诊疗",
			Tags:            []string{"诊疗", "规范", "指南"},
			Department:      "医务部",
			Status:          "active",
			BuildStatus:     "completed",
			DocumentCount:   128,
			ChunkCount:      0,
			RetrievalTopK:   5,
			SimilarityFloor: 0.2,
			CreatedAt:       now,
			UpdatedAt:       now.Add(-48 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			Name:            "药品信息库",
			Description:     "药品说明书、用药指南等信息",
			Scenario:        "药学服务",
			Tags:            []string{"药品", "说明书"},
			Department:      "药剂科",
			Status:          "active",
			BuildStatus:     "completed",
			DocumentCount:   356,
			RetrievalTopK:   5,
			SimilarityFloor: 0.2,
			CreatedAt:       now,
			UpdatedAt:       now.Add(-72 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			Name:            "医保政策库",
			Description:     "医保政策、报销指南等",
			Scenario:        "医保管理",
			Tags:            []string{"医保", "政策"},
			Department:      "医保办",
			Status:          "active",
			BuildStatus:     "completed",
			DocumentCount:   89,
			RetrievalTopK:   5,
			SimilarityFloor: 0.2,
			CreatedAt:       now,
			UpdatedAt:       now.Add(-96 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			Name:            "检验检查参考库",
			Description:     "各项检查检验的参考范围和解读",
			Scenario:        "检验检查",
			Tags:            []string{"检验", "参考范围"},
			Department:      "检验科",
			Status:          "active",
			BuildStatus:     "building",
			DocumentCount:   234,
			RetrievalTopK:   5,
			SimilarityFloor: 0.2,
			CreatedAt:       now,
			UpdatedAt:       now.Add(-120 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			Name:            "护理操作规范库",
			Description:     "护理操作流程和规范",
			Scenario:        "护理管理",
			Tags:            []string{"护理", "操作规范"},
			Department:      "护理部",
			Status:          "active",
			BuildStatus:     "completed",
			DocumentCount:   167,
			RetrievalTopK:   5,
			SimilarityFloor: 0.2,
			CreatedAt:       now,
			UpdatedAt:       now.Add(-168 * time.Hour),
		},
	}
	records := make([]any, 0, len(kbs))
	for _, kb := range kbs {
		records = append(records, kb)
	}
	if _, err := s.db.Collection("knowledge_bases").InsertMany(ctx, records); err != nil {
		return err
	}

	demoDocID := primitive.NewObjectID()
	_, err = s.db.Collection("documents").InsertOne(ctx, models.Document{
		ID:              demoDocID,
		KnowledgeBaseID: kbs[0].ID,
		FileName:        "2型糖尿病防治指南2023.pdf",
		FileType:        ".pdf",
		SizeBytes:       2400 * 1024,
		StoragePath:     "seed://diabetes-guide",
		Status:          "completed",
		CreatedAt:       now.Add(-48 * time.Hour),
		UpdatedAt:       now.Add(-48 * time.Hour),
	})
	if err != nil {
		return err
	}
	chunk := models.Chunk{
		ID:              primitive.NewObjectID(),
		KnowledgeBaseID: kbs[0].ID,
		DocumentID:      demoDocID,
		Text:            "2型糖尿病治疗应包括生活方式干预、血糖监测、个体化药物治疗和定期随访。常用指标包括空腹血糖、餐后2小时血糖、糖化血红蛋白、血脂、肾功能等。",
		Section:         "诊疗建议",
		ChunkIndex:      0,
		VectorID:        "seed-diabetes-0",
		Checksum:        "seed",
		CreatedAt:       now,
	}
	_, err = s.db.Collection("chunks").InsertOne(ctx, chunk)
	return err
}
