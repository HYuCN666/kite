/**
 * 格式化字节大小 (B, KB, MB, GB, TB)
 */
export function formatBytes(bytes: number = 0, decimals = 2): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

/**
 * 格式化速率 (B/s, KB/s, MB/s, GB/s)
 */
export function formatSpeed(bytesPerSec: number = 0, decimals = 2): string {
  if (!bytesPerSec || bytesPerSec === 0) return '0 B/s'
  return `${formatBytes(bytesPerSec, decimals)}/s`
}

/**
 * 格式化日期时间
 */
export function formatDateTime(dateStrOrTs?: string | number | null): string {
  if (!dateStrOrTs) return '-'
  const date = new Date(dateStrOrTs)
  if (isNaN(date.getTime())) return String(dateStrOrTs)
  const pad = (n: number) => (n < 10 ? '0' + n : n)
  const y = date.getFullYear()
  const m = pad(date.getMonth() + 1)
  const d = pad(date.getDate())
  const h = pad(date.getHours())
  const min = pad(date.getMinutes())
  const s = pad(date.getSeconds())
  return `${y}-${m}-${d} ${h}:${min}:${s}`
}

/**
 * 复制文本到剪贴板
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // fallback
    }
  }
  const textArea = document.createElement('textarea')
  textArea.value = text
  textArea.style.position = 'fixed'
  textArea.style.opacity = '0'
  document.body.appendChild(textArea)
  textArea.focus()
  textArea.select()
  try {
    const successful = document.execCommand('copy')
    document.body.removeChild(textArea)
    return successful
  } catch (err) {
    document.body.removeChild(textArea)
    return false
  }
}
