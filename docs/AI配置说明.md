# EQS AI 大模型配置说明（智谱 GLM）

> **文档版本**：V1.0
> **更新日期**：2026-08-11
> **适用**：EQS 后端 AI 项目分析功能（全量分析 + 单项目问题解析）

---

## 1. 功能说明

EQS 后端 AI 分析接口（`ai.go`）采用**可插拔**设计：

| 配置 | 行为 |
|------|------|
| **未配置 Key** | 走**规则分析**（内置逻辑：进度/风险/问题/建议），无需外部依赖 |
| **已配置 Key** | 调用**智谱 GLM** 大模型生成文字摘要与优化建议，叠加规则结果 |

## 2. 需要配置的环境变量

在服务器 systemd 服务（`/etc/systemd/system/eqs-server.service`）的 `[Service]` 段添加：

| 变量 | 说明 | 示例 |
|------|------|------|
| `ZHIPU_API_KEY` | 智谱开放平台 API 密钥 | `your-zhp-api-key` |
| `ZHIPU_BASE_URL` | API 地址（可选，默认内置） | `https://open.bigmodel.cn/api/paas/v4/chat/completions` |
| `ZHIPU_MODEL` | 模型名（可选，默认取环境或内置） | `glm-4-flash` |

> 不配置 `ZHIPU_BASE_URL` / `ZHIPU_MODEL` 时，代码使用默认值（`ai.go` 内置）。

## 3. 获取智谱 API Key

1. 注册登录 [智谱开放平台](https://open.bigmodel.cn)
2. 控制台 → API 密钥 → 创建新密钥 → 复制 `API Key`（形如 `xxxxx.xxxxx`）
3. 将 Key 配置到服务器（见第 4 节）

## 4. 配置步骤（服务器）

```bash
# 编辑 systemd 服务
sudo vi /etc/systemd/system/eqs-server.service

# 在 [Service] 段添加：
# Environment=ZHIPU_API_KEY=你的key
# Environment=ZHIPU_MODEL=glm-4-flash

# 重载并重启
sudo systemctl daemon-reload
sudo systemctl restart eqs-server
```

## 5. 验证

```bash
# 登录获取 token
TOKEN=$(curl -s -X POST http://127.0.0.1:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"13900003333","code":"123456","user_type":3}' | python3 -c "import sys,json;print(json.load(sys.stdin)['token'])")

# 全量分析（generated_by 应变为 ai）
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8090/api/v1/admin/ai/project-analysis | python3 -m json.tool
```

**预期**：配置成功后 `generated_by` 字段从 `rules` 变为 `ai`，`summary` 返回大模型生成的中文总结。

## 6. 注意事项

- 智谱 API 按 token 计费，免费额度有限，建议分析频率受限（前端手动触发）。
- Key 属于敏感信息，**只存服务器环境变量**，绝不提交到代码仓库或 GitHub Secrets 明文。
- 网络不可达智谱时，接口自动降级为规则分析（`generated_by: rules`），不影响功能。

---

**文档结束**
