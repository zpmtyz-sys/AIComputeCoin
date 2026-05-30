# ComputeCoin developer commands. See README.md and CLAUDE.md.
.PHONY: help up down logs build ps restart dev test lint fmt secrets install

help:
	@echo "ComputeCoin make targets:"
	@echo "  make secrets   生成随机密钥写入 .env (首次部署先跑这个)"
	@echo "  make up        docker compose 后台拉起 (postgres+redis+app)"
	@echo "  make down      停止并移除容器"
	@echo "  make logs      跟踪 app 日志"
	@echo "  make ps        查看服务状态"
	@echo "  make restart   重启 app"
	@echo "  make dev       本地(非docker)起开发服务器 (SQLite)"
	@echo "  make install   安装本地开发依赖"
	@echo "  make test      运行 pytest"
	@echo "  make lint      ruff + black --check"
	@echo "  make fmt       自动格式化"

up:
	docker compose up -d --build
	@echo "ComputeCoin: http://localhost:$${APP_HOST_PORT:-8788}  (API docs: /docs)"

down:
	docker compose down

logs:
	docker compose logs -f app

build:
	docker compose build

ps:
	docker compose ps

restart:
	docker compose restart app

install:
	pip install -r backend/requirements-dev.txt

dev:
	cd backend && uvicorn app.main:app --reload --port 8000

test:
	python -m pytest

lint:
	ruff check backend && black --check backend

fmt:
	ruff check --fix backend && black backend

secrets:
	python3 scripts/gen_secrets.py
