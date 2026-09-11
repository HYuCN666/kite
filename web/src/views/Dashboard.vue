<template>
  <div class="dashboard-container">
    <!-- 头部操作与刷新状态 -->
    <div class="page-header">
      <div>
        <h2 class="section-title">概览与实时监控</h2>
        <p class="section-desc">实时观测 Xray 核心运行状态、宿主机资源水位与全网流量吞吐</p>
      </div>
      <div class="header-actions">
        <a-tag :color="wsConnected ? 'green' : 'orange'" bordered size="small" class="status-tag">
          <template #icon>
            <icon-check-circle-fill v-if="wsConnected" />
            <icon-exclamation-circle-fill v-else />
          </template>
          {{ wsConnected ? '实时数据流已连接' : '实时连接中...' }}
        </a-tag>
        <a-button type="outline" size="small" :loading="loading" @click="fetchData">
          <template #icon><icon-refresh /></template>
          刷新
        </a-button>
      </div>
    </div>

    <!-- 顶部四个核心指标卡片 -->
    <a-row :gutter="[16, 16]" class="stat-cards-row">
      <!-- 1. 运行状态 -->
      <a-col :xs="24" :sm="12" :md="6">
        <div class="custom-card stat-card">
          <div class="stat-header">
            <span class="stat-label">运行状态</span>
            <div class="stat-icon-box status-icon" :class="{ active: nodeStatus?.running }">
              <icon-poweroff />
            </div>
          </div>
          <div class="stat-main">
            <div class="stat-value-badge">
              <span class="status-dot" :class="nodeStatus?.running ? 'dot-running' : 'dot-stopped'"></span>
              <span class="value-text">{{ nodeStatus?.running ? '运行中' : '已停止' }}</span>
            </div>
          </div>
          <div class="stat-sub">
            <span>Core: {{ nodeStatus?.installed ? '已就绪' : '未安装' }}</span>
          </div>
        </div>
      </a-col>

      <!-- 2. Xray 版本 -->
      <a-col :xs="24" :sm="12" :md="6">
        <div class="custom-card stat-card">
          <div class="stat-header">
            <span class="stat-label">核心版本</span>
            <div class="stat-icon-box">
              <icon-code-sandbox />
            </div>
          </div>
          <div class="stat-main">
            <div class="mono-font stat-main-text">{{ nodeStatus?.version || 'v1.8.x' }}</div>
          </div>
          <div class="stat-sub">
            <span>官方最新稳定分支</span>
          </div>
        </div>
      </a-col>

      <!-- 3. 入站节点数 -->
      <a-col :xs="24" :sm="12" :md="6">
        <div class="custom-card stat-card">
          <div class="stat-header">
            <span class="stat-label">入站节点</span>
            <div class="stat-icon-box">
              <icon-storage />
            </div>
          </div>
          <div class="stat-main">
            <div class="stat-main-text">{{ inboundCount }}</div>
            <span class="stat-unit">个配置端口</span>
          </div>
          <div class="stat-sub">
            <span>支持 VLESS / TCP / WS</span>
          </div>
        </div>
      </a-col>

      <!-- 4. 活跃/订阅用户数 -->
      <a-col :xs="24" :sm="12" :md="6">
        <div class="custom-card stat-card">
          <div class="stat-header">
            <span class="stat-label">订阅用户</span>
            <div class="stat-icon-box">
              <icon-user-group />
            </div>
          </div>
          <div class="stat-main">
            <div class="stat-main-text">{{ userCount }}</div>
            <span class="stat-unit">位有效客户</span>
          </div>
          <div class="stat-sub">
            <span>在线活跃: {{ statsOverview?.active_users || 0 }} 人</span>
          </div>
        </div>
      </a-col>
    </a-row>

    <!-- 宿主机资源使用率仪表板 (CPU/内存/磁盘/系统负载) -->
    <div class="section-divider-title">系统宿主性能水位</div>
    <a-row :gutter="[16, 16]">
      <!-- CPU -->
      <a-col :xs="12" :sm="12" :md="6">
        <div class="custom-card metric-card">
          <div class="metric-top">
            <span class="metric-title">CPU 使用率</span>
            <span class="metric-val mono-font">{{ (nodeStatus?.system?.cpu ?? 0).toFixed(1) }}%</span>
          </div>
          <a-progress
            :percent="(nodeStatus?.system?.cpu || 0) / 100"
            :color="getUsageColor(nodeStatus?.system?.cpu)"
            :show-text="false"
            class="metric-progress"
          />
        </div>
      </a-col>

      <!-- 内存 -->
      <a-col :xs="12" :sm="12" :md="6">
        <div class="custom-card metric-card">
          <div class="metric-top">
            <span class="metric-title">内存使用率</span>
            <span class="metric-val mono-font">{{ (nodeStatus?.system?.mem ?? 0).toFixed(1) }}%</span>
          </div>
          <a-progress
            :percent="(nodeStatus?.system?.mem || 0) / 100"
            :color="getUsageColor(nodeStatus?.system?.mem)"
            :show-text="false"
            class="metric-progress"
          />
        </div>
      </a-col>

      <!-- 磁盘 -->
      <a-col :xs="12" :sm="12" :md="6">
        <div class="custom-card metric-card">
          <div class="metric-top">
            <span class="metric-title">磁盘空间</span>
            <span class="metric-val mono-font">{{ (nodeStatus?.system?.disk ?? 0).toFixed(1) }}%</span>
          </div>
          <a-progress
            :percent="(nodeStatus?.system?.disk || 0) / 100"
            :color="getUsageColor(nodeStatus?.system?.disk)"
            :show-text="false"
            class="metric-progress"
          />
        </div>
      </a-col>

      <!-- 系统负载 -->
      <a-col :xs="12" :sm="12" :md="6">
        <div class="custom-card metric-card">
          <div class="metric-top">
            <span class="metric-title">平均负载 (Load)</span>
            <span class="metric-val mono-font">{{ nodeStatus?.system?.load ?? '0.00' }}</span>
          </div>
          <div class="metric-sub-tip">1 / 5 / 15 分钟系统调度均值</div>
        </div>
      </a-col>
    </a-row>

    <!-- 实时流量图表与汇总 -->
    <div class="chart-section-wrapper">
      <div class="custom-card chart-card">
        <div class="chart-header">
          <div>
            <span class="chart-title">实时网络吞吐量 (Throughput)</span>
            <span class="chart-subtitle">每秒 WebSocket 实时推送</span>
          </div>
          <div class="traffic-metrics-badge">
            <div class="badge-item">
              <span class="badge-dot up"></span>
              <span class="badge-name">当前上行:</span>
              <span class="badge-value mono-font">{{ formatSpeed(currentUplink) }}</span>
            </div>
            <div class="badge-item">
              <span class="badge-dot down"></span>
              <span class="badge-name">当前下行:</span>
              <span class="badge-value mono-font">{{ formatSpeed(currentDownlink) }}</span>
            </div>
            <div class="badge-divider"></div>
            <div class="badge-item total-summary">
              <span class="badge-name">累计下行:</span>
              <span class="badge-value mono-font">{{ formatBytes(statsOverview?.total_downlink || 0) }}</span>
            </div>
            <div class="badge-item total-summary">
              <span class="badge-name">累计上行:</span>
              <span class="badge-value mono-font">{{ formatBytes(statsOverview?.total_uplink || 0) }}</span>
            </div>
          </div>
        </div>

        <div ref="chartContainer" class="chart-body"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import echarts from '@/utils/echarts'
