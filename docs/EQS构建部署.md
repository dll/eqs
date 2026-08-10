# EQS 多端应用构建部署文档（GitHub Actions 全自动）

> **文档版本**：V3.0
> **更新日期**：2026-08-10
> **适用项目**：工程快捷服务 (EQS)——monorepo 三端一后端（Go / Vue3 / uni-app）
> **核心原则**：所有应用产物（H5、微信小程序、Android APK、iOS IPA）**全部由 GitHub Actions 自动生成**，不依赖任何本地构建工具（不装 HBuilderX、不做本地云打包、不装 Android Studio / Xcode）。
> **版本策略**：**全端统一版本号** `X.Y.Z`，重构时迭加（`0.1.0 +1 → 0.1.1`）。iOS 仅维护 `init`（调试基座）与 `release`（上架）两个版本；其余应用（H5 / 小程序 / Android）每次 push 均可多次构建发布。

---

## 1. 项目架构与多端产物

### 1.1 仓库结构

```
eqs/
├── packages/
│   ├── server/            # Go + Gin + GORM 后端（cmd/server/main.go）
│   ├── admin/             # Vue3 + Element Plus 管理后台（vite）
│   └── client/            # Uni-app（vue3，H5 / 小程序 / App）
├── shared/                # 共享代码
├── deploy/                # nginx / systemd / deploy.sh / docker-compose.prod.yml
├── .github/workflows/     # ci.yml / cd.yml / ios.yml
└── apps/                  # DCloud 离线打包原生工程（Android + iOS，可选）
```

### 1.2 多端产物矩阵

| 端 | CI 构建命令 | 产物 | 交付形态 |
|----|-------------|------|----------|
| 用户端 H5 | `uni build --platform h5` | `dist/build/h5` | COS 静态托管 + nginx |
| 微信小程序 | `uni build --platform mp-weixin` | `dist/build/mp-weixin` | 微信开发者工具上传体验版 |
| Android APK | `uni build --platform app` → 离线 SDK gradle 打包 | `*.apk` | GitHub Artifact / 分发链接 |
| iOS init 基座 | `uni build --platform app` → 离线 SDK xcodebuild | `*.ipa`（调试基座） | 真机安装（自定义基座） |
| iOS release | 同上 + Distribution 签名 | `*.ipa`（App Store 版） | TestFlight / 上架 |
| 管理后台 | `pnpm --filter @eqs/admin build` | `dist/` | COS 静态托管 |

### 1.3 触发策略

| 端 | 触发方式 | 说明 |
|----|----------|------|
| H5 / 小程序 / Android / 后台 | **push 即触发** | 每次提交可多次构建发布 |
| iOS init | 手动 `workflow_dispatch` | 仅首次 / 基座更新时触发 |
| iOS release | 手动 `workflow_dispatch` + `version_type=release` | 仅发版时触发，选中 Secrets 为 Distribution 证书 |

> iOS 因 macOS runner 计费 10 倍且证书受控，刻意设计为**仅手动、两个版本**；其余应用不限制，可反复构建。

---

## 3. 统一版本号策略

### 3.1 版本号定义

全端（根 / client / admin / manifest）使用**同一个**语义化版本 `MAJOR.MINOR.PATCH`：

| 版本位 | 含义 | 变更时机 |
|--------|------|----------|
| MAJOR | 不兼容的大改动 | 重构架构、破坏性升级 |
| MINOR | 新增功能（向后兼容） | 发版新功能 |
| PATCH | 缺陷修复 / 小重构 | **迭加规则：`0.1.0 +1 → 0.1.1`** |

### 3.2 迭加规则（重构时）

**每次重构 / 迭代发版，`PATCH +1`，直到 MINOR 增位时归零**：

| 当前版本 | 重构 1 次后 | 重构 2 次后 | 新增功能后 |
|----------|-------------|-------------|------------|
| 0.1.0 | 0.1.1 | 0.1.2 | 0.2.0 |
| 1.2.3 | 1.2.4 | 1.2.5 | 1.3.0 |
| 2.0.0 | 2.0.1 | 2.0.2 | 2.1.0 |

