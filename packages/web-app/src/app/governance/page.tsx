const MOCK_PROPOSALS = [
  {
    id: "PROP-001",
    title: "Increase staking rewards from 5% to 7% APY",
    status: "active" as const,
    votesFor: 1234567,
    votesAgainst: 456789,
    endDate: "2024-02-01",
  },
  {
    id: "PROP-002",
    title: "Add FPGA compute units to spot market",
    status: "passed" as const,
    votesFor: 2345678,
    votesAgainst: 123456,
    endDate: "2024-01-20",
  },
  {
    id: "PROP-003",
    title: "Reduce maker fee from 0.10% to 0.08%",
    status: "active" as const,
    votesFor: 890123,
    votesAgainst: 780456,
    endDate: "2024-02-05",
  },
  {
    id: "PROP-004",
    title: "Implement cross-chain bridge to Ethereum",
    status: "rejected" as const,
    votesFor: 345678,
    votesAgainst: 1234567,
    endDate: "2024-01-10",
  },
];

function StatusBadge({ status }: { status: "active" | "passed" | "rejected" }) {
  const styles = {
    active: "bg-accent/10 text-accent",
    passed: "bg-positive/10 text-positive",
    rejected: "bg-negative/10 text-negative",
  };

  return (
    <span
      className={`px-2 py-0.5 rounded-full text-xs font-medium capitalize ${styles[status]}`}
    >
      {status}
    </span>
  );
}

export default function GovernancePage() {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Governance</h1>

      <div className="space-y-3">
        {MOCK_PROPOSALS.map((proposal) => {
          const totalVotes = proposal.votesFor + proposal.votesAgainst;
          const forPct =
            totalVotes > 0 ? (proposal.votesFor / totalVotes) * 100 : 0;

          return (
            <div
              key={proposal.id}
              className="rounded-lg bg-card-bg border border-border p-4"
            >
              <div className="flex items-start justify-between mb-3">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs text-gray-400 font-mono">
                      {proposal.id}
                    </span>
                    <StatusBadge status={proposal.status} />
                  </div>
                  <h3 className="text-sm font-medium">{proposal.title}</h3>
                </div>
                <span className="text-xs text-gray-400">
                  Ends {proposal.endDate}
                </span>
              </div>

              {/* Vote Progress */}
              <div className="mb-3">
                <div className="flex justify-between text-xs text-gray-400 mb-1">
                  <span>For: {(proposal.votesFor / 1000000).toFixed(2)}M</span>
                  <span>
                    Against: {(proposal.votesAgainst / 1000000).toFixed(2)}M
                  </span>
                </div>
                <div className="h-2 bg-background rounded-full overflow-hidden">
                  <div
                    className="h-full bg-positive rounded-full"
                    style={{ width: `${forPct}%` }}
                  />
                </div>
              </div>

              {proposal.status === "active" && (
                <div className="flex gap-2">
                  <button className="flex-1 py-1.5 text-xs font-medium rounded border border-positive text-positive hover:bg-positive/10 transition-colors">
                    Vote For
                  </button>
                  <button className="flex-1 py-1.5 text-xs font-medium rounded border border-negative text-negative hover:bg-negative/10 transition-colors">
                    Vote Against
                  </button>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
