# EQS 构建部署文档（CI/CD/CD 全流程）

> **文档版本**：V1.0  
> **创建日期**：2026-08-10  
> **适用项目**：工程快捷服务 (EQS)——monorepo 三端一后端  
> **配套文件**：`.github/workflows/build.yml`（CI/CD 流水线惯例）、`deploy/`（CVM 部署脚本）、`packages/client`（uni-app）  
> **参考基线**：GitHub 私有仓库 + Ubuntu 打包安卓 / iOS 仅初始化与发版手动触发

---

## 1. 项目架构与产物概览

### 1.1 仓库结构

```
eqs/
├── packages/
│   ├── server/            # Go + Gin + GORM 后端（cmd/server/main.go）
│   ├── admin/             # Vue3 + Element Plus 管理后台（vite）
│   └── client/            # Uni-app（vue3，H5 / 小程序 / App）
├── shared/                # 共享代码
├── deploy/                # nginx / systemd / deploy.sh / docker-compose.prod.yml
├── pnpm-workspace.yaml    # pnpm monorepo
└── docker-compose.yml     # 本地 MySQL/Redis/MinIO/Server
```

### 1.2 三端构建命令与产物目录

| 端 | 命令 | 产物目录 | 说明 |
|----|------|----------|------|
| 用户端 H5 | `pnpm --filter @eqs/client build:h5` | `packages/client/dist/build/h5` | 生产 H5 静态站 |
| 微信小程序 | `pnpm --filter @eqs/client build:mp-weixin` | `packages/client/dist/build/mp-weixin` | 微信开发者工具导入 |
| App 资源 | `pnpm --filter @eqs/client build:app` | `packages/client/dist/build/app` | uni-app **vue** 构建产物（app-service.js / app-config.js / manifest.json / uni-app-view.umd.js），**非最终 APK/IPA** |
| 管理后台 | `pnpm --filter @eqs/admin build` | `packages/admin/dist` | vue-tsc + vite build |
| 后端 | `go build -o server cmd/server/main.go` | `packages/server/server` | Go 1.22 二进制 |

### 1.3 平台特性声明（重要）

EQS 的 App 端是**普通 uni-app（vue）**，不是 uni-app x（uts）。
`uni build -p app` 只产出 App **资源包**，最终 APK/IPA 有两条路径：

| 路径 | 工具 | 平台 | 成本 | 自动化程度 |
|------|------|------|------|-----------|
| **云打包**（推荐日常） | HBuilderX 云打包 或 DCloud 云打包 API | 提交资源包即可出 APK | 免费额度内 | 半自动（需上传资源/点一下） |
| **离线打包** | DCloud 离线 SDK + Android Studio / Xcode | 需自行集成原生工程 | 高 | 可用 CI 自动（需自持原生工程） |

> 结论：**CI 流水线负责把 `build:app` 资源产物产出并归档**；APK 实包默认走 HBuilderX 云打包；只有申请到并维护 DCloud 离线 SDK 原生工程后，才可在 Ubuntu/macOS runner 上出 APK/IPA 实包（见步骤 6 可选方案）。

### 1.4 成本控制策略（GitHub 私有仓库额度）

| 项 | 规则 |
|----|------|
| macOS-latest | 计费 **10 倍**分钟 |
| ubuntu-latest | 1 倍分钟 |
| 免费额度 | 私有仓库 2000 min/月 |
| EQS 策略 | **push 只跑 Ubuntu**（后端测试+三端构建）；**iOS 仅 `workflow_dispatch` 手动触发** |

---

## 2. 整体流程一览

```
                    +----------------------------+
                    | 开发者 push origin main     |
                    +-------------+--------------+
                                  |
                                  v
                 +------------------------------+
                 | GitHub Actions (ubuntu)      |
                 | 1. 后端 go vet+test+build    |
                 | 2. admin build               |
                 | 3. client H5 + mp + app      |
                 | 4. 归档构建产物 (artifact)   |
                 | 5. 可选：SSH 自动部署后端    |
                 +-------------+----------------+
                               |
                    +----------+-----------+
                    | 网页手动 Run workflow |
                    +----------+-----------+
                               |
                       +-------v--------+      +---------------------------+
                       | iOS job (macos) |      | APK 实包（HBuilderX 云打包）|
                       | fastlane→IPA    |      | 上传 dist/build/app 资源   |
                       +-----------------+      +---------------------------+
```

