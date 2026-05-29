import { FastifyInstance, FastifyReply, FastifyRequest } from "fastify";
import fp from "fastify-plugin";

export class AppError extends Error {
  public statusCode: number;
  public code: string;

  constructor(message: string, statusCode: number, code: string) {
    super(message);
    this.statusCode = statusCode;
    this.code = code;
    this.name = "AppError";
  }
}

export class ValidationError extends AppError {
  constructor(message: string) {
    super(message, 400, "VALIDATION_ERROR");
    this.name = "ValidationError";
  }
}

export class AuthError extends AppError {
  constructor(message: string) {
    super(message, 401, "AUTH_ERROR");
    this.name = "AuthError";
  }
}

export class NotFoundError extends AppError {
  constructor(message: string) {
    super(message, 404, "NOT_FOUND");
    this.name = "NotFoundError";
  }
}

async function errorHandlerPlugin(app: FastifyInstance): Promise<void> {
  app.setErrorHandler(
    (error: Error, _request: FastifyRequest, reply: FastifyReply) => {
      if (error instanceof AppError) {
        return reply.status(error.statusCode).send({
          error: error.code,
          message: error.message,
        });
      }

      // Handle fastify rate limit errors
      if ("statusCode" in error && (error as { statusCode: number }).statusCode === 429) {
        return reply.status(429).send({
          error: "RATE_LIMIT_EXCEEDED",
          message: "Too many requests",
        });
      }

      // Handle JWT errors
      if (error.name === "UnauthorizedError" || error.message === "Unauthorized") {
        return reply.status(401).send({
          error: "AUTH_ERROR",
          message: "Unauthorized",
        });
      }

      // Default server error
      app.log.error(error);
      return reply.status(500).send({
        error: "INTERNAL_ERROR",
        message: "Internal server error",
      });
    }
  );
}

export const errorHandler = fp(errorHandlerPlugin, {
  name: "error-handler",
});
