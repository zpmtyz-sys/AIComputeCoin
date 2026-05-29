import { FastifyInstance, FastifyRequest, FastifyReply } from "fastify";
import { config } from "../config.js";
import { userStore } from "../services/user-store.js";
import {
  RegisterBodySchema,
  LoginBodySchema,
  RefreshBodySchema,
  ApiKeyBodySchema,
} from "../lib/schemas.js";
import { ValidationError, AuthError, NotFoundError } from "../lib/errors.js";

export async function authRoutes(app: FastifyInstance): Promise<void> {
  // POST /register
  app.post("/register", async (request: FastifyRequest, reply: FastifyReply) => {
    const parsed = RegisterBodySchema.safeParse(request.body);
    if (!parsed.success) {
      throw new ValidationError(parsed.error.errors[0].message);
    }

    const { email, password, name } = parsed.data;

    try {
      const user = await userStore.createUser(email, password, name);
      const accessToken = app.jwt.sign(
        { userId: user.id, email: user.email, role: user.role },
        { expiresIn: config.accessTokenExpiry }
      );
      const refreshToken = app.jwt.refresh.sign(
        { userId: user.id, email: user.email, role: user.role, type: "refresh" },
        { expiresIn: config.refreshTokenExpiry }
      );

      return reply.status(201).send({
        user: {
          id: user.id,
          email: user.email,
          name: user.name,
          role: user.role,
        },
        accessToken,
        refreshToken,
      });
    } catch (err) {
      if (err instanceof Error && err.message === "Email already registered") {
        throw new ValidationError("Email already registered");
      }
      throw err;
    }
  });

  // POST /login
  app.post("/login", async (request: FastifyRequest, reply: FastifyReply) => {
    const parsed = LoginBodySchema.safeParse(request.body);
    if (!parsed.success) {
      throw new ValidationError(parsed.error.errors[0].message);
    }

    const { email, password } = parsed.data;
    const user = userStore.findByEmail(email);
    if (!user) {
      throw new AuthError("Invalid credentials");
    }

    const valid = await userStore.validatePassword(user, password);
    if (!valid) {
      throw new AuthError("Invalid credentials");
    }

    const accessToken = app.jwt.sign(
      { userId: user.id, email: user.email, role: user.role },
      { expiresIn: config.accessTokenExpiry }
    );
    const refreshToken = app.jwt.refresh.sign(
      { userId: user.id, email: user.email, role: user.role, type: "refresh" },
      { expiresIn: config.refreshTokenExpiry }
    );

    return reply.send({
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
        role: user.role,
      },
      accessToken,
      refreshToken,
    });
  });

  // POST /refresh
  app.post("/refresh", async (request: FastifyRequest, reply: FastifyReply) => {
    const parsed = RefreshBodySchema.safeParse(request.body);
    if (!parsed.success) {
      throw new ValidationError(parsed.error.errors[0].message);
    }

    const { refreshToken } = parsed.data;

    try {
      const decoded = app.jwt.refresh.verify<{
        userId: string;
        email: string;
        role: string;
        type?: string;
      }>(refreshToken);

      if (decoded.type !== "refresh") {
        throw new AuthError("Invalid refresh token");
      }

      const accessToken = app.jwt.sign(
        { userId: decoded.userId, email: decoded.email, role: decoded.role },
        { expiresIn: config.accessTokenExpiry }
      );

      return reply.send({ accessToken });
    } catch (err) {
      if (err instanceof AuthError) throw err;
      throw new AuthError("Invalid refresh token");
    }
  });

  // GET /api-keys (requires auth)
  app.get(
    "/api-keys",
    { preHandler: app.authenticate },
    async (request: FastifyRequest, reply: FastifyReply) => {
      const keys = userStore.listApiKeys(request.user.userId);
      return reply.send({
        apiKeys: keys.map((k) => ({
          id: k.id,
          name: k.name,
          key: k.key,
          permissions: k.permissions,
          createdAt: k.createdAt,
        })),
      });
    }
  );

  // POST /api-keys (requires auth)
  app.post(
    "/api-keys",
    { preHandler: app.authenticate },
    async (request: FastifyRequest, reply: FastifyReply) => {
      const parsed = ApiKeyBodySchema.safeParse(request.body);
      if (!parsed.success) {
        throw new ValidationError(parsed.error.errors[0].message);
      }

      const { name, permissions } = parsed.data;
      const apiKey = userStore.createApiKey(
        request.user.userId,
        name,
        permissions || ["read", "trade"]
      );

      return reply.status(201).send({ apiKey });
    }
  );

  // DELETE /api-keys/:id (requires auth)
  app.delete(
    "/api-keys/:id",
    { preHandler: app.authenticate },
    async (request: FastifyRequest, reply: FastifyReply) => {
      const { id } = request.params as { id: string };
      const deleted = userStore.deleteApiKey(request.user.userId, id);
      if (!deleted) {
        throw new NotFoundError("API key not found");
      }

      return reply.status(204).send();
    }
  );
}
