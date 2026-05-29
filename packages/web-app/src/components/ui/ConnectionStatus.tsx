"use client";

interface ConnectionStatusProps {
  isConnected: boolean;
}

export default function ConnectionStatus({ isConnected }: ConnectionStatusProps) {
  return (
    <div className="flex items-center gap-1.5 text-xs">
      <span
        className={`w-2 h-2 rounded-full ${
          isConnected ? "bg-positive" : "bg-negative"
        }`}
      />
      <span className="text-gray-400">
        {isConnected ? "Connected" : "Disconnected"}
      </span>
    </div>
  );
}
