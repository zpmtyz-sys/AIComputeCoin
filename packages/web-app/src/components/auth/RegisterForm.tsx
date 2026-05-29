"use client";

import { useState } from "react";
import Link from "next/link";
import { z } from "zod";
import { useAuth } from "@/lib/auth";

const registerSchema = z
  .object({
    email: z.string().email("Invalid email address"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirmPassword: z.string(),
    termsAccepted: z.literal(true, {
      errorMap: () => ({ message: "You must accept the terms" }),
    }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

export default function RegisterForm() {
  const { register } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [termsAccepted, setTermsAccepted] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});

    const result = registerSchema.safeParse({
      email,
      password,
      confirmPassword,
      termsAccepted,
    });
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.errors.forEach((err) => {
        const field = err.path[0] as string;
        fieldErrors[field] = err.message;
      });
      setErrors(fieldErrors);
      return;
    }

    setIsSubmitting(true);
    try {
      await register(email, password);
    } catch {
      setErrors({ form: "Registration failed. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="reg-email" className="block text-sm text-gray-400 mb-1">
          Email
        </label>
        <input
          id="reg-email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@example.com"
          className="w-full px-3 py-2.5 bg-background border border-border rounded-md text-sm focus:outline-none focus:border-accent"
        />
        {errors.email && (
          <span className="text-xs text-negative mt-1 block">{errors.email}</span>
        )}
      </div>

      <div>
        <label htmlFor="reg-password" className="block text-sm text-gray-400 mb-1">
          Password
        </label>
        <input
          id="reg-password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Min 8 characters"
          className="w-full px-3 py-2.5 bg-background border border-border rounded-md text-sm focus:outline-none focus:border-accent"
        />
        {errors.password && (
          <span className="text-xs text-negative mt-1 block">{errors.password}</span>
        )}
      </div>

      <div>
        <label htmlFor="reg-confirm" className="block text-sm text-gray-400 mb-1">
          Confirm Password
        </label>
        <input
          id="reg-confirm"
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          placeholder="Repeat password"
          className="w-full px-3 py-2.5 bg-background border border-border rounded-md text-sm focus:outline-none focus:border-accent"
        />
        {errors.confirmPassword && (
          <span className="text-xs text-negative mt-1 block">
            {errors.confirmPassword}
          </span>
        )}
      </div>

      <div className="flex items-start gap-2">
        <input
          id="terms"
          type="checkbox"
          checked={termsAccepted}
          onChange={(e) => setTermsAccepted(e.target.checked)}
          className="mt-1"
        />
        <label htmlFor="terms" className="text-xs text-gray-400">
          I agree to the Terms of Service and Privacy Policy
        </label>
      </div>
      {errors.termsAccepted && (
        <span className="text-xs text-negative block">{errors.termsAccepted}</span>
      )}

      {errors.form && (
        <div className="text-sm text-negative text-center">{errors.form}</div>
      )}

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full py-2.5 bg-accent hover:bg-accent/90 text-white font-medium rounded-md text-sm transition-colors disabled:opacity-50"
      >
        {isSubmitting ? "Creating account..." : "Create Account"}
      </button>

      <p className="text-center text-sm text-gray-400">
        Already have an account?{" "}
        <Link href="/login" className="text-accent hover:underline">
          Sign In
        </Link>
      </p>
    </form>
  );
}
