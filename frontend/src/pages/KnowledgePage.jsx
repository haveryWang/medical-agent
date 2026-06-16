import KnowledgeFilters from '../features/knowledge/KnowledgeFilters.jsx';
import KnowledgeSideNav from '../features/knowledge/KnowledgeSideNav.jsx';
import KnowledgeTable from '../features/knowledge/KnowledgeTable.jsx';
import UploadPanel from '../features/knowledge/UploadPanel.jsx';
import { useKnowledgeWorkspace } from '../features/knowledge/useKnowledgeWorkspace.js';

export default function KnowledgePage() {
  const knowledge = useKnowledgeWorkspace();

  return (
    <section className="knowledge-layout">
      <KnowledgeSideNav />
      <section className="kb-main">
        <div className="breadcrumb">知识库管理 / <b>知识库列表</b></div>
        <KnowledgeFilters
          filters={knowledge.filters}
          setFilters={knowledge.setFilters}
          onSearch={() => knowledge.load(knowledge.filters)}
          onUpload={knowledge.openUploader}
        />
        <KnowledgeTable items={knowledge.items} selected={knowledge.selected} onChoose={knowledge.choose} />
        <div className="pager">共 {knowledge.items.length} 条　&lt;　1　&gt;　10条/页　跳至 1 页</div>
      </section>
      <UploadPanel
        documents={knowledge.documents}
        inputRef={knowledge.uploadInputRef}
        uploading={knowledge.uploading}
        onUpload={knowledge.upload}
      />
    </section>
  );
}
