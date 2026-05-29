export interface TradingPair {
  symbol: string;
  base: string;
  quote: string;
  status: string;
  minQuantity: string;
  maxQuantity: string;
  tickSize: string;
}

export interface PriceLevel {
  price: string;
  quantity: string;
  orderCount: number;
}

export interface OrderbookData {
  pair: string;
  bids: PriceLevel[];
  asks: PriceLevel[];
  sequenceId: number;
}

export interface TradeData {
  id: string;
  pair: string;
  price: string;
  quantity: string;
  side: string;
  tradedAt: number;
}

export interface KlineData {
  openTime: number;
  open: string;
  high: string;
  low: string;
  close: string;
  volume: string;
  closeTime: number;
}

export interface TickerData {
  pair: string;
  lastPrice: string;
  priceChange: string;
  priceChangePercent: string;
  high24h: string;
  low24h: string;
  volume24h: string;
  quoteVolume24h: string;
}

const TRADING_PAIRS: TradingPair[] = [
  {
    symbol: "BTC-USDT",
    base: "BTC",
    quote: "USDT",
    status: "active",
    minQuantity: "0.00001",
    maxQuantity: "100.0",
    tickSize: "0.01",
  },
  {
    symbol: "ETH-USDT",
    base: "ETH",
    quote: "USDT",
    status: "active",
    minQuantity: "0.001",
    maxQuantity: "1000.0",
    tickSize: "0.01",
  },
  {
    symbol: "CC-USDT",
    base: "CC",
    quote: "USDT",
    status: "active",
    minQuantity: "1.0",
    maxQuantity: "100000.0",
    tickSize: "0.0001",
  },
  {
    symbol: "CC-BTC",
    base: "CC",
    quote: "BTC",
    status: "active",
    minQuantity: "1.0",
    maxQuantity: "100000.0",
    tickSize: "0.00000001",
  },
];

export class MarketDataService {
  getPairs(): TradingPair[] {
    return TRADING_PAIRS;
  }

  getOrderbook(pair: string, depth: number): OrderbookData {
    const basePrice = this.getBasePrice(pair);
    const bids: PriceLevel[] = [];
    const asks: PriceLevel[] = [];

    for (let i = 0; i < depth; i++) {
      bids.push({
        price: (basePrice - (i + 1) * basePrice * 0.001).toFixed(2),
        quantity: (Math.random() * 10 + 0.1).toFixed(4),
        orderCount: Math.floor(Math.random() * 10) + 1,
      });
      asks.push({
        price: (basePrice + (i + 1) * basePrice * 0.001).toFixed(2),
        quantity: (Math.random() * 10 + 0.1).toFixed(4),
        orderCount: Math.floor(Math.random() * 10) + 1,
      });
    }

    return {
      pair,
      bids,
      asks,
      sequenceId: Date.now(),
    };
  }

  getTrades(pair: string, limit: number): TradeData[] {
    const basePrice = this.getBasePrice(pair);
    const trades: TradeData[] = [];
    const now = Date.now();

    for (let i = 0; i < limit; i++) {
      trades.push({
        id: `trade-${now}-${i}`,
        pair,
        price: (basePrice + (Math.random() - 0.5) * basePrice * 0.01).toFixed(
          2
        ),
        quantity: (Math.random() * 5 + 0.01).toFixed(4),
        side: Math.random() > 0.5 ? "buy" : "sell",
        tradedAt: now - i * 1000,
      });
    }

    return trades;
  }

  getKlines(
    pair: string,
    interval: string,
    limit: number,
    start?: number,
    end?: number
  ): KlineData[] {
    const basePrice = this.getBasePrice(pair);
    const klines: KlineData[] = [];
    const intervalMs = this.intervalToMs(interval);
    const endTime = end || Date.now();
    const startTime = start || endTime - limit * intervalMs;

    let currentTime = startTime;
    let currentPrice = basePrice;

    for (let i = 0; i < limit && currentTime < endTime; i++) {
      const change = (Math.random() - 0.5) * currentPrice * 0.02;
      const open = currentPrice;
      const close = currentPrice + change;
      const high = Math.max(open, close) + Math.random() * currentPrice * 0.005;
      const low = Math.min(open, close) - Math.random() * currentPrice * 0.005;

      klines.push({
        openTime: currentTime,
        open: open.toFixed(2),
        high: high.toFixed(2),
        low: low.toFixed(2),
        close: close.toFixed(2),
        volume: (Math.random() * 100 + 1).toFixed(4),
        closeTime: currentTime + intervalMs - 1,
      });

      currentPrice = close;
      currentTime += intervalMs;
    }

    return klines;
  }

  getTicker(pair: string): TickerData {
    const basePrice = this.getBasePrice(pair);
    const change = (Math.random() - 0.5) * basePrice * 0.05;

    return {
      pair,
      lastPrice: basePrice.toFixed(2),
      priceChange: change.toFixed(2),
      priceChangePercent: ((change / basePrice) * 100).toFixed(2),
      high24h: (basePrice + Math.abs(change) * 2).toFixed(2),
      low24h: (basePrice - Math.abs(change) * 2).toFixed(2),
      volume24h: (Math.random() * 10000 + 100).toFixed(4),
      quoteVolume24h: (Math.random() * 1000000 + 10000).toFixed(2),
    };
  }

  private getBasePrice(pair: string): number {
    switch (pair) {
      case "BTC-USDT":
        return 65000;
      case "ETH-USDT":
        return 3500;
      case "CC-USDT":
        return 2.5;
      case "CC-BTC":
        return 0.0000385;
      default:
        return 100;
    }
  }

  private intervalToMs(interval: string): number {
    switch (interval) {
      case "1m":
        return 60000;
      case "5m":
        return 300000;
      case "15m":
        return 900000;
      case "1h":
        return 3600000;
      case "4h":
        return 14400000;
      case "1d":
        return 86400000;
      default:
        return 3600000;
    }
  }
}

export const marketDataService = new MarketDataService();
