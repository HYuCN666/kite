<template>
  <div class="users-container">
    <!-- 顶部操作与筛选 -->
    <div class="page-header">
      <div>
        <h2 class="section-title">订阅用户管理</h2>
        <p class="section-desc">维护独立用户的节点订阅、流量配额限制、速率策略以及到期时间</p>
      </div>
      <div class="header-actions">
        <a-select
          v-model="filterInboundId"
          placeholder="按所属入站节点筛选"
          allow-clear
          style="width: 200px"
          @change="fetchUsers"
        >
          <a-option
            v-for="inb in inboundOptions"
            :key="inb.id"
            :value="inb.id"
            :label="`${inb.remark} (Port: ${inb.port})`"
          />
        </a-select>
        <a-button type="primary" @click="handleOpenCreate">
          <template #icon><icon-user-add /></template>
          新增订阅用户
        </a-button>
      </div>
    </div>

    <!-- 用户列表表格 -->
    <div class="custom-card table-card">
      <a-table
        :data="userList"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
      >
        <template #columns>
          <!-- 账号标识 / 备注 -->
          <a-table-column title="用户标识 / 备注">
            <template #cell="{ record }">
              <div class="user-name-cell">
                <span class="user-title">{{ record.remark || '未命名用户' }}</span>
                <span class="user-id mono-font">UID: {{ record.id }}</span>
              </div>
            </template>
          </a-table-column>

          <!-- 所属节点 -->
          <a-table-column title="绑定入站">
            <template #cell="{ record }">
              <a-tag size="small" color="blue">
                {{ getInboundName(record.inbound_id) }}
              </a-tag>
            </template>
          </a-table-column>

          <!-- 流量使用进度与配额 -->
          <a-table-column title="流量消耗 / 配额" :width="240">
            <template #cell="{ record }">
              <div class="traffic-progress-cell">
                <div class="traffic-text">
                  <span class="mono-font">{{ formatBytes((record.used_uplink || 0) + (record.used_downlink || 0)) }}</span>
                  <span class="quota-divider">/</span>
                  <span class="quota-total mono-font">{{ record.quota_bytes ? formatBytes(record.quota_bytes) : '无限制' }}</span>
                </div>
                <a-progress
                  :percent="calculateTrafficPercent(record)"
                  :status="getTrafficStatus(record)"
                  :stroke-width="6"
                  :show-text="false"
                />
                <div class="traffic-sub mono-font">
                  <span>↑ {{ formatBytes(record.used_uplink || 0) }}</span>
                  <span>↓ {{ formatBytes(record.used_downlink || 0) }}</span>
                </div>
              </div>
            </template>
          </a-table-column>

          <!-- 限速配置 -->
          <a-table-column title="限速 (上/下行)" :width="140">
            <template #cell="{ record }">
              <div class="mono-font speed-limit-text">
                <span v-if="record.speed_limit_uplink || record.speed_limit_downlink">
                  {{ record.speed_limit_uplink ? formatSpeed(record.speed_limit_uplink) : '不限' }} /
                  {{ record.speed_limit_downlink ? formatSpeed(record.speed_limit_downlink) : '不限' }}
                </span>
                <span v-else class="text-muted">不限速</span>
              </div>
            </template>
          </a-table-column>

          <!-- 到期时间 -->
          <a-table-column title="到期时间" :width="160">
            <template #cell="{ record }">
              <span class="mono-font date-text" :class="{ 'date-expired': isExpired(record.expire_at) }">
                {{ record.expire_at ? formatDateTime(record.expire_at) : '永久有效' }}
              </span>
            </template>
          </a-table-column>

          <!-- 启停状态 -->
          <a-table-column title="状态" :width="90">
            <template #cell="{ record }">
              <a-switch
                :model-value="record.enabled"
                :loading="statusLoadingId === record.id"
                @change="(val) => handleToggleStatus(record, val as boolean)"
              />
            </template>
          </a-table-column>

          <!-- 快捷操作栏 -->
          <a-table-column title="操作" :width="230" align="right">
            <template #cell="{ record }">
              <a-space size="mini">
                <!-- 获取并复制订阅 -->
                <a-tooltip content="复制订阅链接">
                  <a-button type="text" size="small" @click="handleCopySubscription(record.id)">
                    <template #icon><icon-link /></template>
                  </a-button>
                </a-tooltip>

                <!-- 重置流量 -->
                <a-popconfirm
                  content="确定清空重置该用户的已用流量数据吗？"
                  type="warning"
                  ok-text="确认重置"
                  cancel-text="取消"
                  @ok="handleResetTraffic(record.id)"
                >
                  <a-tooltip content="重置用量">
                    <a-button type="text" size="small">
                      <template #icon><icon-sync /></template>
                    </a-button>
                  </a-tooltip>
                </a-popconfirm>

                <!-- 编辑 -->
                <a-tooltip content="编辑参数">
                  <a-button type="text" size="small" @click="handleEdit(record)">
                    <template #icon><icon-edit /></template>
                  </a-button>
                </a-tooltip>

                <!-- 删除 -->
                <a-popconfirm
                  content="确定删除此订阅用户？此操作不可撤销。"
                  type="warning"
                  ok-text="删除"
                  cancel-text="取消"
                  :ok-button-props="{ status: 'danger' }"
                  @ok="handleDelete(record.id)"
                >
                  <a-tooltip content="删除用户">
                    <a-button type="text" status="danger" size="small">
                      <template #icon><icon-delete /></template>
                    </a-button>
                  </a-tooltip>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <!-- 创建 / 编辑用户抽屉表单 -->
    <a-drawer
      :visible="drawerVisible"
      :title="isEdit ? '编辑用户配置' : '新增订阅用户'"
      width="520px"
      :ok-loading="submitLoading"
      ok-text="确认提交"
      cancel-text="取消"
      @ok="handleDrawerSubmit"
      @cancel="drawerVisible = false"
    >
      <a-form ref="formRef" :model="formData" :rules="formRules" layout="vertical">
        <!-- 绑定节点 -->
        <a-form-item field="inbound_id" label="绑定入站节点" :rules="[{ required: true, message: '请选择所属节点' }]">
          <a-select v-model="formData.inbound_id" placeholder="选择入站节点">
            <a-option
              v-for="inb in inboundOptions"
              :key="inb.id"
              :value="inb.id"
              :label="`${inb.remark} (Port: ${inb.port})`"
            />
          </a-select>
        </a-form-item>

        <!-- 备注 / 客户名 -->
        <a-form-item field="remark" label="用户备注名称" :rules="[{ required: true, message: '请输入用户标识' }]">
          <a-input v-model="formData.remark" placeholder="如：user-alice-iphone" />
        </a-form-item>

        <!-- 总流量配额 (GB 输入换算) -->
        <a-form-item field="quota_gb" label="流量配额 (GB)" help="输入 0 或留空表示不限制总流量">
          <a-input-number
            v-model="quotaGB"
            :min="0"
            :step="10"
            placeholder="例如 100 代表 100 GB"
          >
            <template #suffix>GB</template>
          </a-input-number>
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="speed_limit_uplink" label="上行限速 (KB/s)" help="0 为不限制">
              <a-input-number
                v-model="speedLimitUpKB"
                :min="0"
                :step="1024"
                placeholder="0"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="speed_limit_downlink" label="下行限速 (KB/s)" help="0 为不限制">
              <a-input-number
                v-model="speedLimitDownKB"
                :min="0"
                :step="1024"
                placeholder="0"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item field="max_devices" label="并发设备数限制" help="0 为不限制，超出将自动停用该订阅">
          <a-input-number v-model="formData.max_devices" :min="0" :step="1" placeholder="0" />
        </a-form-item>

        <!-- 到期时间 -->
        <a-form-item field="expire_at" label="到期时间" help="留空表示永久有效">
          <a-date-picker
            v-model="formData.expire_at"
            show-time
            format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
            placeholder="选择到期时间"
          />
        </a-form-item>

        <!-- 是否启用 -->
        <a-form-item label="账号启用状态">
          <a-switch v-model="formData.enabled" />
        </a-form-item>
      </a-form>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import type { FormInstance } from '@arco-design/web-vue'
