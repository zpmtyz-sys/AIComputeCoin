#!/bin/bash
set -euo pipefail

# ComputeCoin Security Scanning Script
# Runs dependency audits, container scanning, and secret detection

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORT_DIR="${PROJECT_ROOT}/security-reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track overall status
TOTAL_ISSUES=0
CRITICAL_ISSUES=0

mkdir -p "$REPORT_DIR"

echo "======================================"
echo " ComputeCoin Security Scan"
echo " $(date)"
echo "======================================"
echo ""

# Function to log results
log_result() {
    local service="$1"
    local status="$2"
    local details="$3"
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}[PASS]${NC} $service: $details"
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}[WARN]${NC} $service: $details"
    else
        echo -e "${RED}[FAIL]${NC} $service: $details"
        TOTAL_ISSUES=$((TOTAL_ISSUES + 1))
    fi
}

# 1. Rust Dependency Audit (matching-engine, blockchain)
echo "--- Rust Dependency Audit ---"
for pkg in matching-engine blockchain; do
    if [ -d "${PROJECT_ROOT}/packages/${pkg}" ]; then
        echo "Scanning packages/${pkg}..."
        if command -v cargo &> /dev/null; then
            if cargo audit --file "${PROJECT_ROOT}/packages/${pkg}/Cargo.lock" 2>&1 | tee "${REPORT_DIR}/cargo_audit_${pkg}_${TIMESTAMP}.txt"; then
                log_result "$pkg" "PASS" "No known vulnerabilities"
            else
                log_result "$pkg" "FAIL" "Vulnerabilities found (see report)"
                CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
            fi
        else
            log_result "$pkg" "WARN" "cargo-audit not installed, skipping"
        fi
    else
        log_result "$pkg" "WARN" "Directory not found, skipping"
    fi
done
echo ""

# 2. Node.js Dependency Audit (api-gateway, web-app)
echo "--- Node.js Dependency Audit ---"
for pkg in api-gateway web-app; do
    if [ -d "${PROJECT_ROOT}/packages/${pkg}" ]; then
        echo "Scanning packages/${pkg}..."
        if command -v npm &> /dev/null; then
            if (cd "${PROJECT_ROOT}/packages/${pkg}" && npm audit --audit-level=moderate 2>&1) | tee "${REPORT_DIR}/npm_audit_${pkg}_${TIMESTAMP}.txt"; then
                log_result "$pkg" "PASS" "No moderate+ vulnerabilities"
            else
                log_result "$pkg" "FAIL" "Vulnerabilities found (see report)"
            fi
        else
            log_result "$pkg" "WARN" "npm not installed, skipping"
        fi
    else
        log_result "$pkg" "WARN" "Directory not found, skipping"
    fi
done
echo ""

# 3. Go Vulnerability Check (trading-service, oracle-service)
echo "--- Go Vulnerability Check ---"
for pkg in trading-service oracle-service; do
    if [ -d "${PROJECT_ROOT}/packages/${pkg}" ]; then
        echo "Scanning packages/${pkg}..."
        if command -v govulncheck &> /dev/null; then
            if (cd "${PROJECT_ROOT}/packages/${pkg}" && govulncheck ./... 2>&1) | tee "${REPORT_DIR}/govuln_${pkg}_${TIMESTAMP}.txt"; then
                log_result "$pkg" "PASS" "No known vulnerabilities"
            else
                log_result "$pkg" "FAIL" "Vulnerabilities found (see report)"
            fi
        else
            log_result "$pkg" "WARN" "govulncheck not installed, skipping"
        fi
    else
        log_result "$pkg" "WARN" "Directory not found, skipping"
    fi
done
echo ""

# 4. Container Image Scanning
echo "--- Container Image Scanning ---"
IMAGES=(
    "computecoin/matching-engine:latest"
    "computecoin/api-gateway:latest"
    "computecoin/trading-service:latest"
    "computecoin/oracle-service:latest"
    "computecoin/blockchain:latest"
    "computecoin/web-app:latest"
)

