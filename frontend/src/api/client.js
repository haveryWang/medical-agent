import { consumeSSE } from './sse.js';

export const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

export function createApiClient({ getToken, onUnauthorized } = {}) {
  async function request(path, options = {}) {
    const headers = new Headers(options.headers || {});
    if (!(options.body instanceof FormData)) headers.set('Content-Type', 'application/json');

    const token = getToken?.();
    if (token) headers.set('Authorization', `Bearer ${token}`);

    const response = await fetch(`${API_BASE}${path}`, { ...options, headers });
    if (response.status === 401) {
      onUnauthorized?.();
    }
    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      throw new Error(body?.error?.message || `请求失败: ${response.status}`);
    }
    if (response.status === 204) return null;
    return response.json();
  }

  async function streamConversationMessage(conversationId, payload, onEvent) {
    const token = getToken?.();
    const response = await fetch(`${API_BASE}/conversations/${conversationId}/messages:stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(payload),
    });
    if (response.status === 401) onUnauthorized?.();
    if (!response.ok || !response.body) throw new Error('流式请求失败');
    await consumeSSE(response.body, onEvent);
  }

  return {
    request,
    login: (account, password) => request('/auth/login', { method: 'POST', body: JSON.stringify({ account, password }) }),
    me: () => request('/auth/me'),
    listKnowledgeBases: (params = {}) => request(`/knowledge-bases?${new URLSearchParams(params)}`),
    createKnowledgeBase: (payload) => request('/knowledge-bases', { method: 'POST', body: JSON.stringify(payload) }),
    updateKnowledgeBase: (id, payload) => request(`/knowledge-bases/${id}`, { method: 'PATCH', body: JSON.stringify(payload) }),
    uploadDocument: (id, file) => {
      const form = new FormData();
      form.append('file', file);
      return request(`/knowledge-bases/${id}/documents`, { method: 'POST', body: form, headers: {} });
    },
    listDocuments: (id) => request(`/knowledge-bases/${id}/documents`),
    listConversations: (keyword = '') => request(`/conversations?${new URLSearchParams({ keyword })}`),
    createConversation: (payload) => request('/conversations', { method: 'POST', body: JSON.stringify(payload) }),
    listMessages: (id) => request(`/conversations/${id}/messages`),
    messageDetails: (id) => request(`/messages/${id}/details`),
    streamConversationMessage,
  };
}
