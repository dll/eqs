import { describe, it, expect } from 'vitest'
import { toTypeCode, toTypeName, toWan, wanToYuan, formatPriceWan, yuanToWan, CODE_NAME } from './service'

describe('服务类型映射', () => {
  it('中文名转 code', () => {
    expect(toTypeCode('造价咨询')).toBe('cost')
    expect(toTypeCode('工程监理')).toBe('supervision')
    expect(toTypeCode('地质勘察')).toBe('geotech')
    expect(toTypeCode('工程设计')).toBe('design')
  })

  it('未知名称默认造价咨询', () => {
    expect(toTypeCode('未知类型')).toBe('cost')
  })

  it('code 转中文名', () => {
    expect(toTypeName('supervision')).toBe('工程监理')
    expect(toTypeName('unknown')).toBe('造价咨询')
    expect(Object.keys(CODE_NAME)).toHaveLength(4)
  })
})

describe('金额换算', () => {
  it('元转万元', () => {
    expect(toWan(50000)).toBe(5)
    expect(toWan(123400)).toBe(12.34)
  })

  it('万元转元', () => {
    expect(wanToYuan(5)).toBe(50000)
  })

  it('格式化为万元展示', () => {
    expect(formatPriceWan(10000)).toBe('1万')
  })

  it('yuanToWan 空值返回空串', () => {
    expect(yuanToWan(0)).toBe('')
    expect(yuanToWan(25000)).toBe(2.5)
  })
})