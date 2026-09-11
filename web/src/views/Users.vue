<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">订阅用户</div>
      <div class="page-desc">一用户一订阅链接，独立流量配额</div>
    </div>

    <a-card>
      <div class="toolbar">
        <a-button type="primary" @click="visible = true">
          <template #icon><icon-plus /></template>新建用户
        </a-button>
      </div>
      <a-table :columns="columns" :data="rows" :pagination="false" row-key="id">
        <template #usage="{ record }">
          <a-progress
            :percent="calcPercent(record)"
            :show-text="false"
            :status="isExceeded(record) ? 'danger' : 'normal'"
          />
        </template>
        <template #enabled="{ record }">
          <a-tag :color="record.enabled ? 'green' : 'red'">
            {{ record.enabled ? '启用' : '停用' }}
          </a-tag>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { IconPlus } from '@arco-design/web-vue/es/icon'

const visible = ref(false)
const rows = ref<Array<Record<string, any>>>([])

const columns = [
  { title: '备注', dataIndex: 'remark' },
  { title: '已用流量', slotName: 'usage' },
  { title: '到期时间', dataIndex: 'expire_at' },
  { title: '状态', slotName: 'enabled' },
]

function calcPercent(record: Record<string, any>): number {
  if (!record.quota_bytes) return 0
  const used = (record.used_uplink ?? 0) + (record.used_downlink ?? 0)
  return Math.min(100, Math.round((used / record.quota_bytes) * 100))
}

function isExceeded(record: Record<string, any>): boolean {
  if (!record.quota_bytes) return false
  const used = (record.used_uplink ?? 0) + (record.used_downlink ?? 0)
  return used >= record.quota_bytes
}
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
