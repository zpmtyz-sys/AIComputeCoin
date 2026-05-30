# ComputeCoin 产品需求文档 (PRD)

> **版本**: v1.0.0
> **作者**: 水镜先生 (zpmtyz@gmail.com)
> **最后更新**: 2026-05-30
> **状态**: MVP 已实现,持续迭代

---

## 1. 产品定位

### 1.1 一句话描述

**ComputeCoin = 把真实交付的 AI 算力标准化计量(CU)→ 铸造为透明代币(CC)→ 在现货交易所上交易。**

### 1.2 目标用户

| 角色 | 描述 | 核心需求 |
|------|------|---------|
| **算力供给方 (Supplier)** | 拥有 GPU/TPU 或 sub2api 网关的个人/团队 | 把闲置算力变现为可交易代币 |
| **算力消费方 (Consumer)** | AI 开发者、科研团队、渲染公司 | 用 CC 按需购买算力,成本透明 |
| **交易者 (Trader)** | 散户/量化/机构 | 在交易所买卖 CC 赚价差 |
| **投资人 (Investor)** | VC / 天使 | 透明、可审计、合规的技术架构与代币分配 |
| **AI 编码代理 (Agent)** | Claude Code / Hermes / Cursor | 能读懂架构、并行开发、不破坏主干 |

### 1.3 与竞品的差异

```
┌─────────────────────────────────────────────────────────────────────┐
│                    竞品对比矩阵                                       │
├──────────────┬──────────┬──────────┬──────────┬─────────────────────┤
│              │ 算力租赁  │ 标准化计量 │ 代币铸造  │ 交易所 + 衍生品     │
├──────────────┼──────────┼──────────┼──────────┼─────────────────────┤
│ Render       │    ✓     │    ✗     │    ✓     │         ✗           │
│ Akash        │    ✓     │    ✗     │    ✓     │         ✗           │
│ io.net       │    ✓     │    △     │    ✓     │         ✗           │
│ Golem        │    ✓     │    ✗     │    ✓     │         ✗           │
│ ComputeCoin  │    ✓     │    ✓     │    ✓     │         ✓           │
└──────────────┴──────────┴──────────┴──────────┴─────────────────────┘

关键差异: 他们做"算力淘宝", 我们做"算力华尔街"(金融化)
```

---

## 2. 功能需求(Feature Requirements)

### 2.1 用户身份与认证 (identity)

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-ID-01 | 邮箱注册 | P0 | ✅ Done | 邮箱唯一, 密码 bcrypt 哈希, 注册即创建账本账户 |
| F-ID-02 | JWT 登录 | P0 | ✅ Done | 返回 access_token, 24h 过期 |
| F-ID-03 | 角色区分 | P0 | ✅ Done | user / supplier / admin 三种角色 |
| F-ID-04 | KYC 状态钩子 | P0 | ✅ Done | kyc_status 字段; `require_kyc` 依赖可拦截敏感接口 |
| F-ID-05 | KYC 第三方接入 | P1 | 🔲 Backlog | 对接 Sumsub/Onfido, 真正校验身份 |
| F-ID-06 | sub2api SSO | P2 | 🔲 Backlog | 以 sub2api 为身份源同步用户 |
| F-ID-07 | 2FA (TOTP) | P1 | 🔲 Backlog | Google Authenticator 兼容 |
| F-ID-08 | 登录风控 | P2 | 🔲 Backlog | 失败锁定, 设备指纹, IP 限速 |

### 2.2 算力能力登记与计量 (relay)

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-RL-01 | 登记 sub2api 网关 | P0 | ✅ Done | kind=sub2api, endpoint 必填, source_attestation 必填 |
| F-RL-02 | 登记 GPU 节点 | P0 | ✅ Done | kind=gpu, gpu_model 必填, 自动关联 TFLOPS |
| F-RL-03 | 来源合法声明 | P0 | ✅ Done | source_attestation ≥10 字, 不填则拒绝 |
| F-RL-04 | 记录 token 交付 | P0 | ✅ Done | model + delivered_tokens → 自动折算 CU |
| F-RL-05 | 记录 GPU-hours 交付 | P0 | ✅ Done | gpu_model + hours → 自动折算 CU |
| F-RL-06 | 真实对接 sub2api usage API | P1 | 🔲 Backlog | 调用 sub2api 的计费/用量接口拉取真实数据 |
| F-RL-07 | 来源审核流程 | P1 | 🔲 Backlog | admin 审核 source, 人工 verify |
| F-RL-08 | 批量导入能力 | P2 | 🔲 Backlog | CSV/API 批量登记 |

