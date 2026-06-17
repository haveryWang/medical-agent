import { useEffect, useMemo, useState } from 'react';
import { App } from 'antd';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext.jsx';
import { normalizeReviewNoteDraft, normalizeSelectedReviewNoteIds, reviewNotePageSize } from './reviewNotes.js';

export function useReviewNotesWorkspace() {
  const { api } = useAuth();
  const { message: toast } = App.useApp();
  const location = useLocation();
  const navigate = useNavigate();
  const [notes, setNotes] = useState([]);
  const [exports, setExports] = useState([]);
  const [counts, setCounts] = useState({ total: 0, unexported: 0 });
  const [draft, setDraft] = useState('');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [selectedNoteIds, setSelectedNoteIds] = useState([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [deletingNoteId, setDeletingNoteId] = useState('');
  const [downloadingExportId, setDownloadingExportId] = useState('');

  useEffect(() => {
    void refresh();
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const source = params.get('draft');
    if (!source) return;
    setDraft(source);
    navigate('/review-notes', { replace: true });
  }, [location.search, navigate]);

  const selectedCount = useMemo(() => selectedNoteIds.length, [selectedNoteIds]);

  async function refresh(nextPage = page) {
    setLoading(true);
    try {
      const [list, nextCounts, exportList] = await Promise.all([
        api.listReviewNotes({ page: nextPage, size: reviewNotePageSize }),
        api.reviewNoteCounts(),
        api.listReviewNoteExports({ limit: 5 }),
      ]);
      const nextItems = list.items || [];
      setNotes(nextItems);
      setPage(list.page || nextPage);
      setTotal(list.total || 0);
      setSelectedNoteIds((current) => {
        const visible = new Set(nextItems.map((item) => String(item.id)));
        return current.filter((id) => visible.has(id));
      });
      setCounts(nextCounts || { total: 0, unexported: 0 });
      setExports(exportList.items || []);
    } catch (err) {
      toast.error(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function submit() {
    const normalized = normalizeReviewNoteDraft(draft);
    if (!normalized.valid) {
      toast.error('请输入复盘笔记内容');
      return false;
    }
    setSaving(true);
    try {
      await api.createReviewNote({ content: normalized.content });
      setDraft('');
      toast.success('复盘笔记已记录');
      await refresh(1);
      return true;
    } catch (err) {
      toast.error(err.message);
      return false;
    } finally {
      setSaving(false);
    }
  }

  async function exportMarkdown() {
    const noteIds = normalizeSelectedReviewNoteIds(selectedNoteIds);
    if (!noteIds.length) {
      toast.error('请先选择要生成文档的记录');
      return;
    }
    setExporting(true);
    try {
      const result = await api.exportReviewNotes(noteIds);
      toast.success(`已生成 ${result?.filename || 'Markdown 文件'}`);
      setSelectedNoteIds([]);
      await refresh();
    } catch (err) {
      toast.error(err.message);
    } finally {
      setExporting(false);
    }
  }

  async function deleteNote(note) {
    if (!note?.id) return;
    setDeletingNoteId(note.id);
    try {
      await api.deleteReviewNote(note.id);
      toast.success('复盘记录已删除');
      setSelectedNoteIds((current) => current.filter((id) => id !== note.id));
      const nextPage = notes.length === 1 && page > 1 ? page - 1 : page;
      await refresh(nextPage);
    } catch (err) {
      toast.error(err.message);
    } finally {
      setDeletingNoteId('');
    }
  }

  async function downloadExport(item) {
    if (!item?.id) return;
    setDownloadingExportId(item.id);
    try {
      await api.downloadReviewNoteExport(item.id, item.filename);
    } catch (err) {
      toast.error(err.message);
    } finally {
      setDownloadingExportId('');
    }
  }

  return {
    counts,
    deletingNoteId,
    draft,
    exporting,
    exports,
    downloadingExportId,
    loading,
    notes,
    page,
    reviewNotePageSize,
    saving,
    selectedCount,
    selectedNoteIds,
    total,
    deleteNote,
    downloadExport,
    exportMarkdown,
    refresh,
    setDraft,
    setPage,
    setSelectedNoteIds,
    submit,
  };
}