- **每次 push**：Ubuntu 自动跑后端 + 三端资源构建 → 产出 Android App 资源产物。
- **网页手动触发**：才额外跑 macOS，产出 IPA（自定义调试基座 / 正式发版）。
- iOS 日常调试：首次手动出**自定义调试基座 IPA**，之后爱思助手/热重载，不再重复消耗 macOS。

---

## 3. 步骤 1：代码托管与仓库准备

1. 创建 **GitHub 私有仓库**（如 `your-org/eqs`）。
2. 推送现有代码：

```bash
git remote add origin git@github.com:your-org/eqs.git
git push -u origin main
```

3. 确认 `.gitignore` 已排除敏感/运行时文件（本项目已含）：
   - `packages/server/*.db`（SQLite 运行时数据库）
   - `packages/client/dist/`、`packages/admin/dist/`
   - `.env*`

---

## 4. 步骤 2：准备签名与证书（Windows 完成，无需 Mac）

### 4.1 Android 签名 keystore（jks）

在 `packages/server` 或仓库外的安全目录执行：

```powershell
# 生成正式签名 keystore（Windows / 任意 JDK 环境）
keytool -genkeypair -v -keystore eqs-release.jks -alias eqs `
  -keyalg RSA -keysize 2048 -validity 36500 `
  -dname "CN=eqs,OU=eqs,O=eqs,L=City,ST=State,C=CN"

# 转 Base64（存入 GitHub Secret）
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("eqs-release.jks"))
```

### 4.2 iOS 证书与描述文件（Windows 用 Appuploader）

| 用途 | 材料 | 说明 |
|------|------|------|
| 开发调试基座 | 开发证书 p12 + 开发描述文件 | 真机调试 |
| Ad-Hoc 测试 | Ad-Hoc p12 + adhoc 描述文件 | 需先在描述文件添加 iPhone UDID |
| 正式上架 | Distribution p12 + AppStore 描述文件 | 发包时启用 |

Windows 用 **Appuploader** 登录开发者账号完成创建/下载（无需 Mac）。
全部转 Base64：

```powershell
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("dev.p12"))
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("dev.mobileprovision"))
```

### 4.3 密钥清单与 Secrets 命名

仓库 → Settings → Secrets and variables → Actions → New repository secret：

| Secret 名 | 内容 | 必填 |
|-----------|------|------|
| `ANDROID_SIGN_BASE64` | jks 文件 Base64 | ✅（离线打包时） |
| `ANDROID_KEYSTORE_PWD` | 密钥库密码 | ✅（离线打包时） |
| `ANDROID_ALIAS` | 别名 | ✅（离线打包时） |
| `ANDROID_ALIAS_PWD` | 别名密码 | ✅（离线打包时） |
| `IOS_P12_BASE64` | p12 证书 Base64（开发/发布二选一视阶段） | ✅ |
| `IOS_P12_PASSWORD` | p12 密码 | ✅ |
| `IOS_PROVISION_BASE64` | mobileprovision Base64 | ✅ |
| `IOS_BUNDLE_ID` | iOS 包名（如 `com.eqs.app`） | ✅ |
| `DEPLOY_HOST` / `DEPLOY_USER` / `DEPLOY_SSH_KEY` | 生产 CVM 部署信息（可选，见步骤 7） | 按需 |

> ⚠️ 绝对禁止把 p12/jks/mobileprovision 明文提交进仓库。

---

## 5. 步骤 3：GitHub Actions 流水线（核心）

文件：`.github/workflows/ci.yml`/`cd.yml`（已有）+ `.github/workflows/ios.yml`（本次新增，见 9.2）

> 项目已有 `ci.yml`（Ubuntu 构建+测试）与 `cd.yml`（腾讯云 CVM/COS 自动部署），日常 push 走这两条即可。下面为完整参考流水线，iOS 手动打包单独落在 `ios.yml`（`workflow_dispatch` 触发）。

### 5.1 逻辑

1. `push` 到 main → **仅 ubuntu**：后端测试+构建、admin 构建、client 三端构建、归档产物。
2. `workflow_dispatch` 手动跑 → 同一批 ubuntu 步骤 + **macos iOS 打包**。
3. iOS job 用 `if: github.event_name == 'workflow_dispatch'` 硬锁定——push 永不触发 macOS，零消耗 10 倍配额。

