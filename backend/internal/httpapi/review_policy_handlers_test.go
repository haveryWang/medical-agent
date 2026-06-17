package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReviewNoteAttachmentHeaderEncodesUTF8Filename(t *testing.T) {
	header := markdownAttachmentHeader("复盘笔记-20260617-093000.md")
	if !strings.Contains(header, `attachment`) {
		t.Fatalf("header = %q, want attachment disposition", header)
	}
	if !strings.Contains(header, `filename*=UTF-8''%E5%A4%8D%E7%9B%98%E7%AC%94%E8%AE%B0-20260617-093000.md`) {
		t.Fatalf("header = %q, want RFC 5987 UTF-8 filename", header)
	}
}

func TestAllowedPolicyFileTypeOnlyAcceptsExcel(t *testing.T) {
	if !allowedPolicyFileType(".xlsx") {
		t.Fatal("expected .xlsx policy import to be allowed")
	}
	for _, ext := range []string{".pdf", ".csv", ".docx", ".txt"} {
		if allowedPolicyFileType(ext) {
			t.Fatalf("expected %s policy import to be rejected", ext)
		}
	}
}

func TestPolicyTemplateAttachmentHeaderEncodesUTF8Filename(t *testing.T) {
	header := excelAttachmentHeader("政策文件库导入模板.xlsx")
	if !strings.Contains(header, `attachment`) {
		t.Fatalf("header = %q, want attachment disposition", header)
	}
	if !strings.Contains(header, `filename*=UTF-8''%E6%94%BF%E7%AD%96%E6%96%87%E4%BB%B6%E5%BA%93%E5%AF%BC%E5%85%A5%E6%A8%A1%E6%9D%BF.xlsx`) {
		t.Fatalf("header = %q, want RFC 5987 UTF-8 filename", header)
	}
}

func TestDownloadPolicyTemplateReturnsWorkbook(t *testing.T) {
	api := &API{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/import-template", nil)
	req = req.WithContext(requestWithUser(context.Background(), testUser()))
	rec := httptest.NewRecorder()

	api.downloadPolicyImportTemplate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") {
		t.Fatalf("content-type = %q", got)
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, "filename*=") {
		t.Fatalf("content-disposition = %q", got)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected template body")
	}
}

func TestNormalizeReviewNoteContentRejectsEmpty(t *testing.T) {
	if _, ok := normalizeReviewNoteContent(" \n\t "); ok {
		t.Fatal("expected whitespace-only content to be invalid")
	}
	got, ok := normalizeReviewNoteContent("  科室复盘  ")
	if !ok || got != "科室复盘" {
		t.Fatalf("normalized = %q, ok = %v", got, ok)
	}
}
