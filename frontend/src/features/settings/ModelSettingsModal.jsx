import { useEffect, useState } from 'react';
import { useAuth } from '../../contexts/AuthContext.jsx';

const emptyForm = {
  deepSeekBaseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
  deepSeekAPIKey: '',
  deepSeekChatModel: 'DeepSeek-V4-flash',
  qwenEmbeddingBaseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
  qwenEmbeddingAPIKey: '',
  qwenEmbeddingModel: 'doubao-embedding-vision-251215',
  qwenEmbeddingDimension: 1024,
};

export default function ModelSettingsModal({ open, onClose }) {
  const { api } = useAuth();
  const [form, setForm] = useState(emptyForm);
  const [status, setStatus] = useState('');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!open) return;
    let mounted = true;
    setError('');
    setStatus('正在读取当前配置...');
    api.getModelConfig()
      .then((cfg) => {
        if (!mounted) return;
        setForm({
          deepSeekBaseUrl: cfg.deepSeekBaseUrl || '',
          deepSeekAPIKey: '',
          deepSeekChatModel: cfg.deepSeekChatModel || 'DeepSeek-V4-flash',
          qwenEmbeddingBaseUrl: cfg.qwenEmbeddingBaseUrl || '',
          qwenEmbeddingAPIKey: '',
          qwenEmbeddingModel: cfg.qwenEmbeddingModel || 'doubao-embedding-vision-251215',
          qwenEmbeddingDimension: cfg.qwenEmbeddingDimension || 1024,
        });
        const deepSeek = cfg.deepSeekAPIKeySet ? `DeepSeek Key 已保存（${cfg.deepSeekAPIKeyPreview}）` : 'DeepSeek Key 未配置';
        const qwen = cfg.qwenEmbeddingAPIKeySet ? `Qwen Key 已保存（${cfg.qwenEmbeddingAPIKeyPreview}）` : 'Qwen Key 未配置';
        setStatus(`${deepSeek}；${qwen}`);
      })
      .catch((err) => {
        if (!mounted) return;
        setStatus('');
        setError(err.message);
      });
    return () => {
      mounted = false;
    };
  }, [open, api]);

  if (!open) return null;

  function update(name, value) {
    setForm((prev) => ({ ...prev, [name]: value }));
  }

  async function save(e) {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      const saved = await api.saveModelConfig({
        ...form,
        qwenEmbeddingDimension: Number(form.qwenEmbeddingDimension),
      });
      const deepSeek = saved.deepSeekAPIKeySet ? `DeepSeek Key 已保存（${saved.deepSeekAPIKeyPreview}）` : 'DeepSeek Key 未配置';
      const qwen = saved.qwenEmbeddingAPIKeySet ? `Qwen Key 已保存（${saved.qwenEmbeddingAPIKeyPreview}）` : 'Qwen Key 未配置';
      setStatus(`${deepSeek}；${qwen}`);
      setForm((prev) => ({ ...prev, deepSeekAPIKey: '', qwenEmbeddingAPIKey: '' }));
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="modal-backdrop" role="presentation">
      <section className="settings-modal" role="dialog" aria-modal="true" aria-labelledby="model-settings-title">
        <header>
          <div>
            <h2 id="model-settings-title">模型配置</h2>
            <p>配置会保存到数据库，后端入库和问答任务执行时会读取这里的值。</p>
          </div>
          <button className="modal-close" onClick={onClose} aria-label="关闭">×</button>
        </header>
        <form onSubmit={save}>
          <div className="settings-grid">
            <label>
              <span>DeepSeek Base URL</span>
              <input value={form.deepSeekBaseUrl} onChange={(e) => update('deepSeekBaseUrl', e.target.value)} placeholder="https://ark.cn-beijing.volces.com/api/v3" />
            </label>
            <label>
              <span>DeepSeek Chat Model</span>
              <input value={form.deepSeekChatModel} onChange={(e) => update('deepSeekChatModel', e.target.value)} placeholder="DeepSeek-V4-flash" />
            </label>
            <label className="full">
              <span>DeepSeek API Key</span>
              <input type="password" value={form.deepSeekAPIKey} onChange={(e) => update('deepSeekAPIKey', e.target.value)} placeholder="留空则保留已保存密钥" autoComplete="off" />
            </label>
            <label>
              <span>火山向量 Base URL</span>
              <input value={form.qwenEmbeddingBaseUrl} onChange={(e) => update('qwenEmbeddingBaseUrl', e.target.value)} placeholder="https://ark.cn-beijing.volces.com/api/v3" />
            </label>
            <label>
              <span>火山向量模型</span>
              <input value={form.qwenEmbeddingModel} onChange={(e) => update('qwenEmbeddingModel', e.target.value)} placeholder="doubao-embedding-vision-251215" />
            </label>
            <label>
              <span>火山引擎 API Key</span>
              <input type="password" value={form.qwenEmbeddingAPIKey} onChange={(e) => update('qwenEmbeddingAPIKey', e.target.value)} placeholder="留空则保留已保存密钥" autoComplete="off" />
            </label>
            <label>
              <span>向量维度</span>
              <input type="number" min="1" value={form.qwenEmbeddingDimension} onChange={(e) => update('qwenEmbeddingDimension', e.target.value)} />
            </label>
          </div>
          {status ? <p className="settings-status">{status}</p> : null}
          {error ? <p className="form-error">{error}</p> : null}
          <footer>
            <button type="button" className="ghost bordered" onClick={onClose}>取消</button>
            <button type="submit" className="primary" disabled={saving}>{saving ? '保存中...' : '保存配置'}</button>
          </footer>
        </form>
      </section>
    </div>
  );
}
