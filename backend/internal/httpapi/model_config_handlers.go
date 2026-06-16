package httpapi

import (
	"net/http"
	"strings"

	"medical-agent/backend/internal/models"
)

type modelConfigResponse struct {
	ID                     string `json:"id,omitempty"`
	DeepSeekBaseURL        string `json:"deepSeekBaseUrl"`
	DeepSeekChatModel      string `json:"deepSeekChatModel"`
	DeepSeekAPIKeySet      bool   `json:"deepSeekAPIKeySet"`
	DeepSeekAPIKeyPreview  string `json:"deepSeekAPIKeyPreview"`
	QwenEmbeddingBaseURL   string `json:"qwenEmbeddingBaseUrl"`
	QwenEmbeddingModel     string `json:"qwenEmbeddingModel"`
	QwenEmbeddingDimension int    `json:"qwenEmbeddingDimension"`
	QwenEmbeddingAPIKeySet bool   `json:"qwenEmbeddingAPIKeySet"`
	QwenAPIKeyPreview      string `json:"qwenEmbeddingAPIKeyPreview"`
	UpdatedAt              string `json:"updatedAt,omitempty"`
}

func (api *API) getModelConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := api.store.GetModelConfig(r.Context(), api.cfg)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "config_read_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, redactModelConfig(cfg))
}

func (api *API) saveModelConfig(w http.ResponseWriter, r *http.Request) {
	var req saveModelConfigRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.DeepSeekBaseURL) == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "DeepSeek Base URL 不能为空")
		return
	}
	if strings.TrimSpace(req.DeepSeekChatModel) == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "DeepSeek 模型不能为空")
		return
	}
	if strings.TrimSpace(req.QwenEmbeddingModel) == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Embedding 模型不能为空")
		return
	}
	if req.QwenEmbeddingDimension <= 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Embedding 维度必须大于 0")
		return
	}
	saved, err := api.store.SaveModelConfig(r.Context(), models.ModelConfig{
		DeepSeekBaseURL:        req.DeepSeekBaseURL,
		DeepSeekAPIKey:         req.DeepSeekAPIKey,
		DeepSeekChatModel:      req.DeepSeekChatModel,
		QwenEmbeddingBaseURL:   req.QwenEmbeddingBaseURL,
		QwenEmbeddingAPIKey:    req.QwenEmbeddingAPIKey,
		QwenEmbeddingModel:     req.QwenEmbeddingModel,
		QwenEmbeddingDimension: req.QwenEmbeddingDimension,
	}, api.cfg)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "config_save_failed", err.Error())
		return
	}
	api.store.Audit(r.Context(), currentUser(r).ID, "system.model_config.update", "model_configs", "success", requestID(r))
	writeJSON(w, http.StatusOK, redactModelConfig(saved))
}

func redactModelConfig(cfg models.ModelConfig) modelConfigResponse {
	response := modelConfigResponse{
		DeepSeekBaseURL:        cfg.DeepSeekBaseURL,
		DeepSeekChatModel:      cfg.DeepSeekChatModel,
		DeepSeekAPIKeySet:      strings.TrimSpace(cfg.DeepSeekAPIKey) != "",
		DeepSeekAPIKeyPreview:  previewSecret(cfg.DeepSeekAPIKey),
		QwenEmbeddingBaseURL:   cfg.QwenEmbeddingBaseURL,
		QwenEmbeddingModel:     cfg.QwenEmbeddingModel,
		QwenEmbeddingDimension: cfg.QwenEmbeddingDimension,
		QwenEmbeddingAPIKeySet: strings.TrimSpace(cfg.QwenEmbeddingAPIKey) != "",
		QwenAPIKeyPreview:      previewSecret(cfg.QwenEmbeddingAPIKey),
	}
	if !cfg.ID.IsZero() {
		response.ID = cfg.ID.Hex()
	}
	if !cfg.UpdatedAt.IsZero() {
		response.UpdatedAt = cfg.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return response
}

func previewSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= 8 {
		return "****"
	}
	return string(runes[:4]) + "****" + string(runes[len(runes)-4:])
}