### 3.3 同步位置

发布前需同步以下 4 处，保证全端一致：

| 位置 | 字段 | 说明 |
|------|------|------|
| 根 `package.json` | `version` | monorepo 总版本 |
| `packages/client/package.json` | `version` | 用户端 |
| `packages/admin/package.json` | `version` | 管理后台 |
| `packages/client/src/manifest.json` | `versionName` | App 显示版本 |
| `packages/client/src/manifest.json` | `versionCode` | App 构建号（PATCH 每 +1，`versionCode +1`） |

### 3.4 发布流程（Tag 触发 Release）

```bash
# 1. 同步三处 package.json 版本（示例：0.1.0 → 0.1.1）
# 修改根/ client / admin 的 version 与 manifest versionName/versionCode

# 2. 提交并打 tag
git add package.json packages/client/package.json packages/admin/package.json packages/client/src/manifest.json
git commit -m "chore: 版本号升至 0.1.1"
git tag -a v0.1.1 -m "EQS v0.1.1"
git push origin main --tags
```

> tag 名 `v{version}` 与全端版本号保持一致，`release.yml` 据此自动构建全部产物并发布 GitHub Release。

---

## 4. 整体流程一览

```
                 +----------------------------------------+
                 | push origin main（每端每次可多次）       |
                 +----------------+-----------------------+
                                  |
              +-------------------+-------------------+
              |                                       |
   +----------v-----------+              +------------v------------+
   | CI (ubuntu)          |              | CD (ubuntu)             |
   | 后端 vet+test+build  |              | CVM: server 部署       |
   | admin build          |              | COS: admin H5 部署     |
   | client H5            |              | COS: client H5 部署    |
   | client mp-weixin     |              +-------------------------+
   | client app 资源      |
   | Android APK 打包签名 |
   | 产物 upload artifact |
   +----------+-----------+
              |
   +----------v-----------+
   | 手动 workflow_dispatch (ios.yml)   |
   | version_type: init|release         |
   | macOS: app资源→xcodebuild→IPA       |
   +------------------------------------+
```

---

## 5. 步骤 1：仓库与 Secrets 准备

### 5.1 代码托管

1. 创建 **GitHub 私有仓库**（如 `your-org/eqs`）。
2. 推送现有代码：

```bash
git remote add origin git@github.com:your-org/eqs.git
git push -u origin main
```

3. 确认 `.gitignore` 排除敏感/运行时文件：
   - `packages/server/*.db`（SQLite 运行时数据库）
   - `packages/client/dist/`、`packages/admin/dist/`
   - `.env*`、`*.jks`、`*.p12`、`*.mobileprovision`

### 5.2 签名材料（一次性准备，仅存 Secrets）

| 材料 | 用途 | 生成方式（不需要 Mac） |
|------|------|------------------------|
| `eqs-release.jks` | Android 签名 | `keytool`（Windows/任意 JDK） |
| 开发 p12 + 开发描述文件 | iOS init 基座 | Appuploader（Windows） |
| Distribution p12 + AppStore 描述文件 | iOS release 上架 | Appuploader（Windows） |

转 Base64 存入仓库 Secrets：

```powershell
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("eqs-release.jks"))
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("dev.p12"))
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("dev.mobileprovision"))
```

### 5.3 Secrets 清单

仓库 → Settings → Secrets and variables → Actions：

**Secrets（Actions secrets）**