import { nodeApi, inboundApi, userApi } from '@/api'
import { useTrafficWebSocket } from '@/utils/useTrafficWs'
import { formatBytes, formatSpeed } from '@/utils/format'
import { useAppStore } from '@/store/app'
import type { NodeStatus, StatsOverview } from '@/types/api'

const appStore = useAppStore()
const loading = ref(false)
const nodeStatus = ref<NodeStatus | null>(null)
const statsOverview = ref<StatsOverview | null>(null)
const inboundCount = ref<number>(0)
const userCount = ref<number>(0)

// WebSocket 实时速率钩子
const {
  isConnected: wsConnected,
  currentUplink,
  currentDownlink,
  trafficHistory
} = useTrafficWebSocket()

// ECharts 实例引用
const chartContainer = ref<HTMLDivElement | null>(null)
let chartInstance: echarts.ECharts | null = null

// 根据使用率返回警示颜色
const getUsageColor = (val: number = 0) => {
  if (val > 85) return 'rgb(var(--danger-6))'
  if (val > 65) return 'rgb(var(--warning-6))'
  return 'var(--primary-6)'
}

// 初始化图表
const initChart = () => {
  if (!chartContainer.value) return
  if (chartInstance) {
    chartInstance.dispose()
  }

  chartInstance = echarts.init(chartContainer.value)
  renderChart()
}

