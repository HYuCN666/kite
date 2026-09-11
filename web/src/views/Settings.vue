<template>
  <div class="page">
    <div class="page-header">
      <div class="page-title">系统设置</div>
      <div class="page-desc">面板与内核配置</div>
    </div>

    <a-card title="面板设置">
      <a-form :model="form" layout="vertical" style="max-width: 480px">
        <a-form-item label="监听地址">
          <a-input v-model="form.bind" placeholder="127.0.0.1" />
        </a-form-item>
        <a-form-item label="面板端口">
          <a-input-number v-model="form.port" :min="1" :max="65535" />
        </a-form-item>
        <a-form-item label="Xray 二进制路径">
          <a-input v-model="form.xray_path" placeholder="/usr/local/bin/xray" />
        </a-form-item>
        <a-button type="primary" :loading="saving" @click="save">保存</a-button>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'

const saving = ref(false)

const form = reactive({
  bind: '127.0.0.1',
  port: 8080,
  xray_path: '/usr/local/bin/xray',
})

async function save() {
  saving.value = true
  try {
    await new Promise((r) => setTimeout(r, 500))
    Message.success('已保存')
  } finally {
    saving.value = false
  }
}
</script>
