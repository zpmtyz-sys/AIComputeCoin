export default function Home() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center px-4">
      <section className="text-center max-w-4xl mx-auto py-20">
        <h1 className="text-5xl font-bold mb-6 bg-gradient-to-r from-compute-blue-400 to-oracle-green-400 bg-clip-text text-transparent">
          The World&apos;s First Compute Power Exchange
        </h1>
        <p className="text-xl text-gray-400 mb-8 max-w-2xl mx-auto">
          Trade GPU compute power like financial assets. Real-time pricing,
          deep liquidity, and oracle-verified performance.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
          <div className="p-6 rounded-lg bg-[var(--card-bg)] border border-[var(--border)]">
            <h3 className="text-lg font-semibold text-compute-blue-400 mb-2">
              Real-Time Trading
            </h3>
            <p className="text-gray-400 text-sm">
              Sub-millisecond matching engine with price-time priority orderbook
            </p>
          </div>
          <div className="p-6 rounded-lg bg-[var(--card-bg)] border border-[var(--border)]">
            <h3 className="text-lg font-semibold text-oracle-green-400 mb-2">
              Oracle Verified
            </h3>
            <p className="text-gray-400 text-sm">
              Three-layer verification: hardware fingerprinting, TEE attestation,
              and economic validation
            </p>
          </div>
          <div className="p-6 rounded-lg bg-[var(--card-bg)] border border-[var(--border)]">
            <h3 className="text-lg font-semibold text-purple-400 mb-2">
              DeFi Native
            </h3>
            <p className="text-gray-400 text-sm">
              On-chain settlement, staking rewards, and governance
            </p>
          </div>
        </div>

        <a
          href="/trade"
          className="inline-block px-8 py-3 bg-compute-blue-600 hover:bg-compute-blue-700 text-white font-semibold rounded-lg transition-colors"
        >
          Start Trading
        </a>
      </section>
    </main>
  );
}