if command -v trivy &> /dev/null; then
    for image in "${IMAGES[@]}"; do
        echo "Scanning ${image}..."
        if trivy image --severity HIGH,CRITICAL --exit-code 1 "$image" 2>&1 | tee "${REPORT_DIR}/trivy_${image//\//_}_${TIMESTAMP}.txt"; then
            log_result "$image" "PASS" "No high/critical vulnerabilities"
        else
            log_result "$image" "FAIL" "Vulnerabilities found"
            CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
        fi
    done
else
    log_result "container-scan" "WARN" "trivy not installed, skipping container scanning"
fi
echo ""

# 5. Hardcoded Secrets Detection
echo "--- Hardcoded Secrets Detection ---"
SECRETS_FOUND=0

# Patterns to search for
PATTERNS=(
    'password\s*=\s*["\x27][^"\x27]{8,}'
    'secret\s*=\s*["\x27][^"\x27]{8,}'
    'api[_-]?key\s*=\s*["\x27][^"\x27]{8,}'
    'private[_-]?key\s*=\s*["\x27][^"\x27]{8,}'
    'aws_access_key_id\s*=\s*[A-Z0-9]{20}'
    'AKIA[0-9A-Z]{16}'
    'ghp_[A-Za-z0-9]{36}'
    'sk-[A-Za-z0-9]{48}'
)

EXCLUDE_DIRS="node_modules|.git|target|vendor|.pnpm-store|security-reports|dist|.next"

for pattern in "${PATTERNS[@]}"; do
    matches=$(grep -rEi "$pattern" "$PROJECT_ROOT" \
        --include="*.ts" --include="*.js" --include="*.go" --include="*.rs" \
        --include="*.yaml" --include="*.yml" --include="*.json" --include="*.toml" \
        --include="*.env" --include="*.cfg" --include="*.conf" \
        | grep -Ev "($EXCLUDE_DIRS)" \
        | grep -Ev "(example|sample|test|mock|fixture|placeholder|CHANGE_ME|TODO)" \
        || true)
    if [ -n "$matches" ]; then
        echo "$matches" >> "${REPORT_DIR}/secrets_scan_${TIMESTAMP}.txt"
        SECRETS_FOUND=$((SECRETS_FOUND + 1))
    fi
done

if [ "$SECRETS_FOUND" -gt 0 ]; then
    log_result "secrets" "FAIL" "${SECRETS_FOUND} potential hardcoded secrets found (see report)"
    CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
else
    log_result "secrets" "PASS" "No hardcoded secrets detected"
fi
echo ""

# 6. Dockerfile Security Check
echo "--- Dockerfile Security Check ---"
find "$PROJECT_ROOT/packages" -name "Dockerfile" | while read -r dockerfile; do
    pkg_name=$(basename "$(dirname "$dockerfile")")

    # Check for running as root
    if ! grep -q "USER\|useradd\|adduser" "$dockerfile"; then
        log_result "$pkg_name Dockerfile" "WARN" "No non-root user defined"
    fi

    # Check for latest tag usage
    if grep -q "FROM.*:latest" "$dockerfile"; then
        log_result "$pkg_name Dockerfile" "WARN" "Using :latest tag (pin to specific version)"
    fi

    # Check for COPY vs ADD
    if grep -q "^ADD " "$dockerfile"; then
        log_result "$pkg_name Dockerfile" "WARN" "Using ADD instead of COPY (prefer COPY)"
    fi
done
echo ""

# Summary
echo "======================================"
echo " Security Scan Summary"
echo "======================================"
echo ""
echo "Total Issues: $TOTAL_ISSUES"
echo "Critical Issues: $CRITICAL_ISSUES"
echo "Reports saved to: $REPORT_DIR"
echo ""

if [ "$CRITICAL_ISSUES" -gt 0 ]; then
    echo -e "${RED}CRITICAL issues found! Immediate action required.${NC}"
    exit 1
elif [ "$TOTAL_ISSUES" -gt 0 ]; then
    echo -e "${YELLOW}Issues found. Review and resolve before deployment.${NC}"
    exit 1
else
    echo -e "${GREEN}All checks passed!${NC}"
    exit 0
fi
