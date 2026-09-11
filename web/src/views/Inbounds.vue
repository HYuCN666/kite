<template>
  <div class="inbounds-container">
    <!-- 顶部操作栏 -->
    <div class="page-header">
      <div>
        <h2 class="section-title">入站节点管理</h2>
        <p class="section-desc">配置 Xray 核心监听协议 (VLESS)、端口与传输层 TLS/WS 策略</p>
      </div>
      <div>
        <a-button type="primary" @click="handleOpenCreate">
          <template #icon><icon-plus /></template>
          新建入站节点
        </a-button>
      </div>
    </div>

    <!-- 节点数据表格 -->
    <div class="custom-card table-card">
      <a-table
        :data="inboundList"
        :loading="loading"
        :pagination="false"
        row-key="id"
      >
        <template #columns>
          <!-- 备注名 -->
          <a-table-column title="节点名称" data-index="remark">
            <template #cell="{ record }">
              <div class="node-name-cell">
                <span class="node-title">{{ record.remark || '未命名节点' }}</span>
                <span class="node-id mono-font">ID: {{ record.id }}</span>
              </div>
            </template>
          </a-table-column>

          <!-- 协议 -->
          <a-table-column title="协议" data-index="protocol" :width="100">
            <template #cell="{ record }">
              <a-tag color="arcoblue" size="small" class="mono-font uppercase-tag">
                {{ record.protocol || 'VLESS' }}
              </a-tag>
            </template>
          </a-table-column>

          <!-- 监听与端口 -->
          <a-table-column title="监听地址与端口">
            <template #cell="{ record }">
              <div class="mono-font address-text">
                {{ record.listen || '0.0.0.0' }}:<span class="port-highlight">{{ record.port }}</span>
              </div>
            </template>
          </a-table-column>

          <!-- 传输流与路径 -->
          <a-table-column title="传输方式">
            <template #cell="{ record }">
              <div class="transport-cell">
                <a-tag size="small" :color="record.transport === 'ws' ? 'purple' : 'gray'" class="mono-font">
                  {{ (record.transport || 'tcp').toUpperCase() }}
                </a-tag>
                <span v-if="record.transport === 'ws' && record.stream_settings?.path" class="path-text mono-font">
                  {{ record.stream_settings.path }}
                </span>
              </div>
            </template>
          </a-table-column>

          <!-- TLS 与 探测 -->
          <a-table-column title="安全与嗅探" :width="160">
            <template #cell="{ record }">
              <div class="security-cell">
                <a-tag v-if="record.tls_enabled" color="green" size="small">
                  <template #icon><icon-lock /></template>
                  TLS
                </a-tag>
                <a-tag v-else color="gray" size="small">None</a-tag>

                <a-tag v-if="record.enable_sniffing" color="cyan" size="small">Sniffing</a-tag>
              </div>
            </template>
          </a-table-column>

          <!-- 启停状态 -->
          <a-table-column title="状态" :width="100">
            <template #cell="{ record }">
              <a-switch
                :model-value="record.enabled"
                :loading="statusLoadingId === record.id"
                @change="(val) => handleToggleStatus(record, val as boolean)"
              />
            </template>
          </a-table-column>

          <!-- 操作 -->
          <a-table-column title="操作" :width="150" align="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="handleEdit(record)">
                  <template #icon><icon-edit /></template>
                  编辑
                </a-button>
                <a-popconfirm
                  content="确定删除此入站节点？如果有关联的用户订阅可能无法正常工作。"
                  type="warning"
                  ok-text="删除"
                  cancel-text="取消"
                  :ok-button-props="{ status: 'danger' }"
                  @ok="handleDelete(record.id)"
                >
                  <a-button type="text" status="danger" size="small">
                    <template #icon><icon-delete /></template>
                    删除
                  </a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <!-- 创建 / 编辑抽屉表单 -->
    <a-drawer
      :visible="drawerVisible"
      :title="isEdit ? '编辑入站节点' : '新建入站节点'"
      width="540px"
      :ok-loading="submitLoading"
      ok-text="确认保存"
      cancel-text="取消"
      @ok="handleDrawerSubmit"
      @cancel="drawerVisible = false"
    >
      <a-form ref="formRef" :model="formData" :rules="formRules" layout="vertical">
        <!-- 所属服务器 -->
        <a-form-item field="server_id" label="部署服务器">
          <a-select v-model="formData.server_id" placeholder="选择服务器">
            <a-option :value="0">本机 (Local)</a-option>
            <a-option v-for="sv in serverOptions" :key="sv.id" :value="sv.id">
              {{ sv.name }} ({{ sv.host }})
            </a-option>
          </a-select>
        </a-form-item>

        <!-- 节点名称 -->
        <a-form-item field="remark" label="节点名称 / 备注">
          <a-input v-model="formData.remark" placeholder="如：HK-Node-Vless-01" />
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="protocol" label="核心协议">
              <a-select v-model="formData.protocol" placeholder="选择协议">
                <a-option value="vless">VLESS (推荐)</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="port" label="监听端口">
              <a-input-number
                v-model="formData.port"
                :min="1"
                :max="65535"
                placeholder="例如 443 或 8443"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="listen" label="监听地址">
              <a-input v-model="formData.listen" placeholder="0.0.0.0" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="transport" label="传输层协议 (Transport)">
              <a-select v-model="formData.transport">
                <a-option value="tcp">TCP</a-option>
                <a-option value="ws">WebSocket (WS)</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <!-- WS 路径 -->
        <a-form-item
          v-if="formData.transport === 'ws'"
          field="stream_settings.path"
          label="WebSocket 路径 (Path)"
          help="以斜杠开头的路径，如 /vless-stream"
        >
          <a-input v-model="formData.stream_settings!.path" placeholder="/vless-stream" />
        </a-form-item>

        <a-divider style="margin: 12px 0 20px 0" />

        <!-- 安全配置 TLS -->
        <a-form-item label="TLS 安全加密">
          <a-space direction="vertical" fill>
            <div class="switch-row">
              <span>启用 TLS 传输加密</span>
              <a-switch v-model="formData.tls_enabled" />
            </div>
            <div class="switch-row">
              <span>启用流量探测嗅探 (Sniffing)</span>
              <a-switch v-model="formData.enable_sniffing" />
            </div>
          </a-space>
        </a-form-item>

        <template v-if="formData.tls_enabled">
          <a-form-item field="tls_server_name" label="TLS SNI / 域名">
            <a-input v-model="formData.tls_server_name" placeholder="例如 node1.yourdomain.com" />
          </a-form-item>

          <a-form-item field="tls_cert" label="TLS 证书路径 (或证书内容)">
            <a-textarea
              v-model="formData.tls_cert"
              placeholder="/etc/ssl/certs/volans.crt 或证书 PEM 内容"
              :auto-size="{ minRows: 2, maxRows: 4 }"
            />
          </a-form-item>

          <a-form-item field="tls_key" label="TLS 私钥路径 (或私钥内容)">
            <a-textarea
              v-model="formData.tls_key"
              placeholder="/etc/ssl/certs/volans.key 或私钥 PEM 内容"
              :auto-size="{ minRows: 2, maxRows: 4 }"
            />
          </a-form-item>
        </template>

        <a-form-item label="节点默认启用状态">
          <a-switch v-model="formData.enabled" />
        </a-form-item>
      </a-form>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import type { FormInstance, FieldRule } from '@arco-design/web-vue'
