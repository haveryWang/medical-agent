import assert from 'node:assert/strict';
import test from 'node:test';
import {
  reviewNotePageSize,
  normalizeSelectedReviewNoteIds,
  normalizeReviewNoteDraft,
  messageToReviewDraft,
} from './reviewNotes.js';

test('normalizeReviewNoteDraft trims content and rejects empty drafts', () => {
  assert.deepEqual(normalizeReviewNoteDraft('  复盘经验  '), { content: '复盘经验', valid: true });
  assert.deepEqual(normalizeReviewNoteDraft(' \n\t '), { content: '', valid: false });
});

test('review note list uses server pagination defaults', () => {
  assert.equal(reviewNotePageSize, 10);
});

test('normalizeSelectedReviewNoteIds keeps valid unique ids in order', () => {
  assert.deepEqual(normalizeSelectedReviewNoteIds(['a', '', 'b', 'a', null]), ['a', 'b']);
});

test('messageToReviewDraft creates editable source text from chat messages', () => {
  assert.equal(
    messageToReviewDraft({ role: 'assistant', content: '  建议先核对政策发布日期。 ' }),
    '【对话复盘】\n角色：助手\n\n建议先核对政策发布日期。',
  );
});
