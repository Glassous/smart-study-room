# 智能共享自习室预约系统（Smart Study Room）

![CI](https://github.com/imicola/smart-study-room/actions/workflows/ci.yml/badge.svg)
![Go](https://img.shields.io/badge/backend-Go%201.27%20%2B%20Gin-00ADD8)
![Vue](https://img.shields.io/badge/frontend-Vue%203%20%2B%20Vite-42b883)
![PostgreSQL](https://img.shields.io/badge/db-PostgreSQL%2018-336791)
![Redis](https://img.shields.io/badge/cache-Redis%207-DC382D)
![License](https://img.shields.io/badge/license-MIT-green)

> 《系统分析与设计》课程设计 —— 面向高校共享自习室场景的座位预约与管理平台。

## 项目简介

随着高校与城市共享自习室的普及，传统"到店找座"模式存在座位资源利用率不均、
占座严重、违约成本低等问题。本系统提供**在线选座预约、智能自动分配、
座位热力图、信用分体系、满座候补**等能力，并引入 **Redis 内存中间件** 支撑高并发分布式锁、旁路缓存防雪崩、接口级限流与 JWT 登出黑名单，
帮助学生高效获得学习座位，帮助管理员数据化运营自习室资源。

## 技术栈

| 层次 | 选型 | 用途 |
| --- | --- | --- |
| 后端 | Go 1.27 + Gin + pgx | RESTful API 服务与业务领域模型 |
| 前端 | Vite + Vue 3 + Element Plus + ECharts + Pinia | 响应式交互界面与图表大屏 |
| 关系数据库 | PostgreSQL 18 | 持久化主库，GiST 排除约束兜底防超卖 |
| 缓存与互斥 | Redis 7（Docker 容器） | 旁路缓存（Cache-Aside）、分布式锁、限流与黑名单 |
| 认证机制 | JWT（HS256）+ bcrypt + Redis 黑名单 | 安全会话与主动注销 |
| 容器与 CI | Docker Compose + GitHub Actions | 一键本地编排与自动化构建测试 |

## 功能总览

- **身份鉴别与认证**：注册 / 登录 / JWT 会话 / 登出主动注销（Redis 黑名单） / 登录防爆破限流
- **座位管理**：自习室与座位增删改查、批量生成、靠窗 / 电源 / 区域属性维护、座位平面图 Redis 极速缓存
- **高并发在线预约**：基于 Redis 分布式互斥锁（Lua 脚本）排队抢座 + PostgreSQL 排除约束最终兜底
- **智能自动分配**：按偏好加权评分自动推荐 / 分配座位（偏好得分 + 削峰填谷热度均衡）
- **座位热力图**：座位 × 时段利用率可视化，Redis 旁路缓存聚合报表（TTL 抖动防雪崩）
- **预约全生命周期**：签到、临时离开、签退、超时违约自动处理
- **高可用后台调度**：基于 Redis 分布式锁竞选 Leader 主备节点，支持多实例水平扩展
- **信用分体系**：违约扣分、履约加分、低分限约与流水追溯
- **满座候补**：候补排队、空位自动递补
- **消息中心**：预约 / 违约 / 递补等事件站内通知
- **统计仪表盘**：使用率趋势、高峰时段、热门座位分析

## 快速开始

> 推荐使用 **Docker Compose** 一键启动 PostgreSQL + Redis + 后端（起库即用）；如需纯本地运行调试，系统具备**优雅降级（Fallback）**能力：若不配置或无法连接 Redis，系统会自动降级为纯 PostgreSQL 模式运行。

### 方式一：Docker 启动（数据库 + Redis + 后端，推荐）

前置条件：已安装并启动 [Docker Desktop](https://www.docker.com/products/docker-desktop/)。

```bash
docker compose up -d --build     # 构建并启动 PostgreSQL、Redis 7 与后端
docker compose ps                # 查看状态：db / redis 应 healthy，backend 应 running
docker compose logs -f backend   # 跟踪后端日志
```

- **后端 API**：http://localhost:8080
- **PostgreSQL 数据库**：宿主机 `localhost:5433`（容器内 5432，数据卷持久化于 `db-data`）
- **Redis 缓存服务**：宿主机 `localhost:6380`（容器内 6379，数据卷持久化于 `redis-data`，避免与宿主机既有 6379 冲突）
- **数据迁移与初始化**：首次启动时已自动导入建表脚本与 14 天演示种子数据
- **停止服务（保留数据）**：`docker compose down`
- **重置数据与镜像**：`docker compose down -v`
- **本地前端联调**：`cd frontend && npm install && npm run dev`，Vite 已将 `/api` 代理到 `http://127.0.0.1:8080`，可直接对接容器内后端

### 方式二：纯本地运行（开发调试用）

#### 1. 初始化数据库（PostgreSQL 用户级集群）

```bash
./scripts/db_init.sh     # 首次：初始化集群并创建 studyroom 库
./scripts/db_start.sh    # 启动
./scripts/apply_migrations.sh
./scripts/apply_seed.sh  # 演示种子数据（含 14 天历史预约）
```

#### 2. 启动后端

```bash
cd backend
go run ./cmd/server    # 默认监听 :8080
```

#### 3. 启动前端

```bash
cd frontend
npm install
npm run dev            # 默认 http://localhost:5173
```

默认管理员账号：`admin / admin123`（演示用）。

### AI 助手配置

学生登录后可通过右下角悬浮球打开 AI 助手。助手支持查询本人资料、信用状态、预约、候补和通知，也可根据自然语言偏好推荐座位；推荐结果必须由学生在卡片中再次确认，实际下单仍执行原有信用、时段和并发冲突校验。

后端通过 OpenAI Chat Completions 兼容接口调用模型，至少配置：

```bash
STUDYROOM_AI_BASE_URL=https://api.openai.com/v1
STUDYROOM_AI_API_KEY=your-key
STUDYROOM_AI_MODEL=gpt-4o-mini
```

兼容服务需支持 SSE 流式输出及 tools/function calling。可选配置包括连接超时、响应超时和历史上下文条数，详见 `.env.example`。未配置 AI 时其他预约功能照常运行。

> AI 等环境变量统一填写在仓库根目录的 `.env`（模板见根目录 `.env.example`）。Docker 部署时由 `docker-compose.yml` 将其注入后端容器；本地直跑（`cd backend && go run ./cmd/server`）同样读取该文件。`backend/` 目录下不再需要、也不再维护单独的 `.env`。

AI 会话保存在 PostgreSQL 的 `ai_conversations`、`ai_messages` 表中。部署升级时需执行 `./scripts/apply_migrations.sh` 以应用 `002_ai_assistant.sql`。

## 项目结构

```
backend/
├── cmd/server/       入口主程序（依赖注入、优雅降级装配）
├── internal/
│   ├── config/       配置加载（PostgreSQL、Redis 等环境变量）
│   ├── handler/      HTTP 适配器（参数解析、响应输出）
│   ├── middleware/   中间件链（CORS、JWT 认证、Redis 黑名单、API 频次限流）
│   ├── model/        领域模型与 DTO 传输对象
│   ├── pkg/
│   │   ├── rediscache/ 旁路缓存助手（Cache-Aside、防雪崩、优雅降级）
│   │   └── redissync/  分布式互斥锁（Lua 脚本原子加锁与解锁）
│   ├── repository/   PostgreSQL 数据访问与 Redis 客户端连接池
│   ├── scheduler/    后台周期任务（基于 Redis 锁竞选 Leader 主备节点）
│   └── service/      核心业务逻辑（预约、分配算法、生命周期、信用、统计）
frontend/             Vue 3 前端应用
docs/                 课程设计文档（可行性研究/范围说明/报告/需求/概要/详细/进度日志）
diagrams/             系统分析设计图（Mermaid × 21：用例/架构/ER/类图/状态/时序/DFD 等）
scripts/              数据库与运维自动化脚本
docker-compose.yml    PostgreSQL 18 + Redis 7 + 后端多容器编排
```

## 课程设计文档索引

> 以下文档均基于课程模板文件（`.doc` / `.docx`）修改生成，保留模板封面、目录、标题样式与页脚格式；文档中图片已按模板要求以 `【插图位置】` 占位符标记。Word 版（`.docx`）可直接提交。

| 文档 | 内容 |
| --- | --- |
| [00_可行性研究报告.docx](docs/00_可行性研究报告.docx) | 技术/经济/操作/社会四维度可行性论证、方案对比、投资效益分析 |
| [00_项目范围说明书.docx](docs/00_项目范围说明书.docx) | 项目目标、工作范围（8 任务）、验收标准、假定与约束 |
| [02_需求规格说明书.docx](docs/02_需求规格说明书.docx) | 数据字典、FR-01~10、非功能需求 |
| [00_可行性研究报告.md](docs/00_可行性研究报告.md) | 上述 Word 文档的 Markdown 源文件 |
| [00_项目范围说明书.md](docs/00_项目范围说明书.md) | 上述 Word 文档的 Markdown 源文件 |
| [01_课程设计报告.md](docs/01_课程设计报告.md) | 前言/系统概述/系统分析/系统设计/系统实现/收获体会 |
| [02_需求规格说明书.md](docs/02_需求规格说明书.md) | 数据字典、FR-01~15、非功能需求 |
| [03_概要设计说明书.md](docs/03_概要设计说明书.md) | 四层架构、接口设计、ER/物理结构 |
| [04_详细设计说明书.md](docs/04_详细设计说明书.md) | 逐模块十要素、算法与测试要点 |
| [05_开发进度日志.md](docs/05_开发进度日志.md) | 28 项活动记录与工时统计 |

文档中的 `【插图位置】` 标记与 `diagrams/` 下 Mermaid 文件一一对应，终稿排版时渲染插入。

## 小组分工

| 成员 | 角色 | 职责 |
| --- | --- | --- |
| imicola | 组长 | 总体架构、数据库设计、CI/CD、代码审查与合并 |
| Glassous | 后端 | 认证与权限、用户与信用分模块 |
| DonGKids | 后端 | 预约核心、自动分配、候补与通知、统计 |
| hjsdjuhv8 | 前端 | 全部页面与交互实现 |
| polaris | 测试/文档 | 单元测试、需求文档、进度管理 |

## 许可证

[MIT](./LICENSE)
