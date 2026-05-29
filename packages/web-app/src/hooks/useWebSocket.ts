"use client";

import { useState, useEffect, useRef } from "react";
import WebSocketManager from "@/lib/websocket";

interface UseWebSocketResult<T = unknown> {
  data: T | null;
  isConnected: boolean;
  error: Error | null;
}

export function useWebSocket<T = unknown>(channel: string): UseWebSocketResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const managerRef = useRef<WebSocketManager | null>(null);

  useEffect(() => {
    try {
      const manager = WebSocketManager.getInstance();
      if (!manager) {
        // Server-side or unsupported environment; no-op
        return;
      }
      managerRef.current = manager;
      setIsConnected(manager.isConnected);

      // Subscribe to connection status changes
      const unsubConnection = manager.subscribe("__connection__", (connected) => {
        setIsConnected(connected as boolean);
      });

      // Subscribe to channel data
      const unsubChannel = manager.subscribe(channel, (message) => {
        setData(message as T);
      });

      return () => {
        unsubConnection();
        unsubChannel();
      };
    } catch (err) {
      setError(err instanceof Error ? err : new Error("WebSocket error"));
    }
  }, [channel]);

  return { data, isConnected, error };
}
