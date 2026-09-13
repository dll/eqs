# CCIT ENV P0 交付清单与 Gateway/EQS Manifest R2

> 完整交付事实、命令、判据、runbook 与阻断：
> `E:\2026-2027\2026-2027-1\CCIT\.openclaw\workspaces\ccit-env\memory\facts\CCIT-ENV-P0-delivery-manifest-R2-20260909.md`

## 交付结论

- 采集时间：2026-09-09 22:24:55（Asia/Shanghai）。
- EQS HEAD：`c1d793c2b72e6339278db1429b2c3e600035eb4a`；branch `feature/archived-project-inbound`。
- Gateway：已确认 `127.0.0.1:18789 Ready`；本轮只读 HTTP 根探针返回 200；PID 35356 单一 LISTEN 归属。
- Feishu：采用已确认输入，8 个连接均 `running/connected`；本轮不发测试消息。
- 工具链：Node 24.19.0、npm 11.17.0、pnpm 10.34.5、Go 1.26.0 windows/amd64、Git 2.55.0.windows.5、OpenClaw 2026.9.2 (3928bad)。
- EQS：server 默认 8080；admin 3001；client 3005；前端 `/api` proxy 指向 8090；DB 默认 MySQL，非生产 SQLite/mock 方案与 env 来源已明确。
- 安全边界：未安装 Docker/WSL，未修改宿主机全局配置、防火墙、服务或进程，未触碰真实密钥/生产，未发送飞书测试。
- 审核状态：**可交 CTO 做 Conditional 审核**；不是生产或隔离生产 Ready。
- 主要阻断：Docker/WSL 不可用且按约束不安装；EQS 8080 与 proxy 8090 不一致；8080/Tomcat、3000/Gitea、5432/PostgreSQL 为外部占用；MySQL/Redis 隔离、生产/真实通道、迁移/备份恢复和对外发布均未验证/未获 Final Approved。
