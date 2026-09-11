<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">系统设置</div>
      <div class="page-desc">面板与订阅配置</div>
    </div>

    <a-row :gutter="16">
      <a-col :span="12">
        <a-card title="面板设置">
          <a-form :model="settings" layout="vertical">
            <a-form-item label="公网地址 / 域名 (用于生成订阅链接)">
              <a-input v-model="settings.public_host" placeholder="如 example.com 或 1.2.3.4" />
            </a-form-item>
            <a-button type="primary" :loading="saving" @click="saveSettings">保存</a-button>
          </a-form>
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="修改密码">
          <a-form :model="pwd" layout="vertical">
            <a-form-item label="旧密码">
              <a-input-password v-model="pwd.old" />
            </a-form-item>
            <a-form-item label="新密码">
              <a-input-password v-model="pwd.new" />
            </a-form-item>
            <a-button type="primary" :loading="changing" @click="changePassword">修改密码</a-button>
          </a-form>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { get, put } from '@/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const saving = ref(false)
const changing = ref(false)
const settings = reactive({ public_host: '' })
const pwd = reactive({ old: '', new: '' })

async function load() {
  const res = await get<Record<string, string>>('/settings')
  settings.public_host = res.data.public_host || ''
}

async function saveSettings() {
  saving.value = true
  try {
    await put('/settings', { public_host: settings.public_host })
    Message.success('已保存')
  } catch {
    Message.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function changePassword() {
  if (pwd.new.length < 8) {
    Message.warning('新密码至少 8 位')
    return
  }
  changing.value = true
  try {
    await put('/auth/password', { old_password: pwd.old, new_password: pwd.new })
    Message.success('密码已修改')
    pwd.old = ''
    pwd.new = ''
    auth.logout()
    window.location.href = '/login'
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '修改失败')
  } finally {
    changing.value = false
  }
}

onMounted(load)
</script>
