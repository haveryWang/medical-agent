import { useEffect, useRef, useState } from 'react';
import { useAuth } from '../../contexts/AuthContext.jsx';

const initialFilters = { scenario: '', tag: '', department: '', keyword: '' };

export function useKnowledgeWorkspace() {
  const { api } = useAuth();
  const uploadInputRef = useRef(null);
  const [items, setItems] = useState([]);
  const [filters, setFilters] = useState(initialFilters);
  const [selected, setSelected] = useState(null);
  const [documents, setDocuments] = useState([]);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    load();
  }, []);

  async function load(next = filters) {
    const res = await api.listKnowledgeBases({ ...next, page: 1, size: 10 });
    setItems(res.items || []);
    if (!selected && res.items?.length) {
      setSelected(res.items[0]);
      const docs = await api.listDocuments(res.items[0].id);
      setDocuments(docs.items || []);
    }
  }

  async function choose(kb) {
    setSelected(kb);
    const docs = await api.listDocuments(kb.id);
    setDocuments(docs.items || []);
  }

  async function upload(file) {
    if (!selected || !file) return;
    setUploading(true);
    try {
      await api.uploadDocument(selected.id, file);
      const docs = await api.listDocuments(selected.id);
      setDocuments(docs.items || []);
      await load();
    } finally {
      setUploading(false);
    }
  }

  function openUploader() {
    uploadInputRef.current?.click();
  }

  return {
    documents,
    filters,
    items,
    selected,
    uploadInputRef,
    uploading,
    choose,
    load,
    openUploader,
    setFilters,
    upload,
  };
}
