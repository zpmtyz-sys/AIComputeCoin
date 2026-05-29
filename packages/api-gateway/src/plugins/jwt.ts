import { FastifyInstance, FastifyRequest, FastifyReply } from "fastify";
import fastifyJwt from "@fastify/jwt";
import fp from "fastify-plugin";
import { config } from "../config.js";

export interface JwtPayload {
  userId: string;
  email: string;
  role: string;
  type?: string;
}

declare module "fastify" {
  interface FastifyInstance {
    authenticate: (
      request: FastifyRequest,
      reply: FastifyReply
    ) => Promise<void>;
  }
}

declare module "@fastify/jwt" {
  interface FastifyJWT {
    payload: JwtPayload;
    user: JwtPayload;
  }

  interface JWT {
    refresh: {
      sign(payload: object, options?: object): string;
      verify<Decoded extends object>(token: string, options?: object): Decoded;
    };
  }
}

async function jwtPlugin(app: FastifyInstance): Promise<void> {
  // Register primary JWT for access tokens
  await app.register(fastifyJwt, {
    secret: config.jwtSecret,
    sign: {
      expiresIn: config.accessTokenExpiry,
    },
  });

  // Register a second JWT instance for refresh tokens with a separate secret
  await app.register(fastifyJwt, {
    secret: config.jwtRefreshSecret,
    namespace: "refresh",
    sign: {
      expiresIn: config.refreshTokenExpiry,
    },
  });

  app.decorate(
    "authenticate",
    async (request: FastifyRequest, reply: FastifyReply): Promise<void> => {
      try {
        await request.jwtVerify();
      } catch {
        reply.status(401).send({
          error: "AUTH_ERROR",
          message: "Unauthorized",
        });
      }
    }
  );
}

export default fp(jwtPlugin, {
  name: "jwt-plugin",
});
