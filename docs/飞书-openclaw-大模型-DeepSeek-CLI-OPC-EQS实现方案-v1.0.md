# 飞书 → OpenClaw → OpenCode/DeepSeek → CLI/工具 → OPC：EQS 实现方案

> 文档状态：方案与现有资源基线已建立；EQS Feishu 账号、模型调用和端到端消息尚未验收
> 更新日期：2026-08-29
> 适用项目：EQS（工程快捷服务）
> 文档归属：EQS 项目 `docs/` 目录
> 敏感信息原则：本文不记录 App Secret、API Key、Token、SSH 私钥或其他私密凭据。

## 1. 项目与目标

EQS 是工程服务业一站式智慧交易与协作平台，项目根目录为：

```text
E:\2026-2027\2026-2027-1\MyProjects\eqs
```

现有项目资料显示其技术栈为 Go/Gin/GORM 后端、Vue 3 管理端、uni-app 用户端，数据库以 SQLite 为主，并包含 CI/CD、外部通道和 AI 分析能力。

目标链路：

```text
飞书 EQS 机器人
  → OpenClaw Gateway（共用现有单实例）
  → feishu/<EQS独立accountId>
  → leader-eqs
  → OpenCode 工具链 / DeepSeek Provider
  → CLI/工具按需执行
  → EQS 项目工作区
  → OPC 多 Agent 协作
```

## 2. 当前已确认资源

OpenClaw 当前已存在 EQS Agent：

```text
leader-eqs
pm-eqs
dev-refactor-eqs
qa-regression-eqs
reviewer-audit-eqs
```

其工作区位于：

```text
E:\2026-2027\2026-2027-1\MyProjects\eqs\.openclaw\workspaces\<agent-id>
```

当前已登记的 DeepSeek Provider；OpenClaw 配置中的 `opencode` 插件也已启用，但二者职责不同：DeepSeek 是模型 Provider，OpenCode 是工具/开发运行链路，不能把 OpenCode 的登录状态等同于 DeepSeek API 调用权限。

当前已登记的 DeepSeek Provider：

```text
Provider：deepseek-openclaw
Base URL：https://api.deepseek.com/v1
可用模型：deepseek-v4-pro、deepseek-v4-flash
API：openai-completions
```

以上仅表示本机 OpenClaw 配置中存在注册信息，不等于模型凭据、余额、权限和实际调用已经验收。

## 3. 建议最终资源命名

| 资源 | 建议值 |
|---|---|
| Feishu accountId | `deepseek-eqs` |
| 主 Agent | `leader-eqs` |
| binding | `feishu/deepseek-eqs → leader-eqs` |
| 主模型（推荐） | `deepseek-openclaw/deepseek-v4-flash` |
| 项目根目录 | `E:\2026-2027\2026-2027-1\MyProjects\eqs` |

选择 Flash 作为默认模型是成本和响应速度优先的建议；需要复杂方案审查时再切换 Pro。最终模型以实际可用性和项目负责人确认结果为准。

## 4. Feishu 配置计划

正式落地前必须取得并在本机安全输入：

```text
EQS Feishu App ID
EQS Feishu App Secret
```

增量配置目标：

```json5
channels.feishu.accounts["deepseek-eqs"]
bindings 增加 feishu/deepseek-eqs → leader-eqs
```

账号初始应保持：

```json5
"enabled": false
```

完成 Secret 填写、飞书应用发布、机器人启用、事件订阅、WebSocket/长连接和入群确认后，再改为 `true`。修改前必须备份：

```text
C:\Users\ldl\.openclaw\openclaw.json
```

不得删除或整体覆盖现有 WXX、VOPC、GPPS 配置，不启动第二个 Gateway。

## 5. OPC Agent 分工

| Agent | 职责 | 写入边界 |
|---|---|---|
| `leader-eqs` | 计划、拆分、协调、汇总 | EQS 根目录及批准范围 |
| `pm-eqs` | 需求梳理、版本计划、任务拆分 | 需求/计划报告 |
| `dev-refactor-eqs` | Go、Vue、uni-app 重构与实现 | 经批准的代码文件 |
| `qa-regression-eqs` | Go 测试、前端检查、集成回归 | 测试报告与测试文件 |
| `reviewer-audit-eqs` | 只读审查、安全、diff 检查 | 审查报告，默认不改业务代码 |

不得让多个 Agent 同时修改同一文件。涉及支付、短信、OCR、推送、电子签、生产部署、数据库清理、密钥和外部发布时，必须人工确认。

## 6. EQS 项目基线

已读取的项目资料表明：

- 后端采用 Go/Gin/GORM，前端包括 Vue 3 管理端和 uni-app 客户端；
- 核心业务包括需求发布、服务商匹配、订单、交付、结算、信用、会员和争议处理；
- AI 配置文档描述了未配置智谱 Key 时降级到规则分析，配置后调用智谱 GLM；这属于 EQS 应用内 AI，与 OpenClaw 的 DeepSeek Agent 链路是两条不同链路，不应混写；
- 外部通道（微信支付、短信、OCR、推送、电子签、CAD）必须分别按其文档完成真实联调，不能用“代码测试通过”替代生产验收；
- `EQS` 项目规范要求中文文档、英文代码命名、规范化 Git 提交和测试覆盖要求。

## 7. 工具与 CLI 使用原则

“DeepSeek 模型”负责理解、计划和结果汇总；OpenCode/CLI 工具链负责在项目工作区执行可审计的命令。工具调用必须：

1. 先读取 `EQS/AGENTS.md` 和相关 `docs/`；
2. 先输出计划，复杂变更等待人工确认；
3. 限定工作目录为 EQS 根目录或指定 workspace；
4. 禁止使用 `--yolo` 或等效方式绕过安全确认；
5. 代码变更后执行 `go test ./...`、前端 lint/typecheck/build 中适用的检查；
6. 检查 Git diff、敏感信息和越界修改；
7. 记录失败原因和可回滚点。

OpenClaw 是否经 CC-switch 或其他本地代理连接模型，必须以实际 Provider `baseUrl`、Gateway 进程环境和请求日志为准，不能仅因使用 CLI 就认定已接入。

## 8. 验收标准

### 8.1 配置验收

```text
Gateway：Connectivity probe: ok
Feishu deepseek-eqs：enabled, configured, running, connected, works
路由：feishu/deepseek-eqs → leader-eqs
模型：leader-eqs → deepseek-openclaw/deepseek-v4-flash
```

### 8.2 消息验收

1. 飞书机器人已发布并加入测试群；
2. 群聊明确 @EQS 机器人；
3. 发送只读项目分析问题；
4. 确认回复来自 `leader-eqs`；
5. 确认 Provider/Model 正确且不返回 401/403；
6. 再进行代码修改类任务；
7. 回归 WXX、VOPC、GPPS 原有路由不变。

## 9. 未完成项

- [ ] EQS Feishu App ID/App Secret 已取得并仅在本机配置；
- [ ] 飞书应用发布、机器人、事件订阅和长连接已确认；
- [ ] `deepseek-eqs` 账号与 binding 已增量写入；
- [ ] DeepSeek 实际调用成功；
- [ ] EQS 端到端消息测试通过；
- [ ] WXX、VOPC、GPPS 回归通过；
- [ ] 生产外部通道分别完成真实联调。

## 10. 版本记录

| 版本 | 日期 | 内容 |
|---|---|---|
| v1.0 | 2026-08-29 | 建立 EQS 独立 OpenCode + DeepSeek + OpenClaw + CLI/工具 + OPC 方案，文档保存于 EQS `docs/`，明确已知资源、隔离边界和未验收项 |
