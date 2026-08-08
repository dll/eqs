<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { useSettingsStore } from '@/store/settings'

onLaunch(async () => {
  const userStore = useUserStore()
  const settingsStore = useSettingsStore()
  await userStore.loadUser()
  await settingsStore.loadSettings()
  settingsStore.checkVersion()
})
</script>

<style>
page {
  background-color: var(--bg-color, #f5f5f5);
  color: var(--text-color, #333333);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}
</style>
