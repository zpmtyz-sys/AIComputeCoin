# 代币经济学 / Tokenomics(透明、可审计)

> 本文是**人类可读**版本。**机器可执行**的唯一事实来源是 [`backend/app/tokenomics.py`](../backend/app/tokenomics.py)。
> 两者若不一致,以代码为准,并应提 PR 修正本文档。改动代币参数 = 改协议,必须走 `CLAUDE.md` 的"高风险共享区"流程。

---

## 1. 代币基本面

| 项目 | 值 | 说明 |
|------|----|----|
| 符号 | `CC` | ComputeCoin |
| 精度 | 8 位小数 | 内部用 `Decimal`,最小单位 1e-8 CC |
| 最大供给(硬顶) | **2,100,000,000 CC**(21 亿) | `MAX_SUPPLY`,任何铸造都不得突破 |
| 创世分配 | 占硬顶的一部分,见下表 | 启动时一次性入账,带 vesting |
| 排放(挖矿) | 剩余部分按"减半曲线"逐步铸造给供给方 | Proof-of-Delivered-Compute |

---

## 2. 创世分配(Genesis Allocation,透明)

启动时铸造 `GENESIS_SUPPLY`,**全部公开**,按桶分配。团队/顾问桶**强制带锁定期(cliff)与线性释放(vesting)**。

| 桶 | 占创世比例 | 锁定/释放 | 用途 |
|----|-----------|----------|------|
| 生态与流动性 Ecosystem & Liquidity | 30% | 部分即时(做市/上市流动性) | 交易所流动性、做市启动 |
| 供给激励储备 Supplier Incentive Reserve | 25% | 随排放释放 | 早期供给方额外激励 |
| 团队 Team | 18% | **12 个月 cliff + 36 个月线性** | 团队与创始人(公开披露) |
| 投资人 Investors(VC) | 15% | 6 个月 cliff + 24 个月线性 | 正规股权/代币融资 |
| 国库 Treasury(治理) | 10% | 多签托管 | 治理支配,经提案使用 |
| 社区与空投 Community | 2% | 分批 | 早期用户/贡献者 |

> 这里没有"创始人暗中独占"的桶。创始人的代币 = `Team` 桶里**公开**的、带 12 个月悬崖期的部分。投资人代币是正规融资稀释。具体百分比在 `tokenomics.py` 的 `GENESIS_ALLOCATIONS`,可在融资条款敲定后调整,但**必须保持公开**。

---

## 3. 排放与减半(Emission & Halving,确定性纯函数)

排放部分 = `MAX_SUPPLY − GENESIS_SUPPLY`,按"周期减半"铸造给当期供给方:

```
INITIAL_DAILY_EMISSION   每日初始排放量(CC/天)
HALVING_PERIOD_DAYS      减半周期(天),默认 540(约 18 个月)
emission_per_day(period) = INITIAL_DAILY_EMISSION / 2**period
其中 period = floor(已流逝天数 / HALVING_PERIOD_DAYS)
```

当期排放在供给方之间的分配:

```
某供给方当期获得 = 当期总排放 × (该供给方本期 verified_CU / 全网本期 verified_CU)
                              × availability_factor
```

**不变量**:
- `emission_per_day` 是纯函数,给定 period 必然得到同一结果(可回放、可审计)。
- 累计铸造(创世 + 排放)永不超过 `MAX_SUPPLY`;接近上限时按剩余额度截断。

---

## 4. 费用与国库(Fees & Treasury)

| 费用 | 默认 | 去向 |
|------|------|------|
| Taker 手续费 | 0.10% | 国库账户(`treasury`) |
| Maker 手续费 | 0.02% | 国库账户 |
| 提现费 | 可配 | 国库账户 |

- 所有手续费通过**账本分录**进入 `treasury` 系统账户,余额公开可查(`GET /ledger/accounts/treasury`)。
- 国库的支出由治理提案决定(MVP 仅记录,治理模块在 backlog)。

---

## 5. 可调参数一览(都在 `tokenomics.py`)

| 常量 | 含义 |
|------|------|
| `DECIMALS` | 代币精度 |
| `MAX_SUPPLY` | 硬顶 |
| `GENESIS_SUPPLY` | 创世铸造量 |
| `GENESIS_ALLOCATIONS` | 创世各桶(名称/比例/vesting/cliff) |
| `INITIAL_DAILY_EMISSION` | 每日初始排放 |
| `HALVING_PERIOD_DAYS` | 减半周期 |
| `TOKEN_TO_CU` | 各模型 token→CU 折算 |
| `GPU_TFLOPS_FP16` | 常见 GPU 的 FP16 算力基准 |
| `TAKER_FEE` / `MAKER_FEE` | 交易手续费 |

> 改这些 → 必跑 `make test`(排放/账本/CU 单测会校验不变量)→ 同步更新本文档。
