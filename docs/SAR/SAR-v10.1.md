# 工程快捷服务（EQS）软件审核报告 V10.1（发布前加固轮）

> 文档版本：V10.1
> 审核日期：2026-08-31
> 审核基线：`docs/PRD/PRD-v12.0.md`、`docs/SAR/SAR-v10.0.md`、当前工作区代码（`main` 分支）
> 审核方式：全量构建/测试/静态检查实测 + 遗留项（R2/R3）逐项复核
> 结论方向：发布产品加固，非功能新增

---

## 0. 执行摘要

**V10.1 结论：V10 遗留的「可做代码项」中，仅 R2-2（配置缓存冷启动并发）为真实缺陷，已修复并通过全量回归；其余遗留项经逐项复核判定为「非缺陷 / 高风险低收益」，予以归档说明，暂不改动。全链路（后端 7 包测试 + 前端双端生产构建）绿通过，达到可发布状态。**

## 1. 本轮改动

### 1.1 已修复：R2-2 配置缓存冷启动并发重复加载

- **文件**：`packages/server/internal/handler/config_cache.go`
- **缺陷**：`getPublicCached` 在空缓存时 `RUnlock` 后调用 `loadPublicCache`，多请求并发冷启动会触发多次重复 DB 查询（`WHERE is_public = true`）。
- **修复**：引入独立单飞锁 `publicCacheLoadMu`（`sync.Mutex`）+ `loadPublicCache` 内双检；命中缓存路径保持无锁读（`RLock`）。
- **为何不用 `sync.Once`**：`invalidatePublicCache` 需要清空缓存并允许下次重新加载；复制/重置已使用的 `sync.Once` 存在数据竞争，故采用独立互斥量 + 双检，规避该隐患。
- **验证**：`go build`、`go vet`、`go test ./...` 7 包全绿；`gofmt` 干净。

## 2. 遗留项复核结论（不改动）

| 编号 | 项 | 复核结论 | 处理 |
|------|----|----------|------|
| R2-1 | scope.go 鉴权 N+1 / 热点索引 | 列表热路径（ListMyOrders/ListMyProjects/ListMyBids）均按 `user_id`/`supplier_id` 直接过滤，**不存在逐行 N+1**；`isOrderParticipant` 仅在详情/单对象路径各触发一次额外查询，非热点 | 不改动，归档 |
| R2-2 | 配置缓存并发 | 真实缺陷 | ✅ 已修复（见 1.1）|
| R3-1 | 状态机魔法数字收敛 | 涉及 auth/order/payment/bid 等多域 ~20 处跨切面改动，触及资金与状态机红线，任何常量映射错误即静默变更业务语义 | 随专项评审推进，发布前不改 |
| R1-5 | 前端大组件拆分 | 纯可读性打磨，非功能缺口 | 不改动 |
| R1-4 | shared 包收编 | 依赖前端改版计划 | 不改动 |

## 3. 全量验证实录

| 项 | 命令 | 结果 |
|----|------|------|
| 后端构建 | `go build ./...` | ✅ EXIT 0 |
| 后端静态检查 | `go vet ./...` | ✅ EXIT 0 |
| 后端全量测试 | `go test ./...` | ✅ 7 包全 ok（cmd/channel/config/dxf/handler/middleware/model）|
| Admin 类型检查 | `vue-tsc --noEmit` | ✅ EXIT 0 |
| Client 类型检查 | `vue-tsc --noEmit` | ✅ EXIT 0 |
| Admin 生产构建 | `pnpm --filter @eqs/admin build` | ✅ built in 41.46s |
| Client 生产构建 | `pnpm --filter @eqs/client build:h5` | ✅ DONE Build complete |

## 4. 结论

代码功能层已完整（133 路由 / 28 模型），本轮加固修复一处真实并发缺陷，其余遗留项经复核判定为非缺陷或高风险低收益、予以归档。全链路验证绿通过，产品达到可发布状态。

剩余未完成项均属**非代码类（凭据/商务/部署）**，由公司级部署运维智能体独立承接，不在本仓库代码范围内。

---

*—— leader-eqs，发布前加固轮（V10.1）。*
