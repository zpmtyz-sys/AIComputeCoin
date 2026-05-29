import { FastifyInstance, FastifyRequest } from "fastify";
import rateLimit from "@fastify/rate-limit";
import fp from "fastify-plugin";
import { config } from "../config.js";

export interface RateLimitPluginOptions {
  enabled?: boolean;
}

async function rateLimitPlugin(
  app: FastifyInstance,
  opts: RateLimitPluginOptions
): Promise<void> {
  if (opts.enabled === false) {
    return;
  }

  await app.register(rateLimit, {
    max: (request: FastifyRequest): number => {
      const user = request.user as { role?: string } | undefined;
      if (user?.role === "market_maker") {
        return config.rateLimits.marketMaker;
      }
      if (user?.role) {
        return config.rateLimits.authenticated;
      }
      return config.rateLimits.public;
    },
    timeWindow: "1 second",
    hook: "preHandler",
    keyGenerator: (request: FastifyRequest): string => {
      const user = request.user as { userId?: string } | undefined;
      if (user?.userId) {
        return user.userId;
      }
      return request.ip;
    },
  });
}

export default fp(rateLimitPlugin, {
  name: "rate-limit-plugin",
});