import { userApi, inboundApi } from '@/api'
import { formatBytes, formatSpeed, formatDateTime, copyToClipboard } from '@/utils/format'
import type { UserItem, InboundItem } from '@/types/api'

const loading = ref(false)
const userList = ref<UserItem[]>([])
const inboundOptions = ref<InboundItem[]>([])
const filterInboundId = ref<number | string | undefined>(undefined)
const statusLoadingId = ref<number | string | null>(null)

const pagination = reactive({
  pageSize: 10,
  showTotal: true
})

// 表单抽屉状态
const drawerVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance | null>(null)

// 针对表单中友好输入的临时响应式变量
const quotaGB = ref<number>(0)
const speedLimitUpKB = ref<number>(0)
const speedLimitDownKB = ref<number>(0)

const defaultFormData = (): UserItem => ({
  inbound_id: '',
  remark: '',
  quota_bytes: 0,
  used_uplink: 0,
  used_downlink: 0,
  speed_limit_uplink: 0,
  speed_limit_downlink: 0,
  max_devices: 0,
  expire_at: '',
  enabled: true
})

const formData = reactive<UserItem>(defaultFormData())

const formRules = {
  inbound_id: [{ required: true, message: '请选择绑定入站节点' }],
  remark: [{ required: true, message: '请输入用户备注标识' }]
}

