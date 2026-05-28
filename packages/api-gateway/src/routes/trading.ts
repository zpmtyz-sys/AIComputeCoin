import { FastifyInstance } from "fastify";

export async function tradingRoutes(app: FastifyInstance) {
  app.post("/orders", async (_request, reply) => {
    reply.status(501).send({ error: "Not implemented" });
  });

  app.delete("/orders/:id", async (_request, reply) => {
    reply.status(501).send({ error: "Not implemented" });
  });
}
