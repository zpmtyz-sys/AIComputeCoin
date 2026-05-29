import { marketDataService } from "./market-data.js";

export interface GrpcOrderbookResponse {
  pair: string;
  bids: { price: string; quantity: string; orderCount: number }[];
  asks: { price: string; quantity: string; orderCount: number }[];
  sequenceId: number;
}

export interface GrpcTradeResponse {
  id: string;
  pair: string;
  price: string;
  quantity: string;
  side: string;
  tradedAt: number;
}

export interface GrpcSubmitOrderResponse {
  order: {
    id: string;
    userId: string;
    pair: string;
    side: string;
    orderType: string;
    price: string;
    quantity: string;
    filledQuantity: string;
    status: string;
    createdAt: number;
    updatedAt: number;
  };
  trades: GrpcTradeResponse[];
}

export interface GrpcCancelOrderResponse {
  order: {
    id: string;
    status: string;
    updatedAt: number;
  };
}

/**
 * gRPC client for matching engine.
 * Falls back to mock data when connection is unavailable.
 */
export class GrpcClient {
  private connected: boolean = false;

  constructor(private address: string) {
    // In a real implementation, we would connect to the matching engine
    // using @grpc/grpc-js and @grpc/proto-loader.
    // For now, we gracefully fall back to mock data.
    this.tryConnect();
  }

  private tryConnect(): void {
    // Attempt connection would happen here in production.
    // Since the matching engine may not be running, we set connected = false.
    this.connected = false;
  }

  isConnected(): boolean {
    return this.connected;
  }

  async submitOrder(params: {
    id: string;
    userId: string;
    pair: string;
    side: string;
    orderType: string;
    price: string;
    quantity: string;
  }): Promise<GrpcSubmitOrderResponse> {
    // Fall back to mock response
    const now = Date.now();
    return {
      order: {
        id: params.id,
        userId: params.userId,
        pair: params.pair,
        side: params.side,
        orderType: params.orderType,
        price: params.price,
        quantity: params.quantity,
        filledQuantity: "0",
        status: "new",
        createdAt: now,
        updatedAt: now,
      },
      trades: [],
    };
  }

  async cancelOrder(
    orderId: string,
    _userId: string
  ): Promise<GrpcCancelOrderResponse> {
    return {
      order: {
        id: orderId,
        status: "cancelled",
        updatedAt: Date.now(),
      },
    };
  }

  async getOrderBook(
    pair: string,
    depth: number
  ): Promise<GrpcOrderbookResponse> {
    const data = marketDataService.getOrderbook(pair, depth);
    return data;
  }

  async getTrades(
    pair: string,
    limit: number
  ): Promise<GrpcTradeResponse[]> {
    const data = marketDataService.getTrades(pair, limit);
    return data;
  }

  getAddress(): string {
    return this.address;
  }
}
