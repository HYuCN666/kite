<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">入站节点</div>
      <div class="page-desc">管理 Xray 入站配置</div>
    </div>

    <a-card>
      <div class="toolbar">
        <a-button type="primary" @click="visible = true">
          <template #icon><icon-plus /></template>新建节点
        </a-button>
      </div>
      <a-table :columns="columns" :data="rows" :pagination="false" row-key="id">
        <template #enabled="{ record }">
          <a-tag :color="record.enabled ? 'green' : 'gray'">
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
const rows = ref<Array<Record<string, unknown>>>([])

const columns = [
  { title: '备注', dataIndex: 'remark' },
  { title: '协议', dataIndex: 'protocol' },
  { title: '端口', dataIndex: 'port' },
  { title: '传输', dataIndex: 'transport' },
  { title: 'TLS', dataIndex: 'tls_enabled' },
  { title: '状态', slotName: 'enabled' },
]
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
