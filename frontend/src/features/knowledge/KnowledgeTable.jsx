import { App, Button, Space, Table, Tag, Typography } from 'antd';
import { DeleteOutlined, EditOutlined, EyeOutlined, PauseCircleOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { formatDate } from '../../utils/format.js';

export default function KnowledgeTable({
  items,
  loading,
  page,
  pageSize,
  selected,
  total,
  onChangePage,
  onChoose,
  onDelete,
  onEdit,
  onStatus,
  onViewDocuments,
}) {
  const { modal } = App.useApp();

  const columns = [
    {
      title: '知识库',
      dataIndex: 'name',
      render: (_, kb) => (
        <span className="kb-name-cell">
          <Typography.Text strong>{kb.name}</Typography.Text>
          <Typography.Text type="secondary" ellipsis className="kb-description">{kb.description || '-'}</Typography.Text>
        </span>
      ),
    },
    { title: '场景', dataIndex: 'scenario', render: (value) => value || '-' },
    { title: '部门', dataIndex: 'department', render: (value) => value || '-' },
    {
      title: '标签',
      dataIndex: 'tags',
      render: (tags = []) => tags.length ? tags.map((tag) => <Tag key={tag}>{tag}</Tag>) : '-',
    },
    {
      title: '文档/分片',
      render: (_, kb) => `${kb.documentCount || 0} / ${kb.chunkCount || 0}`,
    },
    {
      title: '状态',
      render: (_, kb) => (
        <Space size={4} wrap>
          <Tag color={kb.status === 'active' ? 'success' : 'default'}>{kb.status === 'active' ? '启用' : kb.status || '-'}</Tag>
          <Tag color={kb.buildStatus === 'building' ? 'processing' : 'blue'}>{buildStatusText(kb.buildStatus)}</Tag>
        </Space>
      ),
    },
    { title: '更新时间', dataIndex: 'updatedAt', render: formatDate },
    {
      title: '操作',
      key: 'actions',
      fixed: 'right',
      width: 340,
      render: (_, kb) => (
        <Space size={4} wrap={false}>
          <Button size="small" icon={<EyeOutlined />} onClick={(event) => action(event, () => onViewDocuments(kb))}>文档</Button>
          <Button size="small" icon={<EditOutlined />} onClick={(event) => action(event, () => onEdit(kb))}>编辑</Button>
          {kb.status === 'active' ? (
            <Button size="small" icon={<PauseCircleOutlined />} onClick={(event) => action(event, () => onStatus(kb, 'disabled'))}>停用</Button>
          ) : (
            <Button size="small" icon={<PlayCircleOutlined />} onClick={(event) => action(event, () => onStatus(kb, 'active'))}>启用</Button>
          )}
          <Button
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={(event) => action(event, () => {
              modal.confirm({
                title: '删除知识库',
                content: `确认删除“${kb.name}”？列表和问答检索将不再使用它。`,
                okText: '删除',
                okButtonProps: { danger: true },
                cancelText: '取消',
                onOk: () => onDelete(kb),
              });
            })}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Table
      rowKey="id"
      size="middle"
      columns={columns}
      dataSource={items}
      loading={loading}
      scroll={{ x: 1320 }}
      rowClassName={(kb) => (selected?.id === kb.id ? 'selected-table-row' : '')}
      onRow={(kb) => ({ onClick: () => onChoose(kb) })}
      pagination={{
        current: page,
        pageSize,
        total,
        showSizeChanger: true,
        showTotal: (value) => `共 ${value} 条`,
        onChange: onChangePage,
      }}
    />
  );
}

function action(event, callback) {
  event.stopPropagation();
  callback();
}

function buildStatusText(status) {
  if (status === 'completed') return '已完成';
  if (status === 'building') return '构建中';
  if (status === 'failed') return '失败';
  return status || '-';
}
