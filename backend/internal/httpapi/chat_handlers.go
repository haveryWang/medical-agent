package httpapi

import (
	"context"
	"net/http"
	"strings"

	"medical-agent/backend/internal/models"
	"medical-agent/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func (api *API) listConversations(w http.ResponseWriter, r *http.Request) {
	items, err := api.store.ListConversations(r.Context(), currentUser(r).ID, r.URL.Query().Get("keyword"))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (api *API) createConversation(w http.ResponseWriter, r *http.Request) {
	var req createConversationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	kbIDs := parseObjectIDs(req.KnowledgeBaseIDs)
	conversation, err := api.store.CreateConversation(r.Context(), currentUser(r).ID, req.Title, kbIDs)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, conversation)
}

func (api *API) updateConversation(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "会话 ID 无效")
		return
	}
	var req updateConversationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	patch := bson.M{}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "会话标题不能为空")
			return
		}
		patch["title"] = title
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if status != "active" && status != "archived" {
			writeError(w, r, http.StatusBadRequest, "validation_error", "会话状态无效")
			return
		}
		patch["status"] = status
	}
	if req.KnowledgeBaseIDs != nil {
		patch["knowledgeBaseIds"] = parseObjectIDs(req.KnowledgeBaseIDs)
	}
	if len(patch) == 0 {
		writeError(w, r, http.StatusBadRequest, "validation_error", "没有可更新字段")
		return
	}
	conversation, err := api.store.UpdateConversation(r.Context(), id, currentUser(r).ID, patch)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "会话不存在")
		return
	}
	writeJSON(w, http.StatusOK, conversation)
}

func (api *API) deleteConversation(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "会话 ID 无效")
		return
	}
	if err := api.store.DeleteConversation(r.Context(), id, currentUser(r).ID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (api *API) listMessages(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "会话 ID 无效")
		return
	}
	if _, err := api.store.GetConversation(r.Context(), id, currentUser(r).ID); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "会话不存在")
		return
	}
	messages, err := api.store.ListMessages(r.Context(), id)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": messages})
}

func (api *API) messageDetails(w http.ResponseWriter, r *http.Request) {
	id, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "消息 ID 无效")
		return
	}
	message, err := api.store.GetMessage(r.Context(), id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "消息不存在")
		return
	}
	if _, err := api.store.GetConversation(r.Context(), message.ConversationID, currentUser(r).ID); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "消息不存在")
		return
	}
	writeJSON(w, http.StatusOK, message)
}

func (api *API) streamMessage(w http.ResponseWriter, r *http.Request) {
	conversationID, err := store.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "会话 ID 无效")
		return
	}
	conversation, err := api.store.GetConversation(r.Context(), conversationID, currentUser(r).ID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "会话不存在")
		return
	}
	var req streamMessageRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "消息不能为空")
		return
	}
	_, _ = api.store.CreateMessage(r.Context(), models.Message{ConversationID: conversationID, Role: "user", Content: content, Status: "completed"})
	kbIDs := conversation.KnowledgeBaseIDs
	if req.KnowledgeBaseIDs != nil {
		kbIDs = parseObjectIDs(req.KnowledgeBaseIDs)
	}
	retrieval, err := api.rag.Retrieve(r.Context(), content, kbIDs)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "retrieval_failed", err.Error())
		return
	}
	assistant, err := api.store.CreateMessage(r.Context(), models.Message{ConversationID: conversationID, Role: "assistant", Status: "streaming", Citations: retrieval.Citations, PromptContext: retrieval.Context, ModelName: api.cfg.DeepSeekChatModel})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "message_failed", err.Error())
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, r, http.StatusInternalServerError, "stream_unsupported", "当前服务不支持流式输出")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	sendSSE(w, flusher, "message.started", map[string]any{"messageId": assistant.ID.Hex(), "conversationId": conversationID.Hex()})
	sendSSE(w, flusher, "retrieval.sources", map[string]any{"sources": retrieval.Citations})
	answer, duration, err := api.rag.StreamAnswer(r.Context(), content, retrieval, func(delta string) error {
		sendSSE(w, flusher, "message.delta", map[string]string{"text": delta})
		return nil
	})
	if err != nil {
		_ = api.store.UpdateMessage(context.Background(), assistant.ID, bson.M{"status": "failed", "content": answer, "durationMs": duration.Milliseconds()})
		sendSSE(w, flusher, "message.error", map[string]string{"code": "model_error", "message": err.Error(), "requestId": requestID(r)})
		return
	}
	_ = api.store.UpdateMessage(context.Background(), assistant.ID, bson.M{"status": "completed", "content": answer, "durationMs": duration.Milliseconds(), "citations": retrieval.Citations, "promptContext": retrieval.Context})
	sendSSE(w, flusher, "message.completed", map[string]any{"messageId": assistant.ID.Hex(), "durationMs": duration.Milliseconds(), "citationCount": len(retrieval.Citations)})
}
