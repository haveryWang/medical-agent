import { fileIcon, formatSize } from '../../utils/format.js';

export default function UploadPanel({ documents, inputRef, uploading, onUpload }) {
  return (
    <aside className="upload-panel">
      <h3>文档上传</h3>
      <label className="drop-zone">
        <input ref={inputRef} type="file" onChange={(e) => onUpload(e.target.files?.[0])} />
        <span>☁</span>
        <b>{uploading ? '上传中...' : '点击或拖拽文件到此处上传'}</b>
        <small>支持 PDF、Word、Excel、Markdown 等格式<br />单个文件不超过 50MB</small>
      </label>
      <h4>已上传文件</h4>
      <div className="uploaded-list">
        {documents.map((doc) => <div key={doc.id}><span>{fileIcon(doc.fileType)}</span><b>{doc.fileName}</b><small>{formatSize(doc.sizeBytes)}</small><em>{doc.status === 'completed' ? '上传成功' : doc.status}</em></div>)}
        {!documents.length && <p className="muted">选择左侧知识库并上传文档。</p>}
      </div>
    </aside>
  );
}
