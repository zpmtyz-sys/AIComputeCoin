import { v4 as uuidv4 } from "uuid";

export interface StoredOrder {
  id: string;
  userId: string;
  pair: string;
  side: string;
  type: string;
  price: string;
  quantity: string;
  filledQuantity: string;
  status: string;
  stopPrice: string;
  createdAt: number;
  updatedAt: number;
}

export class OrderStore {
  private orders: Map<string, StoredOrder> = new Map();
  private userOrders: Map<string, string[]> = new Map();

  createOrder(params: {
    userId: string;
    pair: string;
    side: string;
    type: string;
    price?: string;
    quantity: string;
    stopPrice?: string;
  }): StoredOrder {
    const id = uuidv4();
    const now = Date.now();
    const order: StoredOrder = {
      id,
      userId: params.userId,
      pair: params.pair,
      side: params.side,
      type: params.type,
      price: params.price || "0",
      quantity: params.quantity,
      filledQuantity: "0",
      status: "new",
      stopPrice: params.stopPrice || "0",
      createdAt: now,
      updatedAt: now,
    };

    this.orders.set(id, order);
    const userOrderIds = this.userOrders.get(params.userId) || [];
    userOrderIds.push(id);
    this.userOrders.set(params.userId, userOrderIds);

    return order;
  }

  getOrder(orderId: string): StoredOrder | undefined {
    return this.orders.get(orderId);
  }

  cancelOrder(orderId: string, userId: string): StoredOrder | undefined {
    const order = this.orders.get(orderId);
    if (!order) return undefined;
    if (order.userId !== userId) return undefined;
    if (order.status !== "new" && order.status !== "partially_filled") {
      return undefined;
    }

    order.status = "cancelled";
    order.updatedAt = Date.now();
    return order;
  }

  listOrders(
    userId: string,
    params: {
      pair?: string;
      status?: string;
      limit: number;
      offset: number;
    }
  ): { orders: StoredOrder[]; total: number } {
    const userOrderIds = this.userOrders.get(userId) || [];
    let orders = userOrderIds
      .map((id) => this.orders.get(id))
      .filter((o): o is StoredOrder => o !== undefined);

    if (params.pair) {
      orders = orders.filter((o) => o.pair === params.pair);
    }
    if (params.status) {
      orders = orders.filter((o) => o.status === params.status);
    }

    // Sort by creation time descending
    orders.sort((a, b) => b.createdAt - a.createdAt);

    const total = orders.length;
    const paged = orders.slice(params.offset, params.offset + params.limit);

    return { orders: paged, total };
  }
}

export const orderStore = new OrderStore();
