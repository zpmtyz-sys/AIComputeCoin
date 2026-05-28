import { FastifyInstance } from "fastify";

export async function authRoutes(app: FastifyInstance) {
  app.post("/register", async (_request, reply) => {
    reply.status(501).send({ error: "Not implemented" });
  });

  app.post("/login", async (_request, reply) => {
    reply.status(501).send({ error: "Not implemented" });
  });
}
