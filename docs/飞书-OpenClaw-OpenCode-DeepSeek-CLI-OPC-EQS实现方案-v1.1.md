# 飞书 → OpenClaw → OpenCode + DeepSeek → CLI/工具 → OPC：EQS 实现方案

> 文档状态：EQS 独立方案已归档；OpenClaw/DeepSeek 资源为配置基线，Feishu 账号与端到端消息尚未验收
> 更新日期：2026-08-29 17:43
> 适用项目：EQS（工程快捷服务）
> 文档归属：`E:\2026-2027\2026-2027-1\MyProjects\eqs\docs`
> 敏感信息原则：本文不记录 App Secret、API Key、Token、SSH 私钥或其他私密凭据。

## 1. 项目定位

EQS 是工程服务业一站式智慧交易与协作平台。项目根目录为：

```text
E:\2026-2027\2026-2027-1\MyProjects\eqs
```

当前项目技术栈包括 Go/Gin/GORM 后端、Vue 3 管理端、uni-app 用户端和 SQLite 数据存储，并包含 AI 分析、外部通道、订单协作和交付管理能力。

本文只描述 EQS 的 OpenClaw、OpenCode、DeepSeek、CLI/工具和 OPC 协作链路。

## 2. 总体架构

```text
Feishu EQS Bot
       |
       v
OpenClaw Gateway (single instance)
       |
       v
Feishu account: deepseek-eqs
       |
       v
leader-eqs
       |
       +--------------------------+
       |                          |
       v                          v
DeepSeek Provider            OpenCode / CLI tools
understanding, plan,          auditable workspace
summary                       execution
       |                          |
       +------------+-------------+
                    v
              EQS workspace
                    |
                    v
             OPC multi-agent flow
```

图表注解：上图中的英文节点用于保证 ASCII 等宽排版；DeepSeek 负责理解、计划和汇总，OpenCode/CLI 负责受控执行，二者职责不可混同。

## 3. OpenClaw 与 Agent 资源

| 资源 | 目标值/职责 |
|---|---|
| Gateway | 与现有项目共用一个实例，监听 `127.0.0.1:18789` |
| 主 Agent | `leader-eqs` |
| 需求 Agent | `pm-eqs` |
| 开发 Agent | `dev-refactor-eqs` |
| 测试 Agent | `qa-regression-eqs` |
| 审查 Agent | `reviewer-audit-eqs` |
| EQS 工作区 | `E:\2026-2027\2026-2027-1\MyProjects\eqs` |
| Agent 工作区 | `E:\2026-2027\2026-2027-1\MyProjects\eqs\.openclaw\workspaces\<agent-id>` |

### 3.1 推荐独立通道资源

```text
Feishu accountId: deepseek-eqs
Binding:          feishu/deepseek-eqs -> leader-eqs
```

该账号和 binding 仅作为 EQS 独立资源规划。未取得并在本机安全配置有效凭据前，不应启用真实外部通道，也不得替换其他已有项目的 Feishu 路由。

## 4. DeepSeek 模型链路

推荐配置基线：

```text
Provider:   deepseek-openclaw
Base URL:   https://api.deepseek.com/v1
API:        openai-completions
Model:      deepseek-v4-flash
备用模型:  deepseek-v4-pro
```

OpenClaw Agent 模型引用：

```text
leader-eqs -> deepseek-openclaw/deepseek-v4-flash
```

以上内容表示方案配置基线，不等同于实际 API 凭据、余额、模型权限或端到端调用已经验收。实际请求是否成功，必须以运行日志和有效回复为准。

## 5. OpenCode、CLI 与 DeepSeek 的职责边界

| 组件 | 职责 | 禁止混淆的事项 |
|---|---|---|
| DeepSeek Provider | 需求理解、技术分析、计划生成、结果汇总 | 不把模型调用当成代码已执行 |
| OpenClaw | Feishu 接入、Agent 路由、会话和工具策略 | 不绕过权限或另起 Gateway |
| OpenCode | 在批准工作区执行开发辅助任务 | 不把登录状态当成 DeepSeek API 认证 |
| CLI/系统工具 | 测试、构建、检查、版本控制等可审计操作 | 不使用 `--yolo` 或等效绕过确认 |
| OPC Agent | 按角色并行协作和汇总 | 不让多个 Agent 同时修改同一文件 |

