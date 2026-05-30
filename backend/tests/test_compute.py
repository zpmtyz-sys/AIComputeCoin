"""CU standardization formula tests."""

from decimal import Decimal

import pytest

from app import tokenomics as tk
from app.core.errors import BadRequest
from app.modules.compute import service as compute


def test_tokens_to_cu_default():
    # 1000 tokens * default coeff (per 1000) = coeff
    assert compute.tokens_to_cu("unknown-model", Decimal("1000")) == tk.TOKEN_TO_CU["default"]


def test_tokens_to_cu_known_model():
    coeff = tk.TOKEN_TO_CU["claude-3.5-sonnet"]
    assert compute.tokens_to_cu("claude-3.5-sonnet", Decimal("2000")) == tk.quantize(coeff * 2)


def test_tokens_to_cu_zero():
    assert compute.tokens_to_cu("gpt-4o", Decimal("0")) == Decimal("0")


def test_gpu_hours_to_cu_h100():
    # 1 H100 for 1 hour ~ its FP16 TFLOPS, with default factors = 1
    assert compute.gpu_hours_to_cu("H100", Decimal("1")) == tk.GPU_TFLOPS_FP16["H100"]


def test_gpu_hours_nvlink_factor():
    base = compute.gpu_hours_to_cu("A100", Decimal("1"))
    nv = compute.gpu_hours_to_cu("A100", Decimal("1"), interconnect="nvlink")
    assert nv == tk.quantize(base * tk.INTERCONNECT_FACTOR["nvlink"])


def test_gpu_hours_unknown_model_raises():
    with pytest.raises(BadRequest):
        compute.gpu_hours_to_cu("NOT_A_GPU", Decimal("1"))