### 2.3 算力预言机与验证 (compute)

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-CP-01 | CU 标准化折算(token) | P0 | ✅ Done | tokens_to_cu 纯函数, 系数在 tokenomics.py |
| F-CP-02 | CU 标准化折算(GPU-hours) | P0 | ✅ Done | gpu_hours_to_cu 纯函数, 含修正系数 |
| F-CP-03 | CU 估算器 API | P0 | ✅ Done | GET /compute/cu/quote 公开接口 |
| F-CP-04 | MVP 验证(自动通过) | P0 | ✅ Done | POST /compute/verify/{id} 标记 verified=True |
| F-CP-05 | 基准挑战验证 | P1 | 🔲 Backlog | 随机下发 benchmark 任务, 校验完成时间 |
| F-CP-06 | 跨节点交叉验证 | P2 | 🔲 Backlog | 多个节点互相验证 |
| F-CP-07 | TEE 远程证明 | P2 | 🔲 Backlog | GPU 序列号 + 固件哈希上链 |
| F-CP-08 | 质押与惩罚(slashing) | P1 | 🔲 Backlog | 作弊扣保证金, 信誉评分降级 |

### 2.4 代币账本 (ledger)

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-LG-01 | 双分录账本 | P0 | ✅ Done | 每笔分录借贷相等; 余额总和恒为 0 |
| F-LG-02 | 创世分配 | P0 | ✅ Done | 启动时铸造 210M CC, 分 6 桶, 一次性 |
| F-LG-03 | 供给上限强制 | P0 | ✅ Done | 任何铸造超过 2.1B 抛 SupplyCapExceeded |
| F-LG-04 | 排放/减半(确定性) | P0 | ✅ Done | daily_emission 是纯函数; 540 天减半 |
| F-LG-05 | 排放调度器 | P0 | ✅ Done | 后台循环, 为已完成天数按 CU 占比铸造 CC |
| F-LG-06 | 转账 | P0 | ✅ Done | user→user, 余额不足拒绝 |
| F-LG-07 | 供给信息公开 | P0 | ✅ Done | GET /ledger/supply 返回总量/已铸造/剩余 |
| F-LG-08 | 测试网水龙头 | P0 | ✅ Done | 非生产环境可 mint USDC (不可 mint CC) |
| F-LG-09 | Alembic 迁移 | P1 | 🔲 Backlog | 替换 create_all, 支持 schema 演进 |
| F-LG-10 | vesting/cliff 释放调度 | P1 | 🔲 Backlog | 按 GENESIS_ALLOCATIONS 参数线性释放 |
| F-LG-11 | 对账任务 | P1 | 🔲 Backlog | 周期性校验 conservation_ok, 不一致告警 |
| F-LG-12 | 出入金 | P1 | 🔲 Backlog | 充值/提现接口 + KYC/AML 拦截 |

### 2.5 现货交易所 (exchange)

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-EX-01 | 限价单 | P0 | ✅ Done | 价格-时间优先撮合, 余额冻结校验 |
| F-EX-02 | 市价单 | P0 | ✅ Done | 吃掉对手盘直到填满或资金用尽 |
| F-EX-03 | 撤单 | P0 | ✅ Done | 只能撤自己的, open/partial 才可撤 |
| F-EX-04 | 订单簿 | P0 | ✅ Done | GET /exchange/orderbook, 聚合价位深度 |
| F-EX-05 | 成交历史 | P0 | ✅ Done | GET /exchange/trades, 最近 50 笔 |
| F-EX-06 | 手续费(taker/maker) | P0 | ✅ Done | 0.10%/0.02%, 进国库, 有分录 |
| F-EX-07 | 自成交防护 | P0 | ✅ Done | 同一用户的订单不互相成交 |
| F-EX-08 | WebSocket 实时推送 | P1 | 🔲 Backlog | 行情/订单簿/个人订单状态 |
| F-EX-09 | Rust 撮合引擎 | P2 | 🔲 Backlog | 微秒级延迟, gRPC 对接 |
| F-EX-10 | 多交易对 | P1 | 🔲 Backlog | CC/BTC, CC/ETH 等 |
| F-EX-11 | 风控前置 | P1 | 🔲 Backlog | 持仓上限, 价格偏离检查, 异常交易标记 |