// 获取节点字典映射
const getInboundName = (inboundId: number | string) => {
  const found = inboundOptions.value.find(i => String(i.id) === String(inboundId))
  return found ? found.remark : `节点 #${inboundId}`
}

// 计算流量消耗百分比
const calculateTrafficPercent = (record: UserItem) => {
  if (!record.quota_bytes || record.quota_bytes === 0) return 0
  const used = (record.used_uplink || 0) + (record.used_downlink || 0)
  const percent = used / record.quota_bytes
  return Math.min(percent, 1)
}

// 获取进度条状态颜色
const getTrafficStatus = (record: UserItem) => {
  const percent = calculateTrafficPercent(record)
  if (percent >= 1) return 'danger'
  if (percent >= 0.85) return 'warning'
  return 'normal'
}

// 检查是否已过期
const isExpired = (expireAt?: string | number) => {
  if (!expireAt) return false
  return new Date(expireAt).getTime() < Date.now()
}

// 获取用户列表
const fetchUsers = async () => {
  try {
    loading.value = true
    const params = filterInboundId.value ? { inbound_id: filterInboundId.value } : undefined
    const res = await userApi.getList(params)
    if (res && res.data) {
      userList.value = Array.isArray(res.data) ? res.data : []
    }
  } catch (err) {
    // 统一处理
  } finally {
    loading.value = false
  }
}

// 获取入站节点选项
const fetchInbounds = async () => {
  try {
    const res = await inboundApi.getList()
    if (res && res.data) {
      inboundOptions.value = Array.isArray(res.data) ? res.data : []
    }
  } catch (err) {
    // 忽略
  }
}

// 打开新建
const handleOpenCreate = () => {
  isEdit.value = false
  Object.assign(formData, defaultFormData())
  quotaGB.value = 0
  speedLimitUpKB.value = 0
  speedLimitDownKB.value = 0
  if (inboundOptions.value.length > 0) {
    formData.inbound_id = inboundOptions.value[0].id || ''
  }
  drawerVisible.value = true
}

