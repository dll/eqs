# EQS L1-A 横切盘点 —— leader-eqs（现库现状、证据、缺口、建议）

> 盘点人：leader-eqs（EQS 项目 L1-A 负责人，承 CTO《EQS-L1-A-角色矩阵与账号落法》+《EQS-L0-验收结论》派工）
> 盘点日期：2026-09-07
> 盘点基线：main （commit 21cf6af 前，133 路由原始；本盘点为只读审计，不动交易域）
> 方法：只读代码取证 + 回归确认；净化环境。先盘点后代码；涉及交易域改动一律走 CTO 门禁。
> 主线：按真人"排除约束尽快上线"，本盘点作为 L1-A 加固的最小依赖输入，为"多角色平台（甲方/服务方/OPC真人/公司侧AI agent）"提供落地锚。

---

## 0. 结论速览（TL;DR）

| 盘点项 | 现状结论 | 首要缺口 | 建议等级 |
|---|---|---|---|
| 授权/RBAC × 四主体 | 已具 user_type(1甲方/2服务方/3平台运营/4专家)，自注册仅甲方/服务方 | **无组织/公司模型**(无 Company/OrgID)，无法"同一项目内多人/多组织隔离"；**无 agent principal** | 橙(改动需门禁) |
| audit | 51 类动作可写(含今日存量项目归档动作) | **无读取/检索端点**，监管/OPC 无法走 API 查审计；无 request-id 关联 | 检索端点 **已补**(0e37b4f)；request-id/agent 维度待后续(橙) |
| 配置公私边界 | 有 is_public + /config/public 缓存 + 管理侧 upsert | 私密边界靠单 admin 语义；环境注入 `os.Getenv` 分散于 config.Load | 黄 |
| 错误契约 / 幂等 | 统一 ok/fail/response helper | 写操作普遍无幂等键；批量我今日已事务化 | 黄 |

每条建议落为《变更影响登记》→ 过 CTO 门禁后做**最小加法**，不碰红区(资金/dispute/AI)与交易状态机。

---

## 1. 授权 / RBAC × 四主体

### 1.1 现状（证据）
- 主体单一维度 `User.UserType int`（`internal/model/user.go`）：1=甲方/需求方，2=服务方，3=平台运营(即 admin/内部监管+seed 运营)，4=专家(争议评审)。
- 自注册白名单 `var publicRoles = map[int]bool{1:true, 2:true}`（auth.go，L34）——平台外只允许甲方/服务方自助账号；3/4 受控创建（demo/后台）。
- 路由侧三重守卫：
  - `RequireAdmin()`（middleware）→ `user_type==3`，服务所有 `/admin/*`（当前 ~20 处 admin 端点的收口）。
  - `isAdmin(c)` helper（scope.go L22）→ `user_type==3`，供各业务文件内联放行（23 处调用）。
  - 资源归属守卫：`X.UserID == c.GetUint("user_id")` 或 `SupplierID == userID`（38 处 owner 判断）——对象级鉴权，覆盖 project/order/bid/case/dispute/qualification/file 等。
- 归类统计（业务文件，排除 _test）：含 `user_id` owner 判断 **38** 处，`isAdmin(c)` 调用 **23** 处，admin handler 函数 **~20** 个。

### 1.2 对照 CTO 四主体需求的缺口
| 主体 | 现行能力 | 缺口 |
|------|----------|------|
| 甲方(需求方) | 自助注册 user_type1，owner=本人 | **无组织**:同公司多人无法共享/隔离"本组织"资源(CTO:"仅本人或本组织") |
| 服务方 | user_type2，owner=本人 supplier 视角 | 同上，"所属组织被授权资源"无组织层 |
| OPC 真人 | 语义上≈user_type3(平台运营) | 3 同时兼"admin 超管"过宽(CTO:OPC 不默认 admin/不持资金与凭据管理)；且 demo 把 3 当"EQS平台运营"。"OPC 监管只读+审批、非 admin"的细分角色不存在 |
| 公司侧 AI agent | **无** | 仓库无 agent principal/服务身份/固定 scope/独立 audit 主体——agent 目前只能"借"某个人类账号，违背 CTO"不得共享 admin"要求 |