// 更新图表数据
const renderChart = () => {
  if (!chartInstance) return

  const isDark = appStore.theme === 'dark'
  const textColor = isDark ? '#94a3b8' : '#64748b'
  const splitLineColor = isDark ? '#1e293b' : '#f1f5f9'

  const times = trafficHistory.value.map(item => item.time)
  const uplinks = trafficHistory.value.map(item => Number((item.uplink / 1024).toFixed(2))) // KB/s
  const downlinks = trafficHistory.value.map(item => Number((item.downlink / 1024).toFixed(2))) // KB/s

  // 默认模拟一些初始空数据保证曲线平滑呈现
  const finalTimes = times.length > 0 ? times : Array(15).fill('').map((_, i) => `00:00:${i * 2}`)
  const finalUplinks = uplinks.length > 0 ? uplinks : Array(15).fill(0)
  const finalDownlinks = downlinks.length > 0 ? downlinks : Array(15).fill(0)

  const option: echarts.EChartsCoreOption = {
    animation: false,
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#1e293b' : '#ffffff',
      borderColor: isDark ? '#334155' : '#e2e8f0',
      textStyle: {
        color: isDark ? '#f8fafc' : '#0f172a',
        fontSize: 12
      },
      formatter: (params: any) => {
        if (!Array.isArray(params)) return ''
        let res = `<div style="font-weight:600;margin-bottom:4px;">${params[0].axisValue}</div>`
        params.forEach((item: any) => {
          const valInBytes = item.value * 1024
          res += `<div style="display:flex;align-items:center;gap:6px;font-size:12px;margin:2px 0;">
            <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${item.color}"></span>
            <span>${item.seriesName}:</span>
            <span style="font-family:monospace;font-weight:600">${formatSpeed(valInBytes)}</span>
          </div>`
        })
        return res
      }
    },
    legend: {
      top: 0,
      right: 10,
      textStyle: {
        color: textColor,
        fontSize: 12
      },
      itemWidth: 12,
      itemHeight: 8
    },
    grid: {
      left: '2%',
      right: '2%',
      top: '12%',
      bottom: '5%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: finalTimes,
      axisLine: { lineStyle: { color: splitLineColor } },
      axisTick: { show: false },
      axisLabel: {
        color: textColor,
        fontSize: 11
      }
    },
    yAxis: {
      type: 'value',
      name: '速率 (KB/s)',
      nameTextStyle: {
        color: textColor,
        fontSize: 11,
        padding: [0, 0, 0, 10]
      },
      splitLine: {
        lineStyle: {
          color: splitLineColor,
          type: 'dashed'
        }
      },
      axisLabel: {
        color: textColor,
        fontSize: 11
      }
    },
    series: [
      {
        name: '下行速率 (IN)',
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2, color: '#6366f1' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(99, 102, 241, 0.25)' },
            { offset: 1, color: 'rgba(99, 102, 241, 0.00)' }
          ])
        },
        data: finalDownlinks
      },
      {
        name: '上行速率 (OUT)',
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2, color: '#10b981' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(16, 185, 129, 0.22)' },
            { offset: 1, color: 'rgba(16, 185, 129, 0.00)' }
          ])
        },
        data: finalUplinks
      }
    ]
  }

  chartInstance.setOption(option)
}

