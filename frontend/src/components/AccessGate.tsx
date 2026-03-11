import { useState, useEffect, useRef, useCallback, type ReactNode } from 'react';
import { api } from '@/services/api';
import { Loader2 } from 'lucide-react';

const KEEPALIVE_INTERVAL = 2 * 60 * 1000; // 2 minutes

export function AccessGate({ children }: { children: ReactNode }) {
  const [state, setState] = useState<'loading' | 'open' | 'setup' | 'verify'>('loading');
  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const activeRef = useRef(false);
  const authEnabledRef = useRef(false);

  // Check auth status on mount
  useEffect(() => {
    api.authStatus()
      .then((res) => {
        authEnabledRef.current = true;
        setState(res.setup ? 'verify' : 'setup');
      })
      .catch(() => setState('open')); // auth not enabled
  }, []);

  // Track user activity (mouse, keyboard, scroll, touch)
  useEffect(() => {
    if (!authEnabledRef.current) return;

    const markActive = () => { activeRef.current = true; };
    const events = ['mousemove', 'mousedown', 'keydown', 'scroll', 'touchstart'] as const;
    events.forEach((e) => window.addEventListener(e, markActive, { passive: true }));
    return () => {
      events.forEach((e) => window.removeEventListener(e, markActive));
    };
  }, [state]);

  // Periodic keepalive: send ping if user was active since last check
  useEffect(() => {
    if (state !== 'open' || !authEnabledRef.current) return;

    const interval = setInterval(() => {
      if (activeRef.current) {
        activeRef.current = false;
        api.authKeepalive().catch(() => {
          // Session expired on backend
          setState('verify');
        });
      }
    }, KEEPALIVE_INTERVAL);

    return () => clearInterval(interval);
  }, [state]);

  // Listen for 401 on any fetch to detect session expiry
  const onSessionExpired = useCallback(() => {
    if (authEnabledRef.current && state === 'open') {
      setState('verify');
    }
  }, [state]);

  useEffect(() => {
    const originalFetch = window.fetch;
    window.fetch = async (...args) => {
      const response = await originalFetch(...args);
      if (response.status === 401) {
        const url = typeof args[0] === 'string' ? args[0] : (args[0] as Request).url;
        // Only trigger for API calls, not auth endpoints
        if (url.includes('/api/') && !url.includes('/api/auth/')) {
          onSessionExpired();
        }
      }
      return response;
    };
    return () => { window.fetch = originalFetch; };
  }, [onSessionExpired]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!code.trim() || submitting) return;
    setSubmitting(true);
    setError('');

    try {
      if (state === 'setup') {
        const res = await api.authSetup(code);
        if (res.error) { setError(res.error); return; }
      } else {
        const res = await api.authVerify(code);
        if (res.error) { setError(res.error); return; }
      }
      setCode('');
      setState('open');
    } catch (err) {
      setError(state === 'verify' ? 'Invalid access code' : 'Failed to set access code');
    } finally {
      setSubmitting(false);
    }
  };

  if (state === 'loading') {
    return (
      <div className="h-screen w-screen flex items-center justify-center bg-dark-bg">
        <Loader2 size={24} className="animate-spin text-accent-primary" />
      </div>
    );
  }

  if (state === 'open') return <>{children}</>;

  return (
    <div className="h-screen w-screen flex items-center justify-center bg-dark-bg">
      <div className="w-80">
        <img src="/qodex_logo.png" alt="Qodex" className="w-24 h-24 mx-auto mb-4 object-contain opacity-80" />
        <img src="/qodex_txt.png" alt="Qodex" className="h-6 mx-auto mb-6" />

        <p className="text-sm text-dark-text-secondary text-center mb-4">
          {state === 'setup'
            ? 'Set an access code to protect this instance.'
            : 'Enter the access code to continue.'}
        </p>

        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="password"
            value={code}
            onChange={(e) => { setCode(e.target.value); setError(''); }}
            placeholder={state === 'setup' ? 'Create access code' : 'Access code'}
            autoFocus
            className="w-full bg-dark-bg-tertiary text-dark-text px-3 py-2 rounded border border-dark-border focus:border-accent-primary focus:outline-none text-sm text-center"
          />

          {error && (
            <p className="text-xs text-accent-error text-center">{error}</p>
          )}

          <button
            type="submit"
            disabled={!code.trim() || submitting}
            className="px-4 py-2 bg-accent-primary text-white rounded text-sm font-medium hover:bg-accent-primary/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
          >
            {submitting && <Loader2 size={14} className="animate-spin" />}
            {state === 'setup' ? 'Set Access Code' : 'Enter'}
          </button>
        </form>
      </div>
    </div>
  );
}