```yaml
name: EQS CI/CD
on:
  push:
    branches: [ main ]
  workflow_dispatch:

concurrency:
  group: eqs-ci-${{ github.ref }}
  cancel-in-progress: true

env:
  NODE_VERSION: 20
  GO_VERSION: '1.22'

jobs:
  # ---------- 后端：测试 + 构建（每次 push） ----------
  server:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: packages/server
    steps:
      - uses: actions/checkout@v4

      - name: 配置 Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: go vet
        run: go vet ./...

      - name: go test
        run: go test ./... -cover

      - name: Go 交叉编译 Linux 二进制
        run: |
          CGO_ENABLED=1 go build -o server cmd/server/main.go

      - name: 上传后端二进制
        uses: actions/upload-artifact@v4
        with:
          name: eqs-server-linux
          path: packages/server/server
          if-no-files-found: error

  # ---------- 前端：admin + client 三端（每次 push） ----------
  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: 配置 Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: pnpm-lock.yaml

      - name: 启用 pnpm
        run: |
          npm i -g corepack
          corepack enable

      - name: 安装依赖
        run: pnpm install --frozen-lockfile

      - name: lint client
        working-directory: packages/client
        run: pnpm exec eslint . && pnpm exec vue-tsc --noEmit

      - name: 构建管理后台
        run: pnpm --filter @eqs/admin build

      - name: 构建 H5
        run: pnpm --filter @eqs/client build:h5

      - name: 构建微信小程序
        run: pnpm --filter @eqs/client build:mp-weixin

      - name: 构建 App 资源
        run: pnpm --filter @eqs/client build:app

      - name: 上传 H5 产物
        uses: actions/upload-artifact@v4
        with:
          name: h5-dist
          path: packages/client/dist/build/h5

      - name: 上传小程序产物
        uses: actions/upload-artifact@v4
        with:
          name: mp-weixin-dist
          path: packages/client/dist/build/mp-weixin

      - name: 上传 App 资源产物
        uses: actions/upload-artifact@v4
        with:
          name: app-resources
          path: packages/client/dist/build/app
          if-no-files-found: error

      - name: 上传 admin 产物
        uses: actions/upload-artifact@v4
        with:
          name: admin-dist
          path: packages/admin/dist

  # ---------- iOS：仅手动触发（workflow_dispatch） ----------
  # EQS 为普通 uni-app(vue)：需 DCloud iOS 离线 SDK 的 Xcode 工程才能出 IPA。
  # 若已维护离线原生工程（apps/ios/），fastlane 归档打包；否则此 job 仅产出 App 资源供 HBuilderX 云打包。
  ios:
    runs-on: macos-latest
    if: github.event_name == 'workflow_dispatch'
    needs: [ server, frontend ]
    steps:
      - uses: actions/checkout@v4

      - name: 配置 Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: 安装依赖
        run: |
          corepack enable
          pnpm install --frozen-lockfile

      - name: 构建 App 资源
        run: pnpm --filter @eqs/client build:app

      - name: 安装 fastlane
        run: gem install fastlane -N

      - name: 还原 iOS 证书与描述文件
        run: |
          echo "${{ secrets.IOS_P12_BASE64 }}" | base64 -d > cert.p12
          echo "${{ secrets.IOS_PROVISION_BASE64 }}" | base64 -d > app.mobileprovision
        shell: bash

      - name: fastlane 归档签名导出 IPA
        uses: maierj/fastlane-action@v3
        with:
          lane: build_ipa
          options: >
            {
              "cert": "cert.p12",
              "provision": "app.mobileprovision"
            }
        env:
          P12_FILE: cert.p12
          P12_PASS: ${{ secrets.IOS_P12_PASSWORD }}
          PROV_FILE: app.mobileprovision
          BUNDLE_ID: ${{ secrets.IOS_BUNDLE_ID }}

      - name: 上传 IPA
        uses: actions/upload-artifact@v4
        with:
          name: ios-ipa
          path: build/ipa/*.ipa
          if-no-files-found: warn
```

> 关键行：`if: github.event_name == 'workflow_dispatch'`——push 时该 job 直接跳过，**macOS 虚拟机不会启动**。

---

## 6. 步骤 4：Fastlane 配置（iOS）

### 6.1 前提

