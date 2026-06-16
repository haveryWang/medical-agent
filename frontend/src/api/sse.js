export async function consumeSSE(stream, onEvent) {
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split('\n\n');
    buffer = events.pop() || '';

    for (const raw of events) {
      const lines = raw.split('\n');
      const event = lines.find((line) => line.startsWith('event:'))?.slice(6).trim();
      const data = lines.find((line) => line.startsWith('data:'))?.slice(5).trim();
      if (event && data) onEvent(event, JSON.parse(data));
    }
  }
}
