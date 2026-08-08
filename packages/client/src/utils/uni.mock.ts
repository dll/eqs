// vitest setupFiles：在任何被测模块加载前注入最小化 uni 全局实现
// storage / request / toast / 路由，测试文件无需再手动 import
const storage = new Map<string, string>()

const uniMock = {
  getStorageSync: (k: string) => storage.get(k) ?? '',
  setStorageSync: (k: string, v: string) => { storage.set(k, String(v)) },
  removeStorageSync: (k: string) => { storage.delete(k) },
  request: (_opts: any) => {},
  showToast: () => {},
  navigateBack: () => {},
  reLaunch: () => {},
  clearStorage: () => storage.clear(),
}

;(globalThis as any).uni = uniMock