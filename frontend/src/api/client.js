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

  async function download(path, filename, options = {}) {
    const headers = new Headers();
    if (options.body) headers.set('Content-Type', 'application/json');
    const token = getToken?.();
    if (token) headers.set('Authorization', `Bearer ${token}`);
    const response = await fetch(`${API_BASE}${path}`, { method: options.method || 'GET', headers, body: options.body });
    if (response.status === 401) onUnauthorized?.();
    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      throw new Error(body?.error?.message || `下载失败: ${response.status}`);
    }
    const blob = await response.blob();
    const disposition = response.headers.get('Content-Disposition') || response.headers.get('content-disposition') || '';
    const resolvedFilename = filename || filenameFromDisposition(disposition) || 'download';
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = resolvedFilename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
    return { filename: resolvedFilename };
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
    health: () => fetch(`${API_BASE.replace(/\/api\/v1$/, '')}/health`).then((response) => {
      if (!response.ok) throw new Error(`服务状态检查失败: ${response.status}`);
      return response.json();
    }),
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
    viewDocument: (kbId, docId) => request(`/knowledge-bases/${kbId}/documents/${docId}`),
    listDocumentChunks: (kbId, docId) => request(`/knowledge-bases/${kbId}/documents/${docId}/chunks`),
    downloadDocument: (kbId, doc) => download(`/knowledge-bases/${kbId}/documents/${doc.id}/download`, doc.fileName),
    deleteDocument: (kbId, docId) => request(`/knowledge-bases/${kbId}/documents/${docId}`, { method: 'DELETE' }),
    listReviewNotes: (params = {}) => request(`/review-notes?${new URLSearchParams(params)}`),
    createReviewNote: (payload) => request('/review-notes', { method: 'POST', body: JSON.stringify(payload) }),
    reviewNoteCounts: () => request('/review-notes/counts'),
    exportReviewNotes: (noteIds = []) => download('/review-notes:export', undefined, { method: 'POST', body: JSON.stringify({ noteIds }) }),
    deleteReviewNote: (id) => request(`/review-notes/${id}`, { method: 'DELETE' }),
    listReviewNoteExports: (params = {}) => request(`/review-notes/exports?${new URLSearchParams(params)}`),
    downloadReviewNoteExport: (id, filename) => download(`/review-notes/exports/${id}/download`, filename),
    listPolicyCategories: () => request('/policies/categories'),
    listPolicies: (params = {}) => request(`/policies?${new URLSearchParams(params)}`),
    deletePolicy: (id) => request(`/policies/${id}`, { method: 'DELETE' }),
    downloadPolicyTemplate: () => download('/policies/import-template', '政策文件库导入模板.xlsx'),
    importPolicies: (file) => {
      const form = new FormData();
      form.append('file', file);
      return request('/policies:import', { method: 'POST', body: form, headers: {} });
    },
    listConversations: (keyword = '') => request(`/conversations?${new URLSearchParams({ keyword })}`),
    createConversation: (payload) => request('/conversations', { method: 'POST', body: JSON.stringify(payload) }),
    updateConversation: (id, payload) => request(`/conversations/${id}`, { method: 'PATCH', body: JSON.stringify(payload) }),
    deleteConversation: (id) => request(`/conversations/${id}`, { method: 'DELETE' }),
    listMessages: (id) => request(`/conversations/${id}/messages`),
    messageDetails: (id) => request(`/messages/${id}/details`),
    getModelConfig: () => request('/system/model-config'),
    saveModelConfig: (payload) => request('/system/model-config', { method: 'PATCH', body: JSON.stringify(payload) }),
    streamConversationMessage,
  };
}

function filenameFromDisposition(disposition) {
  const utf8 = disposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8?.[1]) {
    try {
      return decodeURIComponent(utf8[1]);
    } catch {
      return utf8[1];
    }
  }
  const ascii = disposition.match(/filename="?([^";]+)"?/i);
  return ascii?.[1] || '';
}
