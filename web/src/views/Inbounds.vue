<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">入站节点</div>
      <div class="page-desc">管理 Xray 入站配置</div>
    </div>

    <a-card>
      <div class="toolbar">
        <a-button type="primary" @click="openCreate">
          <template #icon><icon-plus /></template>新建节点
        </a-button>
      </div>
      <a-table :columns="columns" :data="rows" :pagination="false" row-key="id" :loading="loading">
        <template #tls="{ record }">
          <a-tag :color="record.tls_enabled ? 'green' : 'gray'">{{ record.tls_enabled ? 'TLS' : '无' }}</a-tag>
        </template>
        <template #enabled="{ record }">
          <a-switch :model-value="record.enabled" @change="(v) => toggle(record, Boolean(v))" />
        </template>
        <template #actions="{ record }">
          <a-button type="text" status="danger" @click="remove(record)">删除</a-button>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:visible="modalVisible" :title="editing ? '编辑节点' : '新建节点'" @ok="save" :ok-loading="saving">
      <a-form :model="form" layout="vertical">
        <a-form-item label="备注">
          <a-input v-model="form.remark" placeholder="如：香港节点" />
        </a-form-item>
        <a-row :gutter="12">
          <a-col :span="12">
            <a-form-item label="协议">
              <a-select v-model="form.protocol">
                <a-option value="vless">VLESS</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="端口">
              <a-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="传输">
          <a-radio-group v-model="form.transport" type="button">
            <a-radio value="tcp">TCP</a-radio>
            <a-radio value="ws">WebSocket</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="form.transport === 'ws'" label="WS 路径">
          <a-input v-model="form.stream_settings.path" placeholder="/ws" />
        </a-form-item>
        <a-form-item label="启用 TLS">
          <a-switch v-model="form.tls_enabled" />
        </a-form-item>
        <template v-if="form.tls_enabled">
          <a-form-item label="证书路径">
            <a-input v-model="form.tls_cert" placeholder="/etc/xray/cert.pem" />
          </a-form-item>
          <a-form-item label="私钥路径">
            <a-input v-model="form.tls_key" placeholder="/etc/xray/key.pem" />
          </a-form-item>
          <a-form-item label="SNI / 域名">
            <a-input v-model="form.tls_server_name" placeholder="example.com" />
          </a-form-item>
        </template>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { IconPlus } from '@arco-design/web-vue/es/icon'
import { get, post, put, patch, del } from '@/api'

const rows = ref<any[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const saving = ref(false)
const editing = ref<number | null>(null)

const columns = [
  { title: '备注', dataIndex: 'remark' },
  { title: '协议', dataIndex: 'protocol' },
  { title: '端口', dataIndex: 'port' },
  { title: '传输', dataIndex: 'transport' },
  { title: 'TLS', slotName: 'tls' },
  { title: '用户数', dataIndex: 'user_count' },
  { title: '状态', slotName: 'enabled' },
  { title: '操作', slotName: 'actions' },
]

const emptyForm = () => ({
  remark: '',
  protocol: 'vless',
  port: 443,
  transport: 'ws',
  stream_settings: { path: '/ws' },
  tls_enabled: false,
  tls_cert: '',
  tls_key: '',
  tls_server_name: '',
  enable_sniffing: true,
  enabled: true,
})

const form = reactive(emptyForm())

async function load() {
  loading.value = true
  try {
    const res = await get<any[]>('/inbounds')
    rows.value = res.data
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, emptyForm())
  modalVisible.value = true
}

async function save() {
  saving.value = true
  try {
    if (editing.value) {
      await put(`/inbounds/${editing.value}`, form)
    } else {
      await post('/inbounds', form)
    }
    Message.success('已保存')
    modalVisible.value = false
    load()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function toggle(record: any, enabled: boolean) {
  try {
    await patch(`/inbounds/${record.id}/status`, { enabled })
    record.enabled = enabled
  } catch {
    Message.error('操作失败')
  }
}

function remove(record: any) {
  Modal.confirm({
    title: '删除节点',
    content: `确定删除「${record.remark}」及其全部用户吗？`,
    onOk: async () => {
      await del(`/inbounds/${record.id}`)
      Message.success('已删除')
      load()
    },
  })
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
</style>
