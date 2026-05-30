# 快速启动指南 / Quickstart (for Agents & Humans)

> 本文假设你在一台有 Docker (20.10+) + Docker Compose v2+ 的 Linux/macOS 机器上操作。
> 你的 8080 端口已被 sub2api 占用。ComputeCoin 默认绑定 **8788**。

---

## 一键部署(推荐)

```bash
# 0. 进入项目根目录
cd AIComputeCoin

# 1. 生成 .env(包含随机密钥)
cp .env.example .env
python3 scripts/gen_secrets.py   # 或: make secrets

# 2. 后台启动(postgres + redis + app)
docker compose up -d --build     # 或: make up

# 3. 确认健康
curl http://localhost:8788/health
# {"status":"ok","service":"ComputeCoin","version":"0.1.0"}

# 4. 打开浏览器
# 仪表盘:  http://localhost:8788
# API 文档: http://localhost:8788/docs  (OpenAPI / Swagger UI)
# 管理员:  .env 中的 ADMIN_EMAIL / ADMIN_PASSWORD
```

首次启动会自动:
- 创建数据库表
- 铸造透明创世分配(210,000,000 CC 分 6 个桶)
- 创建管理员(如果配了 ADMIN_EMAIL)
- 创建做市商并种子流动性(开发环境)
- 启动排放调度器(每小时检查是否有待发放的每日排放)

---

## 换端口

编辑 `.env`:

```env
APP_HOST_PORT=9090       # 把 8788 换成你想要的
POSTGRES_HOST_PORT=5434  # 如果 5433 也被占
REDIS_HOST_PORT=6381     # 如果 6380 也被占
```

然后 `docker compose up -d` 即可。

---

## 停止 / 销毁

```bash
make down                    # 停止容器
docker compose down -v       # 停止 + 删除数据卷(重来)
```

---

## 本地开发(不用 Docker)

```bash
# 安装 Python 3.11+
python3.11 -m venv .venv && source .venv/bin/activate
pip install -r backend/requirements-dev.txt

# 用 SQLite(零配置),不需要 Postgres/Redis
cd backend
uvicorn app.main:app --reload --port 8000
# 浏览器 http://localhost:8000
```

---

## 运行测试

```bash
make test    # 或: pytest
make lint    # ruff + black --check
make fmt     # 自动格式化
```

---

## 接入你本地的 sub2api

在 `.env` 加:

```env
SUB2API_BASE_URL=http://host.docker.internal:8080
# SUB2API_ADMIN_TOKEN=<如果 sub2api 配了 token>
```

然后重启: `docker compose restart app`。

`relay` 模块会通过 HTTP 适配器与你本机 8080 的 sub2api 通信,计量交付的 token → 折算 CU → 排放 CC。

---

## 给 Claude Code / Hermes 的一句话指令

```
读 CLAUDE.md,然后按 docs/workflow.md 的任务 backlog 认领任务。
每个任务完成后跑 make lint && make test,全绿即可提 PR。
```

---

## 常见问题

| 问题 | 解决 |
|------|------|
| 端口 8788 被占 | 改 `.env` 的 `APP_HOST_PORT` |
| `docker compose` 不存在 | 安装 Docker Compose v2 (`docker compose` 是 v2 子命令) |
| 数据库连接失败 | 确认 `db` 容器健康: `docker compose ps` |
| 首次启动慢 | 拉 postgres/redis 镜像 + pip install,第二次秒启 |
| sub2api 连不上 | 确认 sub2api 跑在 host:8080 且 `host.docker.internal` 可达 |
