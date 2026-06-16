package ingestion

import (
	"strings"
	"testing"
)

func TestChunkDocumentSplitsByHeadingAndParagraphUnderLimit(t *testing.T) {
	text := "总则\n" +
		strings.Repeat("甲", 59) + "。" +
		strings.Repeat("乙", 59) + "！" +
		strings.Repeat("丙", 59) + "？" +
		strings.Repeat("丁", 59) + "。"

	chunks := chunkDocument(text, "诊疗规范.pdf", 200)

	if len(chunks) != 2 {
		t.Fatalf("expected sentence aware chunks, got %#v", chunks)
	}
	if got := len([]rune(chunks[0].Text)); got != 180 {
		t.Fatalf("first chunk has %d runes, want 180", got)
	}
	if got := len([]rune(chunks[1].Text)); got != 60 {
		t.Fatalf("second chunk has %d runes, want 60", got)
	}
	if !strings.HasSuffix(chunks[0].Text, "？") {
		t.Fatalf("expected first chunk to end on a sentence boundary, got %q", chunks[0].Text)
	}
	if !strings.HasPrefix(chunks[1].Text, "丁") {
		t.Fatalf("expected second chunk to start with the next full sentence, got %q", chunks[1].Text)
	}
	for _, chunk := range chunks {
		if got := len([]rune(chunk.Text)); got > 500 {
			t.Fatalf("chunk %q has %d runes, want <= 500", chunk.Text, got)
		}
		if chunk.Section != "总则" {
			t.Fatalf("expected section 总则, got %q", chunk.Section)
		}
	}
}

func TestChunkDocumentKeepsShortRowsUnderExplicitSection(t *testing.T) {
	chunks := chunkDocument("工作表: 门诊\n项目 结果\n血压 120/80\n血糖 6.1", "门诊数据.xlsx", 200)

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

func TestChunkDocumentCapsVeryLongSentenceAtHardLimit(t *testing.T) {
	chunks := chunkDocument(strings.Repeat("长", 620)+"。"+"短句。", "长文档.txt", 200)

	if len(chunks) != 3 {
		t.Fatalf("expected long sentence to be capped and trailing sentence kept, got %#v", chunks)
	}
	for _, chunk := range chunks {
		if chunk.Section != "长文档" {
			t.Fatalf("expected section 长文档, got %q", chunk.Section)
		}
		if got := len([]rune(chunk.Text)); got > 500 {
			t.Fatalf("chunk has %d runes, want <= 500", got)
		}
	}
	if got := len([]rune(chunks[0].Text)); got != 500 {
		t.Fatalf("first long-sentence chunk has %d runes, want 500", got)
	}
	if !strings.HasSuffix(chunks[1].Text, "。") {
		t.Fatalf("expected remainder to preserve sentence punctuation, got %q", chunks[1].Text)
	}
	if chunks[2].Text != "短句。" {
		t.Fatalf("expected trailing sentence to remain intact, got %q", chunks[2].Text)
	}
}
