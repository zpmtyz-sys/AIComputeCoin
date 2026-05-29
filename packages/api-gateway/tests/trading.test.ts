import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { FastifyInstance } from "fastify";
import { buildApp } from "../src/app.js";

describe("Trading Routes", () => {
  let app: FastifyInstance;
  let accessToken: string;

  beforeAll(async () => {
    app = await buildApp({ logger: false, rateLimit: false });
    await app.ready();

    // Register and login to get a token
    await app.inject({
      method: "POST",
      url: "/api/v1/auth/register",
      payload: {
        email: "trader@example.com",
        password: "securePassword123",
        name: "Trader",
      },
    });

    const loginRes = await app.inject({
      method: "POST",
      url: "/api/v1/auth/login",
      payload: {
        email: "trader@example.com",
        password: "securePassword123",
      },
    });

    accessToken = JSON.parse(loginRes.body).accessToken;
  });

  afterAll(async () => {
    await app.close();
  });

  describe("POST /api/v1/orders", () => {
    it("should submit a limit order", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
        payload: {
          pair: "BTC-USDT",
          side: "buy",
          type: "limit",
          price: "65000.00",
          quantity: "0.1",
        },
      });

      expect(res.statusCode).toBe(201);
      const body = JSON.parse(res.body);
      expect(body.order.pair).toBe("BTC-USDT");
      expect(body.order.side).toBe("buy");
      expect(body.order.type).toBe("limit");
      expect(body.order.status).toBe("new");
    });

    it("should submit a market order", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
        payload: {
          pair: "ETH-USDT",
          side: "sell",
          type: "market",
          quantity: "1.5",
        },
      });

      expect(res.statusCode).toBe(201);
      const body = JSON.parse(res.body);
      expect(body.order.pair).toBe("ETH-USDT");
      expect(body.order.type).toBe("market");
    });

    it("should reject order without authentication", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/orders",
        payload: {
          pair: "BTC-USDT",
          side: "buy",
          type: "limit",
          price: "65000.00",
          quantity: "0.1",
        },
      });

      expect(res.statusCode).toBe(401);
    });

    it("should reject limit order without price", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
        payload: {
          pair: "BTC-USDT",
          side: "buy",
          type: "limit",
          quantity: "0.1",
        },
      });

      expect(res.statusCode).toBe(400);
    });

    it("should reject order with invalid pair", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
        payload: {
          pair: "INVALID-PAIR",
          side: "buy",
          type: "market",
          quantity: "0.1",
        },
      });

      expect(res.statusCode).toBe(400);
    });
  });

  describe("DELETE /api/v1/orders/:id", () => {
    it("should cancel an existing order", async () => {
      // Create an order first
      const createRes = await app.inject({
        method: "POST",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
        payload: {
          pair: "BTC-USDT",
          side: "buy",
          type: "limit",
          price: "60000.00",
          quantity: "0.5",
        },
      });

      const { order } = JSON.parse(createRes.body);

      const res = await app.inject({
        method: "DELETE",
        url: `/api/v1/orders/${order.id}`,
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.order.status).toBe("cancelled");
    });

    it("should return 404 for non-existent order", async () => {
      const res = await app.inject({
        method: "DELETE",
        url: "/api/v1/orders/non-existent-id",
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(404);
    });
  });

  describe("GET /api/v1/orders", () => {
    it("should list user orders", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.orders).toBeDefined();
      expect(Array.isArray(body.orders)).toBe(true);
      expect(body.total).toBeDefined();
    });

    it("should reject unauthenticated access", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/orders",
      });

      expect(res.statusCode).toBe(401);
    });
  });

  describe("GET /api/v1/orders/:id", () => {
    it("should get a specific order", async () => {
      // Create an order
      const createRes = await app.inject({
        method: "POST",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
        payload: {
          pair: "CC-USDT",
          side: "buy",
          type: "market",
          quantity: "100",
        },
      });

      const { order } = JSON.parse(createRes.body);

      const res = await app.inject({
        method: "GET",
        url: `/api/v1/orders/${order.id}`,
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.order.id).toBe(order.id);
    });
  });

  describe("GET /api/v1/balance", () => {
    it("should return user balance", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/balance",
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.balance).toBeDefined();
      expect(body.balance.USDT).toBeDefined();
      expect(body.balance.BTC).toBeDefined();
    });

    it("should reject unauthenticated access", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/balance",
      });

      expect(res.statusCode).toBe(401);
    });
  });

  describe("GET /api/v1/positions", () => {
    it("should return positions list", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/positions",
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.positions).toBeDefined();
      expect(Array.isArray(body.positions)).toBe(true);
    });
  });
});
