"use client";

import { useState } from "react";

export default function TwoFactorSetup() {
  const [code, setCode] = useState("");
  const [isVerifying, setIsVerifying] = useState(false);

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsVerifying(true);
    // Simulated verification
    await new Promise((resolve) => setTimeout(resolve, 1000));
    setIsVerifying(false);
  };

  return (
    <div className="max-w-md mx-auto">
      <h2 className="text-xl font-bold mb-4">Two-Factor Authentication</h2>
      <p className="text-sm text-gray-400 mb-6">
        Scan the QR code with your authenticator app to enable 2FA.
      </p>

      {/* QR Code Placeholder */}
      <div className="w-48 h-48 mx-auto mb-6 rounded-lg bg-white flex items-center justify-center">
        <div className="text-center text-gray-800 text-xs p-4">
          <div className="w-32 h-32 bg-gray-200 rounded flex items-center justify-center">
            QR Code
          </div>
        </div>
      </div>

      <form onSubmit={handleVerify} className="space-y-4">
        <div>
          <label htmlFor="totp-code" className="block text-sm text-gray-400 mb-1">
            Enter 6-digit code
          </label>
          <input
            id="totp-code"
            type="text"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
            placeholder="000000"
            maxLength={6}
            className="w-full px-3 py-2.5 bg-background border border-border rounded-md text-center text-lg font-mono tracking-widest focus:outline-none focus:border-accent"
          />
        </div>
        <button
          type="submit"
          disabled={code.length !== 6 || isVerifying}
          className="w-full py-2.5 bg-accent hover:bg-accent/90 text-white font-medium rounded-md text-sm transition-colors disabled:opacity-50"
        >
          {isVerifying ? "Verifying..." : "Verify & Enable"}
        </button>
      </form>
    </div>
  );
}
