package httpapi

import (
	"net/http"
	"strings"

	"medical-agent/backend/internal/security"
)

func (api *API) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := api.store.FindUserByAccount(r.Context(), strings.TrimSpace(req.Account))
	if err != nil || !security.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "账号或密码错误")
		return
	}
	token, err := security.NewToken()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "token_error", "生成会话失败")
		return
	}
	if _, err := api.store.CreateSession(r.Context(), user.ID, token, api.cfg.SessionTTL); err != nil {
		writeError(w, r, http.StatusInternalServerError, "session_error", "保存会话失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func (api *API) logout(w http.ResponseWriter, r *http.Request) {
	_ = api.store.RevokeSession(r.Context(), currentToken(r))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (api *API) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"user": currentUser(r)})
}