普通 uni-app(vue) 的 iOS 离线打包需要 DCloud **iOS 离线 SDK** 生成 `.xcworkspace`。
在仓库维护离线原生工程（如 `apps/ios/EQS.xcworkspace`），或按 DCloud 文档由资源手动集成。
`Fastfile` 放仓库根 `fastlane/Fastfile`：

```ruby
default_platform(:ios)

platform :ios do
  desc "签名导出 IPA（CI 使用）"
  lane :build_ipa do |options|
    import_certificate(
      certificate_path: options[:cert],
      certificate_password: ENV["P12_PASS"]
    )
    import_provisioning_profile(path: options[:provision])

    xcodebuild(
      workspace: "apps/ios/EQS.xcworkspace",
      scheme: "EQS",
      configuration: "Release",
      archive_path: "./build/archive.xcarchive"
    )

    export_archive(
      archive_path: "./build/archive.xcarchive",
      export_options: {
        method: ENV["IPA_METHOD"] || "ad-hoc",   # ad-hoc 测试 / app-store 发版
        provisioningProfiles: {
          ENV["BUNDLE_ID"] => "EQS"
        }
      },
      output_directory: "./build/ipa"
    )
  end
end
```

### 6.2 HBuilderX 云打包（日常 iOS 调试基座，免 macOS runner）
若暂不维护离线工程，日常 iOS 基座直接走 HBuilderX 云打包：
用 CI 产出的 `app-resources` 导入 HBuilderX → 云打包 → 自定义调试基座 IPA → 爱思助手真机安装。

---

## 7. 步骤 5：后端自动部署（CVM · 可选自动触发）

项目已有 `deploy/scripts/deploy.sh`（腾讯云 CVM + systemd `eqs-server`）。
CI 中可选加一个部署 job，push 成功后 SSH 到服务器执行：

```yaml
  deploy:
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    needs: [ server, frontend ]
    steps:
      - uses: actions/checkout@v4

      - name: SSH 部署到 CVM
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.DEPLOY_HOST }}
          username: ${{ secrets.DEPLOY_USER }}
          key: ${{ secrets.DEPLOY_SSH_KEY }}
          script: |
            cd /opt/eqs && git pull origin main
            bash deploy/scripts/deploy.sh
```

> `deploy.sh` 内部行为：`go build` → `build:admin` → `build:h5` → `systemctl restart eqs-server`。
> 不启用时直接删除该 job，后端保持手/定时部署。

---

## 8. 步骤 6（可选）：Ubuntu 出 APK 实包（离线打包）

若申请到 **DCloud 离线打包 SDK（Android）**并在仓库维护原生 Gradle 工程（`apps/android/`），可将 `build:app` 产物拷贝进 assets，用 Gradle + 签名一键出 APK：

```yaml
  build_apk:
    runs-on: ubuntu-latest
    needs: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: temurin
          java-version: '17'

      # 下载 App 资源产物
      - uses: actions/download-artifact@v4
        with:
          name: app-resources
          path: app-resources

      # 还原签名
      - run: echo "${{ secrets.ANDROID_SIGN_BASE64 }}" | base64 -d > eqs-release.jks

      - name: Android Gradle 打包签名
        working-directory: apps/android
        run: |
          cp -r ../app-resources app/src/main/assets/apps 2>/dev/null || true
          ./gradlew assembleRelease

      - uses: actions/upload-artifact@v4
        with:
          name: eqs-release.apk
          path: apps/android/app/build/outputs/apk/release/*.apk
```

> 因未维护离线原生工程，此 job 为**可选加挂**。当前推荐路径：CI 出 App 资源 → HBuilderX 云打包出 APK。

---

## 9. 步骤 7：日常开发与发布流程

### 9.1 日常迭代（零 macOS 消耗）

1. 本地启动：`pnpm dev:server` / `pnpm dev:admin` / `pnpm dev:client`
2. 改代码 → `git push origin main`
3. Actions 自动：go vet/test → admin/h5/mp/app 构建 → 归档产物
4. 安卓测试：下载 `app-resources` → HBuilderX 云打包 → 安装 APK
5. iOS 热重载调试：使用**已生成的自定义调试基座**，无需重打 IPA

### 9.2 手动触发 iOS（仅两类场景）

> GitHub 仓库 → Actions → `EQS CI/CD` → Run workflow

