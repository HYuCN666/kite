<template>
  <div class="servers-container">
    <div class="page-header">
      <div>
        <h2 class="section-title">节点服务器</h2>
        <p class="section-desc">集中管理多台节点的 SSH 连接、内核安装与运行状态</p>
      </div>
      <div>
        <a-button type="primary" @click="openCreate">
          <template #icon><icon-plus /></template>
          添加服务器
        </a-button>
      </div>
    </div>

    <div class="custom-card table-card">
      <a-table :data="servers" :loading="loading" :pagination="false" row-key="id">
        <template #columns>
          <a-table-column title="名称" data-index="name">
            <template #cell="{ record }">
              <div class="name-cell">
                <span class="name-title">{{ record.name }}</span>
                <span class="name-sub" v-if="record.id !== 0">{{ record.username }}@{{ record.host }}:{{ record.port }}</span>
                <a-tag v-else size="small" color="arcoblue">本机</a-tag>
              </div>
            </template>
          </a-table-column>

          <a-table-column title="主机地址" data-index="host" :width="200">
            <template #cell="{ record }">
              <span class="mono-font">{{ record.id === 0 ? '127.0.0.1' : record.host }}</span>
            </template>
          </a-table-column>

          <a-table-column title="状态" :width="200">
            <template #cell="{ record }">
              <template v-if="record.id === 0">
                <a-tag color="arcoblue">本地节点</a-tag>
              </template>
              <template v-else>
                <a-tag :color="remoteStatus[record.id]?.running ? 'green' : 'gray'">
                  {{ remoteStatus[record.id]?.running ? '运行中' : '未运行' }}
                </a-tag>
                <span class="version-text mono-font" v-if="remoteStatus[record.id]?.version">
                  {{ remoteStatus[record.id].version }}
                </span>
              </template>
            </template>
          </a-table-column>

          <a-table-column title="操作" :width="280" align="right">
            <template #cell="{ record }">
              <template v-if="record.id === 0">
                <span class="text-muted">由本机面板直接管理</span>
              </template>
              <template v-else>
                <a-space size="mini">
                  <a-tooltip content="测试连接"><a-button type="text" size="small" @click="testServer(record)"><template #icon><icon-wifi /></template></a-button></a-tooltip>
                  <a-tooltip content="查看状态"><a-button type="text" size="small" @click="fetchStatus(record)"><template #icon><icon-eye /></template></a-button></a-tooltip>
                  <a-tooltip content="安装内核"><a-button type="text" size="small" @click="installServer(record)"><template #icon><icon-download /></template></a-button></a-tooltip>
                  <a-dropdown>
                    <a-button type="text" size="small"><template #icon><icon-poweroff /></template></a-button>
                    <template #content>
                      <a-doption @click="control(record, 'start')">启动</a-doption>
                      <a-doption @click="control(record, 'stop')">停止</a-doption>
                      <a-doption @click="control(record, 'restart')">重启</a-doption>
                    </template>
                  </a-dropdown>
                  <a-tooltip content="编辑"><a-button type="text" size="small" @click="openEdit(record)"><template #icon><icon-edit /></template></a-button></a-tooltip>
                  <a-popconfirm content="确定删除此服务器？" type="warning" @ok="remove(record)">
                    <a-tooltip content="删除"><a-button type="text" status="danger" size="small"><template #icon><icon-delete /></template></a-button></a-tooltip>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <a-drawer :visible="drawerVisible" :title="isEdit ? '编辑服务器' : '添加服务器'" width="520px" :ok-loading="submitLoading" ok-text="保存" @ok="save" @cancel="drawerVisible = false">
      <a-form ref="formRef" :model="form" :rules="rules" layout="vertical">
        <a-form-item field="name" label="名称"><a-input v-model="form.name" placeholder="如 HK-Node-01" /></a-form-item>
        <a-form-item field="host" label="主机地址"><a-input v-model="form.host" placeholder="如 1.2.3.4 或 node.example.com" /></a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="port" label="SSH 端口"><a-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="username" label="SSH 用户"><a-input v-model="form.username" placeholder="root" /></a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="auth_type" label="认证方式">
          <a-radio-group v-model="form.auth_type" type="button">
            <a-radio value="password">密码</a-radio>
            <a-radio value="key">私钥</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="form.auth_type === 'password'" field="password" label="密码">
          <a-input-password v-model="form.password" placeholder="SSH 密码" />
        </a-form-item>
        <a-form-item v-else field="private_key" label="私钥">
          <a-textarea v-model="form.private_key" placeholder="粘贴 OpenSSH 私钥内容" :auto-size="{ minRows: 4, maxRows: 8 }" />
        </a-form-item>
        <a-form-item label="启用">
          <a-switch v-model="form.enabled" />
        </a-form-item>
      </a-form>
      <div class="drawer-tip" v-if="!isEdit">
        <a-button type="outline" size="small" :loading="testing" @click="testForm">
          <template #icon><icon-wifi /></template>测试连接
        </a-button>
      </div>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import type { FormInstance } from '@arco-design/web-vue'
