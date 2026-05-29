import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { FastifyInstance } from "fastify";
import { buildApp } from "../src/app.js";

describe("Auth Routes", () => {
  let app: FastifyInstance;

  beforeAll(async () => {
    app = await buildApp({ logger: false, rateLimit: false });
    await app.ready();
  });

  afterAll(async () => {
    await app.close();
  });

  describe("POST /api/v1/auth/register", () => {
    it("should register a new user and return tokens", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/register",
        payload: {
          email: "test@example.com",
          password: "securePassword123",
          name: "Test User",
        },
      });

      expect(res.statusCode).toBe(201);
      const body = JSON.parse(res.body);
      expect(body.user.email).toBe("test@example.com");
      expect(body.user.name).toBe("Test User");
      expect(body.user.role).toBe("user");
      expect(body.accessToken).toBeDefined();
      expect(body.refreshToken).toBeDefined();
    });

    it("should reject duplicate email registration", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/register",
        payload: {
          email: "test@example.com",
          password: "anotherPassword123",
          name: "Another User",
        },
      });

      expect(res.statusCode).toBe(400);
      const body = JSON.parse(res.body);
      expect(body.error).toBe("VALIDATION_ERROR");
    });

    it("should reject invalid email format", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/register",
        payload: {
          email: "invalid-email",
          password: "securePassword123",
          name: "Test User",
        },
      });

      expect(res.statusCode).toBe(400);
    });

    it("should reject password shorter than 8 chars", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/register",
        payload: {
          email: "short@example.com",
          password: "short",
          name: "Test User",
        },
      });

      expect(res.statusCode).toBe(400);
    });
  });

  describe("POST /api/v1/auth/login", () => {
    it("should login with valid credentials", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/login",
        payload: {
          email: "test@example.com",
          password: "securePassword123",
        },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.user.email).toBe("test@example.com");
      expect(body.accessToken).toBeDefined();
      expect(body.refreshToken).toBeDefined();
    });

    it("should reject invalid password", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/login",
        payload: {
          email: "test@example.com",
          password: "wrongPassword",
        },
      });

      expect(res.statusCode).toBe(401);
      const body = JSON.parse(res.body);
      expect(body.error).toBe("AUTH_ERROR");
    });

    it("should reject non-existent email", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/login",
        payload: {
          email: "nonexistent@example.com",
          password: "securePassword123",
        },
      });

      expect(res.statusCode).toBe(401);
    });
  });

  describe("POST /api/v1/auth/refresh", () => {
    it("should return new access token with valid refresh token", async () => {
      // First login to get refresh token
      const loginRes = await app.inject({
        method: "POST",
        url: "/api/v1/auth/login",
        payload: {
          email: "test@example.com",
          password: "securePassword123",
        },
      });

      const { refreshToken } = JSON.parse(loginRes.body);

      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/refresh",
        payload: { refreshToken },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.accessToken).toBeDefined();
    });

    it("should reject invalid refresh token", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/refresh",
        payload: { refreshToken: "invalid-token" },
      });

      expect(res.statusCode).toBe(401);
    });
  });

  describe("API Keys", () => {
    let accessToken: string;

    beforeAll(async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/login",
        payload: {
          email: "test@example.com",
          password: "securePassword123",
        },
      });
      accessToken = JSON.parse(res.body).accessToken;
    });

    it("should create an API key", async () => {
      const res = await app.inject({
        method: "POST",
        url: "/api/v1/auth/api-keys",
        headers: { authorization: `Bearer ${accessToken}` },
        payload: { name: "Test Key", permissions: ["read", "trade"] },
      });

      expect(res.statusCode).toBe(201);
      const body = JSON.parse(res.body);
      expect(body.apiKey.name).toBe("Test Key");
      expect(body.apiKey.key).toBeDefined();
    });

    it("should list API keys", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/auth/api-keys",
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(200);
      const body = JSON.parse(res.body);
      expect(body.apiKeys).toHaveLength(1);
      expect(body.apiKeys[0].name).toBe("Test Key");
    });

    it("should delete an API key", async () => {
      // Get keys first
      const listRes = await app.inject({
        method: "GET",
        url: "/api/v1/auth/api-keys",
        headers: { authorization: `Bearer ${accessToken}` },
      });
      const { apiKeys } = JSON.parse(listRes.body);
      const keyId = apiKeys[0].id;

      const res = await app.inject({
        method: "DELETE",
        url: `/api/v1/auth/api-keys/${keyId}`,
        headers: { authorization: `Bearer ${accessToken}` },
      });

      expect(res.statusCode).toBe(204);

      // Verify deletion
      const listRes2 = await app.inject({
        method: "GET",
        url: "/api/v1/auth/api-keys",
        headers: { authorization: `Bearer ${accessToken}` },
      });
      expect(JSON.parse(listRes2.body).apiKeys).toHaveLength(0);
    });

    it("should reject unauthenticated access to API keys", async () => {
      const res = await app.inject({
        method: "GET",
        url: "/api/v1/auth/api-keys",
      });

      expect(res.statusCode).toBe(401);
    });
  });
});
