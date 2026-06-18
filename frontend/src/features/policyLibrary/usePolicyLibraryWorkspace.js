import { useEffect, useMemo, useState } from 'react';
import { App } from 'antd';
import { useAuth } from '../../contexts/AuthContext.jsx';
import {
  POLICY_CATEGORIES,
  POLICY_FILTER_DEBOUNCE_MS,
  POLICY_TEMPLATE_FILENAME,
  buildPolicyListParams,
  normalizePolicyFacets,
  preparePolicyImportFile,
} from './policyLibrary.js';

export function usePolicyLibraryWorkspace() {
  const { api, user } = useAuth();
  const { message: toast } = App.useApp();
  const [category, setCategory] = useState(POLICY_CATEGORIES[0]);
  const [date, setDate] = useState('');
  const [keyword, setKeyword] = useState('');
  const [debouncedKeyword, setDebouncedKeyword] = useState('');
  const [facets, setFacets] = useState(() => normalizePolicyFacets());
  const [items, setItems] = useState([]);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [deletingPolicyId, setDeletingPolicyId] = useState('');
  const [downloadingTemplate, setDownloadingTemplate] = useState(false);
  const [importing, setImporting] = useState(false);
  const [lastImport, setLastImport] = useState(null);

  const canImport = useMemo(() => (user?.permissions || []).includes('policy:write'), [user]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebouncedKeyword(keyword);
      setPage(1);
    }, POLICY_FILTER_DEBOUNCE_MS);
    return () => window.clearTimeout(timer);
  }, [keyword]);

  useEffect(() => {
    void load({ category, date, keyword: debouncedKeyword, page, pageSize });
  }, [category, date, debouncedKeyword, page, pageSize]);

  async function load(filters = { category, date, keyword: debouncedKeyword, page, pageSize }) {
    setLoading(true);
    try {
      const res = await api.listPolicies(buildPolicyListParams(filters));
      setItems(res.items || []);
      setPage(res.page || filters.page || 1);
      setTotal(res.total || 0);
      setFacets(normalizePolicyFacets(res.facets));
    } catch (err) {
      toast.error(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function importFile(file) {
    const prepared = preparePolicyImportFile(file);
    if (prepared.error) {
      toast.error(prepared.error);
      return;
    }
    setImporting(true);
    try {
      const res = await api.importPolicies(prepared.file);
      setLastImport(res.report || null);
      toast.success(`导入完成：${res.report?.imported || 0} 条`);
      await load({ category, date, keyword: debouncedKeyword, page: 1, pageSize });
    } catch (err) {
      toast.error(err.message);
    } finally {
      setImporting(false);
    }
  }

  async function downloadTemplate() {
    setDownloadingTemplate(true);
    try {
      await api.downloadPolicyTemplate();
      toast.success(`${POLICY_TEMPLATE_FILENAME} 已开始下载`);
    } catch (err) {
      toast.error(err.message);
    } finally {
      setDownloadingTemplate(false);
    }
  }

  async function deletePolicy(item) {
    if (!item?.id) return;
    setDeletingPolicyId(item.id);
    try {
      await api.deletePolicy(item.id);
      toast.success('政策记录已删除');
      const nextPage = items.length === 1 && page > 1 ? page - 1 : page;
      await load({ category, date, keyword: debouncedKeyword, page: nextPage, pageSize });
    } catch (err) {
      toast.error(err.message);
    } finally {
      setDeletingPolicyId('');
    }
  }

  return {
    canImport,
    categories: POLICY_CATEGORIES,
    category,
    date,
    keyword,
    dateFacets: facets.dates,
    categoryFacets: facets.categories,
    deletingPolicyId,
    downloadingTemplate,
    importing,
    items,
    lastImport,
    loading,
    page,
    pageSize,
    total,
    deletePolicy,
    downloadTemplate,
    importFile,
    load,
    refresh: () => load({ category, date, keyword, page, pageSize }),
    setCategory: (value) => {
      setPage(1);
      setCategory(value);
    },
    setDate: (value) => {
      setPage(1);
      setDate(value);
    },
    setKeyword: (value) => {
      setPage(1);
      setKeyword(value);
    },
    setPage,
  };
}
