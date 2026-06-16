import { useEffect, useState } from 'react';
import { Alert, App, Form, Input, InputNumber, Modal, Typography } from 'antd';
import { useAuth } from '../../contexts/AuthContext.jsx';

const defaults = {
  deepSeekBaseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
  deepSeekAPIKey: '',
  deepSeekChatModel: 'deepseek-v4-flash-260425',
  qwenEmbeddingBaseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
  qwenEmbeddingAPIKey: '',
  qwenEmbeddingModel: 'doubao-embedding-vision-251215',
  qwenEmbeddingDimension: 2048,
};

export default function ModelSettingsModal({ open, onClose }) {
  const { api } = useAuth();
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [keyStatus, setKeyStatus] = useState('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!open) return;
    let mounted = true;
    setLoading(true);
    api.getModelConfig()
      .then((cfg) => {
        if (!mounted) return;
        form.setFieldsValue({
          deepSeekBaseUrl: cfg.deepSeekBaseUrl || defaults.deepSeekBaseUrl,
          deepSeekAPIKey: '',
          deepSeekChatModel: cfg.deepSeekChatModel || defaults.deepSeekChatModel,
          qwenEmbeddingBaseUrl: cfg.qwenEmbeddingBaseUrl || defaults.qwenEmbeddingBaseUrl,
          qwenEmbeddingAPIKey: '',
          qwenEmbeddingModel: cfg.qwenEmbeddingModel || defaults.qwenEmbeddingModel,
          qwenEmbeddingDimension: cfg.qwenEmbeddingDimension || defaults.qwenEmbeddingDimension,
        });
        const chatKey = cfg.deepSeekAPIKeySet ? `对话 Key 已保存（${cfg.deepSeekAPIKeyPreview}）` : '对话 Key 未配置';
        const embeddingKey = cfg.qwenEmbeddingAPIKeySet ? `向量 Key 已保存（${cfg.qwenEmbeddingAPIKeyPreview}）` : '向量 Key 未配置';
        setKeyStatus(`${chatKey}；${embeddingKey}`);
      })
      .catch((err) => message.error(err.message))
      .finally(() => mounted && setLoading(false));
    return () => {
      mounted = false;
    };
  }, [open, api, form, message]);

  async function save() {
    const values = await form.validateFields();
    setSaving(true);
    try {
      const saved = await api.saveModelConfig({
        ...values,
        qwenEmbeddingDimension: Number(values.qwenEmbeddingDimension),
      });
      const chatKey = saved.deepSeekAPIKeySet ? `对话 Key 已保存（${saved.deepSeekAPIKeyPreview}）` : '对话 Key 未配置';
      const embeddingKey = saved.qwenEmbeddingAPIKeySet ? `向量 Key 已保存（${saved.qwenEmbeddingAPIKeyPreview}）` : '向量 Key 未配置';
      setKeyStatus(`${chatKey}；${embeddingKey}`);
      form.setFieldsValue({ deepSeekAPIKey: '', qwenEmbeddingAPIKey: '' });
      message.success('模型配置已保存到数据库');
      onClose();
    } catch (err) {
      message.error(err.message);
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal
      title="模型配置"
      open={open}
      onCancel={onClose}
      onOk={save}
      confirmLoading={saving}
      okText="保存配置"
      cancelText="取消"
      width={820}
      destroyOnHidden
      loading={loading}
    >
      <Typography.Paragraph type="secondary">
        配置保存到数据库，后端问答和向量任务执行时会读取当前值。API Key 留空表示保留数据库中已保存的密钥。
      </Typography.Paragraph>
      {keyStatus ? <Alert type="info" showIcon message={keyStatus} className="settings-alert" /> : null}
      <Form form={form} layout="vertical" initialValues={defaults} className="settings-form">
        <div className="settings-form-grid">
          <Form.Item name="deepSeekBaseUrl" label="对话 Base URL" rules={[{ required: true, message: '请输入对话 Base URL' }]}>
            <Input placeholder="https://ark.cn-beijing.volces.com/api/v3" />
          </Form.Item>
          <Form.Item name="deepSeekChatModel" label="对话模型" rules={[{ required: true, message: '请输入对话模型' }]}>
            <Input placeholder="deepseek-v4-flash-260425" />
          </Form.Item>
          <Form.Item name="deepSeekAPIKey" label="对话 API Key" className="settings-form-full">
            <Input.Password placeholder="留空则保留已保存密钥" autoComplete="off" />
          </Form.Item>
          <Form.Item name="qwenEmbeddingBaseUrl" label="向量 Base URL" rules={[{ required: true, message: '请输入向量 Base URL' }]}>
            <Input placeholder="https://ark.cn-beijing.volces.com/api/v3" />
          </Form.Item>
          <Form.Item name="qwenEmbeddingModel" label="向量模型" rules={[{ required: true, message: '请输入向量模型' }]}>
            <Input placeholder="doubao-embedding-vision-251215" />
          </Form.Item>
          <Form.Item name="qwenEmbeddingAPIKey" label="向量 API Key">
            <Input.Password placeholder="留空则保留已保存密钥" autoComplete="off" />
          </Form.Item>
          <Form.Item name="qwenEmbeddingDimension" label="向量维度" rules={[{ required: true, message: '请输入向量维度' }]}>
            <InputNumber min={1} precision={0} className="full-width" />
          </Form.Item>
        </div>
      </Form>
    </Modal>
  );
}
