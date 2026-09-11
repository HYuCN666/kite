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
            <a-tag :color="running ? 'green' : 'red'">
              {{ running ? '运行中' : '已停止' }}
            </a-tag>
          </div>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <div class="stat-label">Xray 版本</div>
          <div class="stat-value">{{ version || '未安装' }}</div>
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

    <a-card title="实时流量" style="margin-top: 16px">
      <div ref="chartRef" style="height: 320px"></div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import * as echarts from 'echarts'

const running = ref(false)
const version = ref('')
const inboundCount = ref(0)
const userCount = ref(0)
const chartRef = ref<HTMLElement>()

let chart: echarts.ECharts | null = null

function initChart() {
  if (!chartRef.value) return
  chart = echarts.init(chartRef.value)
  chart.setOption({
    grid: { left: 40, right: 20, top: 20, bottom: 30 },
    xAxis: { type: 'category', data: [] },
    yAxis: { type: 'value' },
    series: [
      {
        name: '上行',
        type: 'line',
        smooth: true,
        areaStyle: { opacity: 0.15 },
        data: [],
      },
      {
        name: '下行',
        type: 'line',
        smooth: true,
        areaStyle: { opacity: 0.15 },
        data: [],
      },
    ],
  })
}

function resize() {
  chart?.resize()
}

onMounted(() => {
  initChart()
  window.addEventListener('resize', resize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resize)
  chart?.dispose()
})
</script>

<style scoped>
.stat-label {
  font-size: 13px;
  color: var(--kite-text-secondary);
}

.stat-value {
  margin-top: 8px;
  font-size: 22px;
  font-weight: 600;
  color: var(--kite-text);
}
</style>