// 刷新状态与计数
const fetchData = async () => {
  try {
    loading.value = true
    const [statusRes, overviewRes, inboundsRes, usersRes] = await Promise.allSettled([
      nodeApi.getStatus(),
      nodeApi.getStatsOverview(),
      inboundApi.getList(),
      userApi.getList()
    ])

    if (statusRes.status === 'fulfilled' && statusRes.value.data) {
      nodeStatus.value = statusRes.value.data
    }
    if (overviewRes.status === 'fulfilled' && overviewRes.value.data) {
      statsOverview.value = overviewRes.value.data
    }
    if (inboundsRes.status === 'fulfilled' && Array.isArray(inboundsRes.value.data)) {
      inboundCount.value = inboundsRes.value.data.length
    }
    if (usersRes.status === 'fulfilled' && Array.isArray(usersRes.value.data)) {
      userCount.value = usersRes.value.data.length
    }
  } catch (err) {
    // 拦截器自动提示
  } finally {
    loading.value = false
  }
}

// 响应式图表尺寸适配
const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// 监听流量数据推流更新折线图
watch(
  () => trafficHistory.value.length,
  () => {
    renderChart()
  }
)

// 监听主题模式变化，更新图表色调
watch(
  () => appStore.theme,
  () => {
    nextTick(() => {
      initChart()
    })
  }
)

onMounted(() => {
  fetchData()
  nextTick(() => {
    initChart()
    window.addEventListener('resize', handleResize)
  })
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})
</script>

<style scoped>
.dashboard-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 4px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: var(--text-main);
}

.section-desc {
  font-size: 12.5px;
  color: var(--text-muted);
  margin: 4px 0 0 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-tag {
  font-size: 12px;
}

/* 指标卡片 */
.stat-card {
  padding: 18px 20px;
  border-radius: 8px;
  background-color: var(--bg-card);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 120px;
  box-sizing: border-box;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stat-label {
  font-size: 13px;
  color: var(--text-muted);
  font-weight: 500;
}

.stat-icon-box {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background-color: rgba(99, 102, 241, 0.08);
  color: var(--primary-6);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
}

.stat-icon-box.status-icon.active {
  background-color: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.stat-main {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin: 6px 0;
}

.stat-main-text {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: -0.5px;
}

.stat-unit {
  font-size: 12px;
  color: var(--text-muted);
}

.stat-value-badge {
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-running {
  background-color: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.6);
}

.dot-stopped {
  background-color: #ef4444;
}

.value-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-main);
}

.stat-sub {
  font-size: 11.5px;
  color: var(--text-muted);
}

.section-divider-title {
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  margin-top: 8px;
}

/* 性能水位条 */
.metric-card {
  padding: 16px;
  background-color: var(--bg-card);
}

.metric-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.metric-title {
  font-size: 13px;
  color: var(--text-muted);
}

.metric-val {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
}

.metric-progress {
  width: 100%;
}

.metric-sub-tip {
  font-size: 11.5px;
  color: var(--text-muted);
  margin-top: 4px;
}

/* 图表卡片 */
.chart-card {
  padding: 20px;
  background-color: var(--bg-card);
}

.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 16px;
}

.chart-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
  display: block;
}

.chart-subtitle {
  font-size: 12px;
  color: var(--text-muted);
}

.traffic-metrics-badge {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.badge-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12.5px;
}

.badge-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.badge-dot.up {
  background-color: #10b981;
}

.badge-dot.down {
  background-color: #6366f1;
}

.badge-name {
  color: var(--text-muted);
}

.badge-value {
  font-weight: 600;
  color: var(--text-main);
}

.badge-divider {
  width: 1px;
  height: 14px;
  background-color: var(--border-subtle);
}

.total-summary .badge-value {
  color: var(--text-muted);
}

.chart-body {
  width: 100%;
  height: 320px;
}
</style>
