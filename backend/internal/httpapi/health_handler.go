package httpapi

import (
	"net/http"
	"time"
)

func (api *API) health(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"status": "ok",
		"time":   time.Now(),
		"checks": map[string]string{
			"mongodb": "ok",
			"qdrant":  "configured",
			"deepseek": func() string {
				if api.cfg.DeepSeekAPIKey == "" {
					return "missing DEEPSEEK_API_KEY, local demo stream enabled"
				}
				return "configured"
			}(),
			"qwenEmbedding": func() string {
				if api.cfg.QwenEmbeddingBaseURL == "" || api.cfg.QwenEmbeddingAPIKey == "" {
					return "missing QWEN_EMBEDDING_*，本地确定性向量用于演示"
				}
				return "configured"
			}(),
		},
	}
	if err := api.store.Ping(r.Context()); err != nil {
		status["status"] = "degraded"
		status["checks"].(map[string]string)["mongodb"] = err.Error()
	}
	writeJSON(w, http.StatusOK, status)
}
