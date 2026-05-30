# sub2api 集成 / Integration

> 目标:**复用** [sub2api](https://github.com/Wei-Shaw/sub2api) 的"用户接入 + 中转/算力分发"能力,但保持**进程隔离**与**许可证隔离**(见 `compliance.md`)。

---

## 1. 拓扑

```
┌──────────────────────────┐        HTTP/JSON        ┌──────────────────────────┐
│  ComputeCoin App          │  ───────────────────►   │  sub2api 网关             │
│  (:8788, 本仓库, MIT)      │   adapter: relay.       │  (:8080, 外部, LGPL)       │
│                           │   sub2api_client        │  - 用户/Key 分发           │
│  relay 模块登记“能力”      │  ◄───────────────────   │  - 多账号调度/按 token 计费 │
│  compute 模块计量 CU       │   delivered tokens等     │  - 真实请求转发到上游       │
└──────────────────────────┘                          └──────────────────────────┘
```

- sub2api 跑在 **8080**(你已部署的实例),ComputeCoin 跑在 **8788**。两者是**两个独立进程**。
- ComputeCoin **不**导入 sub2api 的 Go 代码,只调用它的 HTTP 接口。

---

## 2. 适配器:`backend/app/modules/relay/sub2api_client.py`

一个薄客户端,封装我们需要的 sub2api 操作。**所有方法都有"未配置则降级为本地模拟"的行为**,这样在没有真实 sub2api 时本仓库依然能跑起来、能测。

| 方法(计划/已留桩) | 对应 sub2api 能力 | 当前实现 |
|----|----|----|
| `health()` | 网关健康 | 真连或返回 mock |
| `register_capacity(...)` | 登记一个上游账号/能力(供给方自有) | 留桩,记录到本地 |
| `fetch_usage(api_key, since)` | 拉取某 Key 的 token 用量(用于计量交付) | 留桩/可对接 sub2api 用量接口 |

> 真实对接时,把 `SUB2API_BASE_URL` / `SUB2API_ADMIN_TOKEN` 配进 `.env`,并按 sub2api 实际接口补全这几个方法即可。接口契约写在该文件 docstring 里,方便 `/workflow` 子代理认领。

---

## 3. "复用注册逻辑"的落地方式

sub2api 本身有完整的用户注册/登录/Key 分发。我们有两种复用模式(可在 `.env` 切换,backlog 完善):

- **模式 A(默认,自有 identity)**:ComputeCoin 自己做注册/登录(`identity` 模块),把 sub2api 当作纯算力后端。简单、解耦。
- **模式 B(SSO 透传)**:以 sub2api 为身份源,ComputeCoin 通过其 API 校验用户 / 同步用户。适合"已有 sub2api 用户群直接迁移"。`identity` 预留了 `external_id` / `auth_provider` 字段以支持这种模式。

> MVP 实现模式 A 并预留模式 B 的字段与钩子,避免一开始就强耦合。

---

## 4. 计量 → CU → 排放 的衔接

```
sub2api.fetch_usage(key) ──► delivered_tokens(by model)
        │
        ▼  compute.service.tokens_to_cu(model, tokens)   # tokenomics.TOKEN_TO_CU
      verified_CU
        │
        ▼  ledger 排放调度器:按各供给方 verified_CU 占比铸造 CC
```

详见 `compute` 与 `ledger` 模块,以及 `docs/architecture.md` 第 4 节。
