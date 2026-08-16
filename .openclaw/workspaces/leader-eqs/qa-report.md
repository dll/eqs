# EQS 重构回归测试报告（qa-report）

> 回归专员：qa-regression-eqs
> 测试日期：2026-08-16
> 被测分支：`refactor/dev-cleanup`（基于 `main`）
> 测试性质：**只读验证**，未修改任何源码与测试语义
> 依据：`pm-checklist.md`（R1-1/R1-3 范围 + B-1~B-9 红线）+ `refactor-notes.md`

---

## 一、重构范围核对（来自 refactor-notes.md）

本次重构仅在 handler 层做**纯文件拆分（只移动不改写）**，涉及两块：

1. **R1-1 config.go（530 行）→ 7 个子文件**：`config_cache.go` / `config_center.go` / `config_user.go` / `config_theme.go` / `config_i18n.go` / `config_platform.go` / `config_version.go`（移出 21 个函数）。
2. **R1-1 file.go（405 行）→ 批注域移入新文件 `annotation.go`**（`AddAnnotationRequest`、`AddAnnotation`、`ListAnnotations`、`ResolveAnnotation`）。

`main.go setupRouter` 的 133 条路由**未改动**（R1-3 评估后不改）。其余大文件（project/order/payment/dispute/qualification/demo）与红线保护域（scope/payment/order/ai/channel 等）一律未触碰。

**commit 清单（相对 main）：**
`6e9fa1a` 重构基线 + gitignore；`f207a66` config.go 拆分；`c09efd0` file.go 批注拆分；`ef508bc` config_i18n.go gofmt 对齐（纯格式化）。

---

## 二、测试环境

| 项 | 值 |
|----|----|
| 系统 | Windows 10 (26200)，pwsh |
| Go | go1.26.0 windows/amd64 |
| 工作区 | `E:\2026-2027\2026-2027-1\MyProjects\eqs\packages\server` |
| 分支校验 | `refactor/dev-cleanup`，`working tree clean` ✓ |

---

## 三、测试命令与结果

### 3.1 代码状态与拆分数值核对
```powershell
cd E:\2026-2027\2026-2027-1\MyProjects\eqs\packages\server
```
- 分支 = `refactor/dev-cleanup`，无未提交改动。✓
- `git diff main..refactor/dev-cleanup --stat`：11 文件变更 = 678 insertions / 627 deletions，改动面仅限 handler 层 config/file 拆分与 .gitignore。✓

### 3.2 构建 / 静态检查 / 测试（全部通过）
| 检查 | 命令 | 结果 |
|------|------|------|
| 构建 | `go build ./...` | ✅ EXIT 0 |
| 静态检查 | `go vet ./...` | ✅ EXIT 0（handler 定向 vet 亦通过）|
| 全量测试（缓存） | `go test ./...` | ✅ 7 包全 ok |
| 全量测试（强制重跑） | `go test -count=1 ./...` | ✅ 7 包全 ok |

**强制重跑明细（-count=1，绕过缓存，确保针对拆分后新文件真实执行）：**
```
ok  github.com/eqs/server/cmd/server        10.689s
ok  github.com/eqs/server/internal/channel  13.992s
ok  github.com/eqs/server/internal/config   6.581s
ok  github.com/eqs/server/internal/dxf      7.425s
ok  github.com/eqs/server/internal/handler  13.655s（含约 180 个 handler 测试）
ok  github.com/eqs/server/internal/middleware 9.858s
ok  github.com/eqs/server/internal/model    4.017s
```

**handler 定向测试实测**（`go test -count=1 -v ./internal/handler/`）：约 180 个测试函数全部 `--- PASS`，**无一 FAIL / SKIP**。覆盖到本次重构直接相关的用例：
- config 域：`TestConfigCenter_AdminCRUD / AdminPermission / ThemeAndI18n / UserPrefs / Version`、`TestPlatformLinks`、`TestVersionLatest`、`TestAdminPublishVersion`、`TestParseConfigValue`、`TestSetProjectTheme(+/Forbidden)`。
- 文件/批注域：`TestFileUploadAndAnnotation`、`TestAddAnnotation_FileNotFound`、`TestResolveAnnotation_NotFound`、`TestListFiles`、`TestPreviewDWG_Unconfigured`、`TestPreviewFilePublic`、`TestConvertCADExternal`、`TestDownloadContract`。

### 3.3 133 条路由绑定核对（最关键回归点）
- **路由条数**：对 `setupRouter` 三个路由组（`api`/`auth`/`admin`）逐行统计 `(GET|POST|PUT|DELETE|PATCH)` 注册：
  - `api`(公开) 16、`auth`(登录态) 90、`admin`(管理端) 27、admin.DELETE 1 ∈ admin —— **TOTAL = 133**，与 pm-checklist 一致（公开~15 / 登录~70 / 管理~25，量级相符）。
