import Fastify from "fastify";
import cors from "@fastify/cors";
import rateLimit from "@fastify/rate-limit";
import { config } from "./config.js";
import { healthRoutes } from "./routes/health.js";
import { authRoutes } from "./routes/auth.js";
import { tradingRoutes } from "./routes/trading.js";
import { marketRoutes } from "./routes/market.js";

const app = Fastify({
  logger: true,
});

async function start() {
  // Register plugins
  await app.register(cors, { origin: true });
  await app.register(rateLimit, { max: 100, timeWindow: "1 minute" });

  // Register routes
  await app.register(healthRoutes);
  await app.register(authRoutes, { prefix: "/api/v1/auth" });
  await app.register(tradingRoutes, { prefix: "/api/v1" });
  await app.register(marketRoutes, { prefix: "/api/v1" });

  // Start server
  try {
    await app.listen({ port: config.port, host: "0.0.0.0" });
    app.log.info(`API Gateway listening on port ${config.port}`);
  } catch (err) {
    app.log.error(err);
    process.exit(1);
  }
}

// Graceful shutdown
const shutdown = async () => {
  app.log.info("Shutting down gracefully...");
  await app.close();
  process.exit(0);
};

process.on("SIGTERM", shutdown);
process.on("SIGINT", shutdown);

start();
