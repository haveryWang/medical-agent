package httpapi

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"medical-agent/backend/internal/ingestion"
	"medical-agent/backend/internal/models"
	"medical-agent/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func (api *API) listKnowledgeBases(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.ParseInt(query.Get("page"), 10, 64)
	size, _ := strconv.ParseInt(query.Get("size"), 10, 64)
	result, err := api.store.ListKnowledgeBases(r.Context(), store.KnowledgeFilter{
		Scenario:   query.Get("scenario"),
		Tag:        query.Get("tag"),
		Department: query.Get("department"),
		Keyword:    query.Get("keyword"),
		Page:       page,
		PageSize:   size,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (api *API) createKnowledgeBase(w http.ResponseWriter, r *http.Request) {
	var req models.KnowledgeBase
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "知识库名称不能为空")
		return
	}
	kb, err := api.store.CreateKnowledgeBase(r.Context(), req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "knowledge.create", kb.ID.Hex(), "success", requestID(r))
	writeJSON(w, http.StatusCreated, kb)
}

func (api *API) updateKnowledgeBase(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "知识库 ID 无效")
		return
	}
	var req map[string]any
	if !decodeJSON(w, r, &req) {
		return
	}
	patch := bson.M{}
	for _, key := range []string{"name", "description", "scenario", "department", "status", "buildStatus", "retrievalTopK", "similarityFloor"} {
		if value, ok := req[key]; ok {
			patch[key] = value
		}
	}
	if value, ok := req["tags"]; ok {
		patch["tags"] = value
	}
	if len(patch) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "没有可更新字段")
		return
	}
	if err := api.store.UpdateKnowledgeBase(r.Context(), id, patch); err != nil {
		writeError(w, r, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "knowledge.update", id.Hex(), "success", requestID(r))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (api *API) listDocuments(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "知识库 ID 无效")
		return
	}
	if _, err := api.store.GetKnowledgeBase(r.Context(), id); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "知识库不存在")
		return
	}
	docs, err := api.store.ListDocuments(r.Context(), id)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": docs})
}

func (api *API) uploadDocument(w http.ResponseWriter, r *http.Request) {
	kbID, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "知识库 ID 无效")
		return
	}
	if _, err := api.store.GetKnowledgeBase(r.Context(), kbID); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "知识库不存在")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, api.cfg.MaxUploadBytes+1024*1024)
	if err := r.ParseMultipartForm(api.cfg.MaxUploadBytes); err != nil {
		writeError(w, r, http.StatusBadRequest, "upload_error", uploadTooLargeMessage())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "upload_error", "请上传 file 字段")
		return
	}
	defer file.Close()
	if header.Size > api.cfg.MaxUploadBytes {
		writeError(w, r, http.StatusBadRequest, "file_too_large", uploadTooLargeMessage())
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == ".doc" {
		writeError(w, r, http.StatusBadRequest, "unsupported_file", "当前支持 Word .docx 文件，.doc 请先转换为 .docx 后上传")
		return
	}
	if !allowedFileType(ext) {
		writeError(w, r, http.StatusBadRequest, "unsupported_file", "当前支持 PDF、Word(.docx)、Excel(.xlsx/.xls)、Markdown、TXT、CSV，单个文件 15MB 以内")
		return
	}
	content, err := readUploadContent(file, api.cfg.MaxUploadBytes)
	if err != nil {
		if errors.Is(err, errUploadTooLarge) {
			writeError(w, r, http.StatusBadRequest, "file_too_large", uploadTooLargeMessage())
			return
		}
		writeError(w, r, http.StatusBadRequest, "upload_error", err.Error())
		return
	}
	doc, job, err := api.store.CreateDocumentAndJob(r.Context(), models.Document{
		KnowledgeBaseID: kbID,
		FileName:        sanitizeFilename(header.Filename),
		FileType:        ext,
		SizeBytes:       int64(len(content)),
		Content:         content,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "document.upload", doc.ID.Hex(), "success", requestID(r))
	go api.worker.RunOnce(context.Background())
	writeJSON(w, http.StatusCreated, map[string]any{"document": doc, "job": job})
}

func (api *API) viewDocumentContent(w http.ResponseWriter, r *http.Request) {
	doc, ok := api.documentFromRoute(w, r)
	if !ok {
		return
	}
	text, err := ingestion.ExtractText(doc)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "preview_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"document": doc,
		"content":  text,
	})
}

