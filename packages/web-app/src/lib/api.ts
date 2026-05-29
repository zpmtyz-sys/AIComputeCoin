// API Client with configurable base URL and mock data fallback

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "/api/v1";

// Mock mode: enabled explicitly via env var, or implicitly in development
const MOCK_MODE =
  process.env.NEXT_PUBLIC_MOCK_MODE === "true" ||
  process.env.NODE_ENV === "development";

// Types
export interface Kline {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface OrderBookLevel {
  price: number;
  quantity: number;
}

export interface OrderBookData {
  asks: OrderBookLevel[];
  bids: OrderBookLevel[];
  lastUpdateId: number;
}

export interface Ticker {
  pair: string;
  lastPrice: number;
  change24h: number;
  changePct24h: number;
  high24h: number;
  low24h: number;
  volume24h: number;
}

export interface Market {
  pair: string;
  baseAsset: string;
  quoteAsset: string;
  lastPrice: number;
  change24h: number;
  changePct24h: number;
  volume24h: number;
  high24h: number;
  low24h: number;
}

export interface OrderSubmission {
  pair: string;
  side: "buy" | "sell";
  type: "limit" | "market" | "stop-limit";
  price?: number;
  stopPrice?: number;
  quantity: number;
}

export interface OrderResponse {
  id: string;
  status: "accepted" | "rejected";
  message: string;
}

export interface Position {
  id: string;
  pair: string;
  side: "long" | "short";
  size: number;
  entryPrice: number;
  markPrice: number;
  unrealizedPnl: number;
  margin: number;
}

export interface Order {
  id: string;
  pair: string;
  side: "buy" | "sell";
  type: string;
  price: number;
  quantity: number;
  filled: number;
  time: string;
  status: string;
}

export interface TradeRecord {
  id: string;
  time: string;
  pair: string;
  side: "buy" | "sell";
  price: number;
  quantity: number;
  fee: number;
  realizedPnl: number;
}

// Mock Data Generators
function mockKlines(count: number): Kline[] {
  const klines: Kline[] = [];
  let time = Math.floor(Date.now() / 1000) - count * 3600;
  let close = 45000;

  for (let i = 0; i < count; i++) {
    const open = close;
    close = open + (Math.random() - 0.48) * 500;
    const high = Math.max(open, close) + Math.random() * 200;
    const low = Math.min(open, close) - Math.random() * 200;
    klines.push({
      time,
      open: Math.round(open * 100) / 100,
      high: Math.round(high * 100) / 100,
      low: Math.round(low * 100) / 100,
      close: Math.round(close * 100) / 100,
      volume: Math.round(Math.random() * 1000 + 200),
    });
    time += 3600;
  }
  return klines;
}

function mockOrderbook(): OrderBookData {
  const asks: OrderBookLevel[] = [];
  const bids: OrderBookLevel[] = [];
  let askPrice = 45050;
  let bidPrice = 45000;

  for (let i = 0; i < 20; i++) {
    asks.push({ price: askPrice, quantity: Math.round((Math.random() * 5 + 0.1) * 1000) / 1000 });
    bids.push({ price: bidPrice, quantity: Math.round((Math.random() * 5 + 0.1) * 1000) / 1000 });
    askPrice += Math.round(Math.random() * 20 + 5);
    bidPrice -= Math.round(Math.random() * 20 + 5);
  }

  return { asks, bids, lastUpdateId: Date.now() };
}

function mockMarkets(): Market[] {
  return [
    { pair: "CU-PERP/USDT", baseAsset: "CU-PERP", quoteAsset: "USDT", lastPrice: 45023.50, change24h: 523.50, changePct24h: 1.18, volume24h: 125430000, high24h: 45500, low24h: 44200 },
    { pair: "GPU-SPOT/USDT", baseAsset: "GPU-SPOT", quoteAsset: "USDT", lastPrice: 12450.00, change24h: -180.00, changePct24h: -1.43, volume24h: 45200000, high24h: 12800, low24h: 12300 },
    { pair: "TPU-PERP/USDT", baseAsset: "TPU-PERP", quoteAsset: "USDT", lastPrice: 9150.00, change24h: 92.00, changePct24h: 1.02, volume24h: 23100000, high24h: 9300, low24h: 8950 },
    { pair: "FPGA-SPOT/USDT", baseAsset: "FPGA-SPOT", quoteAsset: "USDT", lastPrice: 3420.00, change24h: -45.00, changePct24h: -1.30, volume24h: 8700000, high24h: 3500, low24h: 3380 },
  ];
}

// API Functions
async function apiFetch<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${endpoint}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json() as Promise<T>;
}

