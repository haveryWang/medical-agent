import { useEffect, useMemo, useState } from 'react';
import { App } from 'antd';
import { useAuth } from '../../contexts/AuthContext.jsx';
import { requireModelConfig } from '../../utils/modelConfig.js';

const GENERIC_TITLES = new Set(['新的问答会话', '新建会话']);

export function useChatWorkspace() {
  const { api } = useAuth();
  const { message: toast } = App.useApp();
  const [conversations, setConversations] = useState([]);
  const [active, setActive] = useState(null);
  const [messages, setMessages] = useState([]);
  const [knowledgeBases, setKnowledgeBases] = useState([]);
  const [selectedKnowledgeBaseIds, setSelectedKnowledgeBaseIds] = useState([]);
  const [input, setInput] = useState('');
  const [searchKeyword, setSearchKeyword] = useState('');
  const [streaming, setStreaming] = useState(false);
  const [detail, setDetail] = useState(null);
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [loadingConversations, setLoadingConversations] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [serviceStatus, setServiceStatus] = useState({ status: 'checking', checks: {}, updatedAt: '' });

  const activeKnowledgeBaseIds = useMemo(
    () => knowledgeBases.filter((kb) => kb.status === 'active').map((kb) => String(kb.id)),
    [knowledgeBases],
  );
  const visibleSelectedKnowledgeBaseIds = useMemo(
    () => selectedKnowledgeBaseIds.filter((id) => activeKnowledgeBaseIds.includes(String(id))),
    [activeKnowledgeBaseIds, selectedKnowledgeBaseIds],
  );
  const activeKnowledgeBaseMap = useMemo(
    () => new Map(knowledgeBases.filter((kb) => kb.status === 'active').map((kb) => [String(kb.id), kb])),
    [knowledgeBases],
  );
  const selectedKnowledgeBaseNames = knowledgeBaseNames(visibleSelectedKnowledgeBaseIds, activeKnowledgeBaseMap);
  const paneHint = selectedKnowledgeBaseNames
    ? `消息将携带知识库：${selectedKnowledgeBaseNames}`
    : '未选择知识库，本次消息不走检索增强';

  useEffect(() => {
    let mounted = true;
    api.listKnowledgeBases({ page: 1, size: 100 })
      .then((res) => mounted && setKnowledgeBases(res.items || []))
      .catch((err) => toast.error(err.message));
    loadConversations('', () => mounted);
    loadServiceStatus(() => mounted);
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    setSelectedKnowledgeBaseIds((prev) => prev.filter((id) => activeKnowledgeBaseIds.includes(String(id))));
  }, [activeKnowledgeBaseIds]);

  useEffect(() => {
    let mounted = true;
    const timer = window.setInterval(() => {
      void loadServiceStatus(() => mounted, { silent: true });
    }, 15000);
    return () => {
      mounted = false;
      window.clearInterval(timer);
    };
  }, []);

  async function loadServiceStatus(isMounted = () => true, options = {}) {
    try {
      const res = await api.health();
      if (!isMounted()) return;
      setServiceStatus({ status: res.status || 'unknown', checks: res.checks || {}, updatedAt: new Date().toISOString() });
    } catch (err) {
      if (!isMounted()) return;
      setServiceStatus({ status: 'down', checks: { api: err.message }, updatedAt: new Date().toISOString() });
      if (!options.silent) toast.error(err.message);
    }
  }

  async function loadConversations(keyword = searchKeyword, isMounted = () => true) {
    setLoadingConversations(true);
    try {
      const res = await api.listConversations(keyword);
      if (!isMounted()) return;
      const list = res.items || [];
      setConversations(list);
      if (!list.length) {
        setActive(null);
        setMessages([]);
        return;
      }
      const nextActive = list.find((item) => item.id === active?.id) || list[0];
      setActive(nextActive);
      setSelectedKnowledgeBaseIds(normalizeKnowledgeBaseIDs(nextActive.knowledgeBaseIds));
      await loadMessages(nextActive, isMounted);
    } catch (err) {
      if (isMounted()) toast.error(err.message);
    } finally {
      if (isMounted()) setLoadingConversations(false);
    }
  }

  async function loadMessages(conv = active, isMounted = () => true) {
    if (!conv) return;
    setLoadingMessages(true);
    try {
      const res = await api.listMessages(conv.id);
      if (isMounted()) setMessages(res.items || []);
    } catch (err) {
      if (isMounted()) toast.error(err.message);
    } finally {
      if (isMounted()) setLoadingMessages(false);
    }
  }

  function closeDetail() {
    setDetailDrawerOpen(false);
    setDetailLoading(false);
    setDetail(null);
  }

  async function newConversation() {
    try {
      const conv = await api.createConversation({
        title: '新的问答会话',
        knowledgeBaseIds: visibleSelectedKnowledgeBaseIds,
      });
      setConversations((prev) => [conv, ...prev]);
      setActive(conv);
      setSelectedKnowledgeBaseIds(normalizeKnowledgeBaseIDs(conv.knowledgeBaseIds));
      setMessages([]);
      closeDetail();
      toast.success('会话已创建');
    } catch (err) {
      toast.error(err.message);
    }
  }

  async function openConversation(conv) {
    setActive(conv);
    setSelectedKnowledgeBaseIds(normalizeKnowledgeBaseIDs(conv.knowledgeBaseIds));
    closeDetail();
    await loadMessages(conv);
  }

  async function renameConversation(conv, title) {
    try {
      const updated = await api.updateConversation(conv.id, { title });
      setConversations((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      if (active?.id === updated.id) setActive(updated);
      toast.success('会话标题已更新');
    } catch (err) {
      toast.error(err.message);
    }
  }

  async function deleteConversation(conv) {
    try {
      await api.deleteConversation(conv.id);
      const remaining = conversations.filter((item) => item.id !== conv.id);
      setConversations(remaining);
      if (active?.id === conv.id) {
        const next = remaining[0] || null;
        setActive(next);
        closeDetail();
        if (next) {
          setSelectedKnowledgeBaseIds(normalizeKnowledgeBaseIDs(next.knowledgeBaseIds));
          await loadMessages(next);
        } else {
          setMessages([]);
          setSelectedKnowledgeBaseIds([]);
        }
      }
      toast.success('会话已删除');
    } catch (err) {
      toast.error(err.message);
    }
  }

  async function ensureConversation(question) {
    if (active) return active;
    const conv = await api.createConversation({
      title: makeTitle(question),
      knowledgeBaseIds: visibleSelectedKnowledgeBaseIds,
    });
    setConversations((prev) => [conv, ...prev]);
    setActive(conv);
    setSelectedKnowledgeBaseIds(normalizeKnowledgeBaseIDs(conv.knowledgeBaseIds));
    return conv;
  }

  async function send(questionOverride = '') {
    const question = (questionOverride || input).trim();
    if (!question || streaming) return;

    const scopeIds = effectiveKnowledgeBaseIds();
    setStreaming(true);
    try {
      await requireModelConfig(api, scopeIds.length ? ['embedding', 'chat'] : ['chat']);
    } catch (err) {
      setStreaming(false);
      toast.error(err.message);
      return;
    }
    setInput('');
    closeDetail();

    let conv;
    try {
      conv = await ensureConversation(question);
    } catch (err) {
      setStreaming(false);
      toast.error(err.message);
      return;
    }

    const assistantID = `stream-${Date.now()}`;
    const userMsg = { id: `local-${Date.now()}`, role: 'user', content: question, status: 'completed', createdAt: new Date().toISOString() };
    const assistant = { id: assistantID, role: 'assistant', content: '', status: 'streaming', citations: [], createdAt: new Date().toISOString() };

    setMessages((prev) => [...prev, userMsg, assistant]);

    try {
      await api.streamConversationMessage(conv.id, { content: question, knowledgeBaseIds: scopeIds }, (event, data) => {
        if (event === 'retrieval.sources') {
          patchMessage(assistantID, { citations: data.sources || [] });
        }
        if (event === 'message.delta') {
          appendMessageContent(assistantID, data.text || '');
        }
        if (event === 'message.completed') {
          patchMessage(assistantID, { id: data.messageId, status: 'completed', durationMs: data.durationMs });
        }
        if (event === 'message.error') {
          patchMessage(assistantID, { status: 'failed', content: data.message || '生成失败' });
        }
      });
      if (GENERIC_TITLES.has(conv.title)) {
        const updated = await api.updateConversation(conv.id, { title: makeTitle(question) });
        setActive(updated);
        setConversations((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      }
      await loadMessages(conv, () => true);
      await refreshConversationList();
    } catch (err) {
      patchMessage(assistantID, { status: 'failed', content: err.message });
      toast.error(err.message);
    } finally {
      setStreaming(false);
    }
  }

  async function regenerate(message) {
    const index = messages.findIndex((item) => item.id === message.id);
    const priorUser = [...messages.slice(0, index)].reverse().find((item) => item.role === 'user');
    if (!priorUser?.content) {
      toast.warning('未找到可重新生成的问题');
      return;
    }
    await send(priorUser.content);
  }

  async function showDetail(message) {
    setDetailDrawerOpen(true);
    if (message.id?.startsWith('stream-') || message.id?.startsWith('local-')) {
      setDetailLoading(false);
      setDetail(message);
      return;
    }
    setDetailLoading(true);
    setDetail(null);
    try {
      const res = await api.messageDetails(message.id);
      setDetail(res);
    } catch (err) {
      toast.error(err.message);
      closeDetail();
    } finally {
      setDetailLoading(false);
    }
  }

  async function updateKnowledgeBaseScope(ids) {
    const nextIDs = normalizeKnowledgeBaseIDs(ids).filter((id) => activeKnowledgeBaseIds.includes(id));
    setSelectedKnowledgeBaseIds(nextIDs);
    if (!active) return;
    try {
      const updated = await api.updateConversation(active.id, { knowledgeBaseIds: nextIDs });
      setActive(updated);
      setConversations((prev) => prev.map((item) => (item.id === updated.id ? updated : item)));
      toast.success(nextIDs.length ? `已选择知识库：${knowledgeBaseNames(nextIDs, activeKnowledgeBaseMap) || '知识库'}` : '已关闭知识库');
    } catch (err) {
      setSelectedKnowledgeBaseIds(normalizeKnowledgeBaseIDs(active.knowledgeBaseIds));
      toast.error(err.message);
    }
  }

  function effectiveKnowledgeBaseIds() {
    return visibleSelectedKnowledgeBaseIds;
  }

  async function refreshConversationList() {
    const res = await api.listConversations(searchKeyword);
    setConversations(res.items || []);
  }

  function patchMessage(id, patch) {
    setMessages((prev) => prev.map((msg) => (msg.id === id ? { ...msg, ...patch } : msg)));
  }

  function appendMessageContent(id, text) {
    setMessages((prev) => prev.map((msg) => (msg.id === id ? { ...msg, content: msg.content + text } : msg)));
  }

  return {
    active,
    conversations,
    closeDetail,
    detail,
    detailDrawerOpen,
    detailLoading,
    input,
    knowledgeBaseMap: activeKnowledgeBaseMap,
    knowledgeBases,
    loadingConversations,
    loadingMessages,
    messages,
    paneHint,
    searchKeyword,
    selectedKnowledgeBaseIds: visibleSelectedKnowledgeBaseIds,
    serviceStatus,
    streaming,
    deleteConversation,
    loadConversations,
    newConversation,
    openConversation,
    regenerate,
    renameConversation,
    send,
    setInput,
    setSearchKeyword,
    showDetail,
    updateKnowledgeBaseScope,
  };
}

function makeTitle(content) {
  return content.trim().slice(0, 24) || '新的问答会话';
}

function normalizeKnowledgeBaseIDs(ids = []) {
  if (!Array.isArray(ids)) return ids ? [String(ids)] : [];
  return ids.map(String).filter(Boolean);
}

function knowledgeBaseNames(ids, knowledgeBaseMap) {
  return ids
    .map((id) => knowledgeBaseMap.get(String(id))?.name)
    .filter(Boolean)
    .join('、');
}
