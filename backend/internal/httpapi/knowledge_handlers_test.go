package httpapi

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestReadUploadContentRejectsOversizedFile(t *testing.T) {
	maxBytes := int64(15 * 1024 * 1024)
	content := bytes.NewReader(bytes.Repeat([]byte("a"), int(maxBytes)+1))

	_, err := readUploadContent(content, maxBytes)
	if !errors.Is(err, errUploadTooLarge) {
		t.Fatalf("expected errUploadTooLarge, got %v", err)
	}
}

func TestReadUploadContentAllowsFileAtLimit(t *testing.T) {
	maxBytes := int64(15 * 1024 * 1024)
	content := bytes.NewReader(bytes.Repeat([]byte("a"), int(maxBytes)))

	data, err := readUploadContent(content, maxBytes)
	if err != nil {
		t.Fatalf("expected file at limit to pass, got %v", err)
	}
	if int64(len(data)) != maxBytes {
		t.Fatalf("expected %d bytes, got %d", maxBytes, len(data))
	}
}

func TestAllowedFileTypeSupportsKnowledgeFormats(t *testing.T) {
	for _, ext := range []string{".pdf", ".docx", ".xlsx", ".xls", ".md", ".markdown", ".txt", ".csv"} {
		if !allowedFileType(ext) {
			t.Fatalf("expected %s to be allowed", ext)
		}
	}
	if allowedFileType(".doc") {
		t.Fatal("expected legacy .doc to be rejected so users get a clear conversion path")
	}
}

func TestDocumentAttachmentHeaderEncodesUTF8Filename(t *testing.T) {
	header := documentAttachmentHeader("门诊 指南.pdf")
	if !strings.Contains(header, `attachment`) {
		t.Fatalf("header = %q, want attachment disposition", header)
	}
	if !strings.Contains(header, `filename*=UTF-8''%E9%97%A8%E8%AF%8A%20%E6%8C%87%E5%8D%97.pdf`) {
		t.Fatalf("header = %q, want RFC 5987 UTF-8 filename", header)
	}
}
