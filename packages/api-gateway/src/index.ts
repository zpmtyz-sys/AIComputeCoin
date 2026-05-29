import { config } from "./config.js";
import { buildApp } from "./app.js";

export { buildApp } from "./app.js";

async function start(): Promise<void> {
  const app = await buildApp();

  // Start server
  try {
    await app.listen({ port: config.port, host: "0.0.0.0" });
    app.log.info(`API Gateway listening on port ${config.port}`);
  } catch (err) {
    app.log.error(err);
    process.exit(1);
  }

  // Graceful shutdown
  const shutdown = async (): Promise<void> => {
    app.log.info("Shutting down gracefully...");
    await app.close();
    process.exit(0);
  };

  process.on("SIGTERM", shutdown);
  process.on("SIGINT", shutdown);
}

start();