import { serverApi } from '@/api'
import type { ServerItem } from '@/types/api'

const servers = ref<ServerItem[]>([])
const loading = ref(false)
const remoteStatus = reactive<Record<string, any>>({})
const drawerVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const testing = ref(false)
const formRef = ref<FormInstance | null>(null)

const emptyForm = () => ({ id: 0, name: '', host: '', port: 22, username: 'root', auth_type: 'password', password: '', private_key: '', enabled: true })
const form = reactive(emptyForm())

const rules = {
  name: [{ required: true, message: '请输入名称' }],
  host: [{ required: true, message: '请输入主机地址' }],
  username: [{ required: true, message: '请输入 SSH 用户' }]
}

async function load() {
  loading.value = true
  try {
    const res = await serverApi.getList()
    servers.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  isEdit.value = false
  Object.assign(form, emptyForm())
  drawerVisible.value = true
}

function openEdit(record: ServerItem) {
  isEdit.value = true
  Object.assign(form, { ...emptyForm(), ...record, password: '', private_key: '' })
  drawerVisible.value = true
}

async function save() {
  const err = await formRef.value?.validate()
  if (err) return
  submitLoading.value = true
  try {
    const payload = {
      name: form.name, host: form.host, port: form.port, username: form.username,
      auth_type: form.auth_type, password: form.password, private_key: form.private_key, enabled: form.enabled
    }
    if (isEdit.value) {
      await serverApi.update(form.id, payload)
    } else {
      await serverApi.create(payload)
    }
    Message.success('已保存')
    drawerVisible.value = false
    load()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '保存失败')
  } finally {
    submitLoading.value = false
  }
}

async function remove(record: ServerItem) {
  await serverApi.delete(record.id)
  Message.success('已删除')
  load()
}

async function testServer(record: ServerItem) {
  try {
    await serverApi.test({ host: record.host, port: record.port, username: record.username, auth_type: record.auth_type })
    Message.success('连接成功')
  } catch {
    Message.error('连接失败')
  }
}

async function testForm() {
  testing.value = true
  try {
    await serverApi.test({ host: form.host, port: form.port, username: form.username, auth_type: form.auth_type, password: form.password, private_key: form.private_key })
    Message.success('连接成功')
  } catch {
    Message.error('连接失败')
  } finally {
    testing.value = false
  }
}

async function fetchStatus(record: ServerItem) {
  try {
    const res = await serverApi.getStatus(record.id)
    remoteStatus[record.id] = res.data
  } catch {
    remoteStatus[record.id] = { installed: false, version: '', running: false }
  }
}

async function installServer(record: ServerItem) {
  Message.loading('正在安装内核...')
  try {
    await serverApi.install(record.id)
    Message.success('安装完成')
    fetchStatus(record)
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '安装失败')
  }
}

async function control(record: ServerItem, action: string) {
  try {
    await serverApi.control(record.id, action)
    Message.success('操作成功')
    fetchStatus(record)
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}

onMounted(load)
</script>

<style scoped>
.servers-container { display: flex; flex-direction: column; gap: 16px; }
.page-header { display: flex; align-items: flex-end; justify-content: space-between; }
.section-title { font-size: 18px; font-weight: 600; margin: 0; color: var(--text-main); }
.section-desc { font-size: 12.5px; color: var(--text-muted); margin: 4px 0 0 0; }
.table-card { padding: 12px; }
.name-cell { display: flex; flex-direction: column; }
.name-title { font-weight: 500; color: var(--text-main); }
.name-sub { font-size: 11px; color: var(--text-muted); }
.version-text { margin-left: 8px; font-size: 12px; color: var(--text-muted); }
.text-muted { color: var(--text-muted); }
.mono-font { font-family: monospace; }
.drawer-tip { margin-top: 8px; }
</style>