import { inboundApi, serverApi } from '@/api'
import type { InboundItem, ServerItem } from '@/types/api'

const loading = ref(false)
const inboundList = ref<InboundItem[]>([])
const serverOptions = ref<ServerItem[]>([])
const statusLoadingId = ref<number | string | null>(null)

// 抽屉弹窗表单状态
const drawerVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance | null>(null)

const defaultFormData = (): InboundItem => ({
  server_id: 0,
  remark: '',
  protocol: 'vless',
  port: 8443,
  listen: '0.0.0.0',
  transport: 'tcp',
  stream_settings: {
    path: '/vless-stream'
  },
  tls_enabled: false,
  tls_cert: '',
  tls_key: '',
  tls_server_name: '',
  enable_sniffing: true,
  enabled: true
})

const formData = reactive<InboundItem>(defaultFormData())

const formRules: Record<string, FieldRule[]> = {
  remark: [{ required: true, message: '请输入节点名称' }],
  port: [{ required: true, message: '请输入有效监听端口' }],
  listen: [{ required: true, message: '请输入监听地址' }]
}

// 获取入站节点列表
const fetchInbounds = async () => {
  try {
    loading.value = true
    const res = await inboundApi.getList()
    if (res && res.data) {
      inboundList.value = Array.isArray(res.data) ? res.data : []
    }
  } catch (err) {
    // 统一拦截器处理
  } finally {
    loading.value = false
  }
}