export async function getKlines(pair: string, interval: string): Promise<Kline[]> {
  try {
    return await apiFetch<Kline[]>(`/klines?pair=${pair}&interval=${interval}`);
  } catch (error) {
    if (MOCK_MODE) return mockKlines(200);
    throw error;
  }
}

export async function getOrderbook(pair: string): Promise<OrderBookData> {
  try {
    return await apiFetch<OrderBookData>(`/orderbook?pair=${pair}`);
  } catch (error) {
    if (MOCK_MODE) return mockOrderbook();
    throw error;
  }
}

export async function getTicker(pair: string): Promise<Ticker> {
  try {
    return await apiFetch<Ticker>(`/ticker?pair=${pair}`);
  } catch (error) {
    if (MOCK_MODE) {
      return {
        pair,
        lastPrice: 45023.5,
        change24h: 523.5,
        changePct24h: 1.18,
        high24h: 45500,
        low24h: 44200,
        volume24h: 125430000,
      };
    }
    throw error;
  }
}

export async function getMarkets(): Promise<Market[]> {
  try {
    return await apiFetch<Market[]>("/markets");
  } catch (error) {
    if (MOCK_MODE) return mockMarkets();
    throw error;
  }
}

export async function submitOrder(order: OrderSubmission): Promise<OrderResponse> {
  try {
    return await apiFetch<OrderResponse>("/orders", {
      method: "POST",
      body: JSON.stringify(order),
    });
  } catch (error) {
    if (MOCK_MODE) {
      return { id: `ord-${Date.now()}`, status: "accepted", message: "Order placed" };
    }
    throw error;
  }
}

export async function cancelOrder(id: string): Promise<{ success: boolean }> {
  try {
    return await apiFetch<{ success: boolean }>(`/orders/${id}`, {
      method: "DELETE",
    });
  } catch (error) {
    if (MOCK_MODE) return { success: true };
    throw error;
  }
}

export async function getPositions(): Promise<Position[]> {
  try {
    return await apiFetch<Position[]>("/positions");
  } catch (error) {
    if (MOCK_MODE) {
      return [
        { id: "1", pair: "CU-PERP/USDT", side: "long", size: 2.5, entryPrice: 44800, markPrice: 45050, unrealizedPnl: 625, margin: 11200 },
        { id: "2", pair: "GPU-SPOT/USDT", side: "short", size: 1.2, entryPrice: 12500, markPrice: 12650, unrealizedPnl: -180, margin: 3000 },
      ];
    }
    throw error;
  }
}

export async function getOrders(): Promise<Order[]> {
  try {
    return await apiFetch<Order[]>("/orders");
  } catch (error) {
    if (MOCK_MODE) {
      return [
        { id: "ord-001", pair: "CU-PERP/USDT", side: "buy", type: "limit", price: 44500, quantity: 1.5, filled: 0, time: "2024-01-15 14:30:22", status: "open" },
        { id: "ord-002", pair: "GPU-SPOT/USDT", side: "sell", type: "limit", price: 12800, quantity: 3.0, filled: 50, time: "2024-01-15 13:15:08", status: "partial" },
      ];
    }
    throw error;
  }
}

export async function getTradeHistory(): Promise<TradeRecord[]> {
  try {
    return await apiFetch<TradeRecord[]>("/trades/history");
  } catch (error) {
    if (MOCK_MODE) {
      return [
        { id: "t1", time: "2024-01-15 14:22:10", pair: "CU-PERP/USDT", side: "sell", price: 45100, quantity: 1.0, fee: 6.77, realizedPnl: 320.5 },
        { id: "t2", time: "2024-01-15 12:10:45", pair: "CU-PERP/USDT", side: "buy", price: 44780, quantity: 1.0, fee: 4.48, realizedPnl: 0 },
      ];
    }
    throw error;
  }
}
