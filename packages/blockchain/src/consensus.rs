use sha2::{Digest, Sha256};

/// Proof of Compute consensus mechanism.
/// Validates CU (Compute Unit) reports from oracle nodes.
pub struct ProofOfCompute {
    pub min_confidence: f64,
    pub max_deviation: f64,
}

/// A compute proof submitted by an oracle node.
pub struct ComputeProof {
    pub node_id: String,
    pub tflops_measured: f64,
    pub hardware_hash: String,
    pub benchmark_hash: String,
    pub confidence: f64,
}

impl ProofOfCompute {
    pub fn new(min_confidence: f64, max_deviation: f64) -> Self {
        Self {
            min_confidence,
            max_deviation,
        }
    }

    /// Verify a compute proof submission.
    /// Returns true if the proof is considered valid.
    pub fn verify_compute_proof(
        &self,
        proof: &ComputeProof,
        expected_tflops: f64,
    ) -> bool {
        // Check confidence threshold
        if proof.confidence < self.min_confidence {
            return false;
        }

        // Check measurement is within acceptable deviation
        if expected_tflops > 0.0 {
            let deviation = (proof.tflops_measured - expected_tflops).abs() / expected_tflops;
            if deviation > self.max_deviation {
                return false;
            }
        }

        // Verify hardware hash is non-empty
        if proof.hardware_hash.is_empty() {
            return false;
        }

        // Verify benchmark hash integrity
        self.verify_benchmark_hash(&proof.benchmark_hash)
    }

    fn verify_benchmark_hash(&self, hash: &str) -> bool {
        // Stub: verify the benchmark hash matches expected format
        !hash.is_empty()
    }

    /// Compute a hash of the oracle report for on-chain storage.
    pub fn hash_report(node_id: &str, tflops: f64, timestamp: u64) -> String {
        let mut hasher = Sha256::new();
        hasher.update(node_id.as_bytes());
        hasher.update(tflops.to_le_bytes());
        hasher.update(timestamp.to_le_bytes());
        hex::encode(hasher.finalize())
    }
}
