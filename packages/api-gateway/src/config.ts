import "dotenv/config";

export const config = {
  port: parseInt(process.env.PORT || "3000", 10),
  jwtSecret: process.env.JWT_SECRET || "dev-secret-change-in-production",
  databaseUrl:
    process.env.DATABASE_URL ||
    "postgresql://computecoin:computecoin_dev@localhost:5432/computecoin",
  redisUrl: process.env.REDIS_URL || "redis://localhost:6379",
  kafkaBrokers: process.env.KAFKA_BROKERS || "localhost:9092",
  matchingEngineUrl: process.env.MATCHING_ENGINE_URL || "localhost:50051",
};