> 证据核心：仓库内无任何 `Company/OrgID/DeptID/组织` 模型或字段；`scope.go` 无组织级隔离逻辑；无 `agent_id`/`principal`/服务账号概念（见代码检索均空）。`CompanyName` 为展示用自由串。

### 1.3 建议（最小、加法、不触交易域）
1. **新增 `Organization` + `UserOrg`/`OrgMember`（组织模型，加法 + AutoMigrate 追加）**：为甲/乙建立"所属组织"，`Project.OwnerOrgID`、`Supplier.OrgID` 以加法收敛到"资源归属 = 本人 或 本组织成员"，替代纯 `==userID` 的原子见点——不改既有单人 owner 语义（保持原有），只加组织层合并。
   > 注：属**新增模型**，会增 AutoMigrate 成员 → 按 L0 规则触发 CTO 门禁；作为 L1-A 建议项列此，不擅自先行。
2. **新增 `AgentPrincipal`（公司侧 AI agent 服务身份，只做加法）**：为 AI agent 记账"独立身份 + 白名单 scope(读取/只读、按域 API) + 逐动作 audit 标注 `actor_type=agent/actor_id`"，禁止 agent 借用人类账号登入 admin。→ CTO 需纳入真实 AI agent 清单后裁定名单。
3. **将 `user_type=3` 从"超管单角色"逐步拆分为 RBAC 权限数组/矩阵**（加法：新增 `user_role`/`Permission` 关系表，保留既有 user_type 不加删），供 OPC"监管只读+审批集"与"admin 系统级"分开授权——**不改红区、不改现有鉴权断言**，作为 L1-A 后续批次。

### 1.4 影响与回滚
- 净新增文件/表 → 回滚=删除新增迁移；对既有鉴权零改，SAR 回归不受损。

---

## 2. Audit 审计链

