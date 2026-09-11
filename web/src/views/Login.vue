<template>
  <div class="login-container">
    <!-- 极简背景网格点缀 -->
    <div class="background-decorations">
      <div class="ambient-glow"></div>
      <div class="grid-overlay"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-card-wrapper">
      <div class="brand-header">
        <div class="brand-icon">
          <svg viewBox="0 0 24 24" width="36" height="36" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h1 class="brand-title">Volans</h1>
        <p class="brand-subtitle">开源自托管 Xray 正向代理部署管理面板</p>
      </div>

      <div class="login-box">
        <a-form :model="form" layout="vertical" @submit="handleSubmit">
          <a-form-item field="username" label="管理员账号" :rules="[{ required: true, message: '请输入管理员账号' }]">
            <a-input
              v-model="form.username"
              placeholder="请输入用户名"
              size="large"
              allow-clear
            >
              <template #prefix>
                <icon-user />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item field="password" label="认证密码" :rules="[{ required: true, message: '请输入管理员密码' }]">
            <a-input-password
              v-model="form.password"
              placeholder="请输入密码"
              size="large"
              allow-clear
            >
              <template #prefix>
                <icon-lock />
              </template>
            </a-input-password>
          </a-form-item>

          <div class="form-footer">
            <a-button
              type="primary"
              html-type="submit"
              size="large"
              long
              :loading="loading"
              class="submit-btn"
            >
              登 录 控制台
            </a-button>
          </div>
        </a-form>
      </div>

      <div class="login-footer">
        <span>Volans Management Console &bull; 极致简洁 &bull; 安全可控</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { useUserStore } from '@/store/user'
import type { LoginParams } from '@/types/api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
const form = reactive<LoginParams>({
  username: '',
  password: ''
})

const handleSubmit = async () => {
  if (!form.username || !form.password) {
    Message.warning('请输入用户名和密码')
    return
  }
  try {
    loading.value = true
    await userStore.login(form)
    Message.success('登录成功')
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (error: any) {
    // 错误在 request 拦截器中已经弹出提示
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  position: relative;
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--bg-app);
  overflow: hidden;
}

.background-decorations {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.ambient-glow {
  position: absolute;
  top: 25%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.12) 0%, rgba(99, 102, 241, 0) 70%);
  filter: blur(40px);
}

.grid-overlay {
  position: absolute;
  inset: 0;
  background-image: linear-gradient(to right, rgba(100, 116, 139, 0.05) 1px, transparent 1px),
                    linear-gradient(to bottom, rgba(100, 116, 139, 0.05) 1px, transparent 1px);
  background-size: 40px 40px;
}

.login-card-wrapper {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  padding: 24px;
}

.brand-header {
  text-align: center;
  margin-bottom: 32px;
}

.brand-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: rgba(99, 102, 241, 0.1);
  color: var(--primary-6);
  border: 1px solid rgba(99, 102, 241, 0.2);
  margin-bottom: 16px;
}

.brand-title {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.5px;
  color: var(--text-main);
}

.brand-subtitle {
  margin-top: 8px;
  margin-bottom: 0;
  font-size: 13px;
  color: var(--text-muted);
}

.login-box {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  padding: 32px 28px;
  box-shadow: 0 8px 24px -4px rgba(0, 0, 0, 0.08);
}

.form-footer {
  margin-top: 28px;
}

.submit-btn {
  background-color: var(--primary-6);
  border-radius: 8px;
  font-weight: 500;
  letter-spacing: 0.5px;
  transition: all 0.2s;
}

.submit-btn:hover {
  background-color: var(--primary-5);
}

.login-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
}
</style>
