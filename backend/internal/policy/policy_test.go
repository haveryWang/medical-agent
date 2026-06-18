package policy

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestPolicyCategoriesAreFixed(t *testing.T) {
	want := []string{"国家医学中心", "科技创新", "医疗服务", "医保医药", "数智治理", "改革监管", "国际合作", "其他"}
	got := Categories()
	if len(got) != len(want) {
		t.Fatalf("categories = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("category[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if !IsValidCategory("医保医药") {
		t.Fatal("expected 医保医药 to be valid")
	}
	if IsValidCategory("国务院文件") {
		t.Fatal("expected unsupported category to be invalid")
	}
}

func TestParseExcelUsesAliasesAndSkipsInvalidRows(t *testing.T) {
	data := policyWorkbook(t, [][]string{
		{"标题", "摘要", "解读", "发布时间", "主题分类"},
		{"国家医学中心建设通知", "围绕医学中心建设提出重点任务", "强调牵头单位和区域协同", "2026-06-08", "国家医学中心"},
		{"无效分类文件", "这一行分类不在固定列表内", "分类应被跳过", "2026-06-09", "国务院文件"},
		{"", "标题为空应跳过", "缺少标题", "2026-06-10", "科技创新"},
	})

	records, report, err := ParseExcel("政策文件.xlsx", data)
	if err != nil {
		t.Fatalf("ParseExcel error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v, want one valid record", records)
	}
	if records[0].Title != "国家医学中心建设通知" {
		t.Fatalf("title = %q", records[0].Title)
	}
	if records[0].Summary != "围绕医学中心建设提出重点任务" {
		t.Fatalf("summary = %q", records[0].Summary)
	}
	if records[0].Interpretation != "强调牵头单位和区域协同" {
		t.Fatalf("interpretation = %q", records[0].Interpretation)
	}
	if records[0].Date != "2026-06-08" {
		t.Fatalf("date = %q", records[0].Date)
	}
	if records[0].Category != "国家医学中心" {
		t.Fatalf("category = %q", records[0].Category)
	}
	if report.Imported != 1 || report.Skipped != 2 {
		t.Fatalf("report = %#v, want imported 1 skipped 2", report)
	}
	joined := strings.Join(report.Errors, "\n")
	if !strings.Contains(joined, "第 3 行") || !strings.Contains(joined, "第 4 行") {
		t.Fatalf("errors = %#v, want row-level messages", report.Errors)
	}
}

func TestParseExcelNormalizesExcelDateCells(t *testing.T) {
	data := policyWorkbookWithExcelDate(t, "46179")

	records, report, err := ParseExcel("政策文件.xlsx", data)
	if err != nil {
		t.Fatalf("ParseExcel error: %v", err)
	}
	if report.Imported != 1 || report.Skipped != 0 {
		t.Fatalf("report = %#v, want imported 1 skipped 0", report)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v, want one valid record", records)
	}
	if records[0].Date != "2026-06-06" {
		t.Fatalf("date = %q, want 2026-06-06", records[0].Date)
	}
}

func TestParseExcelNormalizesLooseDateText(t *testing.T) {
	data := policyWorkbook(t, [][]string{
		{"标题", "摘要", "解读", "日期", "分类标签"},
		{"斜杠日期", "摘要", "解读", "2026/6/6", "国际合作"},
		{"点分日期", "摘要", "解读", "2026.6.7", "国际合作"},
		{"中文日期", "摘要", "解读", "2026年6月8日", "国际合作"},
		{"紧凑日期", "摘要", "解读", "20260609", "国际合作"},
		{"月份日期", "摘要", "解读", "2026/6", "国际合作"},
	})

	records, report, err := ParseExcel("政策文件.xlsx", data)
	if err != nil {
		t.Fatalf("ParseExcel error: %v", err)
	}
	if report.Imported != 5 || report.Skipped != 0 {
		t.Fatalf("report = %#v, want imported 5 skipped 0", report)
	}
	want := []string{"2026-06-06", "2026-06-07", "2026-06-08", "2026-06-09", "2026-06"}
	for index, value := range want {
		if records[index].Date != value {
			t.Fatalf("record[%d].Date = %q, want %q", index, records[index].Date, value)
		}
	}
}

func TestTemplateWorkbookUsesExpectedHeaders(t *testing.T) {
	data, err := TemplateWorkbook()
	if err != nil {
		t.Fatalf("TemplateWorkbook error: %v", err)
	}
	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open template workbook: %v", err)
	}
	defer file.Close()

	rows, err := file.GetRows(file.GetSheetName(0))
	if err != nil {
		t.Fatalf("get rows: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("template workbook has no header row")
	}
	want := []string{"标题", "摘要", "解读", "日期", "分类标签"}
	if len(rows[0]) < len(want) {
		t.Fatalf("header = %#v, want %#v", rows[0], want)
	}
	for index, value := range want {
		if rows[0][index] != value {
			t.Fatalf("header[%d] = %q, want %q", index, rows[0][index], value)
		}
	}
}

func TestParseExcelRejectsMissingRequiredHeaders(t *testing.T) {
	data := policyWorkbook(t, [][]string{
		{"标题", "摘要"},
		{"缺少日期和分类", "内容"},
	})

	_, _, err := ParseExcel("政策文件.xlsx", data)
	if err == nil {
		t.Fatal("expected missing required headers to fail")
	}
	if !strings.Contains(err.Error(), "标题、摘要、日期、分类") {
		t.Fatalf("error = %v", err)
	}
}

func policyWorkbook(t *testing.T, rows [][]string) []byte {
	t.Helper()
	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				t.Fatalf("cell name: %v", err)
			}
			if err := file.SetCellValue(sheet, cell, value); err != nil {
				t.Fatalf("set cell: %v", err)
			}
		}
	}
	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return buf.Bytes()
}

func policyWorkbookWithExcelDate(t *testing.T, excelDate string) []byte {
	t.Helper()
	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	headers := []string{"标题", "摘要", "解读", "日期", "分类标签"}
	values := []string{"国际合作政策", "国际合作政策摘要", "国际合作政策解读", excelDate, "国际合作"}
	for colIndex, value := range headers {
		cell, err := excelize.CoordinatesToCellName(colIndex+1, 1)
		if err != nil {
			t.Fatalf("header cell name: %v", err)
		}
		if err := file.SetCellValue(sheet, cell, value); err != nil {
			t.Fatalf("set header cell: %v", err)
		}
	}
	for colIndex, value := range values {
		cell, err := excelize.CoordinatesToCellName(colIndex+1, 2)
		if err != nil {
			t.Fatalf("value cell name: %v", err)
		}
		if err := file.SetCellValue(sheet, cell, value); err != nil {
			t.Fatalf("set value cell: %v", err)
		}
	}
	if err := file.SetCellFloat(sheet, "D2", 46179, 0, 64); err != nil {
		t.Fatalf("set excel date value: %v", err)
	}
	style, err := file.NewStyle(&excelize.Style{NumFmt: 14})
	if err != nil {
		t.Fatalf("date style: %v", err)
	}
	if err := file.SetCellStyle(sheet, "D2", "D2", style); err != nil {
		t.Fatalf("set date style: %v", err)
	}
	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return buf.Bytes()
}
