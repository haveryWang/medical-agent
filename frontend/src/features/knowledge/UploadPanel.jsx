import { App, Button, Empty, Popconfirm, Space, Spin, Tag, Tooltip, Typography, Upload } from 'antd';
import { DeleteOutlined, DownloadOutlined, FileSearchOutlined, FileTextOutlined, InboxOutlined, PartitionOutlined } from '@ant-design/icons';
import { formatDateTime, formatSize } from '../../utils/format.js';

const MAX_KB_FILE_BYTES = 15 * 1024 * 1024;
const TOO_LARGE_MESSAGE = '单个知识库文件不能超过 15MB，请切割单文件尺寸后再上传';
const SUPPORTED_EXTENSIONS = ['.pdf', '.docx', '.xlsx', '.xls', '.md', '.markdown', '.txt', '.csv'];
const ACCEPTED_UPLOAD_TYPES = SUPPORTED_EXTENSIONS.join(',');

export default function UploadPanel({ documents, loading, selected, uploading, onDelete, onDownload, onUpload, onViewChunks, onViewDocument }) {
  const { message } = App.useApp();

  return (
    <aside className="document-panel">
      <header>
        <Typography.Title level={5}>文档管理</Typography.Title>
        <Typography.Text type="secondary">{selected ? selected.name : '请选择知识库'}</Typography.Text>
      </header>
      <Upload.Dragger
        name="file"
        multiple={false}
        accept={ACCEPTED_UPLOAD_TYPES}
        showUploadList={false}
        disabled={!selected || uploading}
        beforeUpload={(file) => {
          if (file.size > MAX_KB_FILE_BYTES) {
            message.error(TOO_LARGE_MESSAGE);
            return Upload.LIST_IGNORE;
          }
          const ext = getFileExtension(file.name);
          if (ext === '.doc') {
            message.error('当前支持 Word .docx 文件，.doc 请先转换为 .docx 后上传');
            return Upload.LIST_IGNORE;
          }
          if (!SUPPORTED_EXTENSIONS.includes(ext)) {
            message.error('当前支持 PDF、Word(.docx)、Excel(.xlsx/.xls)、Markdown、TXT、CSV');
            return Upload.LIST_IGNORE;
          }
          return true;
        }}
        customRequest={async ({ file, onError, onSuccess }) => {
          try {
            await onUpload(file);
            onSuccess?.({});
          } catch (err) {
            onError?.(err);
          }
        }}
      >
        <p className="ant-upload-drag-icon"><InboxOutlined /></p>
        <p className="ant-upload-text">{uploading ? '上传中...' : '点击或拖拽文件上传'}</p>
        <p className="ant-upload-hint">支持 PDF、Word(.docx)、Excel(.xlsx/.xls)、Markdown、TXT、CSV，单个文件 15MB 以内，原始文件会保存到数据库并进入预处理与向量入库流程。</p>
      </Upload.Dragger>
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

function getFileExtension(name = '') {
  const index = name.lastIndexOf('.');
  return index >= 0 ? name.slice(index).toLowerCase() : '';
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