### 2.6 做市模块 (marketmaker)

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-MM-01 | 透明报价 | P0 | ✅ Done | 公开 bid/ask/mid/spread, GET /mm/quotes |
| F-MM-02 | 自动挂单 | P0 | ✅ Done | 启动时 + 可调度刷新 |
| F-MM-03 | Avellaneda-Stoikov 策略 | P1 | 🔲 Backlog | 库存风险敞口控制 |
| F-MM-04 | 做市账户资金透明 | P0 | ✅ Done | 走正常账本分录, 可查余额 |

### 2.7 前端仪表盘

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-UI-01 | 注册/登录 | P0 | ✅ Done | 表单 + JWT 存 localStorage |
| F-UI-02 | 概览(供给信息/余额/公开账户) | P0 | ✅ Done | |
| F-UI-03 | 交易市场(订单簿/下单/成交) | P0 | ✅ Done | |
| F-UI-04 | 贡献算力(登记/验证/CU 估算) | P0 | ✅ Done | |
| F-UI-05 | 账户信息 | P0 | ✅ Done | |
| F-UI-06 | 测试 USDC 水龙头 | P0 | ✅ Done | |
| F-UI-07 | Next.js + TradingView 专业界面 | P2 | 🔲 Backlog | K 线, 深度图, 专业交易 |

### 2.8 合规与安全

| ID | 功能 | 优先级 | 状态 | 验收标准 |
|----|------|--------|------|---------|
| F-SC-01 | LGPL 隔离(sub2api) | P0 | ✅ Done | 进程隔离 + HTTP 适配器, MIT 不受传染 |
| F-SC-02 | 透明代币分配 | P0 | ✅ Done | 写死在 tokenomics.py, 公开文档 |
| F-SC-03 | KYC 钩子 | P0 | ✅ Done | require_kyc 依赖, 配置开关 |
| F-SC-04 | 代币功能型设计 | P0 | ✅ Done | 用 CC 消费算力, 不承诺收益 |
| F-SC-05 | 账本审计追踪 | P0 | ✅ Done | 每笔流动有 LedgerEntry + ref |
| F-SC-06 | 牌照路径文档 | P1 | 🔲 Backlog | MAS / VARA / FINMA 选择 |
| F-SC-07 | 冷热钱包分离 | P2 | 🔲 Backlog | |
| F-SC-08 | 多签治理 | P2 | 🔲 Backlog | |
| F-SC-09 | 安全审计 | P2 | 🔲 Backlog | 第三方渗透测试 + 账本/合约审计 |

---

## 3. 非功能需求(Non-Functional Requirements)

### 3.1 性能

| 指标 | MVP 目标 | 生产目标 |
|------|---------|---------|
| 撮合延迟 | < 50ms (Python, DB 事务内) | < 10us (Rust 内存引擎) |
| API 响应 | < 200ms (P95) | < 50ms (P95) |
| 并发用户 | 100 同时在线 | 10,000+ |
| 每日交易量 | $100K | $100M |

### 3.2 可用性

| 指标 | 目标 |
|------|------|
| 服务 SLA | 99.9% (生产) |
| 数据持久化 | PostgreSQL WAL + 定期备份 |
| 容灾 | Docker Compose 单机(MVP) → K8s 多区域(生产) |

### 3.3 安全

| 维度 | 实现 |
|------|------|
| 认证 | JWT (HS256), bcrypt 密码哈希 |
| 授权 | 角色 + KYC 状态钩子 |
| 传输 | HTTPS (反向代理层, backlog) |
| 存储 | 密钥在 .env, 不进 git |
| 审计 | 双分录账本 = 天然审计追踪 |

### 3.4 可维护性

