package ingestion

import (
	"strings"
	"testing"
)

func TestChunkDocumentSplitsByHeadingAndParagraphUnderLimit(t *testing.T) {
	longParagraph := strings.Repeat("糖尿病患者需要评估血糖血压血脂并结合并发症风险制定随访计划。", 8)
	text := "总则\n" +
		"第一段强调建立个人健康档案并完成初始评估。\n\n" +
		longParagraph + "\n\n" +
		"诊疗建议\n" +
		"根据病情分层选择生活方式干预、药物治疗和复诊频率。"

	chunks := chunkDocument(text, 200)

	if len(chunks) < 4 {
		t.Fatalf("expected heading and paragraph aware chunks, got %#v", chunks)
	}
	for _, chunk := range chunks {
		if got := len([]rune(chunk.Text)); got > 200 {
			t.Fatalf("chunk %q has %d runes, want <= 200", chunk.Text, got)
		}
		if strings.Contains(chunk.Text, "第一段强调") && chunk.Section != "总则" {
			t.Fatalf("expected first paragraph section 总则, got %q", chunk.Section)
		}
		if strings.Contains(chunk.Text, "根据病情分层") && chunk.Section != "诊疗建议" {
			t.Fatalf("expected treatment paragraph section 诊疗建议, got %q", chunk.Section)
		}
	}
	if strings.Contains(chunks[0].Text, "诊疗建议") {
		t.Fatalf("expected chunks to split before next heading, got %q", chunks[0].Text)
	}
}

func TestChunkDocumentKeepsShortRowsUnderExplicitSection(t *testing.T) {
	chunks := chunkDocument("工作表: 门诊\n项目 结果\n血压 120/80\n血糖 6.1", 200)

	if len(chunks) != 1 {
		t.Fatalf("expected short rows to stay in one paragraph chunk, got %#v", chunks)
	}
	if chunks[0].Section != "工作表: 门诊" {
		t.Fatalf("expected sheet section, got %q", chunks[0].Section)
	}
	for _, wanted := range []string{"项目 结果", "血压 120/80", "血糖 6.1"} {
		if !strings.Contains(chunks[0].Text, wanted) {
			t.Fatalf("expected chunk to contain %q, got %q", wanted, chunks[0].Text)
		}
	}
}