1. **首次 / 基座更新**：出自定义调试基座 IPA → 爱思助手真机安装
2. **正式发版**：Secrets 切换为 Distribution p12 + AppStore 描述文件，`IPA_METHOD=app-store`，手动跑 → fastlane 后接 `deliver` / 上传 TestFlight

### 9.3 发版灰度（小程序/H5）

- H5：产物 `h5-dist` 上传至 nginx/COS 静态托管
- 小程序：`mp-weixin-dist` 用微信开发者工具上传体验版 → 提审

---

## 10. 计费与优化技巧

| 场景 | 消耗 |
|------|------|
| 日常 push（ubuntu） | 1 倍/次，2000min 免费额度覆盖足够 |
| iOS 手动打包 1 次 | ≈ 10~12min × 10 倍 = 100~120min |
| 每月仅 3~5 次 iOS | ≈ 300~600min，远在免费额度内 |

1. iOS 中间迭代**不要**重复打基座——热重载即可。
2. ad-hoc 新增设备：只更新描述文件，手动打包一次。
3. 备选省钱方案：日常 iOS 调试基座全走 HBuilderX 云打包，macOS 流水线仅正式发包触发。

---

## 11. 避坑清单

1. **签名绝不进仓库**：p12/jks/mobileprovision 只存 GitHub Secrets。
2. **macOS runner 无缓存**：每次全新虚拟机，fastlane 每次必须重新 import 证书/描述文件。
3. **普通 uni-app ≠ uni-app x**：产物目录与原生工程不同。
   - 本仓库：`uni build -p app` → `dist/build/app`（vue 资源）
   - uni-app x：`uni build -p app-android` → 直接原生工程（路径参考模板，EQS 不适用）
4. **ad-hoc 包必须预登记 UDID**，否则真机无法安装。
5. **iOS push 禁止触发**：一律 `workflow_dispatch` 手动管控，依赖 `if: github.event_name == 'workflow_dispatch'`。
6. **pnpm monorepo**：CI 中必须 `corepack enable` + `pnpm install --frozen-lockfile`，且验证 `pnpm-lock.yaml` 一致（`pnpm install` 后在 actions 里 diff，防 lockfile 漂移）。
7. **Go CGO 关键**：server 使用 `gorm.io/driver/sqlite`（底层 `github.com/mattn/go-sqlite3`，**依赖 CGO**）。

   - 本地/Windows：正常 `go build` 即可
   - Linux/CI：必须 CGO 开启（Ubuntu runner 预装 gcc，直接构建）：
   - 现有 `packages/server/Dockerfile` 使用 `CGO_ENABLED=0` 会导致 sqlite 编译失败，需改为 CGO 构建。
   - 若坚持纯静态二进制：将 driver 换成 `modernc.org/sqlite`（纯 Go），或改用 MySQL 模式（`DBDriver=mysql`）。

   ```bash
   # Ubuntu runner 直接构建
   CGO_ENABLED=1 go build -o server cmd/server/main.go
   ```

   ```dockerfile
   # 修复后的 Dockerfile builder 段（musl 交叉编译，产物可在 alpine 运行）
   FROM golang:1.22-alpine AS builder
   RUN apk add --no-cache gcc musl-dev git
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN CGO_ENABLED=1 GOOS=linux go build -tags "sqlite_omit_load_extension" -o server cmd/server/main.go
   ```
8. **concurrency 撤单**：用 `concurrency: cancel-in-progress` 避免冗余构建。

---

## 12. 附录：关键文件索引

| 文件 | 作用 |
|------|------|
| `.github/workflows/ci.yml` | CI 构建测试（已有） |
| `.github/workflows/cd.yml` | 腾讯云 CVM/COS 部署（已有） |
| `.github/workflows/ios.yml` | iOS 手动打包 macOS（本次新增，`workflow_dispatch` 触发） |
| `fastlane/Fastfile` | iOS 签名归档（仓库需离线原生工程 `apps/ios/` 时才启用） |
| `deploy/scripts/deploy.sh` | CVM 后端部署脚本（已有） |
| `deploy/systemd/eqs-server.service` | 后端常驻服务（已有） |
| `deploy/nginx/prod.conf` | 生产反向代理（已有） |
| `packages/client/package.json` | uni-app 构建脚本（已有） |

---

**文档结束**