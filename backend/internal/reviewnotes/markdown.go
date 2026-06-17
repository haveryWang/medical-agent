package reviewnotes

import (
	"fmt"
	"strings"
	"time"

	"medical-agent/backend/internal/models"
)

func RenderMarkdown(notes []models.ReviewNote, generatedAt time.Time) string {
	var builder strings.Builder
	builder.WriteString("# 复盘笔记导出\n\n")
	builder.WriteString("生成时间：")
	builder.WriteString(formatTime(generatedAt))
	builder.WriteString("\n\n")
	builder.WriteString(fmt.Sprintf("记录数：%d\n\n", len(notes)))
	location := generatedAt.Location()
	for index, note := range notes {
		builder.WriteString(fmt.Sprintf("## %d. %s\n\n", index+1, formatTime(note.CreatedAt.In(location))))
		builder.WriteString(strings.TrimSpace(note.Content))
		builder.WriteString("\n\n")
	}
	return strings.TrimSpace(builder.String()) + "\n"
}

func ExportFilename(generatedAt time.Time) string {
	return fmt.Sprintf("复盘笔记-%s.md", generatedAt.Format("20060102-150405"))
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02 15:04:05")
}
