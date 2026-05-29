import Fastify, { FastifyInstance } from "fastify";
import cors from "@fastify/cors";
import { errorHandler } from "./lib/errors.js";
import jwtPlugin from "./plugins/jwt.js";
import rateLimitPlugin from "./plugins/rate-limit.js";
import { healthRoutes } from "./routes/health.js";
import { authRoutes } from "./routes/auth.js";
import { tradingRoutes } from "./routes/trading.js";
import { marketRoutes } from "./routes/market.js";
import { websocketPlugin } from "./websocket/index.js";

export interface BuildAppOptions {
  logger?: boolean;
  rateLimit?: boolean;
}

export async function buildApp(opts?: BuildAppOptions): Promise<FastifyInstance> {
  const app = Fastify({
    logger: opts?.logger ?? true,
  });

  // Register error handler
  await app.register(errorHandler);

  // Register plugins
  await app.register(cors, { origin: true });
  await app.register(jwtPlugin);
  await app.register(rateLimitPlugin, { enabled: opts?.rateLimit ?? true });
  await app.register(websocketPlugin);

  // Register routes
  await app.register(healthRoutes);
  await app.register(authRoutes, { prefix: "/api/v1/auth" });
  await app.register(tradingRoutes, { prefix: "/api/v1" });
  await app.register(marketRoutes, { prefix: "/api/v1" });

  return app;
}
