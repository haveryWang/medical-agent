import { useEffect, useRef } from 'react';
import { App, Avatar, Button, Empty, Space, Spin, Tag, Tooltip, Typography } from 'antd';
import { CopyOutlined, FileSearchOutlined, AppstoreAddOutlined, RedoOutlined, RobotOutlined, UserOutlined } from '@ant-design/icons';
import { formatDuration } from '../../utils/format.js';
import { messageToReviewDraft } from '../reviewNotes/reviewNotes.js';

export default function MessageList({ loading, messages, onRegenerate, onSendToReview, onShowDetail }) {
  const { message: toast } = App.useApp();
  const bottomRef = useRef(null);
  const latestMessage = messages[messages.length - 1];

  useEffect(() => {
    if (loading || messages.length === 0) return;
    bottomRef.current?.scrollIntoView({ block: 'end' });
  }, [loading, messages.length, latestMessage?.content, latestMessage?.status]);

  async function copy(content) {
    await navigator.clipboard.writeText(content || '');
    toast.success('内容已复制');
  }

  if (loading) {
    return (
      <div className="messages loading-state">
        <Spin />
      </div>
    );
  }

  return (
    <div className="messages">
      {messages.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="暂无消息，输入问题后会基于数据库中的知识库检索生成回答"
          className="message-empty"
        />
      ) : null}
      {messages.map((msg) => {
        const isAssistant = msg.role === 'assistant';
        return (
          <article key={msg.id} className={isAssistant ? 'message assistant' : 'message user'}>
            <Avatar className={isAssistant ? 'assistant-avatar' : 'user-avatar'} icon={isAssistant ? <RobotOutlined /> : <UserOutlined />} />
            <div className="bubble">
              <Typography.Paragraph>{msg.content || (msg.status === 'streaming' ? '正在生成...' : '')}</Typography.Paragraph>
              <Space size={6} wrap className="message-meta">
                {msg.status ? <Tag color={statusColor(msg.status)}>{statusText(msg.status)}</Tag> : null}
                {isAssistant && msg.modelName ? <Tag>{msg.modelName}</Tag> : null}
                {isAssistant && msg.durationMs ? <Typography.Text type="secondary">{formatDuration(msg.durationMs)}</Typography.Text> : null}
              </Space>
              {msg.citations?.length > 0 ? (
                <div className="citations">
                  <Typography.Text strong>引用来源</Typography.Text>
                  {msg.citations.slice(0, 4).map((cite, idx) => (
                    <Typography.Text key={`${cite.chunkId}-${idx}`} type="secondary">
                      {idx + 1}. {cite.title || cite.documentId}
                    </Typography.Text>
                  ))}
                </div>
              ) : null}
              <Space size={4} className="message-actions">
                <Tooltip title="复制内容">
                  <Button size="small" icon={<CopyOutlined />} onClick={() => copy(msg.content)} />
                </Tooltip>
                <Tooltip title="添加到复盘笔记">
                  <Button size="small" icon={<AppstoreAddOutlined />} disabled={!msg.content} onClick={() => onSendToReview(messageToReviewDraft(msg))} />
                </Tooltip>
                {isAssistant ? (
                  <Tooltip title="重新生成">
                    <Button size="small" icon={<RedoOutlined />} disabled={msg.status === 'streaming'} onClick={() => onRegenerate(msg)} />
                  </Tooltip>
                ) : null}
                {isAssistant ? (
                  <Tooltip title="查看详情">
                    <Button size="small" icon={<FileSearchOutlined />} onClick={() => onShowDetail(msg)} />
                  </Tooltip>
                ) : null}
              </Space>
            </div>
          </article>
        );
      })}
      <div ref={bottomRef} aria-hidden="true" />
    </div>
  );
}

function statusColor(status) {
  if (status === 'completed') return 'success';
  if (status === 'failed') return 'error';
  if (status === 'streaming') return 'processing';
  return 'default';
}

function statusText(status) {
  if (status === 'completed') return '已完成';
  if (status === 'failed') return '失败';
  if (status === 'streaming') return '生成中';
  return status;
}
