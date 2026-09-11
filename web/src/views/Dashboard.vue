<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">仪表盘</div>
      <div class="page-desc">节点与流量概览</div>
    </div>

    <a-row :gutter="16">
      <a-col :span="6">
        <a-card>
          <div class="stat-label">运行状态</div>
          <div class="stat-value">
            <a-tag :color="status.running ? 'green' : 'red'">
              {{ status.running ? '运行中' : status.installed ? '已停止' : '未安装' }}
            </a-tag>
          </div>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <div class="stat-label">Xray 版本</div>
          <div class="stat-value">{{ status.version || '未安装' }}</div>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <div class="stat-label">入站节点</div>
          <div class="stat-value">{{ inboundCount }}</div>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <div class="stat-label">订阅用户</div>
          <div class="stat-value">{{ userCount }}</div>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="6">
        <a-card><div class="stat-label">CPU</div><div class="stat-value">{{ sys.cpu.toFixed(1) }}%</div></a-card>
      </a-col>
      <a-col :span="6">
        <a-card><div class="stat-label">内存</div><div class="stat-value">{{ sys.mem.toFixed(1) }}%</div></a-card>
      </a-col>
      <a-col :span="6">
        <a-card><div class="stat-label">磁盘</div><div class="stat-value">{{ sys.disk.toFixed(1) }}%</div></a-card>
      </a-col>
      <a-col :span="6">
        <a-card><div class="stat-label">负载</div><div class="stat-value">{{ sys.load.toFixed(2) }}</div></a-card>
      </a-col>
    </a-row>

    <a-card title="实时流量" style="margin-top: 16px">
      <div ref="chartRef" style="height: 320px"></div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, reactive } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsType } from 'echarts/core'
import { get } from '@/api'
import { useAuthStore } from '@/stores/auth'

echarts.use([LineChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const auth = useAuthStore()
const chartRef = ref<HTMLElement>()

const status = reactive({ running: false, installed: false, version: '' })
const sys = reactive({ cpu: 0, mem: 0, disk: 0, load: 0 })
const inboundCount = ref(0)
const userCount = ref(0)

let chart: EChartsType | null = null
let ws: WebSocket | null = null
const times: string[] = []
const upData: number[] = []
const downData: number[] = []

async function load() {
  try {
    const [s, ibs, us] = await Promise.all([
      get<any>('/node/status'),
      get<any[]>('/inbounds'),
      get<any[]>('/users'),
    ])
    const st = s.data
    status.running = st.running
    status.installed = st.installed
    status.version = st.version
    if (st.system) {
      sys.cpu = st.system.cpu
      sys.mem = st.system.mem
      sys.disk = st.system.disk
      sys.load = st.system.load
    }
    inboundCount.value = ibs.data.length
    userCount.value = us.data.length
  } catch {
    /* 静默失败，保持默认值 */
  }
}

function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${proto}//${location.host}/ws?token=${auth.token}`)
  ws.onmessage = (ev) => {
    const msg = JSON.parse(ev.data)
    if (msg.type !== 'traffic') return
    const d = msg.data
    times.push(new Date().toLocaleTimeString())
    upData.push(d.uplink)
    downData.push(d.downlink)
    if (times.length > 60) {
      times.shift()
      upData.shift()
      downData.shift()
    }
    render()
  }
}

function render() {
  chart?.setOption({
    xAxis: { type: 'category', data: times },
    series: [
      { name: '上行', type: 'line', smooth: true, data: upData },
      { name: '下行', type: 'line', smooth: true, data: downData },
    ],
  })
}

function initChart() {
  if (!chartRef.value) return
  chart = echarts.init(chartRef.value)
  chart.setOption({
    grid: { left: 50, right: 20, top: 30, bottom: 30 },
    legend: { data: ['上行', '下行'] },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: times },
    yAxis: { type: 'value' },
    series: [
      { name: '上行', type: 'line', smooth: true, areaStyle: { opacity: 0.12 }, data: upData },
      { name: '下行', type: 'line', smooth: true, areaStyle: { opacity: 0.12 }, data: downData },
    ],
  })
}

function resize() {
  chart?.resize()
}

onMounted(() => {
  load()
  initChart()
  connectWS()
  window.addEventListener('resize', resize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resize)
  ws?.close()
  chart?.dispose()
})
</script>

<style scoped>
.stat-label {
  font-size: 13px;
  color: var(--volans-text-secondary);
}

.stat-value {
  margin-top: 8px;
  font-size: 22px;
  font-weight: 600;
  color: var(--volans-text);
}
</style>
