const MODEL_REQUIREMENTS = {
  chat: [
    ['deepSeekBaseUrl', '对话 Base URL'],
    ['deepSeekAPIKeySet', '对话 API Key'],
    ['deepSeekChatModel', '对话模型'],
  ],
  embedding: [
    ['qwenEmbeddingBaseUrl', '向量 Base URL'],
    ['qwenEmbeddingAPIKeySet', '向量 API Key'],
    ['qwenEmbeddingModel', '向量模型'],
    ['qwenEmbeddingDimension', '向量维度'],
  ],
};

export async function requireModelConfig(api, scopes) {
  const cfg = await api.getModelConfig();
  const missing = [];
  for (const scope of scopes) {
    for (const [key, label] of MODEL_REQUIREMENTS[scope] || []) {
      if (!hasConfigValue(cfg[key])) missing.push(label);
    }
  }
  if (missing.length) {
    throw new Error(`模型配置不完整，请先在系统设置中配置：${[...new Set(missing)].join('、')}`);
  }
  return cfg;
}

function hasConfigValue(value) {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value > 0;
  return String(value || '').trim() !== '';
}
