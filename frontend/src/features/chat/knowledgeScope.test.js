import assert from 'node:assert/strict';
import test from 'node:test';
import {
  conversationScopeText,
  knowledgeBaseNames,
  normalizeKnowledgeBaseIDs,
  scopeHint,
} from './knowledgeScope.js';

test('knowledge scope helpers keep multiple knowledge bases visible', () => {
  const map = new Map([
    ['kb-1', { name: '心内科' }],
    ['kb-2', { name: '内分泌科' }],
  ]);

  assert.deepEqual(normalizeKnowledgeBaseIDs(['kb-1', 'kb-2']), ['kb-1', 'kb-2']);
  assert.equal(knowledgeBaseNames(['kb-1', 'kb-2'], map), '心内科、内分泌科');
  assert.equal(scopeHint(true, ['kb-1', 'kb-2'], map), '当前知识库：心内科、内分泌科');
  assert.equal(conversationScopeText(['kb-1', 'kb-2'], map), '知识库：心内科、内分泌科');
});