| 维度 | 实现 |
|------|------|
| 模块化 | 6 个独立模块, service 层互调 |
| 测试 | 34+ 自动化测试, CI 可集成 |
| 文档 | CLAUDE.md + 8 份 docs/ 文档 |
| 代码规范 | ruff + black 强制 |
| Agent 友好 | workflow.md 任务 backlog, 可并行 |

### 3.5 可扩展性

| 维度 | 当前 | 路线 |
|------|------|------|
| 数据库 | 单 PostgreSQL | 读写分离 → 分库 |
| 撮合 | Python 单进程 | Rust 服务 + gRPC |
| 缓存 | Redis 单实例 | Redis Cluster |
| 部署 | Docker Compose | K8s + Helm |
| 链 | 链下账本 | L1(Cosmos/Substrate) + L2 Rollup |

---

## 4. 数据模型(Entity Relationship)

```
┌────────────┐       ┌──────────────────┐       ┌────────────────┐
│   User     │       │ CapacitySource   │       │ DeliveryRecord │
├────────────┤       ├──────────────────┤       ├────────────────┤
│ id (PK)    │──┐    │ id (PK)          │──┐    │ id (PK)        │
│ email      │  │    │ user_id (FK)     │  │    │ source_id (FK) │
│ password_  │  └───►│ kind             │  └───►│ model          │
│   hash     │       │ name             │       │ delivered_     │
│ role       │       │ endpoint         │       │   tokens       │
│ kyc_status │       │ gpu_model        │       │ gpu_hours      │
│ auth_      │       │ source_          │       │ cu             │
│   provider │       │   attestation    │       │ verified       │
│ external_id│       │ verified         │       │ period         │
│ created_at │       │ reputation       │       │ created_at     │
└────────────┘       │ created_at       │       └────────────────┘
      │              └──────────────────┘
      │
      │       ┌────────────┐       ┌──────────────────┐
      │       │  Account   │       │   LedgerEntry    │
      │       ├────────────┤       ├──────────────────┤
      │       │ id (PK)    │◄──┐   │ id (PK)          │
      └──────►│ ref        │   │   │ ref              │
              │ owner_type │   ├───│ debit_account_id │
              │ kind       │   ├───│ credit_account_id│
              │ created_at │   │   │ asset            │
              └────────────┘   │   │ amount           │
                    │          │   │ kind             │
                    ▼          │   │ memo             │
              ┌────────────┐   │   │ created_at       │
              │  Balance   │   │   └──────────────────┘
              ├────────────┤   │
              │ id (PK)    │   │   ┌──────────────────┐
              │ account_id │───┘   │ EmissionEpoch    │
              │ asset      │       ├──────────────────┤
              │ amount     │       │ day_index (PK)   │
              └────────────┘       │ halving_period   │
                                   │ minted           │
              ┌────────────┐       │ processed_at     │
              │   Order    │       └──────────────────┘
              ├────────────┤
              │ id (PK)    │       ┌──────────────────┐
              │ user_id    │       │     Trade        │
              │ pair       │       ├──────────────────┤
              │ side       │       │ id (PK)          │
              │ type       │◄──────│ taker_order_id   │
              │ price      │◄──────│ maker_order_id   │
              │ qty        │       │ pair             │
              │ filled     │       │ price            │
              │ status     │       │ qty              │
              │ created_at │       │ taker/maker_fee  │
              └────────────┘       │ created_at       │
                                   └──────────────────┘
```

---

## 5. API 接口清单

### 5.1 公开接口(无需登录)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/docs` | OpenAPI / Swagger UI |
| GET | `/api/ledger/supply` | 代币供给信息(总量/已铸造/剩余) |
| GET | `/api/exchange/orderbook` | 订单簿 |
| GET | `/api/exchange/trades` | 最近成交 |
| GET | `/api/mm/quotes` | 做市报价(透明) |
| GET | `/api/compute/cu/quote` | CU 估算器 |
| GET | `/api/relay/sub2api/health` | sub2api 网关健康检查 |

