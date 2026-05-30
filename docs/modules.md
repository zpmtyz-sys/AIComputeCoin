# 模块说明 / Modules

每个模块位于 `backend/app/modules/<name>/`,通常含:
- `service.py` —— **纯业务逻辑**(可单测,跨模块只通过它互调)
- `router.py` —— 薄 HTTP 层(校验入参 → 调 service → 返回 schema)
- `__init__.py`

模型在 `backend/app/models/`,出入参 schema 在 `backend/app/schemas/`。

> 修改某模块"细节"时:优先改其 `service.py` 的纯函数与 `tokenomics.py`/`config.py` 的参数;保持模型与跨模块接口稳定。

---

## identity — 用户与鉴权
- **职责**:注册、登录(JWT)、角色(user/supplier/admin)、KYC 状态、对标并可对接 sub2api 用户体系。
- **关键模型**:`User(id, email, password_hash, role, kyc_status, auth_provider, external_id, created_at)`
- **关键接口(service)**:`register()`, `authenticate()`, `get_by_id()`, `set_kyc_status()`
- **HTTP**:`POST /auth/register`, `POST /auth/login`, `GET /auth/me`
- **不变量**:密码只存哈希(bcrypt);注册即创建对应**账本账户**(见 ledger)。

## relay — sub2api 集成 + 算力能力登记/计量
- **职责**:供给方登记"我合法拥有的算力/中转能力";拉取交付用量;喂给 compute 计量。
- **关键模型**:`CapacitySource(id, user_id, kind[sub2api|gpu], endpoint, source_attestation, verified, reputation, created_at)`、`DeliveryRecord(id, source_id, model, delivered_tokens, gpu_hours, period, created_at)`
- **关键接口**:`register_source()`(强制 `source_attestation`)、`record_delivery()`、`list_sources()`
- **HTTP**:`POST /relay/sources`, `GET /relay/sources`, `POST /relay/sources/{id}/delivery`
- **不变量**:`source_attestation` 必填;`verified=False` 时交付不计入排放。

## compute — 算力预言机(CU + PoDC)
- **职责**:CU 标准化折算;Proof-of-Delivered-Compute 验证(基准挑战 + 信誉);输出"本期各供给方 verified_CU"。
- **关键接口**:`tokens_to_cu(model, tokens)`、`gpu_hours_to_cu(gpu, hours, factors)`、`run_verification(source)`、`verified_cu_by_supplier(period)`
- **HTTP**:`POST /compute/verify/{source_id}`, `GET /compute/cu/quote`(给定模型/算力估 CU)
- **不变量**:折算系数全部来自 `tokenomics.py`;纯函数、有单测。

## ledger — 代币账本(排放/减半/双分录/国库)
- **职责**:账户、余额、双分录分录、创世分配、按公开曲线排放铸造、国库手续费归集。
- **关键模型**:`Account(id, owner_type, owner_ref, kind)`、`LedgerEntry(id, ref, debit_account, credit_account, amount, kind, created_at)`、`EmissionEpoch(period, minted, processed_at)`
- **关键接口**:`get_balance()`, `transfer()`, `mint()`(检查上限)、`emission_for_period(period)`(纯函数)、`run_emission(period)`、`seed_genesis()`
- **HTTP**:`GET /ledger/accounts/{ref}`, `GET /ledger/balance`, `GET /ledger/supply`
- **不变量**:双分录守恒;总供给 ≤ `MAX_SUPPLY`;排放纯函数确定性。

## exchange — 现货撮合引擎
- **职责**:下单/撤单、价格-时间优先撮合、成交、手续费(进国库)、订单簿/行情。
- **关键模型**:`Order(id, user_id, pair, side, type, price, qty, filled, status, created_at)`、`Trade(id, pair, price, qty, taker_order, maker_order, taker_fee, maker_fee, created_at)`
- **关键接口**:`place_order()`(下单即在事务内撮合)、`cancel_order()`、`order_book(pair)`、`recent_trades(pair)`
- **HTTP**:`POST /exchange/orders`, `DELETE /exchange/orders/{id}`, `GET /exchange/orderbook`, `GET /exchange/trades`
- **不变量**:价格优先、同价时间优先;撮合后资产 delta 之和=0(除显式手续费);下单前冻结余额校验。

## marketmaker — 透明做市
- **职责**:围绕参考价/库存给出公开的双边报价,提供流动性。策略**公开**、参数在 config。
- **关键接口**:`quote(pair, mid, inventory)`(返回买卖报价)、`rebalance()`
- **HTTP**:`GET /mm/quotes`(展示当前做市报价,透明)
- **不变量**:策略与参数公开;做市账户资金流同样走账本分录。

---

## 系统账户(account ref 约定)
| ref | 含义 |
|-----|------|
| `genesis:<bucket>` | 创世各桶 |
| `treasury` | 国库(手续费归集) |
| `emission` | 排放铸造的来源账户(铸造时贷方) |
| `user:<user_id>` | 用户账户 |
| `mm` | 做市账户 |
