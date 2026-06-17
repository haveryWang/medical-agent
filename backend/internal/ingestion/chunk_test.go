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

func TestChunkDocumentPreservesShortChineseListItems(t *testing.T) {
	text := strings.Join([]string{
		"会场保障",
		"提前15分钟到，叮嘱会务至少提前30分钟开设备空调",
		"设备调试：话筒声音、PPT、211自开设备",
		"第一排矿水",
		"会议时间地点更改：及时通知调研团内的领导们（7人）",
		"电梯的选择偏好：16号楼电梯10梯，8号楼13梯（因侧面提车），20号楼3楼通到21号楼，20号楼外面1-3梯（1梯到6楼检验科），里面4-6梯（到高层）",
		"保卫保障：停车偏好11号楼侧面",
		"与部门负责人沟通",
		"初次沟通：了解到场人员数量、“雷点”",
		"业务科室联系建议询问主任科室对接人",
		"临时变更时的沟通方式",
	}, "\n")

	chunks := chunkDocument(text, "陪同经验总结.docx", 200)
	if len(chunks) < 2 {
		t.Fatalf("expected Chinese list items to split across multiple chunks, got %#v", chunks)
	}
	allText := ""
	for _, chunk := range chunks {
		allText += chunk.Text + "\n"
	}

	for _, wanted := range []string{
		"提前15分钟到，叮嘱会务至少提前30分钟开设备空调",
		"设备调试：话筒声音、PPT、211自开设备",
		"第一排矿水",
		"会议时间地点更改：及时通知调研团内的领导们（7人）",
		"保卫保障：停车偏好11号楼侧面",
		"业务科室联系建议询问主任科室对接人",
	} {
		if !strings.Contains(allText, wanted) {
			t.Fatalf("expected chunk text to preserve %q, got %#v", wanted, chunks)
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
