# 工作流与任务积压 / Workflow & Backlog(给 /workflow 与并行子代理)

本文件是把 ComputeCoin 从"合规 MVP"推进到"可上线"的**可并行任务图**。
直接喂给 Claude Code 的 `/workflow`,让多个子代理并行认领。**铁律见 `CLAUDE.md` 第 1、4 节**。

---

## 0. 并行规则(必须遵守)

- **高风险共享区(串行,单独 PR)**:`backend/app/core/**`、`backend/app/models/**`、`backend/app/tokenomics.py`。同一时间最多一个任务在改。
- **可并行区(按模块)**:`backend/app/modules/<x>/**`、`backend/tests/<x>/**`、`web/**`、`docs/**`(除非任务显式改共享区)。
- **每个任务 DoD**:`make lint && make test` 全绿;不违反任何不变量;若改了协议参数/接口,更新对应 `docs/`。
- **一个文件同一时刻只允许一个在途任务**。任务领取时在 PR 标题标注 `[<task-id>]`。

---

## 1. Epic 总览与依赖

```
E0 基础设施(已在 MVP 完成,后续仅加固)
   └─► E1 身份与合规钩子
   └─► E2 算力计量与验证(compute/relay)
            └─► E3 账本与排放(ledger)
                     └─► E4 交易所与做市(exchange/mm)
   E5 前端与机构 API   (依赖 E1..E4 的接口)
   E6 链上结算与治理   (依赖 E3)
   E7 衍生品           (依赖 E4)
   E8 运维/安全/合规上线(贯穿)
```

---

## 2. 任务积压(可认领)

### E1 身份与合规
- **T1.1** 接入第三方 KYC 供应商(Sumsub/Onfido)到 `identity` 的 KYC 钩子。影响:`modules/identity/*`, `api/deps.py`。DoD:`require_kyc` 真正拦截未过 KYC 的敏感接口。
- **T1.2** sub2api SSO 模式 B:用 sub2api 作身份源同步用户。影响:`modules/relay/sub2api_client.py`, `modules/identity/service.py`。
- **T1.3** 2FA(TOTP)+ 登录风控(失败锁定、设备指纹)。

### E2 算力计量与验证
- **T2.1** 真实对接 `sub2api.fetch_usage`,把交付 token 周期入账。影响:`modules/relay/sub2api_client.py`, `modules/relay/service.py`。
- **T2.2** PoDC 基准挑战:随机下发 benchmark,校验完成时间;跨节点交叉验证。影响:`modules/compute/service.py`。
- **T2.3** 质押与惩罚(slashing):作弊扣保证金,纳入信誉评分。影响:`modules/compute/*`, `modules/ledger/*`(共享区,串行)。
- **T2.4** TEE 远程证明对接(硬件指纹/固件哈希)。研究 + 接口。

### E3 账本与排放
- **T3.1** Alembic 迁移体系替换 `create_all`(生产必需)。影响:`backend/migrations/**`, `core/database.py`(共享区)。
- **T3.2** vesting/cliff 释放调度器:按 `GENESIS_ALLOCATIONS` 的锁定参数线性释放团队/投资人桶。影响:`modules/ledger/service.py`, `tokenomics.py`(共享区)。
- **T3.3** 排放调度器幂等化 + 分布式锁(Redis),防重复铸造。影响:`modules/ledger/service.py`, `core/*`。
- **T3.4** 账本对账任务:周期性校验"余额和 = 净铸造",不一致告警。

### E4 交易所与做市
- **T4.1** WebSocket 实时行情/订单簿推送。影响:`modules/exchange/*`, `main.py`。
- **T4.2** 撮合引擎抽出为独立 Rust 服务(gRPC),账本仍在主程序。影响:新建 `services/matching-rs/`。
- **T4.3** 做市策略:Avellaneda-Stoikov 库存做市 + 风险敞口上限。影响:`modules/marketmaker/service.py`, `config.py`。
- **T4.4** 风控前置:下单前持仓/自成交/价格偏离检查。影响:`modules/exchange/service.py`。

### E5 前端与机构 API
- **T5.1** 用 Next.js + TradingView 重写专业交易界面(替换静态仪表盘)。影响:新建 `web-next/`。
- **T5.2** 机构 REST/FIX API + API Key 管理 + 限频。影响:`api/*`, `modules/identity/*`。

### E6 链上结算与治理
- **T6.1** 选链(Cosmos SDK / Substrate)发 CC,链下账本作缓存,定期结算上链。
- **T6.2** 治理模块:质押投票、国库提案与支出。

### E7 衍生品
- **T7.1** 永续合约(资金费率、强平引擎、保险基金)。
- **T7.2** CU 期货(现金/实物交割)、期权、算力指数。

### E8 运维 / 安全 / 合规上线
- **T8.1** K8s 部署 + 多区域 + 监控(Prometheus/Grafana)+ 日志(Loki)。
- **T8.2** 冷热钱包分离、多签、密钥管理(KMS/HSM)。
- **T8.3** 安全审计 + 渗透测试 + 账本/合约审计。
- **T8.4** 法律:牌照、用户协议、KYC/AML 制度文件(见 `compliance.md` 清单)。

---

## 3. 给 /workflow 的示例指令

```text
/workflow 读取 docs/workflow.md。先并行执行 E2(T2.1, T2.2)与 E4(T4.3, T4.4),
因为它们分属 compute/relay 与 exchange/marketmaker,不冲突。
共享区任务 T3.1、T3.2、T3.3 串行执行、各自独立 PR。
每个任务完成必须 make lint && make test 全绿,且不违反 CLAUDE.md 的不变量。
完成后用 [task-id] 标注 PR 标题。
```

> 建议先做 E1+E2+E3+E4 把"注册→贡献算力→计量→发币→交易"的主闭环加固到生产级,再推进 E5/E6/E7。