| Secret 名 | 内容 | 必填 |
|-----------|------|------|
| `ANDROID_SIGN_BASE64` | Android jks 的 Base64 | ✅ 出 APK 时 |
| `ANDROID_KEYSTORE_PWD` | 密钥库密码 | ✅ |
| `ANDROID_ALIAS` | 别名 | ✅ |
| `ANDROID_ALIAS_PWD` | 别名密码 | ✅ |
| `IOS_DEV_P12_BASE64` | iOS 开发证书 p12 Base64 | ✅ init |
| `IOS_DEV_PROVISION_BASE64` | 开发描述文件 Base64 | ✅ init |
| `IOS_DIST_P12_BASE64` | Distribution p12 Base64 | ✅ release |
| `IOS_DIST_PROVISION_BASE64` | AppStore 描述文件 Base64 | ✅ release |
| `IOS_P12_PASSWORD` | p12 密码（开发/发布共用） | ✅ |
| `IOS_BUNDLE_ID` | iOS Bundle ID（如 `com.eqs.app`） | ✅ |
| `CVM_HOST` | CVM 公网 IP | ✅ **已配**（129.211.223.113） |
| `CVM_USERNAME` | SSH 用户名 | ✅ **已配**（root） |
| `CVM_SSH_KEY` | SSH 私钥内容 | ✅ **已配**（`~/.ssh/wxx_deploy.pem`） |
| `TENCENT_CLOUD_SECRET_ID` | 腾讯云 API 密钥 ID | ⏳ 待获取 |
| `TENCENT_CLOUD_SECRET_KEY` | 腾讯云 API 密钥 Key | ⏳ 待获取 |

**Variables（Actions variables）**

| 变量名 | 内容 | 必填 |
|--------|------|------|
| `ENV` | 环境标识（如 `prod`），决定 COS bucket 名 `eqs-admin-prod` | ⏳ 待设置 |
| `VITE_API_BASE_URL` | admin 构建时后端 API 地址 | ⏳ 待设置 |

> **获取腾讯云 API 密钥**：腾讯云控制台 → 右上角头像 → 访问管理 (CAM) → 访问密钥 → API 密钥管理 → 新建密钥（SecretKey 仅显示一次）。

---

## 6. 步骤 2：CI 全自动构建（H5 / 小程序 / Android / 后台）

> 现有 `ci.yml` 覆盖 lint + test + H5 构建；`cd.yml` 负责部署。
> 下面补全 **Android APK 自动打包** 与 **多端产物归档**，均 `push` 触发、可多次构建。

### 6.1 构建矩阵

| 端 | 触发 | Workflow | 命令 | 产物 |
|----|------|----------|------|------|
| 后台 Admin | push | ci.yml | `pnpm --filter @eqs/admin build` | `dist/` |
| H5 | push | ci.yml | `pnpm --filter @eqs/client build:h5` | `dist/build/h5` |
| 微信小程序 | push | ci.yml | `pnpm --filter @eqs/client build:mp-weixin` | `dist/build/mp-weixin` |
| App 资源 | push | android.yml | `uni build --platform app` | `dist/build/app` |
| Android APK | push | android.yml | 离线 SDK Gradle 签名 | `*.apk` |
| iOS init / release | 手动 | ios.yml | 离线 SDK xcodebuild | `*.ipa` |

### 6.2 补齐 ci.yml：多端产物上传

在 ci.yml 的 `build-client` job 追加小程序与 App 资源产物上传：

```yaml
      - name: Build client mp-weixin
        working-directory: packages/client
        run: pnpm --filter @eqs/client build:mp-weixin

      - name: Build client app resources
        working-directory: packages/client
        run: pnpm --filter @eqs/client build:app

      - name: Upload mp-weixin
        uses: actions/upload-artifact@v4
        with:
          name: mp-weixin-dist
          path: packages/client/dist/build/mp-weixin

      - name: Upload app resources
        uses: actions/upload-artifact@v4
        with:
          name: app-resources
          path: packages/client/dist/build/app
```

### 6.3 Android APK 全自动打包（Ubuntu + 离线 SDK，push 触发）

前置：在仓库维护 `apps/android/`（DCloud 离线 SDK Gradle 原生工程，一次性放入即可）。
CI 自动完成：构建 App 资源 → 拷贝进 assets → Gradle 构建签名 APK。

