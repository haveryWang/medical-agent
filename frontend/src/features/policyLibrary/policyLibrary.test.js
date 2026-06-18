import assert from 'node:assert/strict';
import test from 'node:test';
import {
  POLICY_CATEGORIES,
  POLICY_FILTER_DEBOUNCE_MS,
  POLICY_TEXT_ELLIPSIS,
  POLICY_TEMPLATE_FILENAME,
  buildPolicyListParams,
  filterPoliciesByCategory,
  normalizePolicyFacets,
  preparePolicyImportFile,
} from './policyLibrary.js';

test('POLICY_CATEGORIES contains the eight fixed categories in display order', () => {
  assert.deepEqual(POLICY_CATEGORIES, ['国家医学中心', '科技创新', '医疗服务', '医保医药', '数智治理', '改革监管', '国际合作', '其他']);
});

test('POLICY_TEXT_ELLIPSIS keeps long policy slices expandable', () => {
  assert.equal(POLICY_TEXT_ELLIPSIS.expandable, true);
  assert.equal(POLICY_TEXT_ELLIPSIS.symbol, '展开全部');
});

test('POLICY_FILTER_DEBOUNCE_MS delays keyword filter requests', () => {
  assert.equal(POLICY_FILTER_DEBOUNCE_MS, 400);
});

test('filterPoliciesByCategory returns only selected category records', () => {
  const policies = [
    { title: 'A', category: '科技创新' },
    { title: 'B', category: '医保医药' },
  ];
  assert.deepEqual(filterPoliciesByCategory(policies, '医保医药'), [{ title: 'B', category: '医保医药' }]);
  assert.deepEqual(filterPoliciesByCategory(policies, ''), policies);
});

test('preparePolicyImportFile only accepts xlsx files', () => {
  assert.deepEqual(preparePolicyImportFile({ name: '政策库.xlsx', size: 1024 }), { file: { name: '政策库.xlsx', size: 1024 }, error: null });
  assert.match(preparePolicyImportFile({ name: '政策库.csv', size: 1024 }).error, /xlsx/);
  assert.match(preparePolicyImportFile({ name: '政策库.xlsx', size: 20 * 1024 * 1024 }).error, /15MB/);
});

test('policy import template filename makes the expected Excel contract visible', () => {
  assert.equal(POLICY_TEMPLATE_FILENAME, '政策文件库导入模板.xlsx');
});

test('buildPolicyListParams includes category and month filters when selected', () => {
  assert.deepEqual(buildPolicyListParams({ category: '医保医药', date: '2026-06', keyword: '医学中心', page: 2, pageSize: 10 }), {
    category: '医保医药',
    date: '2026-06',
    keyword: '医学中心',
    page: 2,
    size: 10,
  });
  assert.deepEqual(buildPolicyListParams({ category: '', date: '' }), {});
  assert.deepEqual(buildPolicyListParams({ keyword: '  医学中心  ' }), { keyword: '医学中心' });
});

test('normalizePolicyFacets keeps fixed categories and date counts usable for filters', () => {
  const facets = normalizePolicyFacets({
    categories: [{ value: '医保医药', count: 3 }],
    dates: [{ value: '2026-06', count: 2 }],
  });
  assert.equal(facets.categories.find((item) => item.value === '医保医药').count, 3);
  assert.equal(facets.categories.find((item) => item.value === '科技创新').count, 0);
  assert.deepEqual(facets.dates, [{ value: '2026-06', label: '2026-06', count: 2 }]);
});
