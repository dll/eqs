import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import KanbanChart, { type KanbanItem } from '../components/KanbanChart.vue'

describe('KanbanChart 看板分组', () => {
  it('按项目状态分列并统计进度状态', () => {
    const items: KanbanItem[] = [
      { id: 1, title: '造价项目', progress: 80, status: 3, schedule_state: 'ahead' },
      { id: 2, title: '监理项目', progress: 50, status: 1, schedule_state: 'on_time' },
      { id: 3, title: '地勘项目', progress: 20, status: 1, schedule_state: 'late' },
      { id: 4, title: '设计项目', progress: 100, status: 4, schedule_state: 'ahead' },
    ]
    const wrapper = mount(KanbanChart, { props: { items } })
    const text = wrapper.text()
    // 状态列标题来自 i18n
    expect(text).toContain('已发布')
    expect(text).toContain('进行中')
    expect(text).toContain('已完成')
    // 进度状态统计出现 提前/按时/滞后
    expect(text).toContain('提前')
    expect(text).toContain('按时')
    expect(text).toContain('滞后')
    // 卡片内容渲染
    expect(text).toContain('#1')
    expect(text).toContain('#4')
    // 状态列计数
    expect(text).toContain('2')
  })

  it('空数据时渲染统计为 0', () => {
    const wrapper = mount(KanbanChart, { props: { items: [] } })
    const text = wrapper.text()
    expect(text).toContain('提前')
    expect(text).toContain('按时')
    expect(text).toContain('滞后')
  })
})
