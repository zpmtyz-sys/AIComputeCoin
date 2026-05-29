import argon2 from "argon2";
import { v4 as uuidv4 } from "uuid";

export interface User {
  id: string;
  email: string;
  name: string;
  passwordHash: string;
  role: string;
  createdAt: string;
  apiKeys: ApiKey[];
  balance: Record<string, string>;
}

export interface ApiKey {
  id: string;
  name: string;
  key: string;
  permissions: string[];
  createdAt: string;
}

export class UserStore {
  private users: Map<string, User> = new Map();
  private emailIndex: Map<string, string> = new Map();

  async createUser(
    email: string,
    password: string,
    name: string
  ): Promise<User> {
    if (this.emailIndex.has(email)) {
      throw new Error("Email already registered");
    }

    const id = uuidv4();
    const passwordHash = await argon2.hash(password);
    const user: User = {
      id,
      email,
      name,
      passwordHash,
      role: "user",
      createdAt: new Date().toISOString(),
      apiKeys: [],
      balance: {
        USDT: "10000.00",
        BTC: "1.00",
        ETH: "10.00",
        CC: "1000.00",
      },
    };

    this.users.set(id, user);
    this.emailIndex.set(email, id);
    return user;
  }

  findByEmail(email: string): User | undefined {
    const id = this.emailIndex.get(email);
    if (!id) return undefined;
    return this.users.get(id);
  }

  findById(id: string): User | undefined {
    return this.users.get(id);
  }

  async validatePassword(user: User, password: string): Promise<boolean> {
    return argon2.verify(user.passwordHash, password);
  }

  createApiKey(userId: string, name: string, permissions: string[]): ApiKey {
    const user = this.users.get(userId);
    if (!user) throw new Error("User not found");

    const apiKey: ApiKey = {
      id: uuidv4(),
      name,
      key: `cc_${uuidv4().replace(/-/g, "")}`,
      permissions,
      createdAt: new Date().toISOString(),
    };

    user.apiKeys.push(apiKey);
    return apiKey;
  }

  listApiKeys(userId: string): ApiKey[] {
    const user = this.users.get(userId);
    if (!user) return [];
    return user.apiKeys;
  }

  deleteApiKey(userId: string, keyId: string): boolean {
    const user = this.users.get(userId);
    if (!user) return false;

    const idx = user.apiKeys.findIndex((k) => k.id === keyId);
    if (idx === -1) return false;

    user.apiKeys.splice(idx, 1);
    return true;
  }

  getBalance(userId: string): Record<string, string> {
    const user = this.users.get(userId);
    if (!user) return {};
    return { ...user.balance };
  }
}

// Singleton instance
export const userStore = new UserStore();
