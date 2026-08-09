<template>
  <el-container class="layout">
    <el-aside width="200px" class="aside">
      <div class="logo">{{ $t('app.title') }}</div>
      <el-menu :default-active="route.path" router background-color="#304156" text-color="#bfcbd9" active-text-color="#409EFF">
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
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <span>{{ $t('menu.' + route.meta.titleKey) }}</span>
        <div>
          <el-select :model-value="lang" size="small" style="width: 100px; margin-right: 12px" @change="onLangChange">
            <el-option label="中文" value="zh-CN" />
            <el-option label="EN" value="en-US" />
          </el-select>
          <el-button type="text" @click="logout">{{ $t('common.logout') }}</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
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

.aside {
  background: #304156;
}

.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
}

.header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.main {
  background: #f0f2f5;
  padding: 20px;
}
</style>
