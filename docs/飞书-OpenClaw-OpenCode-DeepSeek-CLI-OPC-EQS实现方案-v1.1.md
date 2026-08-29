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
Bot name:         EQS / DeepSeek-EQS（以飞书平台实际名称为准）
```

该账号和 binding 仅作为 EQS 独立资源规划。未取得并在本机安全配置有效凭据前，不应启用真实外部通道，也不得替换其他已有项目的 Feishu 路由。

### 3.2 飞书应用与机器人配置

EQS 应使用独立的飞书自建应用和机器人，不复用其他项目的 App、`accountId`、Secret、群会话或 binding。当前方案只记录脱敏后的逻辑标识；真实 App Secret 只能通过本机安全配置输入，不能写入本文、Git、日志或飞书消息。

| 配置项 | EQS 目标值 | 验证要求 |
|---|---|---|
| 应用类型 | 飞书自建应用 | 在飞书开放平台确认应用归属 |
| 应用名称 | EQS / DeepSeek-EQS | 与测试群中的机器人名称一致 |
| App ID | 以本机安全配置为准 | 文档不记录完整 Secret；App ID 可按脱敏策略记录 |
| App Secret | 仅保存于本机安全配置 | 不粘贴到聊天、文档或仓库 |
| 机器人能力 | 启用机器人 | 确认机器人已发布并可被添加 |
| 事件接收方式 | WebSocket/长连接 | 与 OpenClaw Feishu 通道配置保持一致 |
| OpenClaw 账号 | `deepseek-eqs` | 与 `channels.feishu.accounts` 的键一致 |
| 默认 Agent | `leader-eqs` | 与 binding 目标一致 |
| 目标工作区 | EQS 根目录或批准的 Agent workspace | 不跨项目写入 |

### 3.3 飞书开放平台配置步骤

1. 创建或选择 EQS 专用飞书自建应用，确认应用名称、应用所有者和租户正确。
2. 启用应用的机器人能力，设置可识别的机器人名称和头像。
3. 按 OpenClaw Feishu 插件实际需要配置权限；至少应覆盖机器人接收事件、读取必要消息上下文、发送消息/卡片和访问群基本信息的能力。最终权限以当前 OpenClaw 版本和飞书开放平台权限页面为准，不凭空扩大权限。
4. 配置事件订阅，选择与 OpenClaw 配置匹配的 WebSocket/长连接方式；确认应用能保持长连接，不将 HTTP 回调地址误配为另一套部署方案。
5. 保存事件订阅配置并检查飞书平台的校验结果、应用状态和事件连接状态。
6. 发布应用版本。未发布的权限、机器人能力或事件配置不能作为生产验收依据。
7. 将机器人加入专用测试群；群聊默认要求明确 @EQS 机器人后才处理消息。
8. 在完成本机认证探针、路由检查和模型检查后，再执行真实群消息验收。

### 3.4 OpenClaw Feishu 增量配置

修改前先备份：

```text
C:\Users\ldl\.openclaw\openclaw.json
```

只能依据当前 JSON5 schema 做最小增量修改，保留已有项目配置。逻辑目标如下，字段名称和 Secret 注入方式以当前配置 schema 为准：

```json5
channels: {
  feishu: {
    accounts: {
      "deepseek-eqs": {
        enabled: false, // 凭据、发布和平台检查完成后再启用
        appId: "<仅本机安全配置>",
        appSecret: "<仅本机安全配置>"
      }
    }
  }
},
bindings: [
  {
    channel: "feishu",
    accountId: "deepseek-eqs",
    agentId: "leader-eqs"
  }
]
```

落地顺序：

```text
备份配置
  -> 校验 JSON5/schema
  -> 增加独立 Feishu account
  -> 增加 feishu/deepseek-eqs -> leader-eqs binding
  -> 保持账号 disabled 进行结构检查
  -> 安全注入凭据并启用
  -> 等待 Gateway 热加载或按既有单实例流程恢复
  -> channels status --probe
```

不得删除或覆盖其他项目既有路由，不得启动第二个 Gateway。EQS 只新增自己的 `deepseek-eqs` 账号和 binding，不改变现有通道的目标 Agent。

### 3.5 Feishu 会话与消息路由

EQS 消息必须沿以下链路处理：

```text
Feishu EQS 群/私聊
  -> Feishu account: deepseek-eqs
  -> OpenClaw Gateway
  -> binding: feishu/deepseek-eqs -> leader-eqs
  -> leader-eqs session
  -> DeepSeek 负责理解、计划和汇总
  -> OpenCode/CLI 在批准的 EQS workspace 执行
  -> 结果返回 Feishu
