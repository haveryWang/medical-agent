package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"medical-agent/backend/internal/models"
)

type contextKey string

const userKey contextKey = "user"
const tokenKey contextKey = "token"
const requestIDKey contextKey = "requestID"

func (api *API) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-Id", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *API) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "请先登录")
			return
		}
		session, err := api.store.FindSessionByToken(r.Context(), token)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "会话无效或已过期")
			return
		}
		user, err := api.store.FindUserByID(r.Context(), session.UserID)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "用户不存在")
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		ctx = context.WithValue(ctx, tokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *API) requirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := currentUser(r)
			for _, current := range user.Permissions {
				if current == permission {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, r, http.StatusForbidden, "forbidden", "当前账号没有操作权限")
		})
	}
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return ""
}

func currentUser(r *http.Request) models.User {
	user, _ := r.Context().Value(userKey).(models.User)
	return user
}

func currentToken(r *http.Request) string {
	token, _ := r.Context().Value(tokenKey).(string)
	return token
}

func requestID(r *http.Request) string {
	value, _ := r.Context().Value(requestIDKey).(string)
	return value
}
