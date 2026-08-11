<template>
  <el-container class="layout">
    <el-aside
      width="220px"
      class="aside"
    >
      <div class="logo">
        <div class="logo-badge">
          ⚡
        </div>
        <div class="logo-text">
          <div class="logo-title">
            {{ $t('app.title') }}
          </div>
          <div class="logo-sub">
            Agile · AI · Engineering
          </div>
        </div>
      </div>
      <el-menu
        :default-active="route.path"
        router
        class="side-menu"
        background-color="transparent"
        text-color="rgba(255,255,255,.72)"
        active-text-color="#fff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>{{ $t('menu.dashboard') }}</span>
        </el-menu-item>
        <el-menu-item index="/audit">
          <el-icon><DocumentChecked /></el-icon>
          <span>{{ $t('menu.audit') }}</span>
        </el-menu-item>
        <el-menu-item index="/project">
          <el-icon><Folder /></el-icon>
          <span>{{ $t('menu.project') }}</span>
        </el-menu-item>
        <el-menu-item index="/order">
          <el-icon><List /></el-icon>
          <span>{{ $t('menu.order') }}</span>
        </el-menu-item>
        <el-menu-item index="/settlement">
          <el-icon><Money /></el-icon>
          <span>{{ $t('menu.settlement') }}</span>
        </el-menu-item>
        <el-menu-item index="/credit">
          <el-icon><Star /></el-icon>
          <span>{{ $t('menu.credit') }}</span>
        </el-menu-item>
        <el-menu-item index="/dispute">
          <el-icon><Warning /></el-icon>
          <span>{{ $t('menu.dispute') }}</span>
        </el-menu-item>
        <el-menu-item index="/user">
          <el-icon><User /></el-icon>
          <span>{{ $t('menu.user') }}</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>{{ $t('menu.settings') }}</span>
        </el-menu-item>
        <el-menu-item index="/log">
          <el-icon><Document /></el-icon>
          <span>{{ $t('menu.log') }}</span>
        </el-menu-item>
      </el-menu>
      <div class="aside-footer">
        <div class="ai-chip">
          <span class="ai-dot" /> AI 引擎在线
        </div>
      </div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <span class="page-title">{{ $t('menu.' + route.meta.titleKey) }}</span>
          <span class="page-subtitle">EQS 工程快捷服务控制台</span>
        </div>
        <div class="header-right">
          <el-select
            :model-value="lang"
            size="small"
            class="lang-select"
            @change="onLangChange"
          >
            <el-option
              label="中文"
              value="zh-CN"
            />
            <el-option
              label="EN"
              value="en-US"
            />
          </el-select>
          <el-button
            type="primary"
            text
            @click="logout"
          >
            <el-icon style="margin-right: 4px">
              <SwitchButton />
            </el-icon>
            {{ $t('common.logout') }}
          </el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view v-slot="{ Component }">
          <transition
            name="fade"
            mode="out-in"
          >
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import { useI18n } from '@/utils/i18n'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { $t, lang, setAdminLang } = useI18n()

const onLangChange = (v: string) => {
  setAdminLang(v)
}

const logout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout {
  height: 100vh;
}

/* 侧边栏：深色工程渐变 + 玻璃细节 */
.aside {
  background: linear-gradient(180deg, #0f1e3d 0%, #1e3a8a 40%, #0e7490 100%);
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
}
.aside::before {
  content: '';
  position: absolute;
  top: -80px;
  right: -80px;
  width: 220px;
  height: 220px;
  background: radial-gradient(circle, rgba(139, 92, 246, .35) 0%, transparent 70%);
  pointer-events: none;
}
.aside::after {
  content: '';
  position: absolute;
  bottom: 60px;
  left: -60px;
  width: 180px;
  height: 180px;
  background: radial-gradient(circle, rgba(6, 182, 212, .30) 0%, transparent 70%);
  pointer-events: none;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 18px;
  position: relative;
  z-index: 1;
}
.logo-badge {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: var(--eqs-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #fff;
  box-shadow: 0 4px 14px rgba(37, 99, 235, .45);
}
.logo-title {
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: .5px;
}
.logo-sub {
  color: rgba(255, 255, 255, .55);
  font-size: 10px;
  letter-spacing: .4px;
  margin-top: 2px;
}

.side-menu {
  border-right: none;
  flex: 1;
  position: relative;
  z-index: 1;
}
.side-menu :deep(.el-menu-item) {
  height: 44px;
  line-height: 44px;
  margin: 3px 10px;
  border-radius: 10px;
  font-size: 14px;
}
.side-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, rgba(37, 99, 235, .55), rgba(6, 182, 212, .45)) !important;
  color: #fff !important;
  box-shadow: 0 4px 12px rgba(37, 99, 235, .35);
}
.side-menu :deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, .10) !important;
}

.aside-footer {
  padding: 14px 18px;
  position: relative;
  z-index: 1;
}
.ai-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(139, 92, 246, .18);
  border: 1px solid rgba(167, 139, 250, .35);
  border-radius: 20px;
  padding: 8px 14px;
  color: rgba(255, 255, 255, .9);
  font-size: 12px;
}
.ai-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #34d399;
  box-shadow: 0 0 8px #34d399;
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: .4; }
}

/* 顶部栏 */
.header {
  background: rgba(255, 255, 255, .85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--eqs-border);
  height: 60px;
}
.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.page-title {
  font-size: 17px;
  font-weight: 700;
  color: var(--eqs-text);
}
.page-subtitle {
  font-size: 12px;
  color: var(--eqs-text-muted);
}
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.lang-select {
  width: 96px;
}

.main {
  background: var(--eqs-bg);
  padding: 20px;
  overflow-y: auto;
}
</style>
