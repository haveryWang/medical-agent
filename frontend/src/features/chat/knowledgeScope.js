export function normalizeKnowledgeBaseIDs(ids = []) {
  if (!Array.isArray(ids)) return ids ? [String(ids)] : [];
  return ids.map(String).filter(Boolean);
}

export function knowledgeBaseNames(ids = [], knowledgeBaseMap) {
  return normalizeKnowledgeBaseIDs(ids)
    .map((id) => knowledgeBaseMap.get(String(id))?.name)
    .filter(Boolean)
    .join('、');
}

export function scopeHint(active, selectedIds = [], knowledgeBaseMap) {
  const ids = normalizeKnowledgeBaseIDs(selectedIds);
  if (!ids.length) {
    return active ? '当前未选择知识库，发送时不走检索增强' : '未选择知识库，发送首条消息时不走检索增强';
  }
  const names = knowledgeBaseNames(ids, knowledgeBaseMap) || '当前知识库';
  return active ? `当前知识库：${names}` : `已选知识库：${names}，发送首条消息时生效`;
}

export function conversationScopeText(knowledgeBaseIds = [], knowledgeBaseMap) {
  const names = knowledgeBaseNames(knowledgeBaseIds, knowledgeBaseMap);
  if (!names) return '未选择知识库';
  return `知识库：${names}`;
}
