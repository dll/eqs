<template>
  <div
    ref="chartEl"
    class="kanban-chart"
  />
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

export interface KanbanItem {
  id: number
  title: string
  progress: number
  status: number
}

const props = defineProps<{ items: KanbanItem[] }>()

const chartEl = ref<HTMLElement>()
let chart: echarts.ECharts | null = null

function buildOption(): echarts.EChartsCoreOption {
  const items = props.items.slice().reverse()
  return {
    grid: { left: 100, right: 50, top: 10, bottom: 20 },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: any) => {
        const p = params[0]
        const item = items[p.dataIndex]
        return `<b>#${item.id} ${item.title}</b><br/>进度：${item.progress}%`
      },
    },
    xAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    yAxis: {
      type: 'category',
      data: items.map((p) => `#${p.id} ${p.title.slice(0, 10)}`),
      axisLabel: { fontSize: 11 },
    },
    series: [
      {
        type: 'bar',
        data: items.map((p) => ({
          value: p.progress,
          itemStyle: { color: p.status === 4 ? '#67C23A' : p.progress >= 50 ? '#409EFF' : '#E6A23C', borderRadius: 5 },
        })),
        barWidth: 14,
        label: {
          show: true,
          position: 'right',
          formatter: '{c}%',
          fontWeight: 'bold',
          fontSize: 11,
        },
        animationDuration: 900,
        animationEasing: 'cubicOut',
      },
    ],
  }
}

function render() {
  if (!chartEl.value) return
  if (!chart) chart = echarts.init(chartEl.value)
  chart.setOption(buildOption(), true)
}

function resize() {
  chart?.resize()
}

onMounted(() => {
  render()
  window.addEventListener('resize', resize)
})
onBeforeUnmount(() => {
  window.removeEventListener('resize', resize)
  chart?.dispose()
})
watch(() => props.items, render, { deep: true })
</script>

<style scoped>
.kanban-chart {
  width: 100%;
  height: 100%;
  min-height: 300px;
}
</style>
