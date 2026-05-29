const MOCK_NODES = [
  {
    id: "node-001",
    name: "NVIDIA A100 Cluster",
    status: "online" as const,
    gpuType: "A100 80GB",
    gpuCount: 8,
    utilization: 87,
    uptime: "99.9%",
    earnings24h: 245.50,
  },
  {
    id: "node-002",
    name: "RTX 4090 Farm",
    status: "online" as const,
    gpuType: "RTX 4090",
    gpuCount: 16,
    utilization: 72,
    uptime: "99.7%",
    earnings24h: 180.20,
  },
  {
    id: "node-003",
    name: "TPU v4 Pod",
    status: "maintenance" as const,
    gpuType: "TPU v4",
    gpuCount: 4,
    utilization: 0,
    uptime: "95.2%",
    earnings24h: 0,
  },
  {
    id: "node-004",
    name: "H100 SXM5 Node",
    status: "online" as const,
    gpuType: "H100 SXM5",
    gpuCount: 4,
    utilization: 95,
    uptime: "99.8%",
    earnings24h: 520.00,
  },
  {
    id: "node-005",
    name: "FPGA Array",
    status: "offline" as const,
    gpuType: "Xilinx U280",
    gpuCount: 12,
    utilization: 0,
    uptime: "88.5%",
    earnings24h: 0,
  },
];

function StatusDot({ status }: { status: "online" | "offline" | "maintenance" }) {
  const colors = {
    online: "bg-positive",
    offline: "bg-negative",
    maintenance: "bg-yellow-500",
  };

  return (
    <div className="flex items-center gap-1.5">
      <span className={`w-2 h-2 rounded-full ${colors[status]}`} />
      <span className="text-xs capitalize text-gray-400">{status}</span>
    </div>
  );
}

export default function ComputePage() {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Compute Nodes</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        {MOCK_NODES.map((node) => (
          <div
            key={node.id}
            className="rounded-lg bg-card-bg border border-border p-4"
          >
            <div className="flex items-start justify-between mb-3">
              <div>
                <h3 className="text-sm font-medium">{node.name}</h3>
                <p className="text-xs text-gray-400 mt-0.5">
                  {node.gpuCount}x {node.gpuType}
                </p>
              </div>
              <StatusDot status={node.status} />
            </div>

            <div className="space-y-2">
              <div className="flex justify-between text-xs">
                <span className="text-gray-400">Utilization</span>
                <span className="font-mono">{node.utilization}%</span>
              </div>
              <div className="h-1.5 bg-background rounded-full overflow-hidden">
                <div
                  className="h-full bg-accent rounded-full transition-all"
                  style={{ width: `${node.utilization}%` }}
                />
              </div>

              <div className="flex justify-between text-xs">
                <span className="text-gray-400">Uptime</span>
                <span className="font-mono">{node.uptime}</span>
              </div>

              <div className="flex justify-between text-xs pt-2 border-t border-border">
                <span className="text-gray-400">24h Earnings</span>
                <span className="font-mono text-positive">
                  ${node.earnings24h.toFixed(2)}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
