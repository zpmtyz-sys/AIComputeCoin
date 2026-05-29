type MessageCallback = (data: unknown) => void;

interface QueuedMessage {
  channel: string;
  data: unknown;
}

class WebSocketManager {
  private static instance: WebSocketManager | null = null;
  private socket: WebSocket | null = null;
  private url: string;
  private subscriptions: Map<string, Set<MessageCallback>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectDelay = 30000;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private messageQueue: QueuedMessage[] = [];
  private _isConnected = false;

  private constructor(url: string) {
    this.url = url;
  }

  static getInstance(url?: string): WebSocketManager {
    if (!WebSocketManager.instance) {
      WebSocketManager.instance = new WebSocketManager(
        url ?? process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/ws"
      );
    }
    return WebSocketManager.instance;
  }

  get isConnected(): boolean {
    return this._isConnected;
  }

  connect(): void {
    if (this.socket?.readyState === WebSocket.OPEN) return;

    try {
      this.socket = new WebSocket(this.url);

      this.socket.onopen = () => {
        this._isConnected = true;
        this.reconnectAttempts = 0;
        this.flushQueue();
        this.notifyConnectionChange();
      };

      this.socket.onmessage = (event: MessageEvent) => {
        try {
          const message = JSON.parse(event.data as string) as {
            channel: string;
            data: unknown;
          };
          this.routeMessage(message.channel, message.data);
        } catch {
          // Silently ignore malformed messages
        }
      };

      this.socket.onclose = () => {
        this._isConnected = false;
        this.notifyConnectionChange();
        this.scheduleReconnect();
      };

      this.socket.onerror = () => {
        this._isConnected = false;
        this.notifyConnectionChange();
      };
    } catch {
      this.scheduleReconnect();
    }
  }

  subscribe(channel: string, callback: MessageCallback): () => void {
    if (!this.subscriptions.has(channel)) {
      this.subscriptions.set(channel, new Set());
    }
    this.subscriptions.get(channel)!.add(callback);

    // Connect if not already connected
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.connect();
    }

    // Send subscribe message
    this.send({ type: "subscribe", channel });

    return () => {
      const callbacks = this.subscriptions.get(channel);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.subscriptions.delete(channel);
          this.send({ type: "unsubscribe", channel });
        }
      }
    };
  }

  private send(data: unknown): void {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(data));
    } else {
      this.messageQueue.push({ channel: "", data });
    }
  }

  private flushQueue(): void {
    while (this.messageQueue.length > 0) {
      const msg = this.messageQueue.shift();
      if (msg) {
        this.send(msg.data);
      }
    }
  }

  private routeMessage(channel: string, data: unknown): void {
    const callbacks = this.subscriptions.get(channel);
    if (callbacks) {
      callbacks.forEach((cb) => cb(data));
    }
  }

  private notifyConnectionChange(): void {
    const callbacks = this.subscriptions.get("__connection__");
    if (callbacks) {
      callbacks.forEach((cb) => cb(this._isConnected));
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;

    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempts),
      this.maxReconnectDelay
    );
    this.reconnectAttempts++;

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.socket?.close();
    this.socket = null;
    this._isConnected = false;
  }
}

export default WebSocketManager;
