<template>
  <div class="login">
    <div class="login-card">
      <h1 class="login-title">Kite</h1>
      <p class="login-sub">可视化代理部署管理面板</p>
      <a-form :model="form" @submit="handleSubmit">
        <a-form-item field="username" hide-label>
          <a-input v-model="form.username" placeholder="用户名" allow-clear>
            <template #prefix><icon-user /></template>
          </a-input>
        </a-form-item>
        <a-form-item field="password" hide-label>
          <a-input-password v-model="form.password" placeholder="密码">
            <template #prefix><icon-lock /></template>
          </a-input-password>
        </a-form-item>
        <a-button type="primary" html-type="submit" long :loading="loading">
          登录
        </a-button>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { IconUser, IconLock } from '@arco-design/web-vue/es/icon'
import { post } from '@/api'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

async function handleSubmit() {
  if (!form.username || !form.password) {
    Message.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const res = await post<{ token: string }>('/auth/login', form)
    auth.setToken(res.data.token)
    auth.setUsername(form.username)
    Message.success('登录成功')
    router.push('/dashboard')
  } catch {
    Message.error('登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-card {
  width: 360px;
  padding: 40px 36px;
  background: var(--kite-surface);
  border: 1px solid var(--kite-border);
  border-radius: 12px;
}

.login-title {
  font-size: 26px;
  font-weight: 700;
  color: var(--kite-text);
}

.login-sub {
  margin: 8px 0 28px;
  font-size: 13px;
  color: var(--kite-text-secondary);
}
</style>
