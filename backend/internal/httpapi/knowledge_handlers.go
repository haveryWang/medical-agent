package httpapi

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	if err := r.ParseMultipartForm(api.cfg.MaxUploadBytes); err != nil {
		writeError(w, r, http.StatusBadRequest, "upload_error", "文件过大或表单无效")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "upload_error", "请上传 file 字段")
		return
	}
	defer file.Close()
	if header.Size > api.cfg.MaxUploadBytes {
		writeError(w, r, http.StatusBadRequest, "file_too_large", "单个文件不能超过配置限制")
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedFileType(ext) {
		writeError(w, r, http.StatusBadRequest, "unsupported_file", "仅支持 PDF、Word、Excel、Markdown、文本文件")
		return
	}
	path, filename, err := api.saveUpload(kbID.Hex(), file, header)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "save_failed", err.Error())
		return
	}
	doc, job, err := api.store.CreateDocumentAndJob(r.Context(), models.Document{
		KnowledgeBaseID: kbID,
		FileName:        filename,
		FileType:        ext,
		SizeBytes:       header.Size,
		StoragePath:     path,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "document.upload", doc.ID.Hex(), "success", requestID(r))
	go api.worker.RunOnce(context.Background())
	writeJSON(w, http.StatusCreated, map[string]any{"document": doc, "job": job})
}

func (api *API) saveUpload(kbID string, file multipart.File, header *multipart.FileHeader) (string, string, error) {
	dir := filepath.Join(api.cfg.UploadDir, kbID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	filename := sanitizeFilename(header.Filename)
	name := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
	path := filepath.Join(dir, name)
	out, err := os.Create(path)
	if err != nil {
		return "", "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return "", "", err
	}
	return path, filename, nil
}

func allowedFileType(ext string) bool {
	switch ext {
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".md", ".txt":
		return true
	default:
		return false
	}
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	base = strings.ReplaceAll(base, string(os.PathSeparator), "_")
	if base == "." || base == "" {
		return "upload"
	}
	return base
}
