<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">订阅用户</div>
      <div class="page-desc">一用户一订阅链接，独立流量配额</div>
    </div>

    <a-card>
      <div class="toolbar">
        <a-button type="primary" @click="openCreate">
          <template #icon><icon-plus /></template>新建用户
        </a-button>
      </div>
      <a-table :columns="columns" :data="rows" :pagination="false" row-key="id" :loading="loading">
        <template #usage="{ record }">
          <a-progress :percent="calcPercent(record)" :show-text="false" :status="isExceeded(record) ? 'danger' : 'normal'" style="width: 120px" />
          <div class="usage-text">{{ formatBytes(record.used_uplink + record.used_downlink) }} / {{ formatBytes(record.quota_bytes) }}</div>
        </template>
        <template #expire="{ record }">
          {{ record.expire_at ? new Date(record.expire_at).toLocaleDateString() : '永久' }}
        </template>
        <template #enabled="{ record }">
          <a-switch :model-value="record.enabled" @change="(v) => toggle(record, Boolean(v))" />
        </template>
        <template #actions="{ record }">
          <a-button type="text" @click="copySub(record)">复制订阅</a-button>
          <a-button type="text" @click="resetTraffic(record)">重置流量</a-button>
          <a-button type="text" status="danger" @click="remove(record)">删除</a-button>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:visible="modalVisible" title="新建用户" @ok="save" :ok-loading="saving">
      <a-form :model="form" layout="vertical">
        <a-form-item label="所属入站">
          <a-select v-model="form.inbound_id" placeholder="选择入站节点">
            <a-option v-for="ib in inbounds" :key="ib.id" :value="ib.id">
              {{ ib.remark || ib.tag }} ({{ ib.port }})
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="备注">
          <a-input v-model="form.remark" placeholder="如：客户A" />
        </a-form-item>
        <a-form-item label="流量配额 (GB，0 为不限)">
          <a-input-number v-model="form.quota_gb" :min="0" style="width: 100%" />
        </a-form-item>
        <a-form-item label="到期时间">
          <a-date-picker v-model="form.expire_at" style="width: 100%" show-time />
        </a-form-item>
        <a-form-item label="上行限速 (MB/s，0 为不限)">
          <a-input-number v-model="form.uplink_mb" :min="0" style="width: 100%" />
        </a-form-item>
        <a-form-item label="下行限速 (MB/s，0 为不限)">
          <a-input-number v-model="form.downlink_mb" :min="0" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { IconPlus } from '@arco-design/web-vue/es/icon'
import { get, post, patch, del } from '@/api'

const rows = ref<any[]>([])
const inbounds = ref<any[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const saving = ref(false)

const columns = [
  { title: '备注', dataIndex: 'remark' },
  { title: '已用 / 配额', slotName: 'usage' },
  { title: '到期时间', slotName: 'expire' },
  { title: '状态', slotName: 'enabled' },
  { title: '操作', slotName: 'actions' },
]

const form = reactive({
  inbound_id: undefined as number | undefined,
  remark: '',
  quota_gb: 0,
  expire_at: '' as string | number | Date,
  uplink_mb: 0,
  downlink_mb: 0,
})

async function load() {
  loading.value = true
  try {
    const [u, ib] = await Promise.all([
      get<any[]>('/users'),
      get<any[]>('/inbounds'),
    ])
    rows.value = u.data
    inbounds.value = ib.data
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, { inbound_id: undefined, remark: '', quota_gb: 0, expire_at: '', uplink_mb: 0, downlink_mb: 0 })
  modalVisible.value = true
}

async function save() {
  if (!form.inbound_id) {
    Message.warning('请选择入站')
    return
  }
  saving.value = true
  try {
    await post('/users', {
      inbound_id: form.inbound_id,
      remark: form.remark,
      quota_bytes: form.quota_gb * 1024 * 1024 * 1024,
      expire_at: form.expire_at ? new Date(form.expire_at).toISOString() : null,
      speed_limit_uplink: form.uplink_mb * 1024 * 1024,
      speed_limit_downlink: form.downlink_mb * 1024 * 1024,
      enabled: true,
    })
    Message.success('已创建')
    modalVisible.value = false
    load()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '创建失败')
  } finally {
    saving.value = false
  }
}

async function toggle(record: any, enabled: boolean) {
  try {
    await patch(`/users/${record.id}/status`, { enabled })
    record.enabled = enabled
  } catch {
    Message.error('操作失败')
  }
}

async function copySub(record: any) {
  const res = await get<{ url: string }>(`/users/${record.id}/subscription`)
  await navigator.clipboard.writeText(res.data.url)
  Message.success('订阅链接已复制')
}

function resetTraffic(record: any) {
  Modal.confirm({
    title: '重置流量',
    content: `确定清零「${record.remark}」的已用流量吗？`,
    onOk: async () => {
      await post(`/users/${record.id}/reset-traffic`)
      Message.success('已重置')
      load()
    },
  })
}

function remove(record: any) {
  Modal.confirm({
    title: '删除用户',
    content: `确定删除「${record.remark}」吗？`,
    onOk: async () => {
      await del(`/users/${record.id}`)
      Message.success('已删除')
      load()
    },
  })
}

function calcPercent(record: any): number {
  if (!record.quota_bytes) return 0
  const used = (record.used_uplink ?? 0) + (record.used_downlink ?? 0)
  return Math.min(100, Math.round((used / record.quota_bytes) * 100))
}

function isExceeded(record: any): boolean {
  if (!record.quota_bytes) return false
  const used = (record.used_uplink ?? 0) + (record.used_downlink ?? 0)
  return used >= record.quota_bytes
}

function formatBytes(n: number): string {
  if (!n) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(1)} ${units[i]}`
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.usage-text {
  font-size: 12px;
  color: var(--volans-text-secondary);
}
</style>
