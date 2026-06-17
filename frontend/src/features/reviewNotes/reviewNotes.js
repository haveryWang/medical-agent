export const reviewNotePageSize = 10;

export function normalizeReviewNoteDraft(value) {
  const content = String(value || '').trim();
  return { content, valid: content.length > 0 };
}

export function normalizeSelectedReviewNoteIds(ids = []) {
  const seen = new Set();
  const result = [];
  for (const value of ids) {
    const id = String(value || '').trim();
    if (!id || seen.has(id)) continue;
    seen.add(id);
    result.push(id);
  }
  return result;
}

export function messageToReviewDraft(message) {
  const role = message?.role === 'assistant' ? '助手' : '用户';
  const content = String(message?.content || '').trim();
  return `【对话复盘】\n角色：${role}\n\n${content}`.trim();
}
