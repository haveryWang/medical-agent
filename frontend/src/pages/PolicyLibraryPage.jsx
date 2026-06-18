import { Button, Empty, Input, List, Pagination, Popconfirm, Select, Space, Tag, Typography, Upload } from 'antd';
import { DeleteOutlined, DownloadOutlined, InboxOutlined, ReloadOutlined, SearchOutlined, UploadOutlined } from '@ant-design/icons';
import { usePolicyLibraryWorkspace } from '../features/policyLibrary/usePolicyLibraryWorkspace.js';
import { POLICY_TEXT_ELLIPSIS } from '../features/policyLibrary/policyLibrary.js';

export default function PolicyLibraryPage() {
  const workspace = usePolicyLibraryWorkspace();

  return (
    <div className="policy-page work-page">
      <header className="page-heading">
        <div>
          <Typography.Title level={3}>政策文件库</Typography.Title>
        </div>
        <Space>
          <Button
            icon={<DownloadOutlined />}
            onClick={() => workspace.downloadTemplate()}
            loading={workspace.downloadingTemplate}
          >
            下载导入模板
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => workspace.refresh()} loading={workspace.loading}>
            刷新
          </Button>
          {workspace.canImport ? (
            <Upload
              accept=".xlsx"
              showUploadList={false}
              beforeUpload={(file) => {
                void workspace.importFile(file);
                return false;
              }}
            >
              <Button icon={<UploadOutlined />} loading={workspace.importing}>
                导入 Excel
              </Button>
            </Upload>
          ) : null}
        </Space>
      </header>

      <main className="policy-content">
        <aside className="policy-categories">
          <div className="category-heading">主题分类</div>
          {workspace.categoryFacets.map((category) => (
            <button
              key={category.value}
              type="button"
              className={workspace.category === category.value ? 'policy-category active' : 'policy-category'}
              onClick={() => workspace.setCategory(category.value)}
            >
              <span>{category.label}</span>
              <span>{category.count}</span>
            </button>
          ))}
        </aside>

        <section className="policy-list-panel">
          <div className="policy-filter-bar">
            <Space size={10} wrap>
              <Typography.Text type="secondary">聚合筛选</Typography.Text>
              <Input
                allowClear
                prefix={<SearchOutlined />}
                placeholder="按名称搜索"
                value={workspace.keyword}
                onChange={(event) => workspace.setKeyword(event.target.value)}
                className="policy-keyword-input"
              />
              <Select
                allowClear
                placeholder="按日期"
                value={workspace.date || undefined}
                onChange={(value) => workspace.setDate(value || '')}
                className="policy-date-select"
                options={workspace.dateFacets.map((item) => ({
                  value: item.value,
                  label: `${item.label} (${item.count})`,
                }))}
              />
            </Space>
            <Typography.Text type="secondary">共 {workspace.total} 条</Typography.Text>
          </div>
          {workspace.lastImport ? (
            <div className="import-summary">
              <InboxOutlined />
              <span>最近导入 {workspace.lastImport.imported || 0} 条，跳过 {workspace.lastImport.skipped || 0} 条</span>
            </div>
          ) : null}
          <List
            loading={workspace.loading}
            dataSource={workspace.items}
            locale={{ emptyText: <Empty description="当前分类暂无政策文件" /> }}
            renderItem={(item) => (
              <List.Item
                className="policy-item"
                actions={workspace.canImport ? [
                  <Popconfirm
                    key="delete"
                    title="删除政策记录"
                    description="确认删除这条政策记录？"
                    okText="删除"
                    cancelText="取消"
                    onConfirm={() => workspace.deletePolicy(item)}
                  >
                    <Button danger size="small" icon={<DeleteOutlined />} loading={workspace.deletingPolicyId === item.id}>
                      删除
                    </Button>
                  </Popconfirm>,
                ] : []}
              >
                <div className="policy-copy">
                  <Space size={8} wrap className="policy-meta">
                    <Tag color="blue">{item.category}</Tag>
                    <span>{item.date || '未标注日期'}</span>
                  </Space>
                  <Typography.Title level={4}>{item.title}</Typography.Title>
                  <Typography.Paragraph ellipsis={POLICY_TEXT_ELLIPSIS}>
                    {item.summary}
                  </Typography.Paragraph>
                  {item.interpretation ? (
                    <Typography.Paragraph className="policy-interpretation" ellipsis={POLICY_TEXT_ELLIPSIS}>
                      {item.interpretation}
                    </Typography.Paragraph>
                  ) : null}
                </div>
              </List.Item>
            )}
          />
          <Pagination
            className="policy-pagination"
            current={workspace.page}
            pageSize={workspace.pageSize}
            total={workspace.total}
            showSizeChanger={false}
            onChange={workspace.setPage}
          />
        </section>
      </main>
    </div>
  );
}
