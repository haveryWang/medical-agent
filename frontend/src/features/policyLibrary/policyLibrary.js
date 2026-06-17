export const POLICY_CATEGORIES = ['国家医学中心', '科技创新', '医疗服务', '医保医药', '数智治理', '改革监管', '其他'];
export const MAX_POLICY_FILE_BYTES = 15 * 1024 * 1024;
export const POLICY_TEMPLATE_FILENAME = '政策文件库导入模板.xlsx';

export function filterPoliciesByCategory(policies, category) {
  if (!category) return policies;
  return policies.filter((item) => item.category === category);
}

export function preparePolicyImportFile(file) {
  if (!file) return { file: null, error: '请选择政策库 Excel 文件' };
  const name = file.name || '';
  if (!name.toLowerCase().endsWith('.xlsx')) {
    return { file: null, error: '政策文件库仅支持 .xlsx 文件' };
  }
  if (file.size > MAX_POLICY_FILE_BYTES) {
    return { file: null, error: '单个政策库文件不能超过 15MB' };
  }
  return { file, error: null };
}

export function buildPolicyListParams(filters = {}) {
  const params = {};
  if (filters.category) params.category = filters.category;
  if (filters.date) params.date = filters.date;
  if (filters.page) params.page = filters.page;
  if (filters.pageSize) params.size = filters.pageSize;
  return params;
}

export function normalizePolicyFacets(facets = {}) {
  const categoryCounts = new Map((facets.categories || []).map((item) => [item.value, item.count || 0]));
  return {
    categories: POLICY_CATEGORIES.map((category) => ({
      value: category,
      label: category,
      count: categoryCounts.get(category) || 0,
    })),
    dates: (facets.dates || []).map((item) => ({
      value: item.value,
      label: item.value,
      count: item.count || 0,
    })),
  };
}
