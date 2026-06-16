import { useEffect, useState } from 'react';
import { useAuth } from '../../contexts/AuthContext.jsx';

export function useChatWorkspace() {
  const { api } = useAuth();
  const [conversations, setConversations] = useState([]);
  const [active, setActive] = useState(null);
  const [messages, setMessages] = useState([]);
  const [knowledgeBases, setKnowledgeBases] = useState([]);
  const [input, setInput] = useState('请问2型糖尿病的最新诊疗规范是什么？');
  const [streaming, setStreaming] = useState(false);
  const [detail, setDetail] = useState(null);

  useEffect(() => {
    let mounted = true;
    api.listKnowledgeBases({ page: 1, size: 50 }).then((res) => {
      if (mounted) setKnowledgeBases(res.items || []);
    });
    loadConversations(() => mounted);
    return () => {
      mounted = false;
    };
  }, []);

  async function loadConversations(isMounted = () => true) {
    const res = await api.listConversations();
    if (!isMounted()) return;
    if (res.items?.length) {
      setConversations(res.items);
      setActive(res.items[0]);
      const messagesRes = await api.listMessages(res.items[0].id);
      if (isMounted()) setMessages(messagesRes.items || []);
      return;
    }

    const conv = await api.createConversation({ title: '糖尿病治疗规范咨询' });
    if (!isMounted()) return;
    setConversations([conv]);
    setActive(conv);
    setMessages([]);
  }

  async function newConversation() {
    const conv = await api.createConversation({ title: '新的问答会话', knowledgeBaseIds: knowledgeBases.slice(0, 1).map((kb) => kb.id) });
    setConversations((prev) => [conv, ...prev]);
    setActive(conv);
    setMessages([]);
    setDetail(null);
  }

  async function openConversation(conv) {
    setActive(conv);
    setDetail(null);
    const res = await api.listMessages(conv.id);
    setMessages(res.items || []);
  }

  async function send() {
    if (!input.trim() || !active || streaming) return;
    const assistantID = `stream-${Date.now()}`;
    const userMsg = { id: `local-${Date.now()}`, role: 'user', content: input, status: 'completed', createdAt: new Date().toISOString() };
    const assistant = { id: assistantID, role: 'assistant', content: '', status: 'streaming', citations: [], createdAt: new Date().toISOString() };
    const question = input;

    setMessages((prev) => [...prev, userMsg, assistant]);
    setStreaming(true);
    setInput('');

    try {
      await api.streamConversationMessage(active.id, { content: question, knowledgeBaseIds: active.knowledgeBaseIds || [] }, (event, data) => {
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
    } catch (err) {
      patchMessage(assistantID, { status: 'failed', content: err.message });
    } finally {
      setStreaming(false);
    }
  }

  async function showDetail(message) {
    if (message.id?.startsWith('stream-') || message.id?.startsWith('local-')) {
      setDetail(message);
      return;
    }
    const res = await api.messageDetails(message.id);
    setDetail(res);
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
    detail,
    input,
    messages,
    streaming,
    newConversation,
    openConversation,
    send,
    setDetail,
    setInput,
    showDetail,
  };
}
