import { ref, onMounted, onUnmounted } from 'vue'
import type { WsMessage, WsTrafficData } from '@/types/api'

export function useTrafficWebSocket() {
  const isConnected = ref(false)
  const currentUplink = ref(0)
  const currentDownlink = ref(0)
  const trafficHistory = ref<Array<{ time: string; uplink: number; downlink: number }>>([])

  let socket: WebSocket | null = null
  let reconnectTimer: any = null
  let pingTimer: any = null

  const maxHistoryLength = 30 // 保留最近30个点

  const connect = () => {
    const token = localStorage.getItem('volans_token')
    if (!token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/ws?token=${token}`

    try {
      socket = new WebSocket(wsUrl)

      socket.onopen = () => {
        isConnected.value = true
        if (reconnectTimer) {
          clearTimeout(reconnectTimer)
          reconnectTimer = null
        }
        // 心跳包
        pingTimer = setInterval(() => {
          if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: 'ping' }))
          }
        }, 20000)
      }

      socket.onmessage = (event) => {
        try {
          const msg: WsMessage = JSON.parse(event.data)
          if (msg.type === 'traffic' && msg.data) {
            const data = msg.data
            currentUplink.value = data.uplink || 0
            currentDownlink.value = data.downlink || 0

            const now = new Date(data.time || Date.now())
            const timeLabel = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`

            trafficHistory.value.push({
              time: timeLabel,
              uplink: data.uplink || 0,
              downlink: data.downlink || 0
            })

            if (trafficHistory.value.length > maxHistoryLength) {
              trafficHistory.value.shift()
            }
          }
        } catch (e) {
          // 忽略格式解析错误
        }
      }

      socket.onclose = () => {
        isConnected.value = false
        if (pingTimer) clearInterval(pingTimer)
        // 自动重连
        reconnectTimer = setTimeout(() => {
          connect()
        }, 5000)
      }

      socket.onerror = () => {
        isConnected.value = false
      }
    } catch (err) {
      isConnected.value = false
    }
  }

  const disconnect = () => {
    if (pingTimer) clearInterval(pingTimer)
    if (reconnectTimer) clearTimeout(reconnectTimer)
    if (socket) {
      socket.close()
      socket = null
    }
    isConnected.value = false
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    isConnected,
    currentUplink,
    currentDownlink,
    trafficHistory,
    reconnect: connect,
    disconnect
  }
}
