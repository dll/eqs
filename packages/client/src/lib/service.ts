// 工程服务类型的中文名称与内部 code 映射（前端共享）
export const PROJECT_TYPES = ['造价咨询', '工程监理', '地质勘察', '工程设计'] as const

export const TYPE_CODE: Record<string, string> = {
  造价咨询: 'cost',
  工程监理: 'supervision',
  地质勘察: 'geotech',
  工程设计: 'design',
}

export const CODE_NAME: Record<string, string> = {
  cost: '造价咨询',
  supervision: '工程监理',
  geotech: '地质勘察',
  design: '工程设计',
}

// 内部 code 对应的 i18n key（供页面 $t 使用）
export const CODE_I18N: Record<string, string> = {
  cost: 'category.cost',
  supervision: 'category.supervision',
  geotech: 'category.geotech',
  design: 'category.design',
}

// code -> i18n key
export const toTypeKey = (code: string): string => CODE_I18N[code] || 'category.cost'

// 人民币：后端金额单位为元，前端展示习惯用万元
export const toWan = (yuan: number): number => +(yuan / 10000).toFixed(2)
export const yuanToWan = (yuan: number): number | '' => (yuan ? +((yuan / 10000).toFixed(2)) : '')
export const wanToYuan = (wan: number): number => Math.round(wan * 10000)

// 中文服务名转内部 code，默认造价咨询
export const toTypeCode = (name: string): string => TYPE_CODE[name] || 'cost'
// 内部 code 转中文名
export const toTypeName = (code: string): string => CODE_NAME[code] || '造价咨询'

// 价格展示：元 -> "x 万元"
export const formatPriceWan = (yuan: number): string => `${toWan(yuan)}万`