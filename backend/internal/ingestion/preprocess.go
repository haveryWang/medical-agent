package ingestion

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"medical-agent/backend/internal/models"

	"github.com/dslipak/pdf"
	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
)

var (
	markdownFenceRe   = regexp.MustCompile("(?s)```[^\\n`]*\\n?(.*?)```")
	markdownImageRe   = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	markdownLinkRe    = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	markdownHeadingRe = regexp.MustCompile(`(?m)^\s{0,3}#{1,6}\s*`)
)

func ExtractText(doc models.Document) (string, error) {
	data, err := documentBytes(doc)
	if err != nil {
		return "", err
	}

	ext := documentExtension(doc)
	var text string
	switch ext {
	case ".txt", ".text":
		text, err = extractPlainText(data)
	case ".md", ".markdown":
		text, err = extractMarkdown(data)
	case ".pdf":
		text, err = extractPDF(data)
	case ".docx":
		text, err = extractDOCX(data)
	case ".xlsx":
		text, err = extractXLSX(data)
	case ".xls":
		text, err = extractXLS(data)
	case ".csv":
		text, err = extractCSV(data)
	case ".doc":
		return "", errors.New("当前仅支持 Word .docx 文件，.doc 请先转换为 .docx 后上传")
	default:
		return "", fmt.Errorf("暂不支持 %s 文件，请上传 PDF、Word(.docx)、Excel(.xlsx/.xls)、Markdown、TXT 或 CSV", ext)
	}
	if err != nil {
		return "", err
	}

	text = normalizeExtractedText(text)
	if text == "" {
		return "", errors.New("当前文件未提取到可向量化文本，请确认文件不是扫描件或空文档")
	}
	if !looksLikeText(text) {
		return "", errors.New("当前文件未提取到可向量化文本，请上传可解析的文本型文档")
	}
	return text, nil
}

func extractText(doc models.Document) (string, error) {
	return ExtractText(doc)
}

func documentBytes(doc models.Document) ([]byte, error) {
	if len(doc.Content) > 0 {
		return doc.Content, nil
	}
	if strings.TrimSpace(doc.StoragePath) != "" {
		data, err := os.ReadFile(doc.StoragePath)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			return data, nil
		}
	}
	return nil, errors.New("当前文件内容为空，无法生成向量")
}

func documentExtension(doc models.Document) string {
	if doc.FileType != "" {
		ext := strings.ToLower(strings.TrimSpace(doc.FileType))
		if strings.HasPrefix(ext, ".") {
			return ext
		}
		return "." + ext
	}
	return strings.ToLower(filepath.Ext(doc.FileName))
}

func extractPlainText(data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", errors.New("当前文本文件不是 UTF-8 编码，请转换编码后重新上传")
	}
	return string(data), nil
}

func extractMarkdown(data []byte) (string, error) {
	text, err := extractPlainText(data)
	if err != nil {
		return "", err
	}
	return cleanMarkdown(text), nil
}

func cleanMarkdown(text string) string {
	text = markdownFenceRe.ReplaceAllString(text, "$1")
	text = markdownImageRe.ReplaceAllString(text, "$1")
	text = markdownLinkRe.ReplaceAllString(text, "$1")
	text = markdownHeadingRe.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "~~", "")

	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		if isMarkdownTableSeparator(line) {
			continue
		}
		line = strings.TrimLeft(line, "> ")
		line = strings.ReplaceAll(line, "|", " ")
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}

func isMarkdownTableSeparator(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.Contains(line, "|") || !strings.Contains(line, "-") {
		return false
	}
	for _, r := range line {
		if !strings.ContainsRune("| :-", r) {
			return false
		}
	}
	return true
}

func extractPDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("PDF 解析失败: %w", err)
	}
	plain, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("PDF 文本提取失败: %w", err)
	}
	text, err := io.ReadAll(plain)
	if err != nil {
		return "", fmt.Errorf("PDF 文本读取失败: %w", err)
	}
	return string(text), nil
}

func extractDOCX(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("Word(.docx) 解析失败: %w", err)
	}

	parts := make([]*zip.File, 0)
	for _, file := range reader.File {
		if isDOCXTextPart(file.Name) {
			parts = append(parts, file)
		}
	}
	sort.SliceStable(parts, func(i, j int) bool {
		pi, pj := docxPartPriority(parts[i].Name), docxPartPriority(parts[j].Name)
		if pi == pj {
			return parts[i].Name < parts[j].Name
		}
		return pi < pj
	})

	var builder strings.Builder
	for _, part := range parts {
		text, err := extractDOCXXMLPart(part)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(text) != "" {
			builder.WriteString(text)
			builder.WriteByte('\n')
		}
	}
	return builder.String(), nil
}

func isDOCXTextPart(name string) bool {
	if name == "word/document.xml" {
		return true
	}
	for _, prefix := range []string{
		"word/header",
		"word/footer",
		"word/footnotes",
		"word/endnotes",
		"word/comments",
	} {
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".xml") {
			return true
		}
	}
	return false
}

