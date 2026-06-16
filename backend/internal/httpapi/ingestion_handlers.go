package httpapi

import (
	"context"
	"net/http"

	"medical-agent/backend/internal/store"

	"github.com/go-chi/chi/v5"
)

func (api *API) getIngestionJob(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "任务 ID 无效")
		return
	}
	job, err := api.store.GetIngestionJob(r.Context(), id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "任务不存在")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (api *API) retryIngestionJob(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "任务 ID 无效")
		return
	}
	if err := api.store.RetryIngestionJob(r.Context(), id); err != nil {
		writeError(w, r, http.StatusInternalServerError, "retry_failed", err.Error())
		return
	}
	go api.worker.RunOnce(context.Background())
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