// 打开编辑
const handleEdit = (record: UserItem) => {
  isEdit.value = true
  Object.assign(formData, record)
  // 将 bytes 转换为 GB / KB 辅助显示
  quotaGB.value = record.quota_bytes ? Math.round(record.quota_bytes / (1024 * 1024 * 1024)) : 0
  speedLimitUpKB.value = record.speed_limit_uplink ? Math.round(record.speed_limit_uplink / 1024) : 0
  speedLimitDownKB.value = record.speed_limit_downlink ? Math.round(record.speed_limit_downlink / 1024) : 0
  drawerVisible.value = true
}

// 提交表单
const handleDrawerSubmit = async () => {
  if (!formRef.value) return
  const errors = await formRef.value.validate()
  if (errors) return

  // 转换单位回 bytes
  formData.quota_bytes = quotaGB.value > 0 ? quotaGB.value * 1024 * 1024 * 1024 : 0
  formData.speed_limit_uplink = speedLimitUpKB.value > 0 ? speedLimitUpKB.value * 1024 : 0
  formData.speed_limit_downlink = speedLimitDownKB.value > 0 ? speedLimitDownKB.value * 1024 : 0

  try {
    submitLoading.value = true
    if (isEdit.value && formData.id) {
      await userApi.update(formData.id, formData)
      Message.success('订阅用户信息已更新')
    } else {
      await userApi.create(formData)
      Message.success('订阅用户添加成功')
    }
    drawerVisible.value = false
    fetchUsers()
  } catch (err) {
    // 失败处理
  } finally {
    submitLoading.value = false
  }
}

// 快速启停状态
const handleToggleStatus = async (record: UserItem, val: boolean) => {
  if (!record.id) return
  try {
    statusLoadingId.value = record.id
    await userApi.updateStatus(record.id, val)
    record.enabled = val
    Message.success(val ? '用户订阅已启用' : '用户订阅已暂停')
  } catch (err) {
    record.enabled = !val
  } finally {
    statusLoadingId.value = null
  }
}

// 重置用量
const handleResetTraffic = async (id?: number | string) => {
  if (!id) return
  try {
    await userApi.resetTraffic(id)
    Message.success('用户流量用量已重置为 0')
    fetchUsers()
  } catch (err) {
    // 失败
  }
}

// 复制订阅链接
const handleCopySubscription = async (id?: number | string) => {
  if (!id) return
  try {
    const res = await userApi.getSubscription(id)
    if (res && res.data?.url) {
      const ok = await copyToClipboard(res.data.url)
      if (ok) {
        Message.success('订阅链接已成功复制到剪贴板')
      } else {
        Message.info(`订阅链接: ${res.data.url}`)
      }
    } else {
      Message.warning('未能获取到订阅 URL')
    }
  } catch (err) {
    // 错误处理
  }
}

// 删除用户
const handleDelete = async (id?: number | string) => {
  if (!id) return
  try {
    await userApi.delete(id)
    Message.success('订阅用户已删除')
    fetchUsers()
  } catch (err) {
    // 错误处理
  }
}

onMounted(() => {
  fetchInbounds()
  fetchUsers()
})
</script>

<style scoped>
.users-container {
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.table-card {
  padding: 12px;
}

.user-name-cell {
  display: flex;
  flex-direction: column;
}

.user-title {
  font-weight: 500;
  color: var(--text-main);
}

.user-id {
  font-size: 11px;
  color: var(--text-muted);
}

.traffic-progress-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.traffic-text {
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.quota-divider {
  color: var(--text-muted);
}

.quota-total {
  color: var(--text-muted);
}

.traffic-sub {
  display: flex;
  gap: 10px;
  font-size: 11px;
  color: var(--text-muted);
}

.speed-limit-text {
  font-size: 12px;
}

.text-muted {
  color: var(--text-muted);
}

.date-text {
  font-size: 12px;
}

.date-expired {
  color: rgb(var(--danger-6));
  text-decoration: line-through;
}
</style>
