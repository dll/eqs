<template>
  <div class="gantt-wrap">
    <div ref="chartEl" class="gantt-chart"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, DataZoomComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([BarChart, GridComponent, TooltipComponent, DataZoomComponent, CanvasRenderer])

interface Milestone {
  id: number
  name: string
  status: string
  ratio: number
  order_id: number
}

export interface GanttItem {
  id: number
  title: string
  service_type: string
  status: number
  status_text: string
  start_date: string | null
  end_date: string | null
  progress: number
  milestones: Milestone[]
}

const props = defineProps<{ projects: GanttItem[] }>()
const emit = defineEmits<{ (e: 'select', id: number): void }>()

const chartEl = ref<HTMLElement>()
let chart: echarts.ECharts | null = null

function buildOption(): echarts.EChartsCoreOption {
  const items = props.projects.slice().reverse() // 底部最新
  const yNames = items.map((p) => `#${p.id} ${p.title.slice(0, 12)}`)

  // 时间范围
  let minT = Infinity
  let maxT = -Infinity
  const dates = items.map((p) => ({ s: p.start_date, e: p.end_date }))
  dates.forEach((d) => {
    if (d.s) {
      const t = new Date(d.s).getTime()
      if (t < minT) minT = t
    }
    if (d.e) {
      const t = new Date(d.e).getTime()
      if (t > maxT) maxT = t
    }
  })
  if (minT === Infinity) minT = Date.now() - 7 * 86400000
  if (maxT === -Infinity) maxT = Date.now() + 7 * 86400000
  const pad = (maxT - minT) * 0.1
  minT -= pad
  maxT += pad

  // 每个项目两条：主进度条（时间线）+ 进度百分比条（0-100 映射到时间轴）
  const barSeries = items.map((p) => {
    const start = p.start_date ? new Date(p.start_date).getTime() : minT
    const end = p.end_date ? new Date(p.end_date).getTime() : maxT - pad
    return {
      type: 'bar' as const,
      xAxisIndex: 0,
      yAxisIndex: 0,
      data: [[start, end]],
      barWidth: 14,
      itemStyle: {
        borderRadius: 7,
        color: ganttColor(p.status),
      },
      emphasis: {
        itemStyle: { shadowBlur: 10, shadowColor: 'rgba(64,158,255,.4)' },
      },
      tooltip: {
        formatter: () => {
          const ms = p.milestones.map((m) => `${m.name}(${m.status})`).join('、')
          return `<b>#${p.id} ${p.title}</b><br/>进度：${p.progress}%<br/>状态：${p.status_text}${ms ? `<br/>里程碑：${ms}` : ''}`
        },
      },
    }
  })

  // 进度百分比条（时间轴右端标注数字）
  const progressSeries = items.map((p) => ({
    type: 'bar' as const,
    xAxisIndex: 0,
    yAxisIndex: 0,
    data: [[minT, minT + (maxT - minT) * (p.progress / 100)]],
    barWidth: 14,
    itemStyle: { color: 'transparent' },
    label: {
      show: true,
      position: 'right',
      formatter: `${p.progress}%`,
      color: '#409EFF',
      fontWeight: 'bold' as const,
      fontSize: 11,
    },
    tooltip: { show: false },
  }))

  // 里程碑点
  const msScatter = items.flatMap((p, i) =>
    p.milestones.map((m) => ({
      type: 'scatter' as const,
      xAxisIndex: 0,
      yAxisIndex: 0,
      data: [[minT + (maxT - minT) * (m.ratio / 100), i]],
      symbolSize: 6,
      itemStyle: { color: msColor(m.status) },
      tooltip: { formatter: `#${p.id} ${m.name}：${m.status}（比例${m.ratio}%）` },
    })),
  )

  return {
    grid: { left: 120, right: 60, top: 10, bottom: 30 },
    tooltip: { trigger: 'item', triggerOn: 'mousemove' },
    xAxis: {
      type: 'time',
      min: minT,
      max: maxT,
      axisLabel: { formatter: '{MM}-{dd}' },
    },
    yAxis: {
      type: 'category',
      data: yNames,
      axisLabel: { fontSize: 11 },
    },
    dataZoom: [{ type: 'inside' }, { type: 'slider', height: 12, bottom: 2 }],
    animation: true,
    animationDuration: 800,
    series: [...barSeries, ...progressSeries, ...msScatter],
  }
}

function ganttColor(status: number): string {
  if (status === 4) return '#67C23A'
  if (status >= 2) return '#409EFF'
  if (status === 1) return '#E6A23C'
  return '#C0C4CC'
}

function msColor(status: string): string {
  if (status === 'settled' || status === 'accepted') return '#67C23A'
  if (status === 'disputed') return '#F56C6C'
  return '#909399'
}

function render() {
  if (!chartEl.value) return
  if (!chart) {
    chart = echarts.init(chartEl.value)
    chart.on('click', (params: any) => {
      const idx = params.dataIndex
      if (typeof idx === 'number' && idx >= 0 && props.projects[idx]) {
        emit('select', props.projects[idx].id)
      }
    })
  }
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

watch(() => props.projects, render, { deep: true })
</script>

<style scoped>
.gantt-wrap {
  width: 100%;
  height: 100%;
  min-height: 320px;
}
.gantt-chart {
  width: 100%;
  height: 100%;
  min-height: 320px;
}
</style>
