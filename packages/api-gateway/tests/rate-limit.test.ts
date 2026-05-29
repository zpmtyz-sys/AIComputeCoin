import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { FastifyInstance } from "fastify";
import { buildApp } from "../src/app.js";

describe("Rate Limiting", () => {
  let app: FastifyInstance;

  beforeAll(async () => {
    app = await buildApp({ logger: false, rateLimit: true });
    await app.ready();
  });

  afterAll(async () => {
    await app.close();
  });

  it("should return 429 after exceeding public rate limit", async () => {
    // Public rate limit is 10 req/s
    const results = [];
    for (let i = 0; i < 12; i++) {
      const res = await app.inject({
        method: "GET",
        url: "/health",
        remoteAddress: "10.0.0.1",
      });
      results.push(res.statusCode);
    }

    // At least one request should be rate limited
    expect(results).toContain(429);
  });

  it("should allow more requests for authenticated users", async () => {
    // Register a user
    await app.inject({
      method: "POST",
      url: "/api/v1/auth/register",
      payload: {
        email: "ratelimit@example.com",
        password: "securePassword123",
        name: "Rate Limit User",
      },
      remoteAddress: "10.0.0.2",
    });

    const loginRes = await app.inject({
      method: "POST",
      url: "/api/v1/auth/login",
      payload: {
        email: "ratelimit@example.com",
        password: "securePassword123",
      },
      remoteAddress: "10.0.0.2",
    });

    const { accessToken } = JSON.parse(loginRes.body);

    // Authenticated rate limit is 50 req/s - send 15 to verify no 429
    const results = [];
    for (let i = 0; i < 15; i++) {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/orders",
        headers: { authorization: `Bearer ${accessToken}` },
        remoteAddress: "10.0.0.2",
      });
      results.push(res.statusCode);
    }

    // All requests should go through (authenticated limit is 50/s)
    const rateLimited = results.filter((s) => s === 429);
    expect(rateLimited.length).toBe(0);
  });
});