// 获取服务器选项
const fetchServers = async () => {
  try {
    const res = await serverApi.getList()
    if (res && res.data) {
      serverOptions.value = res.data.filter((s: ServerItem) => s.id !== 0)
    }
  } catch (err) {
    // 忽略
  }
}

// 打开新建
const handleOpenCreate = () => {
  isEdit.value = false
  Object.assign(formData, defaultFormData())
  drawerVisible.value = true
}

// 打开编辑
const handleEdit = (record: InboundItem) => {
  isEdit.value = true
  Object.assign(formData, {
    ...record,
    stream_settings: record.stream_settings || { path: '/vless-stream' }
  })
  drawerVisible.value = true
}

// 提交抽屉表单
const handleDrawerSubmit = async () => {
  if (!formRef.value) return
  const errors = await formRef.value.validate()
  if (errors) return

  try {
    submitLoading.value = true
    if (isEdit.value && formData.id) {
      await inboundApi.update(formData.id, formData)
      Message.success('节点更新成功')
    } else {
      await inboundApi.create(formData)
      Message.success('节点创建成功')
    }
    drawerVisible.value = false
    fetchInbounds()
  } catch (err) {
    // 请求失败提示
  } finally {
    submitLoading.value = false
  }
}

// 快速启停状态
const handleToggleStatus = async (record: InboundItem, val: boolean) => {
  if (!record.id) return
  try {
    statusLoadingId.value = record.id
    await inboundApi.updateStatus(record.id, val)
    record.enabled = val
    Message.success(val ? '节点已启用' : '节点已暂停')
  } catch (err) {
    // 恢复状态
    record.enabled = !val
  } finally {
    statusLoadingId.value = null
  }
}

// 删除节点
const handleDelete = async (id?: number | string) => {
  if (!id) return
  try {
    await inboundApi.delete(id)
    Message.success('节点已成功删除')
    fetchInbounds()
  } catch (err) {
    // 错误处理
  }
}

onMounted(() => {
  fetchInbounds()
  fetchServers()
})
</script>

<style scoped>
.inbounds-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
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

.table-card {
  padding: 12px;
}

.node-name-cell {
  display: flex;
  flex-direction: column;
}

.node-title {
  font-weight: 500;
  color: var(--text-main);
}

.node-id {
  font-size: 11px;
  color: var(--text-muted);
}

.uppercase-tag {
  font-weight: 600;
  letter-spacing: 0.5px;
}

.address-text {
  font-size: 13px;
}

.port-highlight {
  color: var(--primary-6);
  font-weight: 600;
}

.transport-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.path-text {
  font-size: 12px;
  color: var(--text-muted);
}

.security-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: 6px;
  background-color: var(--color-fill-1);
  font-size: 13px;
}
</style>
