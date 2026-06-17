package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"medical-agent/backend/internal/reviewnotes"
	"medical-agent/backend/internal/store"

	"github.com/go-chi/chi/v5"
)

func (api *API) createReviewNote(w http.ResponseWriter, r *http.Request) {
	var req createReviewNoteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	content, ok := normalizeReviewNoteContent(req.Content)
	if !ok {
		writeError(w, r, http.StatusBadRequest, "validation_error", "复盘笔记内容不能为空")
		return
	}
	note, err := api.store.CreateReviewNote(r.Context(), currentUser(r).ID, content)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "review_note.create", note.ID.Hex(), "success", requestID(r))
	writeJSON(w, http.StatusCreated, note)
}

func (api *API) listReviewNotes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.ParseInt(query.Get("page"), 10, 64)
	size, _ := strconv.ParseInt(query.Get("size"), 10, 64)
	result, err := api.store.ListReviewNotes(r.Context(), store.ReviewNoteFilter{Page: page, PageSize: size})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (api *API) reviewNoteCounts(w http.ResponseWriter, r *http.Request) {
	counts, err := api.store.ReviewNoteCounts(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "count_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, counts)
}

func (api *API) exportReviewNotes(w http.ResponseWriter, r *http.Request) {
	var req exportReviewNotesRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	noteIDs := parseObjectIDs(req.NoteIDs)
	if len(noteIDs) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "请选择要导出的复盘笔记")
		return
	}
	now := time.Now()
	filename := reviewnotes.ExportFilename(now)
	notes, batch, err := api.store.ClaimSelectedReviewNotes(r.Context(), currentUser(r).ID, noteIDs, filename)
	if err != nil {
		if errors.Is(err, errNoReviewNotes) || strings.Contains(err.Error(), "没有未导出") {
			writeError(w, r, http.StatusBadRequest, "validation_error", "请选择要导出的复盘笔记")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "export_failed", err.Error())
		return
	}
	markdown := batch.Content
	if markdown == "" {
		markdown = reviewnotes.RenderMarkdown(notes, batch.CreatedAt)
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", markdownAttachmentHeader(batch.Filename))
	w.Header().Set("Content-Length", strconv.Itoa(len([]byte(markdown))))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(markdown))
	api.store.Audit(r.Context(), currentUser(r).ID, "review_note.export", batch.ID.Hex(), "success", requestID(r))
}

func (api *API) deleteReviewNote(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "复盘笔记 ID 无效")
		return
	}
	if err := api.store.DeleteReviewNote(r.Context(), id); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "复盘笔记不存在")
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "review_note.delete", id.Hex(), "success", requestID(r))
	writeJSON(w, http.StatusOK, map[string]any{"id": id.Hex(), "deleted": true})
}

func (api *API) listReviewNoteExports(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	exports, err := api.store.ListReviewNoteExports(r.Context(), limit)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": exports})
}

func (api *API) downloadReviewNoteExport(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "导出批次 ID 无效")
		return
	}
	batch, markdown, err := api.store.GetReviewNoteExport(r.Context(), id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "导出批次不存在")
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", markdownAttachmentHeader(batch.Filename))
	w.Header().Set("Content-Length", strconv.Itoa(len([]byte(markdown))))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(markdown))
}

var errNoReviewNotes = errors.New("没有未导出的复盘笔记")

func normalizeReviewNoteContent(value string) (string, bool) {
	content := strings.TrimSpace(value)
	return content, content != ""
}

func markdownAttachmentHeader(filename string) string {
	filename = sanitizeFilename(filename)
	ascii := strings.NewReplacer("\\", "_", `"`, "_", "\r", "_", "\n", "_").Replace(filename)
	return `attachment; filename="` + ascii + `"; filename*=UTF-8''` + url.PathEscape(filename)
}