func docxPartPriority(name string) int {
	switch {
	case name == "word/document.xml":
		return 0
	case strings.HasPrefix(name, "word/header"):
		return 1
	case strings.HasPrefix(name, "word/footer"):
		return 2
	default:
		return 3
	}
}

func extractDOCXXMLPart(part *zip.File) (string, error) {
	rc, err := part.Open()
	if err != nil {
		return "", fmt.Errorf("Word(.docx) 读取 %s 失败: %w", part.Name, err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("Word(.docx) 读取 %s 失败: %w", part.Name, err)
	}
	return extractWordXML(data)
}

func extractWordXML(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var builder strings.Builder
	inText := false

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("Word(.docx) XML 解析失败: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab":
				builder.WriteByte(' ')
			case "br", "cr":
				builder.WriteByte('\n')
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p", "tr":
				builder.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				builder.Write([]byte(t))
			}
		}
	}
	return builder.String(), nil
}

func extractXLSX(data []byte) (string, error) {
	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("Excel(.xlsx) 解析失败: %w", err)
	}
	defer file.Close()

	var builder strings.Builder
	for _, sheet := range file.GetSheetList() {
		rows, err := file.GetRows(sheet)
		if err != nil {
			return "", fmt.Errorf("Excel(.xlsx) 读取工作表 %s 失败: %w", sheet, err)
		}
		writeSheetRows(&builder, sheet, rows)
	}
	return builder.String(), nil
}

func extractXLS(data []byte) (string, error) {
	workbook, err := xls.OpenReader(bytes.NewReader(data), "utf-8")
	if err != nil {
		return "", fmt.Errorf("Excel(.xls) 解析失败: %w", err)
	}
	if workbook == nil {
		return "", errors.New("Excel(.xls) 解析失败: 未读取到工作簿")
	}

	var builder strings.Builder
	for i := 0; i < workbook.NumSheets(); i++ {
		sheet := workbook.GetSheet(i)
		if sheet == nil {
			continue
		}
		builder.WriteString("工作表: ")
		builder.WriteString(sheet.Name)
		builder.WriteByte('\n')
		for rowIndex := 0; rowIndex <= int(sheet.MaxRow); rowIndex++ {
			row := safeXLSRow(sheet, rowIndex)
			if row == nil {
				continue
			}
			cells := make([]string, 0)
			for colIndex := row.FirstCol(); colIndex < row.LastCol(); colIndex++ {
				cells = append(cells, strings.TrimSpace(safeXLSCell(row, colIndex)))
			}
			if hasNonEmptyCell(cells) {
				builder.WriteString(strings.Join(cells, "\t"))
				builder.WriteByte('\n')
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String(), nil
}

func safeXLSRow(sheet *xls.WorkSheet, rowIndex int) (row *xls.Row) {
	defer func() {
		if recover() != nil {
			row = nil
		}
	}()
	return sheet.Row(rowIndex)
}

func safeXLSCell(row *xls.Row, colIndex int) (cell string) {
	defer func() {
		if recover() != nil {
			cell = ""
		}
	}()
	return row.Col(colIndex)
}

func extractCSV(data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", errors.New("CSV 文件不是 UTF-8 编码，请转换编码后重新上传")
	}
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1

	var builder strings.Builder
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("CSV 解析失败: %w", err)
		}
		for i := range record {
			record[i] = strings.TrimSpace(record[i])
		}
		if hasNonEmptyCell(record) {
			builder.WriteString(strings.Join(record, "\t"))
			builder.WriteByte('\n')
		}
	}
	return builder.String(), nil
}

func writeSheetRows(builder *strings.Builder, sheet string, rows [][]string) {
	builder.WriteString("工作表: ")
	builder.WriteString(sheet)
	builder.WriteByte('\n')
	for _, row := range rows {
		for i := range row {
			row[i] = strings.TrimSpace(row[i])
		}
		if hasNonEmptyCell(row) {
			builder.WriteString(strings.Join(row, "\t"))
			builder.WriteByte('\n')
		}
	}
	builder.WriteByte('\n')
}

func hasNonEmptyCell(cells []string) bool {
	for _, cell := range cells {
		if strings.TrimSpace(cell) != "" {
			return true
		}
	}
	return false
}

func normalizeExtractedText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.Map(func(r rune) rune {
		if r == 0 {
			return -1
		}
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, text)

	lines := strings.Split(text, "\n")
	normalized := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			if !blank && len(normalized) > 0 {
				normalized = append(normalized, "")
				blank = true
			}
			continue
		}
		normalized = append(normalized, line)
		blank = false
	}
	return strings.TrimSpace(strings.Join(normalized, "\n"))
}

func looksLikeText(text string) bool {
	var total, printable int
	for _, r := range text {
		total++
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			printable++
		}
	}
	return total > 0 && float64(printable)/float64(total) >= 0.85
}
