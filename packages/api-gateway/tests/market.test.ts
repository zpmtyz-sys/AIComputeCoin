import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { FastifyInstance } from "fastify";
import { buildApp } from "../src/app.js";

describe("Market Routes", () => {
  let app: FastifyInstance;

  beforeAll(async () => {
    app = await buildApp({ logger: false, rateLimit: false });
    await app.ready();
  });

  afterAll(async () => {
    await app.close();
  });

  describe("GET /api/v1/markets/pairs", () => {
    it("should return trading pairs list", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/pairs",
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.pairs).toBeDefined();
      expect(Array.isArray(body.pairs)).toBe(true);
      expect(body.pairs.length).toBeGreaterThan(0);
      expect(body.pairs[0].symbol).toBeDefined();
      expect(body.pairs[0].base).toBeDefined();
      expect(body.pairs[0].quote).toBeDefined();
      expect(body.pairs[0].status).toBe("active");
    });
  });

  describe("GET /api/v1/markets/:pair/orderbook", () => {
    it("should return orderbook with bids and asks", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/BTC-USDT/orderbook",
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.pair).toBe("BTC-USDT");
      expect(body.bids).toBeDefined();
      expect(body.asks).toBeDefined();
      expect(Array.isArray(body.bids)).toBe(true);
      expect(Array.isArray(body.asks)).toBe(true);
      expect(body.bids.length).toBeGreaterThan(0);
      expect(body.bids[0].price).toBeDefined();
      expect(body.bids[0].quantity).toBeDefined();
    });

    it("should respect depth parameter", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/BTC-USDT/orderbook?depth=5",
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.bids).toHaveLength(5);
      expect(body.asks).toHaveLength(5);
    });

    it("should return 404 for invalid pair", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/INVALID-PAIR/orderbook",
      });

      expect(res.statusCode).toBe(404);
    });
  });

  describe("GET /api/v1/markets/:pair/trades", () => {
    it("should return recent trades", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/BTC-USDT/trades",
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.trades).toBeDefined();
      expect(Array.isArray(body.trades)).toBe(true);
      expect(body.trades.length).toBeGreaterThan(0);
      expect(body.trades[0].price).toBeDefined();
      expect(body.trades[0].quantity).toBeDefined();
      expect(body.trades[0].side).toBeDefined();
    });

    it("should respect limit parameter", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/ETH-USDT/trades?limit=10",
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.trades).toHaveLength(10);
    });
  });

  describe("GET /api/v1/markets/:pair/klines", () => {
    it("should return kline data", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/BTC-USDT/klines?interval=1h&limit=10",
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.klines).toBeDefined();
      expect(Array.isArray(body.klines)).toBe(true);
      expect(body.klines.length).toBeLessThanOrEqual(10);
      if (body.klines.length > 0) {
        expect(body.klines[0].open).toBeDefined();
        expect(body.klines[0].high).toBeDefined();
        expect(body.klines[0].low).toBeDefined();
        expect(body.klines[0].close).toBeDefined();
        expect(body.klines[0].volume).toBeDefined();
      }
    });
  });

  describe("GET /api/v1/markets/:pair/ticker", () => {
    it("should return 24h ticker data", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/BTC-USDT/ticker",
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.pair).toBe("BTC-USDT");
      expect(body.lastPrice).toBeDefined();
      expect(body.priceChange).toBeDefined();
      expect(body.priceChangePercent).toBeDefined();
      expect(body.high24h).toBeDefined();
      expect(body.low24h).toBeDefined();
      expect(body.volume24h).toBeDefined();
    });

    it("should return 404 for invalid pair", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/markets/INVALID/ticker",
      });

      expect(res.statusCode).toBe(404);
    });
  });
});
