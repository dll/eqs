# EQS v0.2.0 部署与多端真实化执行日志

> 任务链：EQS-MP-20260831-01（开发）+ EQS-OPS-20260831-01（运维）+ OPC 2026-09-01 批示
> 执行：ccit-ceo 协调，ccit-ops / leader-eqs 产出，CEO 持审批执行生产写
> 时间：2026-09-01 07:07-07:20 GMT+8

## 审批①：生产 .env 写入 WX 凭据（07:10）

- 凭据来源：OPC 提供于 `eqs\.env`（本地，gitignore 覆盖已验证 `check-ignore` ✅）
- 疑点排除：appid `wx8861be57512cdd59` 与 wxx 小程序 `wx811d1225e67b8f38`（wxx/docs/AGENTS.md 明载）不同，确认为 EQS 独立新小程序；两仓库代码均无此 appid（全新启用）
- 操作：sed dedupe → append 3 键（WX_MINI_APPID / WX_MINI_SECRET / **WX_MINI_MOCK=0**）；属主 root:root 600 前后一致；值经 temp 片段文件中转，未回显、未入 Git

## 审批②：tag v0.2.0 → CD 自动部署（07:14-07:16）

- tag 内容：小程序 code2session 真实登录（8d815f6）+ R2-2 缓存并发修复（04a1b9f）+ CD tag 授权制（4e189e5）
- CD 触发：`git push origin v0.2.0` → GitHub Actions（首用新 tag 授权制，push main 不部署已先行实证）
- 生产结果：二进制 mtime 09-01 07:16（原 08-16 14:33）；eqs-server 重启 active（新 pid 2669507，原 1545774）；.env 3 键完好

## 部署后验证（07:17-07:18）

| 项 | 结果 |
|---|---|
| HTTPS 公网 /api/v1/config/public | 200 |
| HTTPS /h5/ | 200 |
| HTTPS /admin/ | 200 |
| 路由注册 | journalctl 确认 `POST /api/v1/auth/wechat-login → handler.WxLogin` |
| 启动日志 error/fail | 无 |
| **真实链路判别** | 假 code POST → **「微信登录失败」**。判别逻辑：mock 模式对任意 code 均签发 token（openid_<code>），绝不会返回此错误；只有真实 code2session 调用微信被拒（invalid code）才触发 → **生产已走真实微信链路** ✅ |

注：第一次判别返回「参数错误」系 SSH 三层引号致 JSON 体损坏（`\"` 未还原），修正引号后复测为准。

## 附带：nginx 配置卫生清理（07:19）

- `sites-enabled/eqs.bak.20260811205051`（普通文件残留，致每次 reload 报 conflicting server name 警告）→ `mv` 归档至 `sites-available/`
- `nginx -t` 通过 → reload 成功 → h5=200 复验；此后 reload 不再有 conflicting 警告

## 遗留事项

1. 微信小程序后台：将服务器 IP 加入 request 合法域名 / 上传体验版（miniprogram-ci 需上传密钥入 GitHub Secrets）
2. 凭据批次待推进：短信/OCR（B1 #5/6）→ 微信支付商户（#7，唯一 P1）
3. Android 签名 4 Secrets（ANDROID_SIGN_BASE64 等）；iOS 待账号（BLOCKED）
4. gh CLI token 失效待 `gh auth login`（仅影响查询）
5. push main 前建议跑 CI（tag 质量门已覆盖 main push，但 PR 流仍推荐）
