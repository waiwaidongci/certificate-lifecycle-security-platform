# 项目 03：证书生命周期与安全配置分发平台

`github.com/acme/certpilot` 是一个从 0 到 1 实现的纯 Go 企业级后端项目，用于管理内部服务证书、TLS 配置模板、轮换计划、到期提醒、签发请求和安全配置下发。项目默认使用 `modernc.org/sqlite`，无需本机 Docker 或 PostgreSQL 即可启动；同时提供 PostgreSQL 迁移和部署示例。

## 架构

项目按业务领域和分层组织：

- `cmd/api`：HTTP API 入口，负责装配依赖、中间件、优雅关闭和后台提醒 worker。
- `cmd/rotator`：轮换 CLI 入口，可生成到期轮换计划并执行到期任务。
- `internal/certificate`：证书档案、签发请求、撤销和到期扫描。
- `internal/issuer`：签发器抽象、模拟签发器与签发器档案。
- `internal/policy`：签发策略、域名策略、有效期策略和环境策略。
- `internal/rotation`：轮换计划、轮换任务和状态流转。
- `internal/distribution`：安全配置模板、配置分发记录、日志适配器和 webhook 适配器。
- `internal/servicecatalog`：内部服务、环境、域名、负责人和证书绑定。
- `internal/notification`：到期提醒扫描、发送适配器和提醒记录。
- `internal/shared`：配置、日志、错误、HTTP、限流、数据库、事件日志等共享能力。

每个核心领域按 `domain`、`application`、`adapter`、`infrastructure` 分层。仓储接口定义在 `domain`，应用服务只依赖接口和 `database.Executor`，通过构造函数注入依赖，并通过 `context`、错误包装和结构化日志传播链路信息。

## 配置

默认配置见 [configs/config.example.yaml](configs/config.example.yaml)。环境变量可覆盖 YAML：

| 环境变量 | 说明 |
| --- | --- |
| `CERTPILOT_SERVER_ADDRESS` | HTTP 监听地址 |
| `CERTPILOT_DATABASE_DRIVER` | `sqlite` 或 `postgres` |
| `CERTPILOT_DATABASE_DSN` | 数据库 DSN |
| `CERTPILOT_ISSUER_DEFAULT_LIFETIME_DAYS` | 默认证书有效期 |
| `CERTPILOT_ROTATION_DEFAULT_ADVANCE_DAYS` | 默认轮换提前天数 |
| `CERTPILOT_LOG_LEVEL` | `debug`、`info`、`warn`、`error` |

SQLite 默认使用 `certpilot.db`，首次启动自动执行内嵌迁移。PostgreSQL 需要先应用 `migrations/postgres/000001_init.up.sql`。

## 迁移

本地开发：

```bash
./scripts/run-dev.sh
```

PostgreSQL：

```bash
docker compose -f deploy/docker-compose.postgres.yml up -d postgres
export CERTPILOT_DATABASE_DRIVER=postgres
export CERTPILOT_DATABASE_DSN='postgres://certpilot:certpilot@localhost:5432/certpilot?sslmode=disable'
./scripts/migrate-postgres.sh
```

## 运行

```bash
make run-dev
# 或
./scripts/run-dev.sh
```

服务默认监听 `:8080`。`/healthz`、`/readyz` 和 `/metrics` 为运维端点。

## 示例 API 请求

1. 登记服务：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/services \
  -H 'Content-Type: application/json' \
  -d '{"name":"orders","environment":"production","domain":"orders.internal","owner":"platform@example.com","region":"cn-east"}'
```

2. 创建策略：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/policies \
  -H 'Content-Type: application/json' \
  -d '{"name":"prod-tls","environments":["production"],"min_validity_days":7,"max_validity_days":365,"allowed_domains":["*.internal"]}'
```

3. 提交证书申请：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/certificates/issue \
  -H 'Content-Type: application/json' \
  -d '{"service_id":"<SERVICE_ID>","common_name":"api.orders.internal","sans":["orders.internal"],"validity_days":30,"idempotency_key":"issue-orders-001"}'
```

4. 创建配置模板：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/config-templates \
  -H 'Content-Type: application/json' \
  -d '{"name":"baseline-tls12","min_tls_version":"TLS1.2","cipher_suites":["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"],"require_mutual_tls":false,"require_full_chain":true}'
```

5. 下发配置：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/distributions \
  -H 'Content-Type: application/json' \
  -d '{"service_id":"<SERVICE_ID>","template_id":"<TEMPLATE_ID>","certificate_id":"<CERT_ID>","target_type":"log","target":"stdout"}'
```

6. 生成并推进轮换：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/rotation/plans/generate \
  -H 'Content-Type: application/json' -d '{"advance_days":30}'

curl -sS -X POST http://127.0.0.1:8080/api/v1/rotation/tasks/<TASK_ID>/transition \
  -H 'Content-Type: application/json' \
  -d '{"target_status":"completed","assigned_to":"operator","version":1}'
```

7. 扫描并发送到期提醒：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/notifications/reminders/scan \
  -H 'Content-Type: application/json' -d '{"advance_days":30,"channel":"log"}'
```

8. 查询事件日志：

```bash
curl -sS 'http://127.0.0.1:8080/api/v1/events?page=1&page_size=20&sort=created_at&sort_dir=desc'
```

## Makefile 与 scripts

- `make fmt`：执行 `gofmt -w .`
- `make test`：执行 `go test ./...`
- `make vet`：执行 `go vet ./...`
- `make build`：执行 `go build ./...`
- `make run-dev`：构建并启动本地 API
- `./scripts/count-go.sh`：统计非测试 Go 文件数和行数
- `./scripts/migrate-postgres.sh`：应用 PostgreSQL 迁移
- `./scripts/smoke.sh`：基础 curl 冒烟流程

## 已实现能力

- REST API、统一错误响应、分页、过滤、排序、乐观锁、幂等申请。
- 复杂写入事务：签发时同时写证书、绑定服务和事件日志；生成轮换时同时写计划和任务。
- HTTP 中间件：请求 ID、访问日志、超时、panic 恢复、限流、认证占位。
- 优雅关闭与 `/healthz`、`/readyz`、`/metrics`。
- 后台 reminder worker 扫描到期证书并发送通知。