- **符号完整性**：`main.go` 中共引用 135 个唯一 `handler.Xxx` 函数（133 条路由 handler + `MonitorMiddleware`/`VersionRateLimit` 中间件引用），逐一在 handler 包源码（排除 *_test.go）中 `func Xxx(` 匹配 —— **135/135 全部存在，无缺失符号**。
- handler 包所有导出函数无重复定义（213 个导出函数无同名 dup）。
- 由于 `go build ./...` 通过，`setupRouter` 中所有路由绑定均编译期解析成功，**B-1 接口契约路由绑定未因拆分丢失或改变**。✓

### 3.4 业务逻辑未退化（纯移动验证）
对拆分前后做**空白不敏感逐字节内容比对**（忽略空格/空行/gofmt 对齐）：

1. **config.go → 7 子文件**：旧 config.go（import 头之后）与新 7 文件合并内容，折叠全部空白后逐行 `Compare-Object` —— **完全一致（IDENTICAL）**。唯一差异是旧版 `config_i18n.go` 内 i18n map 的空格对齐，属 gofmt（commit `ef508bc`）纯格式，非逻辑变化。
2. **file.go 批注 → annotation.go**：旧 file.go 中 `AddAnnotationRequest` 结构体至文件末尾段，与新 annotation.go 折叠空白后比对 —— 仅 6 处差异，全部为**文件头/import 块 + 注释分隔行**（`//====文件批注====` vs `// ===== 文件批注 =====`），**函数体（struct 字段、3 个批注函数逻辑）逐字节一致**。

结论：文件拆分确为**纯移动**，函数名、签名、导出路由引用、业务逻辑均未改变。✓

---

## 四、回归风险点（评估）

| 风险点 | 评估 | 结论 |
|--------|------|------|
| 拆分误删函数/改签名 → 路由绑定失效 | `go build` 编译期全解析通过；135/135 符号存在；handler 包无重复导出 | ✅ 无风险残留 |
| 拆分子文件 import 冗余/缺失 | `go vet` + `go build` 通过 | ✅ 无 |
| 纯移动中夹带逻辑改动 | 空白不敏感逐字节比对：config 全部 IDENTICAL；annotation 仅文件头/注释差异 | ✅ 无逻辑变更 |
| 既有测试语义失效（B-9） | 全量测试 `-count=1` 强跑 7 包全 ok，handler 约 180 测试全 PASS | ✅ 未失效 |
| i18n gofmt 对齐破坏文案 | 折叠空白后 map 键值对内容一致（仅对齐空格变化） | ✅ 无 |
| 前端/迁移/通道层受影响 | 本次仅改后端 handler 层 2 个文件，client/admin/channel/migrations 零改动 | ✅ 无波及 |

**未纳入本次验证（超出重构范围，供评审留意）：**
- R2-1 N+1 / 索引（scope.go）、R3-1 状态常量收敛、R1-5 前端拆分等**遗留建议项**未在本次重构实施，不存在对应回归。

---

## 五、结论

> **✅ 可进入评审。**

- 工作区处于 `refactor/dev-cleanup`，干净无脏。
- `go build ./...`、`go vet ./...`、`go test -count=1 ./...` **全量通过**（7 包全 ok，handler 约 180 测试全 PASS）。
- **133 条接口路由绑定完整无丢失、无改变**；135 个 main.go 引用的 handler 符号全部定义解析成功。
- config.go/file.go 拆分经**空白不敏感逐字节比对**证实为**纯移动（只移动不改写）**，业务逻辑零变化。
- 既有 `*_test.go` 断言语义未失效。
- 未修改任何源码与测试语义；未触发任何红线（B-1~B-9）。

---

## 六、遗留问题 / 建议

1. **无阻塞性问题**。所有回归通过。
2. **建议（非回归）**：refactor-notes §6 遗留项（R2-1 N+1/索引、R1-5 前端大组件、R1-4 shared 包收编、R3-1 状态常量收敛）均属后续独立专项，建议由 PM 排期，与本重构解耦评审。
3. **验收提示**：本次验证覆盖编译/单测/路由绑定/纯移动比对；未运行运行时冒烟（需 DB + 凭据），若评审要求更高置信可再补一次手工冒烟（启动 server + 抽查公开接口）。
4. **可回滚性**：`f207a66`/`c09efd0`/`ef508bc` 均独立 commit，评审若需回退可单条 `git revert`，不影响其他改动。

---

*完成：2026-08-16，qa-regression-eqs（只读回归）。*
