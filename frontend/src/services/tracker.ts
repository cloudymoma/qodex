interface UIEvent {
  action: string;
  target?: string;
  value?: string;
}

const buffer: UIEvent[] = [];
let flushTimer: ReturnType<typeof setTimeout> | null = null;
const FLUSH_INTERVAL = 3000; // 3 seconds

function scheduleFlush() {
  if (flushTimer) return;
  flushTimer = setTimeout(flush, FLUSH_INTERVAL);
}

function flush() {
  flushTimer = null;
  if (buffer.length === 0) return;

  const events = buffer.splice(0);
  fetch('/api/events', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ events }),
  }).catch(() => {
    // silently ignore — logging should not break the app
  });
}

/** Track a UI event. Events are batched and sent every 3 seconds. */
export function track(action: string, target?: string, value?: string) {
  buffer.push({ action, target, value });
  // Also log to browser console for debugging
  console.log(`[track] ${action}`, target ?? '', value ?? '');
  scheduleFlush();
}

// Flush on page unload
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', flush);
}
