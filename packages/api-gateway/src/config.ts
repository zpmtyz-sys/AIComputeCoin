import "dotenv/config";

export interface RateLimitConfig {
  public: number;
  authenticated: number;
  marketMaker: number;
}

export interface Config {
  port: number;
  jwtSecret: string;
  jwtRefreshSecret: string;
  accessTokenExpiry: string;
  refreshTokenExpiry: string;
  databaseUrl: string;
  redisUrl: string;
  kafkaBrokers: string;
  matchingEngineUrl: string;
  rateLimits: RateLimitConfig;
}

export const config: Config = {
  port: parseInt(process.env.PORT || "3000", 10),
  jwtSecret: process.env.JWT_SECRET || "dev-secret-change-in-production",
  jwtRefreshSecret:
    process.env.JWT_REFRESH_SECRET || "dev-refresh-secret-change-in-production",
  accessTokenExpiry: process.env.ACCESS_TOKEN_EXPIRY || "15m",
  refreshTokenExpiry: process.env.REFRESH_TOKEN_EXPIRY || "7d",
  databaseUrl:
    process.env.DATABASE_URL ||
    "postgresql://computecoin:computecoin_dev@localhost:5432/computecoin",
  redisUrl: process.env.REDIS_URL || "redis://localhost:6379",
  kafkaBrokers: process.env.KAFKA_BROKERS || "localhost:9092",
  matchingEngineUrl: process.env.MATCHING_ENGINE_URL || "localhost:50051",
  rateLimits: {
    public: parseInt(process.env.RATE_LIMIT_PUBLIC || "10", 10),
    authenticated: parseInt(process.env.RATE_LIMIT_AUTH || "50", 10),
    marketMaker: parseInt(process.env.RATE_LIMIT_MARKET_MAKER || "200", 10),
  },
};
