export const MAX_KB_FILE_BYTES = 15 * 1024 * 1024;
export const TOO_LARGE_MESSAGE = '单个知识库文件不能超过 15MB，请切割单文件尺寸后再上传';
export const SUPPORTED_EXTENSIONS = ['.pdf', '.docx', '.xlsx', '.xls', '.md', '.markdown', '.txt', '.csv'];
export const ACCEPTED_UPLOAD_TYPES = SUPPORTED_EXTENSIONS.join(',');

export function prepareUploadBatch(files = []) {
  return Array.from(files || []).reduce(
    (batch, file) => {
      const result = validateUploadFile(file);
      if (result.ok) {
        batch.validFiles.push(file);
      } else {
        batch.errors.push({ file, message: result.message });
      }
      return batch;
    },
    { validFiles: [], errors: [] },
  );
}

export function validateUploadFile(file) {
  if (!file) return { ok: false, message: '请选择需要上传的文件' };
  if (Number(file.size || 0) > MAX_KB_FILE_BYTES) {
    return { ok: false, message: TOO_LARGE_MESSAGE };
  }
  const ext = getFileExtension(file.name);
  if (ext === '.doc') {
    return { ok: false, message: '当前支持 Word .docx 文件，.doc 请先转换为 .docx 后上传' };
  }
  if (!SUPPORTED_EXTENSIONS.includes(ext)) {
    return { ok: false, message: '当前支持 PDF、Word(.docx)、Excel(.xlsx/.xls)、Markdown、TXT、CSV' };
  }
  return { ok: true };
}

export function getFileExtension(name = '') {
  const index = name.lastIndexOf('.');
  return index >= 0 ? name.slice(index).toLowerCase() : '';
}
