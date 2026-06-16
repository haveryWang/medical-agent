import { useRef } from 'react';
import { App, Button, Empty, Popconfirm, Space, Spin, Tag, Tooltip, Typography, Upload } from 'antd';
import { DeleteOutlined, DownloadOutlined, FileSearchOutlined, FileTextOutlined, InboxOutlined, PartitionOutlined } from '@ant-design/icons';
import { formatDateTime, formatSize } from '../../utils/format.js';
import { ACCEPTED_UPLOAD_TYPES, prepareUploadBatch } from './uploadBatch.js';

export default function UploadPanel({
  documents,
  loading,
  selected,
  uploadQueue = [],
  uploading,
  onDelete,
  onDownload,
  onUpload,
  onViewChunks,
  onViewDocument,
}) {
  const { message } = App.useApp();
  const uploadTimer = useRef(null);

  function scheduleUpload(fileList) {
    window.clearTimeout(uploadTimer.current);
    uploadTimer.current = window.setTimeout(async () => {
      const { validFiles, errors } = prepareUploadBatch(fileList);
      errors.forEach(({ file, message: errorMessage }) => {
        message.error(`${file?.name || '文件'}：${errorMessage}`);
      });
      if (!validFiles.length) return;
      try {
        await onUpload(validFiles);
      } catch {
        // useKnowledgeWorkspace 已经统一 toast，这里只负责阻止 Upload 默认错误弹层重复出现。
      }
    }, 0);
  }

  return (
    <aside className="document-panel">
      <header>
        <Typography.Title level={5}>文档管理</Typography.Title>
        <Typography.Text type="secondary">{selected ? selected.name : '请选择知识库'}</Typography.Text>
      </header>
      <Upload.Dragger
        name="file"
        multiple
        accept={ACCEPTED_UPLOAD_TYPES}
        showUploadList={false}
        disabled={!selected || uploading}
        beforeUpload={(_, fileList) => {
          scheduleUpload(fileList);
          return Upload.LIST_IGNORE;
        }}
      >
        <p className="ant-upload-drag-icon"><InboxOutlined /></p>
        <p className="ant-upload-text">{uploading ? `正在上传 ${uploadQueue.length || ''} 个文件` : '点击或拖拽文件批量上传'}</p>
        <p className="ant-upload-hint">支持 PDF、Word(.docx)、Excel(.xlsx/.xls)、Markdown、TXT、CSV，单个文件 15MB 以内</p>
      </Upload.Dragger>
      {uploading && uploadQueue.length ? (
        <div className="upload-queue">
          {uploadQueue.map((name, index) => (
            <Tag key={`${name}-${index}`}>{name}</Tag>
          ))}
        </div>
      ) : null}
      <div className="document-list">
        {loading ? (
          <div className="loading-state"><Spin /></div>
        ) : (
          <div className="document-items">
            {documents.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无文档" /> : null}
            {documents.map((doc) => (
              <div className="document-item" key={doc.id}>
                <FileTextOutlined className="doc-icon" />
                <div className="document-copy">
                  <Typography.Text ellipsis>{doc.fileName}</Typography.Text>
                  <Space size={8} wrap>
                    <Typography.Text type="secondary">{formatSize(doc.sizeBytes)}</Typography.Text>
                    <Typography.Text type="secondary">{formatDateTime(doc.updatedAt || doc.createdAt)}</Typography.Text>
                  </Space>
                  {doc.failureReason ? <Typography.Text type="danger">{doc.failureReason}</Typography.Text> : null}
                  <Space size={4} wrap>
                    <Tooltip title="查看文档内容">
                      <Button size="small" icon={<FileSearchOutlined />} onClick={() => onViewDocument(doc)} />
                    </Tooltip>
                    <Tooltip title="查看分片">
                      <Button size="small" icon={<PartitionOutlined />} onClick={() => onViewChunks(doc)} />
                    </Tooltip>
                    <Tooltip title="下载原始文件">
                      <Button size="small" icon={<DownloadOutlined />} onClick={() => onDownload(doc)} />
                    </Tooltip>
                    <Popconfirm
                      title="删除文档"
                      description="会删除原始文件、分片和向量索引，确认继续？"
                      okText="删除"
                      cancelText="取消"
                      okButtonProps={{ danger: true }}
                      onConfirm={() => onDelete(doc)}
                    >
                      <Tooltip title="删除文档">
                        <Button danger size="small" icon={<DeleteOutlined />} />
                      </Tooltip>
                    </Popconfirm>
                  </Space>
                </div>
                <Tag color={docStatusColor(doc.status)}>{docStatusText(doc.status)}</Tag>
              </div>
            ))}
          </div>
        )}
      </div>
    </aside>
  );
}

function docStatusColor(status) {
  if (status === 'completed') return 'success';
  if (status === 'failed') return 'error';
  if (status === 'pending') return 'processing';
  return 'default';
}

function docStatusText(status) {
  if (status === 'completed') return '已入库';
  if (status === 'failed') return '失败';
  if (status === 'pending') return '待处理';
  return status || '-';
}