func (api *API) downloadDocument(w http.ResponseWriter, r *http.Request) {
	doc, ok := api.documentFromRoute(w, r)
	if !ok {
		return
	}
	content, err := rawDocumentContent(doc)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "download_failed", err.Error())
		return
	}
	contentType := mime.TypeByExtension(doc.FileType)
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", documentAttachmentHeader(doc.FileName))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (api *API) listDocumentChunks(w http.ResponseWriter, r *http.Request) {
	doc, ok := api.documentFromRoute(w, r)
	if !ok {
		return
	}
	chunks, err := api.store.ListDocumentChunks(r.Context(), doc.ID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "chunks_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"document": doc, "items": chunks})
}

func (api *API) deleteDocument(w http.ResponseWriter, r *http.Request) {
	doc, ok := api.documentFromRoute(w, r)
	if !ok {
		return
	}
	result, err := api.store.DeleteDocumentCascade(r.Context(), doc)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	vectorCleanup := true
	if api.vector != nil && len(result.VectorIDs) > 0 {
		if err := api.vector.DeletePoints(r.Context(), result.VectorIDs); err != nil {
			vectorCleanup = false
			if api.logger != nil {
				api.logger.Printf("删除文档 %s 后清理向量失败: %v", doc.ID.Hex(), err)
			}
		}
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "document.delete", doc.ID.Hex(), "success", requestID(r))
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"documentId":    result.DocumentID.Hex(),
		"deletedChunks": result.DeletedChunks,
		"vectorCleanup": vectorCleanup,
	})
}

func (api *API) documentFromRoute(w http.ResponseWriter, r *http.Request) (models.Document, bool) {
	kbID, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "知识库 ID 无效")
		return models.Document{}, false
	}
	docID, err := store.ObjectIDFromHex(chi.URLParam(r, "documentId"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "文档 ID 无效")
		return models.Document{}, false
	}
	if _, err := api.store.GetKnowledgeBase(r.Context(), kbID); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "知识库不存在")
		return models.Document{}, false
	}
	doc, err := api.store.GetDocument(r.Context(), docID)
	if err != nil || doc.KnowledgeBaseID != kbID {
		writeError(w, r, http.StatusNotFound, "not_found", "文档不存在")
		return models.Document{}, false
	}
	return doc, true
}

func rawDocumentContent(doc models.Document) ([]byte, error) {
	if len(doc.Content) > 0 {
		return doc.Content, nil
	}
	if strings.TrimSpace(doc.StoragePath) != "" {
		content, err := os.ReadFile(doc.StoragePath)
		if err != nil {
			return nil, err
		}
		if len(content) > 0 {
			return content, nil
		}
	}
	return nil, errors.New("当前文档没有可下载的原始内容")
}

func documentAttachmentHeader(filename string) string {
	filename = sanitizeFilename(filename)
	ascii := strings.NewReplacer("\\", "_", `"`, "_", "\r", "_", "\n", "_").Replace(filename)
	return `attachment; filename="` + ascii + `"; filename*=UTF-8''` + url.PathEscape(filename)
}

var errUploadTooLarge = errors.New("upload too large")

func readUploadContent(file io.Reader, maxBytes int64) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > maxBytes {
		return nil, errUploadTooLarge
	}
	if len(content) == 0 {
		return nil, errors.New("上传文件不能为空")
	}
	return content, nil
}

func uploadTooLargeMessage() string {
	return "单个知识库文件不能超过 15MB，请切割单文件尺寸后再上传"
}

func allowedFileType(ext string) bool {
	switch ext {
	case ".pdf", ".docx", ".xlsx", ".xls", ".md", ".markdown", ".txt", ".csv":
		return true
	default:
		return false
	}
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "" {
		return "upload"
	}
	return base
}
