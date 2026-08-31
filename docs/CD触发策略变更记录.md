# CD 触发策略变更记录（v2）

> 日期：2026-09-01 | 决策人：OPC（刘东良）| 执行：ccit-ceo
> 背景任务：EQS-OPS-20260831-01 B3 决策项（ccit-ops 提出：自动 CD 与"生产写需审批"制度冲突）

## 变更内容

| 项 | 变更前 | 变更后 |
|---|---|---|
| cd.yml 触发 | `workflow_run`（CI 成功即自动部署 main push） | **`push: tags: ['v*']` + `workflow_dispatch`**（仅 tag/手动触发） |
| ci.yml 触发 | push main + PR | 追加 `tags: ['v*']`（tag 也跑一遍 CI 供 CD 的 workflow_run 依赖，保持质量门） |

## 生效规则

1. **push main ≠ 部署**：合并后 main 只跑 CI（lint+test+build），生产零影响。
2. **打 tag = 生产授权**：`git tag v0.2.0 && git push origin v0.2.0` 触发完整 CD（build→scp→restart eqs-server + admin/h5 上传 + nginx reload）。tag 即审批凭证，由 OPC 授权或 CEO 持明确批令执行。
3. **手动兜底**：GitHub Actions 页面 `workflow_dispatch` 可手动触发（应急回滚重部署用）。

## 关联事项

- R2-2 升级（04a1b9f，线上 8/16 二进制缺失）随下次 tag 部署一并生效。
- 小程序真实化（8d815f6）不影响线上行为（mock 降级）；WX 凭据到位后无需再发版即可切真实链路（读环境变量）。
- 建议首 tag：`v0.2.0`（多端真实化 + R2-2 + CD 策略）。

## 回滚

cd.yml/ci.yml 变更本身回滚 = git revert 本次提交；生产回滚 = 保留 8/16 二进制（服务器 backup 目录）+ 手动恢复。
