import { useAuthStore } from '../store/authStore';

type MessagePayload = { type: string; payload?: any };

class Realtime {
  private ws: WebSocket | null = null;
  private reconnectMs = 1000;
  private maxReconnectMs = 30000;
  private listeners: Map<string, Set<(payload: any) => void>> = new Map();
  private wildcardListeners: Set<(msg: MessagePayload) => void> = new Set();

  private getUrl() {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host; // respects forwarded host/port
    const token = useAuthStore.getState().token;
    const params = token ? `?token=${encodeURIComponent(token)}` : '';
    return `${proto}//${host}/ws${params}`;
  }

  connect() {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) return;
    const url = this.getUrl();
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.reconnectMs = 1000;
    };

    this.ws.onmessage = (ev) => {
      try {
        const msg: MessagePayload = JSON.parse(ev.data);
        this.wildcardListeners.forEach((l) => l(msg));
        const set = this.listeners.get(msg.type);
        if (set) {
          set.forEach((l) => l(msg.payload));
        }
      } catch (err) {
        console.warn('Realtime: invalid message', err);
      }
    };

    this.ws.onclose = () => {
      this.ws = null;
      setTimeout(() => this.tryReconnect(), this.reconnectMs);
      this.reconnectMs = Math.min(this.maxReconnectMs, this.reconnectMs * 1.5);
    };

    this.ws.onerror = () => {
      // let onclose handle reconnection
      try {
        this.ws?.close();
      } catch {}
    };
  }

  private tryReconnect() {
    if (!this.ws) this.connect();
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
  }

  send(type: string, payload?: any) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return false;
    const msg = JSON.stringify({ type, payload });
    this.ws.send(msg);
    return true;
  }

  on(type: string, cb: (payload: any) => void) {
    if (type === '*') {
      this.wildcardListeners.add(cb as any);
      return () => this.wildcardListeners.delete(cb as any);
    }
    const set = this.listeners.get(type) || new Set();
    set.add(cb);
    this.listeners.set(type, set);
    return () => set.delete(cb);
  }

  off(type: string, cb: (payload: any) => void) {
    if (type === '*') {
      this.wildcardListeners.delete(cb as any);
      return;
    }
    const set = this.listeners.get(type);
    if (set) set.delete(cb);
  }
}

const realtime = new Realtime();
export default realtime;
