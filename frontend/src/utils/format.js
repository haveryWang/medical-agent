export function fileIcon(type) {
  if (type?.includes('pdf')) return '📕';
  if (type?.includes('doc')) return '📘';
  if (type?.includes('xls')) return '📗';
  return '📄';
}

export function formatSize(bytes = 0) {
  if (bytes > 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)}MB`;
  if (bytes > 1024) return `${Math.round(bytes / 1024)}KB`;
  return `${bytes}B`;
}
