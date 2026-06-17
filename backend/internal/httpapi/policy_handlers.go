package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"medical-agent/backend/internal/policy"
	"medical-agent/backend/internal/store"

	"github.com/go-chi/chi/v5"
)

func (api *API) policyCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": policy.Categories()})
}

func (api *API) listPolicies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	category := strings.TrimSpace(query.Get("category"))
	if category != "" && !policy.IsValidCategory(category) {
		writeError(w, r, http.StatusBadRequest, "validation_error", "政策分类不在固定分类列表中")
		return
	}
	page, _ := strconv.ParseInt(query.Get("page"), 10, 64)
	size, _ := strconv.ParseInt(query.Get("size"), 10, 64)
	result, err := api.store.ListPolicyDocuments(r.Context(), store.PolicyFilter{
		Category: category,
		Date:     strings.TrimSpace(query.Get("date")),
		Keyword:  query.Get("keyword"),
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	facets, err := api.store.ListPolicyFacets(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": result.Items, "total": result.Total, "page": result.Page, "size": result.PageSize, "facets": facets})
}

func (api *API) deletePolicy(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "政策记录 ID 无效")
		return
	}
	if err := api.store.DeletePolicyDocument(r.Context(), id); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "政策记录不存在")
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "policy.delete", id.Hex(), "success", requestID(r))
	writeJSON(w, http.StatusOK, map[string]any{"id": id.Hex(), "deleted": true})
}

func (api *API) downloadPolicyImportTemplate(w http.ResponseWriter, r *http.Request) {
	content, err := policy.TemplateWorkbook()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "template_failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", excelAttachmentHeader("政策文件库导入模板.xlsx"))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (api *API) importPolicies(w http.ResponseWriter, r *http.Request) {
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
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedPolicyFileType(ext) {
		writeError(w, r, http.StatusBadRequest, "unsupported_file", "政策文件库仅支持 Excel .xlsx 文件")
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
	docs, report, err := policy.ParseExcel(sanitizeFilename(header.Filename), content)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	batch, err := api.store.ImportPolicyDocuments(r.Context(), currentUser(r).ID, sanitizeFilename(header.Filename), docs, report.Imported, report.Skipped, report.Errors)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "import_failed", err.Error())
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "policy.import", batch.ID.Hex(), "success", requestID(r))
	writeJSON(w, http.StatusCreated, map[string]any{"batch": batch, "report": report})
}

func allowedPolicyFileType(ext string) bool {
	return strings.ToLower(ext) == ".xlsx"
}

func excelAttachmentHeader(filename string) string {
	filename = sanitizeFilename(filename)
	ascii := strings.NewReplacer("\\", "_", `"`, "_", "\r", "_", "\n", "_").Replace(filename)
	return `attachment; filename="` + ascii + `"; filename*=UTF-8''` + url.PathEscape(filename)
}
