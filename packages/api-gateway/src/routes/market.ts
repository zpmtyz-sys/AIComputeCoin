import { FastifyInstance } from "fastify";

export async function marketRoutes(app: FastifyInstance) {
  app.get("/markets", async (_request, reply) => {
    reply.status(501).send({ error: "Not implemented" });
  });

  app.get("/markets/:pair/orderbook", async (_request, reply) => {
    reply.status(501).send({ error: "Not implemented" });
  });
}
