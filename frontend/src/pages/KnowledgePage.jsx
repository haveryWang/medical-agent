import { useEffect, useState } from 'react';
import { Drawer, Form, Input, InputNumber, Layout, Modal, Typography } from 'antd';
import KnowledgeFilters from '../features/knowledge/KnowledgeFilters.jsx';
import KnowledgeTable from '../features/knowledge/KnowledgeTable.jsx';
import UploadPanel from '../features/knowledge/UploadPanel.jsx';
import { useKnowledgeWorkspace } from '../features/knowledge/useKnowledgeWorkspace.js';

const { Content } = Layout;

export default function KnowledgePage() {
  const knowledge = useKnowledgeWorkspace();

  return (
    <Layout className="knowledge-layout">
      <Content className="kb-main">
        <header className="page-heading">
          <div>
            <Typography.Title level={3}>知识库管理</Typography.Title>
          </div>
        </header>
        <KnowledgeFilters
          filters={knowledge.filters}
          options={knowledge.filterOptions}
          setFilters={knowledge.setFilters}
          onCreate={knowledge.createKnowledgeBase}
          onReset={knowledge.resetFilters}
          onSearch={knowledge.search}
        />
        <KnowledgeTable
          items={knowledge.items}
          loading={knowledge.loading}
          page={knowledge.page}
          pageSize={knowledge.pageSize}
          selected={knowledge.selected}
          total={knowledge.total}
          onChangePage={(nextPage, nextSize) => knowledge.load(knowledge.filters, nextPage, nextSize)}
          onChoose={knowledge.choose}
          onDelete={knowledge.deleteKnowledgeBase}
          onEdit={knowledge.editKnowledgeBase}
          onStatus={knowledge.setKnowledgeStatus}
          onViewDocuments={(kb) => knowledge.choose(kb, true)}
        />
      </Content>
      <Drawer
        title={knowledge.selected ? `知识库文档 · ${knowledge.selected.name}` : '知识库文档'}
        size={820}
        open={knowledge.documentDrawerOpen}
        onClose={() => knowledge.setDocumentDrawerOpen(false)}
      >
        <UploadPanel
          documents={knowledge.documents}
          loading={knowledge.documentsLoading}
          selected={knowledge.selected}
          uploading={knowledge.uploading}
          onDelete={knowledge.deleteDocument}
          onDownload={knowledge.downloadDocument}
          onUpload={knowledge.upload}
          onViewChunks={knowledge.viewChunks}
          onViewDocument={knowledge.viewDocument}
        />
      </Drawer>
      <Modal
        title={knowledge.documentViewer.title || '文档内容'}
        open={knowledge.documentViewer.open}
        onCancel={() => knowledge.setDocumentViewer({ open: false, title: '', content: '', loading: false })}
        footer={null}
        width={920}
        loading={knowledge.documentViewer.loading}
        destroyOnHidden
      >
        <pre className="document-preview-content">{knowledge.documentViewer.content}</pre>
      </Modal>
      <KnowledgeEditorModal
        editing={knowledge.editing}
        open={knowledge.editorOpen}
        onCancel={() => knowledge.setEditorOpen(false)}
        onSave={knowledge.saveKnowledgeBase}
      />
    </Layout>
  );
}

function KnowledgeEditorModal({ editing, open, onCancel, onSave }) {
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!open) return;
    form.setFieldsValue({
      name: editing?.name || '',
      description: editing?.description || '',
      scenario: editing?.scenario || '',
      department: editing?.department || '',
      tags: (editing?.tags || []).join('，'),
      retrievalTopK: editing?.retrievalTopK || 5,
      similarityFloor: editing?.similarityFloor ?? 0.2,
    });
  }, [editing, form, open]);

  async function submit() {
    const values = await form.validateFields();
    setSaving(true);
    try {
      await onSave(values);
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal
      title={editing ? '编辑知识库' : '新建知识库'}
      open={open}
      onCancel={onCancel}
      onOk={submit}
      confirmLoading={saving}
      okText="保存"
      cancelText="取消"
      destroyOnHidden
      width={720}
    >
      <Form form={form} layout="vertical" requiredMark={false}>
        <Form.Item name="name" label="知识库名称" rules={[{ required: true, message: '请输入知识库名称' }]}>
          <Input placeholder="例如：内分泌科诊疗规范" maxLength={80} showCount />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea placeholder="说明知识库覆盖的文档范围和使用场景" autoSize={{ minRows: 3, maxRows: 5 }} maxLength={300} showCount />
        </Form.Item>
        <div className="form-grid">
          <Form.Item name="scenario" label="场景">
            <Input placeholder="临床诊疗" />
          </Form.Item>
          <Form.Item name="department" label="所属部门">
            <Input placeholder="医务部" />
          </Form.Item>
          <Form.Item name="retrievalTopK" label="检索 TopK">
            <InputNumber min={1} max={50} precision={0} className="full-width" />
          </Form.Item>
          <Form.Item name="similarityFloor" label="相似度阈值">
            <InputNumber min={0} max={1} step={0.01} className="full-width" />
          </Form.Item>
        </div>
        <Form.Item name="tags" label="标签">
          <Input placeholder="多个标签用逗号或空格分隔" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