```yaml
name: Build Android APK

on:
  push:
    branches: [main]
  workflow_dispatch:

concurrency:
  group: eqs-android-${{ github.ref }}
  cancel-in-progress: true

jobs:
  apk:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Setup pnpm
        uses: pnpm/action-setup@v4
        with:
          version: 9

      - name: Install dependencies
        run: pnpm install

      - name: Build App resources
        working-directory: packages/client
        run: pnpm --filter @eqs/client build:app

      - name: Upload app resources
        uses: actions/upload-artifact@v4
        with:
          name: app-resources
          path: packages/client/dist/build/app

      - name: Copy resources into native project
        if: hashFiles('apps/android/**') != ''
        run: |
          mkdir -p apps/android/app/src/main/assets/apps
          rm -rf apps/android/app/src/main/assets/apps/*
          cp -r packages/client/dist/build/app/* apps/android/app/src/main/assets/apps/

      - name: Restore signing keystore
        if: hashFiles('apps/android/**') != ''
        run: echo "$ANDROID_SIGN" | base64 -d > apps/android/eqs-release.jks
        env:
          ANDROID_SIGN: ${{ secrets.ANDROID_SIGN_BASE64 }}

      - uses: actions/setup-java@v4
        if: hashFiles('apps/android/**') != ''
        with:
          distribution: temurin
          java-version: '17'

      - name: Build signed APK
        if: hashFiles('apps/android/**') != ''
        working-directory: apps/android
        run: ./gradlew assembleRelease
        env:
          ANDROID_KEYSTORE_PWD: ${{ secrets.ANDROID_KEYSTORE_PWD }}
          ANDROID_ALIAS: ${{ secrets.ANDROID_ALIAS }}
          ANDROID_ALIAS_PWD: ${{ secrets.ANDROID_ALIAS_PWD }}

      - name: Upload APK
        if: hashFiles('apps/android/**') != ''
        uses: actions/upload-artifact@v4
        with:
          name: eqs-release
          path: apps/android/app/build/outputs/apk/release/*.apk
```

> 若离线 SDK 工程（`apps/android`）尚未维护，该 job 经 `if: hashFiles` 自动跳过——不报错、不占额度。补齐工程后 push 即自动出 APK。

---

## 7. 步骤 3：iOS 两版本（仅手动触发）

> 文件：`.github/workflows/ios.yml`（已提供）。
> `push` 永不触发 iOS；只允许网页手动 `Run workflow`，规避 macOS 10 倍计费与证书泄露。

### 7.1 版本对比

| 项 | init（调试基座） | release（上架） |
|----|------------------|-----------------|
| 用途 | 真机调试、热重载 | TestFlight / App Store |
| 签名证书 | 开发证书（dev p12） | Distribution p12 |
| 描述文件 | 开发描述文件（含 UDID） | AppStore 描述文件 |
| 导出方式 | ad-hoc | app-store |
| 触发 | 手动，首次/基座更新 | 手动，发版时 |
| 所需 Secrets | `IOS_DEV_*` | `IOS_DIST_*` |

### 7.2 触发步骤

1. 仓库 → **Actions** → **iOS Build** → **Run workflow**。
2. 选择 `version_type`：
   - `init`：生成调试基座 IPA → 下载用爱思助手真机安装。
   - `release`：生成上架 IPA → Transporter 上传 TestFlight。
3. Artifact 自动区分：`ios-ipa-init` / `ios-ipa-release`。

### 7.3 证书自动切换

`ios.yml` 依据 `version_type` 选择 Secrets 与导出方式：

| 分支 | 证书 | 描述文件 | export method |
|------|------|----------|---------------|
| init | `IOS_DEV_P12_BASE64` | `IOS_DEV_PROVISION_BASE64` | ad-hoc |
| release | `IOS_DIST_P12_BASE64` | `IOS_DIST_PROVISION_BASE64` | app-store |

---

## 8. 步骤 4：部署与分发

### 8.1 后端与 Web 自动部署（cd.yml，push 自动）

| Job | 目标 | 内容 |
|-----|------|------|
| `deploy-server` | CVM | `git pull` → `go build`（CGO_ENABLED=1）→ 重启 `eqs-server` |
| `deploy-admin` | COS | `coscli sync` → `eqs-admin-${ENV}` |
| `deploy-client` | COS | `coscli sync` → `eqs-client-${ENV}`（相机源 `dist/build/h5`） |

### 8.2 移动端分发

- **Android**：GitHub Actions → Actions 页 → `eqs-release` Artifact 下载 → 安装/分发给测试
- **iOS init**：`ios-ipa-init` 下载 → 爱思助手真机安装
- **iOS release**：`ios-ipa-release` 下载 → Transporter 上传 TestFlight → 提交审核

