import assert from 'node:assert/strict';
import test from 'node:test';
import { MAX_KB_FILE_BYTES, prepareUploadBatch } from './uploadBatch.js';

function file(name, size = 1024) {
  return { name, size };
}

test('prepareUploadBatch keeps supported files and reports invalid files', () => {
  const batch = prepareUploadBatch([
    file('指南.pdf'),
    file('病历.doc'),
    file('大文件.txt', MAX_KB_FILE_BYTES + 1),
    file('记录.md'),
  ]);

  assert.deepEqual(batch.validFiles.map((item) => item.name), ['指南.pdf', '记录.md']);
  assert.equal(batch.errors.length, 2);
  assert.match(batch.errors[0].message, /docx/);
  assert.match(batch.errors[1].message, /15MB/);
});
