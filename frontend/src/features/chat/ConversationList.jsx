import { useState } from 'react';
import { App, Button, Dropdown, Empty, Input, Modal, Select, Space, Tag, Typography } from 'antd';
import { DeleteOutlined, EditOutlined, MoreOutlined, PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { formatDateTime } from '../../utils/format.js';

export default function ConversationList({
  active,
  conversations,
  knowledgeBases,
  keyword,
  loading,
  onDelete,
  onKnowledgeBaseChange,
  onKeywordChange,
  onNew,
  onOpen,
  onRename,
  onSearch,
  selectedKnowledgeBaseId,
  serviceStatus,
}) {
  const { modal } = App.useApp();
  const [renameTarget, setRenameTarget] = useState(null);
  const [title, setTitle] = useState('');
  const activeKnowledgeBases = (knowledgeBases || []).filter((kb) => kb.status === 'active');
  const knowledgeBaseMap = new Map(activeKnowledgeBases.map((kb) => [String(kb.id), kb]));

  function openRename(conv) {
    setRenameTarget(conv);
    setTitle(conv.title || '');
  }

  async function submitRename() {
    if (!renameTarget || !title.trim()) return;
    await onRename(renameTarget, title.trim());
    setRenameTarget(null);
  }

  function confirmDelete(conv) {
    modal.confirm({
      title: '删除会话',
      content: `确认删除“${conv.title}”？历史消息会从列表中移除。`,
      okText: '删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: () => onDelete(conv),
    });
  }

  return (
    <aside className="conversation-pane">
      <Button type="primary" icon={<PlusOutlined />} block onClick={onNew}>
        新建对话
      </Button>
      <Input
        allowClear
        prefix={<SearchOutlined />}
        placeholder="搜索会话标题"
        value={keyword}
        onChange={(event) => onKeywordChange(event.target.value)}
        onPressEnter={onSearch}
      />
      <div className="retrieval-scope">
        <Typography.Text strong>知识库</Typography.Text>
        <Select
          allowClear
          showSearch
          value={selectedKnowledgeBaseId || undefined}
          placeholder="不选则关闭知识库"
          optionFilterProp="label"
          disabled={!activeKnowledgeBases.length}
          onChange={onKnowledgeBaseChange}
          options={activeKnowledgeBases.map((kb) => ({
            value: String(kb.id),
            label: kb.name,
            title: kb.description || kb.name,
            kb,
          }))}
          filterOption={(input, option) => {
            const kb = option.kb || option.data?.kb || {};
            const text = `${kb.name || option.label || ''} ${kb.description || ''}`.toLowerCase();
            return text.includes(input.trim().toLowerCase());
          }}
          optionRender={(option) => {
            const kb = option.data?.kb || {};
            return (
              <div className="kb-option">
                <Typography.Text>{option.label}</Typography.Text>
                <Typography.Text type="secondary">
                  {kb.documentCount || 0} 文档 / {kb.chunkCount || 0} 分片
                </Typography.Text>
              </div>
            );
          }}
        />
        <Typography.Text type="secondary" className="scope-hint">
          {scopeHint(active, selectedKnowledgeBaseId, knowledgeBaseMap)}
        </Typography.Text>
        {activeKnowledgeBases.length === 0 ? <Tag color="warning">暂无可用知识库</Tag> : null}
      </div>
      <div className={loading ? 'conversation-list loading' : 'conversation-list'}>
        {!loading && conversations.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无会话" /> : null}
        {conversations.map((conv) => (
          <div
            key={conv.id}
            tabIndex={0}
            className={active?.id === conv.id ? 'conversation-item active' : 'conversation-item'}
            onClick={() => onOpen(conv)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') onOpen(conv);
            }}
          >
            <span className="conversation-copy">
              <Typography.Text ellipsis strong>{conv.title || '未命名会话'}</Typography.Text>
              <Typography.Text type="secondary">{formatDateTime(conv.updatedAt || conv.createdAt)}</Typography.Text>
              <Typography.Text type="secondary">
                {conversationScopeText(conv.knowledgeBaseIds, knowledgeBaseMap)}
              </Typography.Text>
            </span>
            <Dropdown
              trigger={['click']}
              menu={{
                items: [
                  { key: 'rename', icon: <EditOutlined />, label: '重命名' },
                  { key: 'delete', icon: <DeleteOutlined />, label: '删除', danger: true },
                ],
                onClick: ({ key, domEvent }) => {
                  domEvent.stopPropagation();
                  if (key === 'rename') openRename(conv);
                  if (key === 'delete') confirmDelete(conv);
                },
              }}
            >
              <Button
                type="text"
                size="small"
                icon={<MoreOutlined />}
                onClick={(event) => event.stopPropagation()}
                aria-label="会话操作"
              />
            </Dropdown>
          </div>
        ))}
      </div>
      <Modal
        title="重命名会话"
        open={Boolean(renameTarget)}
        onCancel={() => setRenameTarget(null)}
        onOk={submitRename}
        okText="保存"
        cancelText="取消"
      >
        <Input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="请输入会话标题" maxLength={60} showCount />
      </Modal>
      <ServiceStatus status={serviceStatus} />
    </aside>
  );
}

function scopeHint(active, selectedId, knowledgeBaseMap) {
  if (!selectedId) {
    return active ? '当前未选择知识库，发送时不走检索增强' : '未选择知识库，发送首条消息时不走检索增强';
  }
  const name = knowledgeBaseMap.get(String(selectedId))?.name || '当前知识库';
  return active ? `当前知识库：${name}` : `已选知识库：${name}，发送首条消息时生效`;
}

function conversationScopeText(knowledgeBaseIds = [], knowledgeBaseMap) {
  const id = knowledgeBaseIds.find((item) => knowledgeBaseMap.has(String(item)));
  if (!id) return '未选择知识库';
  return `知识库：${knowledgeBaseMap.get(String(id))?.name || '未知知识库'}`;
}

function ServiceStatus({ status }) {
  const state = status?.status || 'checking';
  const checks = status?.checks || {};
  return (
    <div className="service-status">
      <Space size={6} align="center">
        <span className={`status-dot ${serviceStatusTone(state)}`} />
        <Typography.Text strong>服务状态</Typography.Text>
        <Tag color={serviceStatusColor(state)}>{serviceStatusText(state)}</Tag>
      </Space>
      <Typography.Text type="secondary">
        MongoDB：{checks.mongodb || '-'} · Qdrant：{checks.qdrant || '-'}
      </Typography.Text>
      <Typography.Text type="secondary">
        对话：{modelCheckText(checks.deepseek)} · 向量：{modelCheckText(checks.qwenEmbedding)}
      </Typography.Text>
    </div>
  );
}

function serviceStatusTone(status) {
  if (status === 'ok') return 'ok';
  if (status === 'checking') return 'checking';
  return 'down';
}

function serviceStatusColor(status) {
  if (status === 'ok') return 'success';
  if (status === 'checking') return 'processing';
  return 'error';
}

function serviceStatusText(status) {
  if (status === 'ok') return '正常';
  if (status === 'checking') return '检查中';
  if (status === 'degraded') return '降级';
  if (status === 'down') return '异常';
  return status || '未知';
}

function modelCheckText(value = '') {
  if (value === 'configured') return '已配置';
  if (!value) return '-';
  return value.includes('missing') ? '未配置' : value;
}
