import { Button, Card, Checkbox, Col, Empty, Input, List, Pagination, Popconfirm, Row, Space, Statistic, Tag, Typography } from 'antd';
import { DeleteOutlined, DownloadOutlined, FileDoneOutlined, FileTextOutlined, ReloadOutlined, SendOutlined } from '@ant-design/icons';
import { useReviewNotesWorkspace } from '../features/reviewNotes/useReviewNotesWorkspace.js';

export default function ReviewNotesPage() {
  const workspace = useReviewNotesWorkspace();

  return (
    <div className="review-page work-page">
      <header className="page-heading">
        <div>
          <Typography.Title level={3}>复盘笔记</Typography.Title>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={workspace.refresh} loading={workspace.loading}>
            刷新
          </Button>
          <Button type="primary" icon={<DownloadOutlined />} onClick={workspace.exportMarkdown} loading={workspace.exporting} disabled={!workspace.selectedCount}>
            生成选中记录
          </Button>
        </Space>
      </header>

      <main className="review-content">
        <Row gutter={[16, 16]} className="review-grid">
          <Col xs={24} xl={9}>
            <Card className="tool-card" title="记录复盘" variant="outlined">
              <div className="note-input-wrap">
                <Input.TextArea
                  value={workspace.draft}
                  onChange={(event) => workspace.setDraft(event.target.value)}
                  placeholder="记录一次可复用的判断、流程口径、沟通经验或风险提醒"
                  autoSize={{ minRows: 8, maxRows: 12 }}
                  maxLength={3000}
                  showCount
                />
              </div>
              <Button className="submit-note" type="primary" icon={<SendOutlined />} loading={workspace.saving} onClick={workspace.submit}>
                提交记录
              </Button>
            </Card>
            <Row gutter={12} className="note-stats">
              <Col span={12}>
                <Card className="metric-card" variant="outlined">
                  <Statistic title="未导出" value={workspace.counts.unexported || 0} suffix="条" />
                </Card>
              </Col>
              <Col span={12}>
                <Card className="metric-card" variant="outlined">
                  <Statistic title="累计记录" value={workspace.counts.total || 0} suffix="条" />
                </Card>
              </Col>
            </Row>
            <Card className="tool-card export-history-card" title="导出记录" variant="outlined">
              {workspace.exports.length ? (
                <List
                  size="small"
                  dataSource={workspace.exports}
                  renderItem={(item) => (
                    <List.Item
                      actions={[
                        <Button
                          key="download"
                          type="link"
                          size="small"
                          icon={<DownloadOutlined />}
                          loading={workspace.downloadingExportId === item.id}
                          onClick={() => workspace.downloadExport(item)}
                        >
                          再次下载
                        </Button>,
                      ]}
                    >
                      <List.Item.Meta
                        avatar={<span className="note-icon"><FileDoneOutlined /></span>}
                        title={item.filename || '复盘笔记导出.md'}
                        description={`${formatDateTime(item.createdAt)} · ${item.noteCount || 0} 条`}
                      />
                    </List.Item>
                  )}
                />
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无导出记录" />
              )}
            </Card>
          </Col>
          <Col xs={24} xl={15}>
            <Card
              className="tool-card note-list-card"
              title="记录列表"
              variant="outlined"
              loading={workspace.loading}
              extra={<Typography.Text type="secondary">已选 {workspace.selectedCount} 条</Typography.Text>}
            >
              {workspace.notes.length ? (
                <>
                  <div className="note-scroll-list">
                    <List
                      dataSource={workspace.notes}
                      renderItem={(note) => (
                        <List.Item
                          actions={[
                            <Popconfirm
                              key="delete"
                              title="删除复盘记录"
                              description="确认删除这条记录？已生成的导出文档不会被自动删除。"
                              okText="删除"
                              cancelText="取消"
                              onConfirm={() => workspace.deleteNote(note)}
                            >
                              <Button danger size="small" icon={<DeleteOutlined />} loading={workspace.deletingNoteId === note.id}>
                                删除
                              </Button>
                            </Popconfirm>,
                          ]}
                        >
                          <Checkbox
                            className="note-select"
                            checked={workspace.selectedNoteIds.includes(note.id)}
                            onChange={(event) => {
                              const checked = event.target.checked;
                              workspace.setSelectedNoteIds((current) => (
                                checked ? [...current, note.id] : current.filter((id) => id !== note.id)
                              ));
                            }}
                          />
                          <List.Item.Meta
                            avatar={<span className="note-icon"><FileTextOutlined /></span>}
                            title={(
                              <Space size={8} wrap>
                                <Typography.Text>{formatDateTime(note.createdAt)}</Typography.Text>
                                <Tag color={note.exported ? 'default' : 'blue'}>{note.exported ? '已导出' : '未导出'}</Tag>
                              </Space>
                            )}
                            description={<Typography.Paragraph className="note-content" ellipsis={{ rows: 4, expandable: true, symbol: '展开' }}>{note.content}</Typography.Paragraph>}
                          />
                        </List.Item>
                      )}
                    />
                  </div>
                  <Pagination
                    className="note-pagination"
                    current={workspace.page}
                    pageSize={workspace.reviewNotePageSize}
                    total={workspace.total}
                    showSizeChanger={false}
                    onChange={(nextPage) => {
                      workspace.setPage(nextPage);
                      void workspace.refresh(nextPage);
                    }}
                  />
                </>
              ) : (
                <Empty description="暂无复盘笔记" />
              )}
            </Card>
          </Col>
        </Row>
      </main>
    </div>
  );
}

function formatDateTime(value) {
  if (!value) return '未记录时间';
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}
