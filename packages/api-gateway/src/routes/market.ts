import { FastifyInstance, FastifyRequest, FastifyReply } from "fastify";
import { marketDataService } from "../services/market-data.js";
import {
  OrderbookQuerySchema,
  TradesQuerySchema,
  KlinesQuerySchema,
  OrderbookQuery,
  TradesQuery,
  KlinesQuery,
} from "../lib/schemas.js";
import { ValidationError, NotFoundError } from "../lib/errors.js";

export async function marketRoutes(app: FastifyInstance): Promise<void> {
  // GET /markets/pairs
  app.get(
    "/markets/pairs",
    async (_request: FastifyRequest, reply: FastifyReply) => {
      const pairs = marketDataService.getPairs();
      return reply.send({ pairs });
    }
  );

  // GET /markets/:pair/orderbook
  app.get(
    "/markets/:pair/orderbook",
    async (
      request: FastifyRequest<{
        Params: { pair: string };
        Querystring: OrderbookQuery;
      }>,
      reply: FastifyReply
    ) => {
      const { pair } = request.params as { pair: string };
      const parsed = OrderbookQuerySchema.safeParse(request.query);
      if (!parsed.success) {
        throw new ValidationError(parsed.error.errors[0].message);
      }

      const validPairs = marketDataService
        .getPairs()
        .map((p) => p.symbol);
      if (!validPairs.includes(pair)) {
        throw new NotFoundError(`Trading pair ${pair} not found`);
      }

      const orderbook = marketDataService.getOrderbook(pair, parsed.data.depth);
      return reply.send(orderbook);
    }
  );

  // GET /markets/:pair/trades
  app.get(
    "/markets/:pair/trades",
    async (
      request: FastifyRequest<{
        Params: { pair: string };
        Querystring: TradesQuery;
      }>,
      reply: FastifyReply
    ) => {
      const { pair } = request.params as { pair: string };
      const parsed = TradesQuerySchema.safeParse(request.query);
      if (!parsed.success) {
        throw new ValidationError(parsed.error.errors[0].message);
      }

      const validPairs = marketDataService
        .getPairs()
        .map((p) => p.symbol);
      if (!validPairs.includes(pair)) {
        throw new NotFoundError(`Trading pair ${pair} not found`);
      }

      const trades = marketDataService.getTrades(pair, parsed.data.limit);
      return reply.send({ trades });
    }
  );

  // GET /markets/:pair/klines
  app.get(
    "/markets/:pair/klines",
    async (
      request: FastifyRequest<{
        Params: { pair: string };
        Querystring: KlinesQuery;
      }>,
      reply: FastifyReply
    ) => {
      const { pair } = request.params as { pair: string };
      const parsed = KlinesQuerySchema.safeParse(request.query);
      if (!parsed.success) {
        throw new ValidationError(parsed.error.errors[0].message);
      }

      const validPairs = marketDataService
        .getPairs()
        .map((p) => p.symbol);
      if (!validPairs.includes(pair)) {
        throw new NotFoundError(`Trading pair ${pair} not found`);
      }

      const { interval, limit, start, end } = parsed.data;
      const klines = marketDataService.getKlines(
        pair,
        interval,
        limit,
        start,
        end
      );
      return reply.send({ klines });
    }
  );

  // GET /markets/:pair/ticker
  app.get(
    "/markets/:pair/ticker",
    async (
      request: FastifyRequest<{ Params: { pair: string } }>,
      reply: FastifyReply
    ) => {
      const { pair } = request.params as { pair: string };

      const validPairs = marketDataService
        .getPairs()
        .map((p) => p.symbol);
      if (!validPairs.includes(pair)) {
        throw new NotFoundError(`Trading pair ${pair} not found`);
      }

      const ticker = marketDataService.getTicker(pair);
      return reply.send(ticker);
    }
  );
}