```

路由验收必须同时核对 `accountId`、`agentId`、会话键和工作区，不能只看到机器人回复就认定路由正确。历史会话中的旧模型或旧 Agent 信息不能覆盖新会话的实际日志；模型验收以当前请求日志和有效回复为准。

### 3.6 群聊、私聊与权限策略

- 群聊默认 `requireMention: true`，必须明确 @EQS 机器人；未 @ 的普通群消息不应触发 Agent。
- 测试群应使用最小成员范围，先验证只读问题，再验证需要人工确认的工作区任务。
- 私聊策略以 EQS 账号的实际 allowlist 配置为准，不得因调试方便扩大到任意用户。
- 不在群聊中发送 App Secret、API Key、Token、服务器密码、数据库密码或用户隐私数据。
- 消息处理应遵循 OpenClaw 的会话隔离；EQS 会话不能复用其他项目会话，其他项目也不能写入 EQS workspace。
- 涉及生产部署、支付、短信、OCR、推送、电子签、数据库清理或批量修改时，必须二次确认。

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

### 9.1 飞书平台验收

| 检查项 | 通过标准 | 证据 |
|---|---|---|
| 应用身份 | EQS 独立自建应用，名称和租户正确 | 飞书开放平台应用页 |
| 机器人 | 机器人能力已启用 | 应用能力页 |
| 权限 | 所需消息接收、群信息读取和发送权限已授权 | 权限/版本页 |
| 事件 | WebSocket/长连接配置成功 | 事件订阅页与 Gateway 日志 |
| 发布 | 当前权限和机器人配置已发布 | 应用版本页 |
| 入群 | EQS 机器人已加入专用测试群 | 群成员列表 |

### 9.2 Gateway、路由和模型验收

```text
Gateway:              Connectivity probe: ok
Feishu deepseek-eqs:  enabled, configured, running, connected, works
Binding:              feishu/deepseek-eqs -> leader-eqs
Agent:                leader-eqs
Model:                deepseek-openclaw/deepseek-v4-flash
```

至少执行并保存脱敏后的结果：

```powershell
openclaw gateway status
openclaw channels status --probe
```

检查点：

- Gateway 只有一个运行实例，监听既有 `127.0.0.1:18789`；
- `deepseek-eqs` 账号状态为 `enabled, configured, running, connected, works`；
- binding 精确指向 `leader-eqs`；
- 日志中的实际 Provider、Model、请求地址和 HTTP 状态与配置一致；
- 不把配置文件存在模型项等同于模型调用成功。

### 9.3 Feishu 端到端验收

1. 在专用测试群明确 @EQS 机器人，发送只读问题，例如“请说明当前 EQS Agent、Provider 和 Model，不要修改文件”。
2. 确认消息由 `deepseek-eqs` 账号接收，并路由到 `leader-eqs`。
3. 确认收到有效中文回复，且实际请求不是 401、403、404、429 或超时。
4. 核对日志中的 Provider、Model、请求 URL、响应状态和会话键；凭据和完整用户隐私内容不得进入验收记录。
5. 新建或重置测试会话后，再执行一次只读工作区检查，确认工作目录为 EQS。
6. 如需验证代码任务，先取得人工确认，再执行最小变更；禁止直接执行破坏性操作。
7. 验证回复返回飞书群，卡片/文本流式输出正常，重复消息不会造成重复执行。
8. 回归确认其他项目的既有 Feishu 路由未改变。

### 9.4 EQS 工程验证

- 后端执行 `go test ./...`；
- 前端按实际目录执行适用的 lint、typecheck 和 build；
- 关键业务流程执行集成测试；
- 发布前按项目要求执行关键路径 E2E 测试；
- 测试、日志和报告中不得保存真实密钥。

### 9.5 失败排查

| 现象 | 优先检查 |
|---|---|
| `app secret invalid` / `10014` | Secret 是否属于 EQS App、是否被轮换、租户是否正确 |
| 账号未连接 | App 是否发布、机器人是否启用、长连接/事件订阅和网络是否正常 |
| 能连接但收不到群消息 | 机器人是否入群、是否明确 @、事件权限和群策略是否正确 |
| 收到消息但 Agent 错误 | binding、`accountId`、Agent ID、会话键和 workspace |
| 返回 401/403 | DeepSeek API Key、模型权限、余额、实际请求 URL 和 Provider 配置 |
| 返回 404 | Base URL 与 OpenAI-compatible 路径是否重复拼接 `/v1` |
| 超时或重复执行 | 会话队列、模型响应时间、消息去重和 Gateway 日志 |
| 回复模型与配置不一致 | 新建会话、检查 fallback/旧会话覆盖，并以实际日志为准 |

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
| v1.2 | 2026-08-29 18:45 | 补充与其他项目同等完整的飞书应用、机器人、权限、长连接、发布、入群、路由、验收和故障排查内容 |