---

## 9. 发布（GitHub Release，Tag 触发）

### 9.1 release.yml 自动发布

文件：`.github/workflows/release.yml`（已落地）。打 tag `v*` 时自动：

| 阶段 | 内容 |
|------|------|
| 构建 | server（CGO）、admin、H5、小程序、App 资源，全端同一版本 |
| 组装 | 统一收进 `release/`：`eqs-server`、`admin-web/`、`eqs-h5/`、`eqs-mp-weixin/`、`eqs-app-resources/` |
| 发布 | 生成 GitHub Release（draft），附带全部产物 + 自动 release notes |

### 9.2 发布步骤（结合统一版本号）

```bash
# 1. 升级版本号（0.1.0 → 0.1.1，见第 3 章），同步 4 处
# 2. 提交并推送 tag
git add package.json packages/client/package.json packages/admin/package.json packages/client/src/manifest.json
git commit -m "chore: 版本号升至 0.1.1"
git tag -a v0.1.1 -m "EQS v0.1.1"
git push origin main --tags
```

### 9.3 发布清单核对

发布前确认：

1. ✅ CI（ci.yml）通过——各端 lint / test / build 全绿
2. ✅ CD（cd.yml）已部署——CVM / COS Secrets 已配置
3. ✅ Android/iOS 签名 Secrets 已配置（如出实包）
4. ✅ `release.yml` 构建的 server 二进制可运行（CGO 静态）

---

## 10. 计费与优化

| 场景 | 消耗 |
|------|------|
| push 自动（ubuntu） | 1 倍/次，免费额度（2000 min/月）覆盖 |
| iOS 手动 1 次 | ≈ 10~12 min × 10 倍 = 100~120 min |
| 每月 3~5 次 iOS | ≈ 300~600 min，远在免费额度内 |

- iOS 日常迭代用热重载，不重复打基座。
- Android / H5 / 小程序可反复 push 构建发布，无额外成本。

---

## 11. 避坑清单

1. **签名绝不进仓库**：p12/jks/mobileprovision 只存 GitHub Secrets。
2. **iOS push 永不上 macOS**：一律手动触发，靠 `workflow_dispatch`。
3. **普通 uni-app ≠ uni-app x**：`uni build -p app` 产出 App 资源，依赖离线 SDK 原生工程（`apps/android`、`apps/ios`）二次打包成 APK/IPA。
4. **Go CGO 关键**：server 依赖 `github.com/mattn/go-sqlite3`（CGO）；cd.yml 与 Dockerfile 须 `CGO_ENABLED=1`（已修正）。
5. **pnpm monorepo**：CI 用 `pnpm/action-setup@v4` 安装（勿用 `setup-node` 的 `cache: pnpm`——pnpm 未装会直接报错；勿用 `npm install -g pnpm`，独立 step 的 PATH 不可靠）。
6. **COS 路径**：client 产物真实路径为 `packages/client/dist/build/h5/`（cd.yml 已修正，勿回改）。
7. **coscli 资产名**：下载地址必须用实际资产名 `coscli-v1.0.8-linux-amd64`（`coscli-linux` 不存在，会 404 导致部署失败）。
8. **部署凭据前置**：CD 的 `deploy-server` / `deploy-admin` / `deploy-client` 依赖 Secrets（CVM_HOST/USERNAME/SSH_KEY、腾讯云 ID/KEY）与 Variables（ENV）。未配置时 job 直接失败（`missing server host`），配置完整才跑通。
9. **ad-hoc 必须预登记 UDID**，否则 init 基座真机无法安装。
10. **Android/iOS 离线 SDK**：仓库未维护原生工程时，android job 自动跳过、iOS 仅产出 App 资源——补齐工程后 push/手动即自动出 APK/IPA。

---

## 12. 演进建议（可选）

- 申请 DCloud 离线打包 SDK，维护 `apps/android` 与 `apps/ios` 原生工程，即可实现全端 CI 自动出安装包。
- 增加 `workflow_call` 复用构建步骤，减少维护成本。

---

**文档结束**



