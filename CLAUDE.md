# CLAUDE.md — Agent Operating Guide for ComputeCoin

> 读者:Claude Code、Hermes,以及任何在本仓库工作的自动化编码代理。
> 目标:让你在**不破坏主干稳定性**的前提下,并行地、增量地把每个模块做深。

---

## 0. 30 秒了解这个仓库

ComputeCoin = **算力计量(CU)→ 透明发币(CC)→ 现货交易所**,在 [sub2api](https://github.com/Wei-Shaw/sub2api) 之上做**合规二次开发**。
代码是一个 **模块化 FastAPI 单体**,位于 `backend/app/`,按业务域(domain)切分模块,每个模块自带 `service`(业务逻辑)+ `router`(HTTP)。

```
backend/app/
  core/        配置、DB、安全、错误、日志  —— 基础设施,改动需谨慎
  models/      SQLAlchemy ORM 模型(持久化层,单一事实来源)
  schemas/     Pydantic 出入参模型(API 契约)
  modules/
    identity/  用户与鉴权
    relay/     sub2api 集成 + 算力能力登记/计量
    compute/   CU 标准化 + Proof-of-Delivered-Compute
    ledger/    代币账本(排放/减半/双分录/国库)
    exchange/  撮合引擎 + 订单簿
    marketmaker/ 透明做市
  api/         路由聚合、依赖注入(deps)
  tokenomics.py 全部代币经济学“常量”(透明、写死)
  main.py      app 工厂 + 生命周期 + 静态仪表盘挂载
backend/tests/ pytest 单测(核心不变量)
web/           静态仪表盘
docs/          架构 / 代币经济学 / 合规 / 模块 / 工作流 backlog
```

---

## 1. 绝不可破坏的不变量(INVARIANTS)

任何改动**必须**保持下列性质。CI/测试会校验其中大部分。改这些等于改协议本身,需在 PR 描述里显式说明并更新 `docs/tokenomics.md`。

1. **账本守恒(双分录)**:每一笔 `LedgerEntry` 的借贷两侧金额相等;全网所有账户余额之和 = 已铸造总量 − 已销毁总量。永远不允许出现"凭空"加余额而无对应分录。
2. **排放上限**:`CC` 总供给硬顶 = `tokenomics.MAX_SUPPLY`。任何铸造前必须检查不会突破上限。
3. **减半确定性**:给定 `(已流逝周期数)`,单位时间排放量是**纯函数**、可复现、与调用时间无关(便于审计与回放)。
4. **创世分配透明**:`tokenomics.GENESIS_ALLOCATIONS` 是唯一创世来源,总和与各桶占比公开;团队桶必须带 `vesting`/`cliff`。不允许新增"隐藏桶"。
5. **撮合公平性**:撮合遵循 **价格优先、同价时间优先**;撮合后买卖双方资产 delta 之和为 0(扣除显式手续费),手续费进国库账户且有分录。
6. **金额用 Decimal/整数 base unit**,**禁止用 float** 表示余额、价格、数量。
7. **来源合法声明**:`relay` 模块登记算力能力时,`source_attestation`(来源合法声明)为必填,且默认 `verified=False` 直到通过验证。
8. **无后门**:不得新增绕过鉴权、绕过余额检查、可单方面增发或转移他人资产的接口。

---

## 2. 编码规范(让代码库保持稳定)

- **语言**:Python 3.11,全量 type hints。异步优先(`async def` + SQLAlchemy async session)。
- **风格**:`ruff` + `black`(行宽 100)。提交前跑 `make lint`。
- **模块边界**:模块之间**只能**通过 `service` 层的公开函数互相调用,**禁止**跨模块直接查对方的表/写对方的模型。需要跨域协作时,在 `service` 暴露明确函数。
- **数据库会话**:统一用 `core.database.get_session` 依赖注入;一个请求一个事务;service 函数接收 `session` 参数,不自行开全局连接。
- **钱/数量**:一律 `Decimal`。序列化为字符串返回前端,避免精度丢失。
- **错误**:抛 `core.errors.AppError` 的子类,由全局异常处理器转 HTTP;不要在 router 里散落 `raise HTTPException`。
- **新增模型**:加到 `models/`,并在 `docs/modules.md` 更新 ER 说明。MVP 用 `create_all` 建表;若改了已存在字段语义,在 PR 注明迁移影响(生产用 Alembic,见 backlog)。
- **测试**:核心逻辑(CU 公式、排放、账本、撮合)必须有单测;新功能至少 1 个 happy path + 1 个边界。
- **不要**为了通过测试而弱化不变量;不要删测试。

---

## 3. 如何安全地"改某个模块细节"

典型诉求:"我想改 CU 的计算系数 / 改减半周期 / 加一个交易对 / 换做市策略"。

- **改参数**:几乎所有可调参数都集中在 `backend/app/tokenomics.py` 和 `backend/app/core/config.py`。优先改这里,不要把魔法数字散落到 service。
- **改算法**:在对应模块的 `service.py` 内修改纯函数(如 `compute/service.py::tokens_to_cu`、`ledger/service.py::emission_for_period`),并同步更新单测与 `docs/`。
- **加模块**:复制 `modules/<existing>` 的结构(`__init__.py` / `service.py` / `router.py`),在 `api/__init__.py` 注册路由。保持"service 纯逻辑、router 薄"。

---

## 4. 工作流(给 /workflow 与并行子代理)

- 任务积压与依赖图在 [`docs/workflow.md`](docs/workflow.md)。每个任务都标了:`id`、依赖、影响文件、验收标准(DoD)。
- 并行规则:**同一文件同一时间只允许一个任务在改**;不同模块的任务可并行。`core/`、`models/`、`tokenomics.py` 属于"高风险共享区",改它们的任务必须串行且单独 PR。
- 每个任务完成的硬性 DoD:`make lint && make test` 全绿,且不违反第 1 节不变量。

---

## 5. 常用命令

```bash
make up        # docker compose 后台拉起
make down      # 停止
make logs      # 看 app 日志
make dev       # 本地(非 docker)起开发服务器(SQLite)
make test      # pytest
make lint      # ruff + black --check
make fmt       # 自动格式化
make secrets   # 生成随机密钥写入 .env
```

---

## 6. 合规红线(代理也要遵守)

- 不实现"隐藏增发 / 暗箱独占收益 / 误导性收益承诺"。
- 不实现"批量汇聚并转售他人 API 凭证";`relay` 只登记**贡献者自有**能力与**交付计量**。
- 涉及证券/资金合规的判断,代码里只放**钩子与 TODO**,正文注明"需法律意见",不假装给出法律结论。

详见 [`docs/compliance.md`](docs/compliance.md)。
