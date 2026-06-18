package policy

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"medical-agent/backend/internal/models"

	"github.com/xuri/excelize/v2"
)

type ImportReport struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

var fixedCategories = []string{"国家医学中心", "科技创新", "医疗服务", "医保医药", "数智治理", "改革监管", "国际合作", "其他"}

func Categories() []string {
	return append([]string(nil), fixedCategories...)
}

func IsValidCategory(value string) bool {
	value = strings.TrimSpace(value)
	for _, category := range fixedCategories {
		if value == category {
			return true
		}
	}
	return false
}

func ParseExcel(filename string, data []byte) ([]models.PolicyDocument, ImportReport, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".xlsx" && ext != ".xlsm" && ext != ".xltx" && ext != ".xltm" {
		return nil, ImportReport{}, errors.New("请上传 Excel .xlsx 文件")
	}
	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, ImportReport{}, fmt.Errorf("Excel 解析失败: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, ImportReport{}, errors.New("Excel 中没有可读取的工作表")
	}
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, ImportReport{}, fmt.Errorf("读取工作表失败: %w", err)
	}
	if len(rows) < 2 {
		return nil, ImportReport{}, errors.New("Excel 至少需要包含表头和一行政策数据")
	}
	columns := resolveColumns(rows[0])
	if !columns.complete() {
		return nil, ImportReport{}, errors.New("Excel 必须包含标题、摘要、日期、分类字段")
	}

	report := ImportReport{}
	records := make([]models.PolicyDocument, 0, len(rows)-1)
	for rowIndex, row := range rows[1:] {
		rowNumber := rowIndex + 2
		record := models.PolicyDocument{
			Title:          cell(row, columns.title),
			Summary:        cell(row, columns.summary),
			Interpretation: cell(row, columns.interpretation),
			Date:           policyDateCell(file, sheets[0], rowNumber, columns.date, cell(row, columns.date)),
			Category:       cell(row, columns.category),
		}
		if strings.TrimSpace(record.Title) == "" || strings.TrimSpace(record.Summary) == "" || strings.TrimSpace(record.Date) == "" || strings.TrimSpace(record.Category) == "" {
			report.Skipped++
			report.Errors = append(report.Errors, fmt.Sprintf("第 %d 行缺少标题、摘要、日期或分类", rowNumber))
			continue
		}
		if !IsValidCategory(record.Category) {
			report.Skipped++
			report.Errors = append(report.Errors, fmt.Sprintf("第 %d 行分类 %q 不在固定分类列表中", rowNumber, record.Category))
			continue
		}
		record.RowChecksum = checksum(record.Title, record.Summary, record.Interpretation, record.Date, record.Category)
		records = append(records, record)
		report.Imported++
	}
	return records, report, nil
}

