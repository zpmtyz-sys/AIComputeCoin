import { z } from "zod";

// Auth schemas
export const RegisterBodySchema = z.object({
  email: z.string().email(),
  password: z.string().min(8).max(128),
  name: z.string().min(1).max(100),
});

export const LoginBodySchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
});

export const RefreshBodySchema = z.object({
  refreshToken: z.string().min(1),
});

export const ApiKeyBodySchema = z.object({
  name: z.string().min(1).max(100),
  permissions: z.array(z.string()).optional(),
});

// Trading schemas
export const OrderBodySchema = z.object({
  pair: z.string().min(1),
  side: z.enum(["buy", "sell"]),
  type: z.enum(["limit", "market", "stop_limit", "ioc", "fok"]),
  price: z.string().optional(),
  quantity: z.string().min(1),
  stopPrice: z.string().optional(),
});

export const OrderQuerySchema = z.object({
  pair: z.string().optional(),
  status: z.string().optional(),
  limit: z.coerce.number().int().min(1).max(100).default(50),
  offset: z.coerce.number().int().min(0).default(0),
});

// Market schemas
export const OrderbookQuerySchema = z.object({
  depth: z.coerce.number().int().min(1).max(100).default(20),
});

export const TradesQuerySchema = z.object({
  limit: z.coerce.number().int().min(1).max(100).default(50),
});

export const KlinesQuerySchema = z.object({
  interval: z.enum(["1m", "5m", "15m", "1h", "4h", "1d"]).default("1h"),
  start: z.coerce.number().optional(),
  end: z.coerce.number().optional(),
  limit: z.coerce.number().int().min(1).max(500).default(100),
});

// Type exports
export type RegisterBody = z.infer<typeof RegisterBodySchema>;
export type LoginBody = z.infer<typeof LoginBodySchema>;
export type RefreshBody = z.infer<typeof RefreshBodySchema>;
export type ApiKeyBody = z.infer<typeof ApiKeyBodySchema>;
export type OrderBody = z.infer<typeof OrderBodySchema>;
export type OrderQuery = z.infer<typeof OrderQuerySchema>;
export type OrderbookQuery = z.infer<typeof OrderbookQuerySchema>;
export type TradesQuery = z.infer<typeof TradesQuerySchema>;
export type KlinesQuery = z.infer<typeof KlinesQuerySchema>;
