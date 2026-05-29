"use client";

import RegisterForm from "@/components/auth/RegisterForm";
import { AuthProvider } from "@/lib/auth";

export default function RegisterPage() {
  return (
    <AuthProvider>
      <div className="min-h-screen flex items-center justify-center px-4 -mt-14">
        <div className="w-full max-w-sm">
          <div className="rounded-lg bg-card-bg border border-border p-6">
            <h1 className="text-xl font-bold text-center mb-6">Create Account</h1>
            <RegisterForm />
          </div>
        </div>
      </div>
    </AuthProvider>
  );
}