标准工作流：

```text
阅读 docs/ 与 AGENTS.md
  -> 输出计划
  -> 人工确认复杂或高风险变更
  -> 在限定工作区执行 OpenCode/CLI
  -> 运行测试、Lint、构建或静态检查
  -> 检查 diff、敏感信息和写入边界
  -> 形成可回滚提交
```

## 6. EQS 应用内 AI 与 OpenClaw DeepSeek 的隔离

EQS 应用内 `ai.go` 的智谱 GLM 链路与 OpenClaw DeepSeek Agent 链路是两套独立能力：

- 应用内 AI 使用 `ZHIPU_API_KEY`、`ZHIPU_BASE_URL`、`ZHIPU_MODEL`；
- 未配置智谱 Key 时，应用可降级到规则分析；
- OpenClaw DeepSeek 用于 Agent 的理解、计划和汇总；
- OpenCode/CLI 用于 EQS 工作区中的可审计执行；
- 不得用应用内智谱调用结果证明 OpenClaw DeepSeek 已验收，反之亦然。

任何真实 Key 只能保存在本机或服务器安全环境变量/SecretRefs 中，不得写入本文、Git、日志或飞书消息。

## 7. 验收标准

### 7.1 配置和通道验收

```text
Gateway:              Connectivity probe: ok
Feishu deepseek-eqs:  enabled, configured, running, connected, works
Binding:              feishu/deepseek-eqs -> leader-eqs
Model:                deepseek-openclaw/deepseek-v4-flash
```

### 7.2 端到端验收

1. EQS Feishu 应用已发布，机器人已启用并加入测试群；
2. 群聊明确 @EQS 机器人，发送只读项目分析问题；
3. 确认回复路由到 `leader-eqs`；
4. 以日志确认实际 Provider、Model、请求地址和 HTTP 状态；
5. 确认没有 401/403，并收到有效中文回复；
6. 再执行一个经人工确认的只读检查或代码修改任务；
7. 执行 `go test ./...`，并按前端实际目录运行适用的 lint、typecheck 或 build；
8. 复核其他项目既有路由未被修改；
9. 记录失败、降级和回滚方式。

## 8. 安全与写入边界

- 不启动第二个 OpenClaw Gateway；
- 修改 OpenClaw 配置前先备份 `C:\Users\ldl\.openclaw\openclaw.json`；
- 依据当前 JSON5 schema 做最小增量修改，不整体覆盖配置；
- EQS Agent 只写入 EQS 根目录及明确批准的工作区；
- `pm-eqs` 输出需求和计划报告；
- `dev-refactor-eqs` 修改经批准的代码文件；
- `qa-regression-eqs` 维护测试和测试报告；
- `reviewer-audit-eqs` 默认只读审查并输出审计报告；
- 多个 Agent 不得同时修改同一文件；
- 支付、短信、OCR、推送、电子签、生产部署、密钥和外部发布必须人工确认。

## 9. 当前状态与后续事项

| 项目 | 状态 |
|---|---|
| EQS 独立方案文档 | 已归档于本项目 `docs/` |
| `leader-eqs` 及职责 Agent | 已有配置基线 |
| DeepSeek Provider 方案 | 已定义，待实际调用验收 |
| EQS Feishu App | 待取得并安全配置独立凭据 |
| `deepseek-eqs` 账号与 binding | 待确认后增量写入 |
| Feishu 发布、入群 | 待平台侧操作 |
| OpenClaw DeepSeek 端到端消息 | 待真实群聊验收 |
| EQS 应用内智谱 GLM | 独立链路，按 `AI配置说明.md` 验收 |
| 外部支付/短信/OCR等通道 | 必须分别进行真实联调 |

## 10. 版本记录

| 版本 | 日期 | 内容 |
|---|---|---|
| v1.0 | 2026-08-29 | 建立 EQS OpenClaw、OpenCode、DeepSeek、CLI/工具与 OPC 协作方案 |
| v1.1 | 2026-08-29 17:43 | 明确 EQS 文档归属路径、OpenCode 与 DeepSeek 职责边界、应用内智谱链路隔离及验收门槛 |