### 2.1 现状（证据）
- 统一落点 `handler.WriteAudit(c, action, targetType, targetID, detail)`（`audit.go`；DB 未初始化/超时 3s 安全跳过不阻断主流程）→ 写 `AuditLog`：`UserID/Action/TargetType/TargetID/Detail(JSON)/IP/CreatedAt`。
- 动作面广：跨 16 个域 **51 类** action（order/pay/dispute/contract/bid/milestone/config/admin/project/qualification/case/member/attendance/tools/user/version/**archive**）。今日存量项目归档动作 `archive.project.*` 亦已挂载。
- `user.login` / `pay.*` 敏感动作均留痕。

### 2.2 缺口（证据）
- **无任何审计读取/检索端点**（main.go 无 audit 路由、handler 无 AuditList/GetAudit）——OPC/监管无法按 API 查审计，只能直连 DB；违背"审计可查看、还原关键状态变更"的落地。
- 无 request-id：`AuditLog` 无请求关联标识列，一笔用户操作若触发多动作，难以串联还原（CTO:请求关联标识）。
- 无 actor 维度（agent/真实用户/脚本）：当前只记 UserID，agent 主体无法区分。

### 2.3 建议（加法、橙色、改动需入登记）
1. **新增审计检索端点**（读 /admin/audit 或受限 /audit/*）：按 action/targetType/targetID/user/时段/actor 过滤 + 分页；**只读聚合**，不泄详情外的原始敏感字段（脱敏仅管理侧）。→ 新增端点=门禁。
2. `AuditLog` 追加字段 `RequestID string` + `ActorType`(`user/agent`)的**加列**，由中间件透传 request id。
   > 加列=影响 AutoMigrate 既有成员吗：gorm AutoMigrate 会为既有表**加新列**(非删改) → 属于"不影响既有 28 模型成员但会改 red区表结构除外"——AuditLog 非红区，属简单加列风险低，仍入登记。
3. 保留现有 WriteAudit 方法签名，仅在其内补齐 RequestID/ActorType（不改调用方=纯加法）。

---

## 3. 配置公私边界与环境注入

### 3.1 现状（证据）
- `SystemConfig{Key,Value,IsPublic}`；`is_public=true` 仅 `/config/public`（走缓存，`config_cache.go` 单飞锁修复过并发）对全端只读；私密其走 `AdminUpsertConfig`/`AdminListConfigs`（admin 组）。
- 环境注入：`internal/config/config.go` 在 `Load()` 内经 `os.Getenv` 读一大组变量（JWT/DB/Redis/WX/密钥/AI 等）。本机开发注意：他项目（wxx）继承变量会污染（见 L0 复跑报告 §3.1 已记录净化）。

### 3.2 缺口
- "公开/私密"边界当前依赖 `is_public` 布尔 + admin 收口，粒度粗；Key 命名无命名空间强制校验，易误配 public。
- 生产弱凭据与 DATA_ENCRYPTION_KEY 强校验在 `main.go` 启动段有（IsProduction），属防护完备点 ✓。

### 3.3 建议（黄色）
- 对 `is_public` upsert 加 Key 前缀白名单/黑名单校验（如 `secret.*`/`*key*`/`password` 禁 public），避免误公开敏感项。
- 不强推环境变量重构；仅把"凭据类不回填文档/聊天"作为纪律（合规已 OK）。改动小、低风险，作为 L1-B 例行。

---

## 4. 错误契约 / 幂等

### 4.1 现状（证据）
- 统一成功/错误 helper（response.go）：`ok/created/fail/badRequest/unauthorized/notFound/forbidden/serverError`；生产不向客户端回显内部 err（serverError 已克制泄露）✓。
- 分页统一 `parsePage`(size≤100)✓；`parseUint` 隔离。
- 幂等：**无通用写幂等键**；多数 POST 为创建型，重复提交可产生重复记录——今日我对 **批量导入已事务化**（缺标题整批回滚）；单条创建仍无防重。

### 4.2 建议（黄色/橙色）
- 对**易重放/涉钱敏感已有红区不改**；对新增端点（如存量归档 create）长期可由前端去重+后端唯一约束兜底。当前不强制引入全局幂等键框架（牵动面大、属 L2），先以"单测覆盖关键创建路径的重复提交防护策略清单"方式列档，待 L2 统一定位。

---

## 5. 各建议落为后续批次（均需 CTO 门禁的登记）

| 建议 | 触碰 | 门禁文件 |
|------|------|----------|
| Organization 组织层（RBAC 落点） | 新增模型+端点 | 《变更影响登记》待开 |
| AgentPrincipal（公司侧身份） | 新增模型+端点 | 待 CTO 提供 agent 名单后开 |
| user_type3 拆 RBAC / OPC 细分角色 | RBAC 加法 | L1-A 后续 |
| 审计检索端点 + RequestID/ActorType 加列 | AuditLog 加列 + 新端点 | 待开 |
| config public 前缀防护 | Upsert 校验加法 | L1-B |
| 幂等策略清单 | 文档/单测 | L2 |

> 规则：每一个改动独立《变更影响登记》；不改红区；不改交易状态机/落库值/既有 AutoMigrate 既有成员(除 AuditLog 加列)；回归净化环境 7 包全绿；QA 留证据；reviewer 专项查凭据与越权。

---

## 6. 交给 CTO / CEO / 真人的决策项（非技术实现，按流程转）
1. 现库**无公司/组织模型**——L1-A 若要"甲方/服务方本组织隔离"，是否立即立项 Organization 组织层？建议：是（多角色平台的根前提）。若暂缓，维持"本人 owner + admin"最低口径先上线，隔离留待二期。
2. 公司侧 AI agent 使用需**独立 agent 身份**——请提供首批需接入的 agent 清单（CCIT 各 leader/运营位）与其需要的作用域(读/写范围)，以便 L1-A 排 `AgentPrincipal` 名单。
3. OPC：`user_type=3` 现含超管语义过宽——若 OPC 用普通账号即够监管/审批，请确认 L1 是否把"监管只读+审批"与"系统 admin"权限分离。

*—— leader-eqs，L1-A 盘点 v1（只读取证），2026-09-07。待补充项可在后续轮次按 CTO 门禁落代码。*
