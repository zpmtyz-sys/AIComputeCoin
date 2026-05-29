import { FastifyInstance, FastifyRequest, FastifyReply } from "fastify";
import { orderStore } from "../services/order-store.js";
import { userStore } from "../services/user-store.js";
import {
  OrderBodySchema,
  OrderQuerySchema,
} from "../lib/schemas.js";
import { ValidationError, NotFoundError, AuthError } from "../lib/errors.js";
import { marketDataService } from "../services/market-data.js";

export async function tradingRoutes(app: FastifyInstance): Promise<void> {
  // All trading routes require authentication
  app.addHook("preHandler", app.authenticate);

  // POST /orders
  app.post("/orders", async (request: FastifyRequest, reply: FastifyReply) => {
    const parsed = OrderBodySchema.safeParse(request.body);
    if (!parsed.success) {
      throw new ValidationError(parsed.error.errors[0].message);
    }

    const { pair, side, type, price, quantity, stopPrice } = parsed.data;

    // Validate pair exists
    const validPairs = marketDataService.getPairs().map((p) => p.symbol);
    if (!validPairs.includes(pair)) {
      throw new ValidationError(`Invalid trading pair: ${pair}`);
    }

    // Validate limit orders have price
    if (
      (type === "limit" || type === "stop_limit") &&
      (!price || price === "0")
    ) {
      throw new ValidationError("Limit orders require a price");
    }

    const order = orderStore.createOrder({
      userId: request.user.userId,
      pair,
      side,
      type,
      price,
      quantity,
      stopPrice,
    });

    return reply.status(201).send({ order });
  });

  // DELETE /orders/:id
  app.delete("/orders/:id", async (request: FastifyRequest, reply: FastifyReply) => {
    const { id } = request.params as { id: string };
    const order = orderStore.cancelOrder(id, request.user.userId);
    if (!order) {
      throw new NotFoundError("Order not found or cannot be cancelled");
    }

    return reply.send({ order });
  });

  // GET /orders
  app.get("/orders", async (request: FastifyRequest, reply: FastifyReply) => {
    const parsed = OrderQuerySchema.safeParse(request.query);
    if (!parsed.success) {
      throw new ValidationError(parsed.error.errors[0].message);
    }

    const { pair, status, limit, offset } = parsed.data;
    const result = orderStore.listOrders(request.user.userId, {
      pair,
      status,
      limit,
      offset,
    });

    return reply.send({
      orders: result.orders,
      total: result.total,
      limit,
      offset,
    });
  });

  // GET /orders/:id
  app.get("/orders/:id", async (request: FastifyRequest, reply: FastifyReply) => {
    const { id } = request.params as { id: string };
    const order = orderStore.getOrder(id);
    if (!order) {
      throw new NotFoundError("Order not found");
    }

    if (order.userId !== request.user.userId) {
      throw new AuthError("Unauthorized");
    }

    return reply.send({ order });
  });

  // GET /positions
  app.get("/positions", async (_request: FastifyRequest, reply: FastifyReply) => {
    // In-memory mock: return empty positions
    return reply.send({ positions: [] });
  });

  // GET /balance
  app.get("/balance", async (request: FastifyRequest, reply: FastifyReply) => {
    const balance = userStore.getBalance(request.user.userId);
    return reply.send({ balance });
  });
}