func TemplateWorkbook() ([]byte, error) {
	file := excelize.NewFile()
	defer file.Close()
	sheet := file.GetSheetName(0)
	if err := file.SetSheetName(sheet, "政策文件库"); err != nil {
		return nil, err
	}
	sheet = "政策文件库"
	headers := []string{"标题", "摘要", "解读", "日期", "分类标签"}
	example := []string{"国家医学中心建设通知", "围绕医学中心建设提出重点任务", "提炼适用范围、执行口径和落地提醒", "2026-06-08", "国家医学中心"}
	for index, value := range headers {
		cellName, err := excelize.CoordinatesToCellName(index+1, 1)
		if err != nil {
			return nil, err
		}
		if err := file.SetCellValue(sheet, cellName, value); err != nil {
			return nil, err
		}
	}
	for index, value := range example {
		cellName, err := excelize.CoordinatesToCellName(index+1, 2)
		if err != nil {
			return nil, err
		}
		if err := file.SetCellValue(sheet, cellName, value); err != nil {
			return nil, err
		}
	}
	if err := file.SetColWidth(sheet, "A", "A", 30); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(sheet, "B", "C", 42); err != nil {
		return nil, err
	}
	if err := file.SetColWidth(sheet, "D", "E", 18); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type columns struct {
	title          int
	summary        int
	interpretation int
	date           int
	category       int
}

func (c columns) complete() bool {
	return c.title >= 0 && c.summary >= 0 && c.date >= 0 && c.category >= 0
}

func resolveColumns(header []string) columns {
	result := columns{title: -1, summary: -1, interpretation: -1, date: -1, category: -1}
	for index, raw := range header {
		name := normalizeHeader(raw)
		switch {
		case contains(headerAliasesTitle, name):
			result.title = index
		case contains(headerAliasesSummary, name):
			result.summary = index
		case contains(headerAliasesInterpretation, name):
			result.interpretation = index
		case contains(headerAliasesDate, name):
			result.date = index
		case contains(headerAliasesCategory, name):
			result.category = index
		}
	}
	return result
}

var (
	headerAliasesTitle          = []string{"标题", "文件标题", "政策标题", "名称", "文件名称"}
	headerAliasesSummary        = []string{"摘要", "政策摘要", "内容摘要", "简介", "概述"}
	headerAliasesInterpretation = []string{"解读", "政策解读", "解读内容", "要点解读"}
	headerAliasesDate           = []string{"日期", "发布时间", "发布日期", "发文日期", "印发日期"}
	headerAliasesCategory       = []string{"分类", "主题分类", "类别", "政策分类", "标签", "分类标签"}
)

func normalizeHeader(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "：")
	value = strings.TrimSuffix(value, ":")
	return strings.Join(strings.Fields(value), "")
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func normalizeDate(value string) string {
	value = strings.TrimSpace(strings.TrimSuffix(value, " 00:00:00"))
	if parsed, ok := parsePolicyDateText(value); ok {
		return parsed
	}
	return value
}

func policyDateCell(file *excelize.File, sheet string, rowNumber int, columnIndex int, fallback string) string {
	cellName, err := excelize.CoordinatesToCellName(columnIndex+1, rowNumber)
	if err != nil {
		return normalizeDate(fallback)
	}
	raw, err := file.GetCellValue(sheet, cellName, excelize.Options{RawCellValue: true})
	if err == nil && raw != "" && isDateCell(file, sheet, cellName) {
		if serial, err := strconv.ParseFloat(raw, 64); err == nil {
			use1904Format := false
			if props, err := file.GetWorkbookProps(); err == nil && props.Date1904 != nil {
				use1904Format = *props.Date1904
			}
			if parsed, err := excelize.ExcelDateToTime(serial, use1904Format); err == nil {
				return parsed.Format("2006-01-02")
			}
		}
	}
	if raw != "" {
		if parsed, ok := parsePolicyDateText(raw); ok {
			return parsed
		}
	}
	return normalizeDate(fallback)
}

func isDateCell(file *excelize.File, sheet string, cellName string) bool {
	styleID, err := file.GetCellStyle(sheet, cellName)
	if err != nil {
		return false
	}
	style, err := file.GetStyle(styleID)
	if err != nil || style == nil {
		return false
	}
	if isBuiltInDateNumFmt(style.NumFmt) {
		return true
	}
	if style.CustomNumFmt == nil {
		return false
	}
	format := strings.ToLower(*style.CustomNumFmt)
	return strings.Contains(format, "y") && strings.Contains(format, "d")
}

func isBuiltInDateNumFmt(numFmt int) bool {
	if numFmt >= 14 && numFmt <= 17 {
		return true
	}
	if numFmt == 22 {
		return true
	}
	if numFmt >= 27 && numFmt <= 36 {
		return true
	}
	return numFmt >= 50 && numFmt <= 58
}

func parsePolicyDateText(value string) (string, bool) {
	value = strings.TrimSpace(strings.TrimSuffix(value, " 00:00:00"))
	if value == "" {
		return "", false
	}
	value = strings.Fields(value)[0]
	normalized := strings.NewReplacer("/", "-", ".", "-", "年", "-", "月", "-", "日", "").Replace(value)
	normalized = strings.TrimSuffix(normalized, "-")
	if len(normalized) == 8 && allDigits(normalized) {
		normalized = normalized[:4] + "-" + normalized[4:6] + "-" + normalized[6:8]
	}
	parts := strings.Split(normalized, "-")
	if len(parts) != 2 && len(parts) != 3 {
		return "", false
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 1000 || year > 9999 {
		return "", false
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return "", false
	}
	if len(parts) == 2 {
		return fmt.Sprintf("%04d-%02d", year, month), true
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil || day < 1 || day > 31 {
		return "", false
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day), true
}

func allDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func checksum(values ...string) string {
	hash := sha1.New()
	for _, value := range values {
		hash.Write([]byte(strings.TrimSpace(value)))
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
