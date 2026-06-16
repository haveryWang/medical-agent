import { Button, Descriptions, Spin, Tag, Typography } from 'antd';
import { formatDuration } from '../../utils/format.js';

export default function AnswerDetail({ detail, loading }) {
  if (loading) {
    return (
      <div className="detail-body loading-state">
        <Spin />
      </div>
    );
  }

  if (!detail) return null;

  return (
    <div className="detail-body">
      <Descriptions size="small" column={1} bordered>
        <Descriptions.Item label="消息 ID">{detail.id || '-'}</Descriptions.Item>
        <Descriptions.Item label="模型">{detail.modelName || '-'}</Descriptions.Item>
        <Descriptions.Item label="耗时">{formatDuration(detail.durationMs)}</Descriptions.Item>
        <Descriptions.Item label="状态">{detail.status || '-'}</Descriptions.Item>
      </Descriptions>
      <section>
        <Typography.Title level={5}>知识引用</Typography.Title>
        <div className="reference-list">
          {(detail.citations || []).length === 0 ? <Typography.Text type="secondary">无引用来源</Typography.Text> : null}
          {(detail.citations || []).map((cite, idx) => (
            <div className="reference-item" key={`${cite.chunkId || cite.documentId}-${idx}`}>
              <div>
                <Typography.Text strong>{idx + 1}. {cite.title || cite.documentId || '未命名来源'}</Typography.Text>
                <Typography.Paragraph type="secondary">{cite.snippet || cite.chunkId}</Typography.Paragraph>
              </div>
              {typeof cite.score === 'number' ? <Tag color="blue">{Math.round(cite.score * 100)}%</Tag> : null}
            </div>
          ))}
        </div>
      </section>
      <section>
        <Typography.Title level={5}>发送给模型的上下文</Typography.Title>
        <pre className="prompt-context">{detail.promptContext || '当前消息未记录上下文。'}</pre>
      </section>
    </div>
  );
}
