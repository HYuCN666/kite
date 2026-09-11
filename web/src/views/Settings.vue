<template>
  <div class="settings-container">
    <!-- 头部说明 -->
    <div class="page-header">
      <div>
        <h2 class="section-title">系统核心配置</h2>
        <p class="section-desc">管理服务器对外公网解析地址、通信策略以及管理员安全凭证</p>
      </div>
    </div>

    <a-row :gutter="[20, 20]">
      <!-- 1. 公网与节点连接配置 -->
      <a-col :xs="24" :md="12">
        <div class="custom-card setting-card">
          <div class="card-header">
            <div class="header-icon">
              <icon-cloud />
            </div>
            <div>
              <span class="header-title">公网访问与节点通信</span>
              <span class="header-desc">生成客户端订阅链接时的外部默认域名或 IP</span>
            </div>
          </div>

          <a-form
            ref="settingsFormRef"
            :model="settingsData"
            layout="vertical"
            class="setting-form"
          >
            <a-form-item
              field="public_host"
              label="服务器公网地址 (Public IP / FQDN)"
              help="如 node1.domain.com 或 123.45.67.89，留空时使用服务端自探测地址"
            >
              <a-input
                v-model="settingsData.public_host"
                placeholder="例如 198.51.100.1 或 proxy.example.com"
                allow-clear
              >
                <template #prefix>
                  <icon-public />
                </template>
              </a-input>
            </a-form-item>

            <div class="form-action-row">
              <a-button
                type="primary"
                :loading="settingsLoading"
                @click="handleSaveSettings"
              >
                <template #icon><icon-save /></template>
                保存网络配置
              </a-button>
            </div>
          </a-form>
        </div>
      </a-col>

      <!-- 2. 安全认证与密码修改 -->
      <a-col :xs="24" :md="12">
        <div class="custom-card setting-card">
          <div class="card-header">
            <div class="header-icon">
              <icon-safe />
            </div>
            <div>
              <span class="header-title">管理员认证安全</span>
              <span class="header-desc">定期轮换密码以确保管理控制台处于高安全状态</span>
            </div>
          </div>

          <a-form
            ref="passwordFormRef"
            :model="passwordData"
            :rules="passwordRules"
            layout="vertical"
            class="setting-form"
          >
            <a-form-item field="old_password" label="当前原密码">
              <a-input-password
                v-model="passwordData.old_password"
                placeholder="请输入当前生效的旧密码"
              >
                <template #prefix><icon-lock /></template>
              </a-input-password>
            </a-form-item>

            <a-form-item field="new_password" label="设定新密码">
              <a-input-password
                v-model="passwordData.new_password"
                placeholder="请输入 6 位以上强密码"
              >
                <template #prefix><icon-key /></template>
              </a-input-password>
            </a-form-item>

            <a-form-item field="confirm_password" label="确认新密码">
              <a-input-password
                v-model="passwordData.confirm_password"
                placeholder="请再次输入新密码"
              >
                <template #prefix><icon-check-circle /></template>
              </a-input-password>
            </a-form-item>

            <div class="form-action-row">
              <a-button
                type="primary"
                status="warning"
                :loading="passwordLoading"
                @click="handleUpdatePassword"
              >
                <template #icon><icon-shield /></template>
                更新管理员密码
              </a-button>
            </div>
          </a-form>
        </div>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message, Modal } from '@arco-design/web-vue'
import type { FormInstance, FieldRule } from '@arco-design/web-vue'
import { settingsApi, authApi } from '@/api'
import { useUserStore } from '@/store/user'
import type { SystemSettings } from '@/types/api'

const router = useRouter()
const userStore = useUserStore()

// 1. 系统网络设置状态
const settingsLoading = ref(false)
const settingsFormRef = ref<FormInstance | null>(null)
const settingsData = reactive<SystemSettings>({
  public_host: ''
})

// 2. 密码修改状态
const passwordLoading = ref(false)
const passwordFormRef = ref<FormInstance | null>(null)
const passwordData = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const passwordRules: Record<string, FieldRule[]> = {
  old_password: [{ required: true, message: '请输入旧密码' }],
  new_password: [
    { required: true, message: '请输入新密码' },
    { minLength: 6, message: '密码长度不得小于 6 个字符' }
  ],
  confirm_password: [
    { required: true, message: '请确认新密码' },
    {
      validator: (value, callback) => {
        if (value !== passwordData.new_password) {
          callback('两次输入的密码不一致')
        } else {
          callback()
        }
      }
    }
  ]
}

// 获取系统配置
const fetchSettings = async () => {
  try {
    const res = await settingsApi.getSettings()
    if (res && res.data) {
      Object.assign(settingsData, res.data)
    }
  } catch (err) {
    // 错误由拦截器统一通知
  }
}

// 保存系统配置
const handleSaveSettings = async () => {
  try {
    settingsLoading.value = true
    await settingsApi.updateSettings(settingsData)
    Message.success('系统配置已成功保存')
  } catch (err) {
    // 失败
  } finally {
    settingsLoading.value = false
  }
}

// 提交密码更新
const handleUpdatePassword = async () => {
  if (!passwordFormRef.value) return
  const errors = await passwordFormRef.value.validate()
  if (errors) return

  try {
    passwordLoading.value = true
    await authApi.updatePassword({
      old_password: passwordData.old_password,
      new_password: passwordData.new_password
    })

    Modal.success({
      title: '密码修改成功',
      content: '安全凭证已成功更新，请使用新密码重新登录。',
      okText: '重新登录',
      onOk: () => {
        userStore.logout()
        router.push('/login')
      }
    })
  } catch (err) {
    // 失败
  } finally {
    passwordLoading.value = false
  }
}

onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.settings-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

.setting-card {
  padding: 24px;
  background-color: var(--bg-card);
  height: 100%;
  box-sizing: border-box;
}

.card-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.header-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background-color: rgba(99, 102, 241, 0.1);
  color: var(--primary-6);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.header-title {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
}

.header-desc {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

.setting-form {
  max-width: 480px;
}

.form-action-row {
  margin-top: 24px;
}
</style>
