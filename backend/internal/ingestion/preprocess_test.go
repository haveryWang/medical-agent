package ingestion

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"medical-agent/backend/internal/models"

	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
)

func TestExtractTextPreprocessesMarkdown(t *testing.T) {
	doc := models.Document{
		FileName: "clinical.md",
		FileType: ".md",
		Content: []byte(`# 发热处理

参考 [指南](https://example.test/guide)，图片：![流程图](https://example.test/image.png)

| 病种 | 处置 |
| --- | --- |
| 发热 | 观察 |

` + "```go" + `
fmt.Println("ok")
` + "```" + `
`),
	}

	text, err := extractText(doc)
	if err != nil {
		t.Fatalf("extractText returned error: %v", err)
	}
	for _, unwanted := range []string{"# 发热处理", "https://example.test", "| --- |", "```"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("expected markdown syntax %q to be removed, got:\n%s", unwanted, text)
		}
	}
	for _, wanted := range []string{"发热处理", "指南", "流程图", "发热 观察", `fmt.Println("ok")`} {
		if !strings.Contains(text, wanted) {
			t.Fatalf("expected extracted markdown text to contain %q, got:\n%s", wanted, text)
		}
	}
}

func TestExtractTextReadsDocxDocumentText(t *testing.T) {
	doc := models.Document{
		FileName: "record.docx",
		FileType: ".docx",
		Content: makeDocx(t, map[string]string{
			"word/document.xml": `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>主诉</w:t></w:r><w:r><w:tab/></w:r><w:r><w:t>头痛三天</w:t></w:r></w:p>
    <w:p><w:r><w:t>处理建议</w:t></w:r><w:r><w:br/></w:r><w:r><w:t>复查血压</w:t></w:r></w:p>
  </w:body>
</w:document>`,
		}),
	}

	text, err := extractText(doc)
	if err != nil {
		t.Fatalf("extractText returned error: %v", err)
	}
	for _, wanted := range []string{"主诉 头痛三天", "处理建议", "复查血压"} {
		if !strings.Contains(text, wanted) {
			t.Fatalf("expected docx text to contain %q, got:\n%s", wanted, text)
		}
	}
}

func TestExtractTextReadsPDFDocumentText(t *testing.T) {
	text, err := extractText(models.Document{
		FileName: "guide.pdf",
		FileType: ".pdf",
		Content:  makeSimplePDF("PDF Fever Advice"),
	})
	if err != nil {
		t.Fatalf("extractText returned error: %v", err)
	}
	if !strings.Contains(text, "PDF Fever Advice") {
		t.Fatalf("expected PDF text to be extracted, got:\n%s", text)
	}
}

func TestExtractTextReadsXLSXSheets(t *testing.T) {
	book := excelize.NewFile()
	defer book.Close()
	index, err := book.NewSheet("门诊")
	if err != nil {
		t.Fatalf("NewSheet failed: %v", err)
	}
	book.SetActiveSheet(index)
	if err := book.SetSheetRow("门诊", "A1", &[]any{"项目", "结果"}); err != nil {
		t.Fatalf("SetSheetRow header failed: %v", err)
	}
	if err := book.SetSheetRow("门诊", "A2", &[]any{"血压", "120/80"}); err != nil {
		t.Fatalf("SetSheetRow body failed: %v", err)
	}
	var buf bytes.Buffer
	if err := book.Write(&buf); err != nil {
		t.Fatalf("Write workbook failed: %v", err)
	}

	text, err := extractText(models.Document{
		FileName: "lab.xlsx",
		FileType: ".xlsx",
		Content:  buf.Bytes(),
	})
	if err != nil {
		t.Fatalf("extractText returned error: %v", err)
	}
	for _, wanted := range []string{"工作表: 门诊", "项目 结果", "血压 120/80"} {
		if !strings.Contains(text, wanted) {
			t.Fatalf("expected xlsx text to contain %q, got:\n%s", wanted, text)
		}
	}
}

func TestExtractTextReadsLegacyXLSFixture(t *testing.T) {
	data := readXLSFixture(t)
	text, err := extractText(models.Document{
		FileName: "legacy.xls",
		FileType: ".xls",
		Content:  data,
	})
	if err != nil {
		t.Fatalf("extractText returned error: %v", err)
	}
	if !strings.Contains(text, "工作表:") {
		t.Fatalf("expected xls text to include sheet names, got:\n%s", text)
	}
}

func TestExtractTextNormalizesLongTXT(t *testing.T) {
	doc := models.Document{
		FileName: "notes.txt",
		FileType: ".txt",
		Content:  []byte("第一段   内容\r\n\r\n\r\n第二段\t内容"),
	}

	text, err := extractText(doc)
	if err != nil {
		t.Fatalf("extractText returned error: %v", err)
	}
	if strings.Contains(text, "\r") || strings.Contains(text, "\n\n\n") || strings.Contains(text, "第一段   内容") {
		t.Fatalf("expected TXT text to be normalized, got:\n%q", text)
	}
	for _, wanted := range []string{"第一段 内容", "第二段 内容"} {
		if !strings.Contains(text, wanted) {
			t.Fatalf("expected TXT text to contain %q, got:\n%s", wanted, text)
		}
	}
}

func makeDocx(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s failed: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s failed: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip failed: %v", err)
	}
	return buf.Bytes()
}

func makeSimplePDF(text string) []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 6)

	writeObject := func(id int, content string) {
		offsets[id] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, content)
	}

	stream := fmt.Sprintf("BT /F1 24 Tf 100 700 Td (%s) Tj ET", escapePDFString(text))
	writeObject(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObject(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObject(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>")
	writeObject(4, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	writeObject(5, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream))

	xrefOffset := buf.Len()
	buf.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for id := 1; id <= 5; id++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[id])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset)
	return buf.Bytes()
}

func readXLSFixture(t *testing.T) []byte {
	t.Helper()
	pc := reflect.ValueOf(xls.OpenReader).Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		t.Fatal("failed to locate xls module source")
	}
	file, _ := fn.FileLine(pc)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "Table.xls"))
	if err != nil {
		t.Fatalf("read xls fixture failed: %v", err)
	}
	return data
}

func escapePDFString(text string) string {
	text = strings.ReplaceAll(text, `\`, `\\`)
	text = strings.ReplaceAll(text, "(", `\(`)
	text = strings.ReplaceAll(text, ")", `\)`)
	return text
}
