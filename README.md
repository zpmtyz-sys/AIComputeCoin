# ComputeCoin (算力币) — 合规版 / Compliant Edition

> **一句话**:把"真实交付的 AI 算力"标准化计量、铸造成代币(CC),并提供一个现货交易所。供给方贡献**自己合法拥有**的算力/中转能力赚 CC,消费方花 CC 调用算力,交易所提供流动性。
>
> **One line**: Standardize *actually delivered* AI compute into a transparent unit (CU), mint it into a token (CC), and run a spot exchange on top. Suppliers contribute compute capacity they **legitimately own** and earn CC; consumers spend CC to access pooled compute; the exchange provides liquidity.

本仓库是 **ComputeCoin** 的合规版实现:一个**模块化 FastAPI 单体**(modular monolith),与开源项目 [sub2api](https://github.com/Wei-Shaw/sub2api) 以"独立服务 + HTTP 适配器"的方式集成,而**不是** fork/改它的源码(原因见 [`docs/compliance.md`](docs/compliance.md))。

---

## ⚠️ 在你继续之前必须读的 3 条合规约束

本项目刻意**不**实现任何"暗中独占收益 / 误导散户 / 大规模汇聚他人凭证转售"的设计。它被设计成可被机构尽调的形态:

1. **代币分配完全透明且写死在代码里**。创世分配、团队归属(vesting/cliff)、国库多签、排放/减半曲线全部公开,见 [`docs/tokenomics.md`](docs/tokenomics.md)。没有"隐藏增发"开关。
2. **供给方只能贡献自己合法拥有的算力/账号能力**。注册流程强制勾选来源合法声明;系统记录的是"交付的算力计量",不池化、不转卖他人凭证。上游 ToS 风险免责声明继承自 sub2api,见 [`docs/compliance.md`](docs/compliance.md)。
3. **交易/出入金接口预留 KYC/AML 钩子**,代币定位为"功能型(utility)"。是否构成证券由**你聘请的证券律师**判定 —— 代码不替你做法律判断,只把结构做成"可合规"的样子。

> 给投资人:商业与融资联系 **水镜先生 / Mr. Shuijing — zpmtyz@gmail.com**。本仓库是技术与架构实现,不构成证券要约或投资建议。

---

## 端口约定(重要)

你本地 **8080 已被 sub2api 占用**。本项目默认使用一组**不冲突**的端口,全部可在 `.env` 改:

| 服务 | 容器内端口 | 宿主机端口(默认) | 说明 |
|------|-----------|------------------|------|
| ComputeCoin App(API + 仪表盘) | 8000 | **8788** | 主入口,浏览器访问 `http://localhost:8788` |
| PostgreSQL | 5432 | **5433** | 避开本机默认 5432 |
| Redis | 6379 | **6380** | 避开本机默认 6379 |

> 想换端口?改 `.env` 里的 `APP_HOST_PORT` / `POSTGRES_HOST_PORT` / `REDIS_HOST_PORT` 即可,无需改代码。

---

## 60 秒拉起(本地 Docker 后台部署)

```bash
# 1. 准备环境变量(会生成随机密钥)
cp .env.example .env
make secrets   # 自动把随机 JWT_SECRET 等写进 .env(或手动编辑)

# 2. 后台启动全部服务(postgres + redis + app)
make up        # 等价于 docker compose up -d --build

# 3. 看日志 / 健康检查
make logs
curl http://localhost:8788/health

# 4. 打开浏览器
#    仪表盘:   http://localhost:8788
#    API 文档:  http://localhost:8788/docs   (OpenAPI / Swagger)

# 5. 停止
make down
```

首次启动会自动:建表 → 写入**透明创世分配**(genesis allocation)→ 启动排放调度器 → 创建管理员(读 `.env` 的 `ADMIN_EMAIL/ADMIN_PASSWORD`)。

---

## 这个项目由什么组成(模块地图)

```
ComputeCoin App (FastAPI, 模块化单体)
├── identity     用户注册/登录/JWT/角色/KYC 状态(对标 sub2api 注册逻辑)
├── relay        sub2api 集成适配器:登记“我合法拥有的算力/中转能力”,计量交付
├── compute      算力预言机:CU 标准化公式 + Proof-of-Delivered-Compute 验证
├── ledger       代币账本:透明排放/减半、双分录、国库、创世分配
├── exchange     现货交易所:价格-时间优先撮合引擎、订单簿、成交、手续费
└── marketmaker  透明做市模块:公开的报价/库存策略(非暗箱)
```

每个模块**职责单一、接口清晰**,可以单独改而不破坏整体。详见 [`docs/modules.md`](docs/modules.md)。

---

## 给 AI 编码代理(Claude Code / Hermes)的入口

- 先读 [`CLAUDE.md`](CLAUDE.md) / [`AGENTS.md`](AGENTS.md):全局约定、不可破坏的不变量(invariants)、编码规范。
- 再读 [`docs/workflow.md`](docs/workflow.md):一个**可并行执行的任务积压(task backlog)与依赖图**,直接喂给 `/workflow` 让大量子智能体并行开发,且保证主干稳定。
- 架构推演见 [`docs/architecture.md`](docs/architecture.md)。

---

## 技术栈

| 层 | 选型 | 理由 |
|----|------|------|
| 后端 | Python 3.11 + FastAPI + SQLAlchemy(async) | 可读性极高、AI agent 友好、迭代快、模块化清晰 |
| 数据库 | PostgreSQL(生产)/ SQLite(本地测试) | 同一份 ORM 代码两边都能跑 |
| 缓存/队列 | Redis | 会话、撮合热数据、排放任务锁 |
| 前端 | 轻量仪表盘(Tailwind CDN + 原生 JS,由后端托管) | 一个容器即可跑起来,零额外构建 |
| 部署 | Docker Compose | `make up` 一键后台拉起 |

> 这是一个**可运行的合规 MVP + 可被大规模并行扩展的骨架**,不是最终生产系统。生产化路线(Rust 撮合、链上结算、L2、独立微服务)写在 [`docs/architecture.md`](docs/architecture.md) 与 `docs/workflow.md` 的 backlog 里。

---

## 许可

当前为 [MIT](LICENSE)(允许后续版本闭源,由你掌控收益)。与 sub2api(LGPL-3.0)的隔离方式见 [`docs/compliance.md`](docs/compliance.md)。
