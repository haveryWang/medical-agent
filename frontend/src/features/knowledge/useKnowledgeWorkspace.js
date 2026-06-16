import { useEffect, useMemo, useState } from 'react';
import { App } from 'antd';
import { useAuth } from '../../contexts/AuthContext.jsx';
import { requireModelConfig } from '../../utils/modelConfig.js';

const initialFilters = { scenario: '', tag: '', department: '', keyword: '' };

export function useKnowledgeWorkspace() {
  const { api } = useAuth();
  const { message: toast } = App.useApp();
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [filters, setFilters] = useState(initialFilters);
  const [selected, setSelected] = useState(null);
  const [documents, setDocuments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [documentsLoading, setDocumentsLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadQueue, setUploadQueue] = useState([]);
  const [documentDrawerOpen, setDocumentDrawerOpen] = useState(false);
  const [documentViewer, setDocumentViewer] = useState({ open: false, title: '', content: '', loading: false });
  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState(null);

  useEffect(() => {
    load();
  }, []);

  const filterOptions = useMemo(() => ({
    scenarios: unique(items.map((item) => item.scenario)),
    tags: unique(items.flatMap((item) => item.tags || [])),
    departments: unique(items.map((item) => item.department)),
  }), [items]);

  const hasBuildingItems = useMemo(
    () => items.some((item) => item.buildStatus === 'building'),
    [items],
  );

  useEffect(() => {
    if (!hasBuildingItems) return undefined;

    const timer = window.setInterval(() => {
      void load(filters, page, pageSize, { silent: true });
    }, 10000);

    return () => window.clearInterval(timer);
  }, [hasBuildingItems, filters, page, pageSize, documentDrawerOpen]);

  async function load(nextFilters = filters, nextPage = page, nextSize = pageSize, options = {}) {
    const { silent = false } = options;
    if (!silent) setLoading(true);
    try {
      const res = await api.listKnowledgeBases({ ...compact(nextFilters), page: nextPage, size: nextSize });
      const nextItems = res.items || [];
      setItems(nextItems);
      setTotal(res.total || 0);
      setPage(res.page || nextPage);
      setPageSize(res.size || nextSize);
      if (!nextItems.length) {
        setSelected(null);
        setDocuments([]);
        return;
      }
      const nextSelected = nextItems.find((item) => item.id === selected?.id) || nextItems[0];
      setSelected(nextSelected);
      if (documentDrawerOpen) {
        await loadDocuments(nextSelected, false);
      }
    } catch (err) {
      if (!silent) toast.error(err.message);
    } finally {
      if (!silent) setLoading(false);
    }
  }

  async function search() {
    setPage(1);
    await load(filters, 1, pageSize);
  }

  async function resetFilters() {
    setFilters(initialFilters);
    setPage(1);
    await load(initialFilters, 1, pageSize);
  }

  async function choose(kb, openDrawer = false) {
    setSelected(kb);
    if (openDrawer) setDocumentDrawerOpen(true);
    await loadDocuments(kb);
  }

  async function loadDocuments(kb = selected, showErrors = true) {
    if (!kb) return;
    setDocumentsLoading(true);
    try {
      const docs = await api.listDocuments(kb.id);
      setDocuments(docs.items || []);
    } catch (err) {
      if (showErrors) toast.error(err.message);
    } finally {
      setDocumentsLoading(false);
    }
  }

  function createKnowledgeBase() {
    setEditing(null);
    setEditorOpen(true);
  }

  function editKnowledgeBase(kb) {
    setEditing(kb);
    setEditorOpen(true);
  }

  async function saveKnowledgeBase(values) {
    const payload = {
      ...values,
      tags: normalizeTags(values.tags),
      retrievalTopK: Number(values.retrievalTopK || 0),
      similarityFloor: Number(values.similarityFloor || 0),
    };
    try {
      if (editing) {
        await api.updateKnowledgeBase(editing.id, payload);
        toast.success('知识库已更新');
      } else {
        await api.createKnowledgeBase(payload);
        toast.success('知识库已创建');
      }
      setEditorOpen(false);
      await load();
    } catch (err) {
      toast.error(err.message);
      throw err;
    }
  }

  async function setKnowledgeStatus(kb, status) {
    try {
      await api.updateKnowledgeBase(kb.id, { status });
      toast.success(status === 'active' ? '知识库已启用' : '知识库已停用');
      await load();
    } catch (err) {
      toast.error(err.message);
    }
  }

  async function deleteKnowledgeBase(kb) {
    try {
      await api.updateKnowledgeBase(kb.id, { status: 'deleted' });
      toast.success('知识库已删除');
      await load();
    } catch (err) {
      toast.error(err.message);
    }
  }

  async function upload(files) {
    if (!selected) return;
    const list = Array.isArray(files) ? files.filter(Boolean) : files ? [files] : [];
    if (!list.length) return;
    setUploading(true);
    setUploadQueue(list.map((file) => file.name));
    try {
      await requireModelConfig(api, ['embedding']);
      for (const file of list) {
        await api.uploadDocument(selected.id, file);
      }
      toast.success('文档已上传，后台正在入库处理');
      await loadDocuments(selected);
      await load();
    } catch (err) {
      toast.error(err.message);
      throw err;
    } finally {
      setUploading(false);
      setUploadQueue([]);
    }
  }

  async function viewDocument(doc) {
    if (!selected || !doc) return;
    setDocumentViewer({ open: true, title: doc.fileName, content: '', loading: true });
    try {
      const res = await api.viewDocument(selected.id, doc.id);
      setDocumentViewer({
        open: true,
        title: res.document?.fileName || doc.fileName,
        content: res.content || '',
        loading: false,
      });
    } catch (err) {
      toast.error(err.message);
      setDocumentViewer((current) => ({ ...current, loading: false, content: err.message }));
    }
  }

  async function viewChunks(doc) {
    if (!selected || !doc) return;
    setDocumentViewer({ open: true, title: `${doc.fileName} · 分片`, content: '', loading: true });
    try {
      const res = await api.listDocumentChunks(selected.id, doc.id);
      const chunks = res.items || [];
      const content = chunks.length
        ? chunks.map((chunk) => `#${chunk.chunkIndex + 1}\n${chunk.text}`).join('\n\n---\n\n')
        : '暂无分片内容';
      setDocumentViewer({ open: true, title: `${doc.fileName} · 分片`, content, loading: false });
    } catch (err) {
      toast.error(err.message);
      setDocumentViewer((current) => ({ ...current, loading: false, content: err.message }));
    }
  }

  async function downloadDocument(doc) {
    if (!selected || !doc) return;
    try {
      await api.downloadDocument(selected.id, doc);
    } catch (err) {
      toast.error(err.message);
    }
  }

  async function deleteDocument(doc) {
    if (!selected || !doc) return;
    try {
      await api.deleteDocument(selected.id, doc.id);
      toast.success('文档已删除');
      await loadDocuments(selected);
      await load(filters, page, pageSize, { silent: true });
    } catch (err) {
      toast.error(err.message);
    }
  }

  return {
    documents,
    documentsLoading,
    documentDrawerOpen,
    documentViewer,
    editing,
    editorOpen,
    filterOptions,
    filters,
    items,
    loading,
    page,
    pageSize,
    selected,
    total,
    uploading,
    uploadQueue,
    choose,
    createKnowledgeBase,
    deleteKnowledgeBase,
    editKnowledgeBase,
    load,
    loadDocuments,
    resetFilters,
    saveKnowledgeBase,
    search,
    setDocumentDrawerOpen,
    setDocumentViewer,
    setEditorOpen,
    setFilters,
    setKnowledgeStatus,
    upload,
    viewChunks,
    viewDocument,
    downloadDocument,
    deleteDocument,
  };
}

function compact(value) {
  return Object.fromEntries(Object.entries(value).filter(([, item]) => item !== '' && item !== undefined && item !== null));
}

function unique(values) {
  return [...new Set(values.filter(Boolean))];
}

function normalizeTags(value) {
  if (Array.isArray(value)) return value.filter(Boolean);
  return String(value || '')
    .split(/[,，\s]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}