### 5.2 需要登录(Bearer JWT)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录 |
| GET | `/api/auth/me` | 当前用户 |
| GET | `/api/ledger/balance` | 我的余额 |
| GET | `/api/ledger/accounts/{ref}` | 查看账户(公开账户或自己的) |
| POST | `/api/ledger/transfer` | 转账 |
| POST | `/api/ledger/faucet` | 领取测试币(非生产) |
| POST | `/api/relay/sources` | 登记算力能力 |
| GET | `/api/relay/sources` | 我的算力源 |
| POST | `/api/relay/sources/{id}/delivery` | 记录交付 |
| GET | `/api/relay/sources/{id}/deliveries` | 交付历史 |
| POST | `/api/compute/verify/{source_id}` | 验证算力源 |
| POST | `/api/exchange/orders` | 下单 |
| GET | `/api/exchange/orders` | 我的订单 |
| DELETE | `/api/exchange/orders/{id}` | 撤单 |

---

## 6. 部署需求

| 项目 | 要求 |
|------|------|
| 运行环境 | Docker 20.10+ / Docker Compose v2+ |
| 最低硬件 | 2 vCPU, 4GB RAM, 20GB SSD |
| 网络 | 端口 8788(app), 5433(postgres), 6380(redis) |
| 操作系统 | Linux / macOS (Windows WSL2 亦可) |
| 一键部署 | `make secrets && make up` |

---

## 7. 里程碑(Milestones)

| 阶段 | 内容 | 状态 |
|------|------|------|
| **M0: 骨架** | 仓库结构 + 文档 + CLAUDE.md | ✅ 完成 |
| **M1: 核心闭环** | identity + relay + compute + ledger + exchange + mm + 测试 + Docker | ✅ 完成 |
| **M2: 加固** | 真实 sub2api 对接 + Alembic + vesting + 对账 | 🔲 下一步 |
| **M3: 体验** | WebSocket + Next.js 前端 + 多交易对 + 2FA | 🔲 计划中 |
| **M4: 合规** | KYC 接入 + 法律文件 + 牌照申请 | 🔲 计划中 |
| **M5: 规模化** | Rust 撮合 + K8s + 链上结算 + 衍生品 | 🔲 远期 |

---

## 8. 验收标准(Definition of Done)

每个功能的 DoD:

1. 实现代码位于正确的模块内(`modules/<x>/service.py` + `router.py`)
2. 不违反 `CLAUDE.md` 的 8 条不变量
3. `make lint` (ruff + black) 全绿
4. `make test` (pytest) 全绿,新功能至少 1 happy path + 1 边界测试
5. 涉及参数变更时同步更新 `docs/tokenomics.md` 或相关文档
6. 涉及接口变更时 Swagger 自动反映(FastAPI 自动生成)

---

## 9. 风险登记(Risk Register)

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|---------|
| 被认定为证券 | 中 | 高 | 功能型设计 + 先拿法律意见 + 选友好辖区 |
| 上游 ToS 封禁 | 中 | 中 | source_attestation + 免责声明 + 多元化供给 |
| 流动性冷启动 | 高 | 中 | 做市模块 + 交易挖矿(backlog) + 种子流动性 |
| 算力验证作弊 | 中 | 中 | 质押 + 随机挑战 + 信誉 + 经济学惩罚 |
| 黑客攻击 | 中 | 高 | 冷热分离 + 多签 + 审计 + 保险基金(backlog) |
| 单点创始人风险 | 低 | 高 | 透明治理 + 多签国库 + 代码公开 |

---

## 10. 术语表(Glossary)

| 术语 | 含义 |
|------|------|
| **CC** | ComputeCoin, 本项目代币符号 |
| **CU** | Compute Unit, 标准化算力单位 (1 TFLOPS * 1h FP16) |
| **PoDC** | Proof-of-Delivered-Compute, 交付证明 |
| **Genesis** | 创世分配, 启动时一次性铸造的代币 |
| **Halving** | 减半, 每 540 天排放速率减半 |
| **Double-entry** | 双分录, 每笔操作同时借记和贷记等额 |
| **Treasury** | 国库, 手续费归集账户 |
| **sub2api** | 开源 AI API 网关, 本项目通过 HTTP 集成 |
| **Vesting** | 归属/释放期, 代币按时间线性解锁 |
| **Cliff** | 悬崖期, 锁定期内完全不可释放 |
