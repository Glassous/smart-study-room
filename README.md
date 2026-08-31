# 智能共享自习室预约系统（Smart Study Room）

> 系统分析与设计课程设计项目 —— 面向高校共享自习室场景的座位预约与管理平台。

## 项目简介

随着高校与城市共享自习室的普及，传统"到店找座"模式存在座位资源利用率不均、
占座严重、违约成本低等问题。本系统提供**在线选座预约、智能自动分配、
座位热力图、信用分体系、满座候补**等能力，帮助学生高效获得学习座位，
帮助管理员数据化运营自习室资源。

## 技术栈

| 层次 | 选型 |
| --- | --- |
| 后端 | Go 1.27 + Gin + pgx |
| 前端 | Vite + Vue 3 + Element Plus + ECharts + Pinia |
| 数据库 | PostgreSQL 18 |
| 认证 | JWT（HS256）+ bcrypt |
| CI | GitHub Actions |

## 功能总览

- 身份鉴别与认证：注册 / 登录 / JWT 会话 / 学生与管理员双角色
- 座位管理：自习室、座位的增删改查与批量生成，靠窗 / 电源 / 区域属性
- 在线预约：座位平面图选座、时段冲突检测
- 智能自动分配：按偏好加权评分自动推荐 / 分配座位
- 座位热力图：座位 × 时段利用率可视化
- 预约全生命周期：签到、临时离开、签退、超时违约自动处理
- 信用分体系：违约扣分、履约加分、低分限约
- 满座候补：候补排队、空位自动递补
- 消息中心：预约 / 违约 / 递补等事件通知
- 统计仪表盘：使用率趋势、高峰时段、热门座位

## 快速开始

### 1. 初始化数据库（PostgreSQL 用户级集群）

```bash
./scripts/db_init.sh     # 首次：初始化集群并创建 studyroom 库
./scripts/db_start.sh    # 启动
./scripts/apply_migrations.sh
./scripts/apply_seed.sh  # 演示种子数据（含 14 天历史预约）
```

### 2. 启动后端

```bash
cd backend
go run ./cmd/server    # 默认监听 :8080
```

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev            # 默认 http://localhost:5173
```

默认管理员账号：`admin / admin123`（演示用）。

## 项目结构

```
backend/    Go 后端（handler / service / repository 分层）
frontend/   Vue 3 前端
docs/       课程设计文档
diagrams/   系统分析设计图（Mermaid）
scripts/    数据库与运维脚本
```

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
