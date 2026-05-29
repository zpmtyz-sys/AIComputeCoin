import { create } from "zustand";

interface OrderBookLevel {
  price: number;
  quantity: number;
}

interface OrderBookState {
  asks: OrderBookLevel[];
  bids: OrderBookLevel[];
}

interface Position {
  id: string;
  pair: string;
  side: "long" | "short";
  size: number;
  entryPrice: number;
  markPrice: number;
  unrealizedPnl: number;
  margin: number;
}

interface OpenOrder {
  id: string;
  pair: string;
  side: "buy" | "sell";
  type: string;
  price: number;
  quantity: number;
  filled: number;
  time: string;
}

interface TradeRecord {
  id: string;
  time: string;
  pair: string;
  side: "buy" | "sell";
  price: number;
  quantity: number;
  fee: number;
  realizedPnl: number;
}

interface TradingState {
  currentPair: string;
  orderbook: OrderBookState;
  openOrders: OpenOrder[];
  positions: Position[];
  tradeHistory: TradeRecord[];
  setCurrentPair: (pair: string) => void;
  updateOrderbook: (orderbook: OrderBookState) => void;
  addOrder: (order: OpenOrder) => void;
  cancelOrder: (orderId: string) => void;
  updatePositions: (positions: Position[]) => void;
}

export const useTradingStore = create<TradingState>((set) => ({
  currentPair: "CU-PERP/USDT",
  orderbook: { asks: [], bids: [] },
  openOrders: [],
  positions: [],
  tradeHistory: [],

  setCurrentPair: (pair) => set({ currentPair: pair }),

  updateOrderbook: (orderbook) => set({ orderbook }),

  addOrder: (order) =>
    set((state) => ({ openOrders: [...state.openOrders, order] })),

  cancelOrder: (orderId) =>
    set((state) => ({
      openOrders: state.openOrders.filter((o) => o.id !== orderId),
    })),

  updatePositions: (positions) => set({ positions }),
}));
