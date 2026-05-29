"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from "react";
import React from "react";

interface User {
  id: string;
  email: string;
  displayName: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Token stored in memory (simulating httpOnly cookie pattern)
let accessToken: string | null = null;

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Auto-check auth on mount
    const checkAuth = async () => {
      try {
        if (accessToken) {
          setUser({
            id: "user-1",
            email: "trader@computecoin.io",
            displayName: "Trader",
          });
        }
      } finally {
        setIsLoading(false);
      }
    };
    checkAuth();
  }, []);

  const login = useCallback(async (email: string, _password: string) => {
    // Simulated login - in production this calls the API
    await new Promise((resolve) => setTimeout(resolve, 500));
    accessToken = "mock-jwt-token";
    setUser({
      id: "user-1",
      email,
      displayName: email.split("@")[0],
    });
  }, []);

  const register = useCallback(async (email: string, _password: string) => {
    // Simulated registration
    await new Promise((resolve) => setTimeout(resolve, 500));
    accessToken = "mock-jwt-token";
    setUser({
      id: "user-1",
      email,
      displayName: email.split("@")[0],
    });
  }, []);

  const logout = useCallback(() => {
    accessToken = null;
    setUser(null);
  }, []);

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    register,
    logout,
  };

  return React.createElement(AuthContext.Provider, { value }, children);
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
