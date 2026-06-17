package reviewnotes

import (
	"strings"
	"testing"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRenderMarkdownIncludesNoteContentAndTimestamps(t *testing.T) {
	generatedAt := time.Date(2026, 6, 17, 9, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	notes := []models.ReviewNote{
		{
			ID:        primitive.NewObjectID(),
			Content:   "晨会复盘：先确认政策适用范围，再回答执行口径。",
			CreatedAt: time.Date(2026, 6, 16, 17, 5, 0, 0, generatedAt.Location()),
		},
		{
			ID:        primitive.NewObjectID(),
			Content:   "医保问题需要保留原始文件出处。",
			CreatedAt: time.Date(2026, 6, 17, 8, 15, 0, 0, generatedAt.Location()),
		},
	}

	markdown := RenderMarkdown(notes, generatedAt)
	for _, want := range []string{
		"# 复盘笔记导出",
		"生成时间：2026-06-17 09:30:00",
		"记录数：2",
		"## 1. 2026-06-16 17:05:00",
		"晨会复盘：先确认政策适用范围，再回答执行口径。",
		"## 2. 2026-06-17 08:15:00",
		"医保问题需要保留原始文件出处。",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown missing %q:\n%s", want, markdown)
		}
	}
}

func TestRenderMarkdownFormatsNoteTimesInGeneratedTimezone(t *testing.T) {
	generatedAt := time.Date(2026, 6, 17, 12, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	notes := []models.ReviewNote{{
		ID:        primitive.NewObjectID(),
		Content:   "跨时区时间应按导出时区展示。",
		CreatedAt: time.Date(2026, 6, 17, 4, 30, 0, 0, time.UTC),
	}}

	markdown := RenderMarkdown(notes, generatedAt)
	if !strings.Contains(markdown, "## 1. 2026-06-17 12:30:00") {
		t.Fatalf("markdown should format note time in generated timezone:\n%s", markdown)
	}
}

func TestExportFilenameUsesGeneratedDate(t *testing.T) {
	generatedAt := time.Date(2026, 6, 17, 9, 30, 0, 0, time.UTC)
	if got := ExportFilename(generatedAt); got != "复盘笔记-20260617-093000.md" {
		t.Fatalf("filename = %q", got)
	}
}
