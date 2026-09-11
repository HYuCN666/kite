<template>
  <div class="admins-container">
    <div class="page-header">
      <div>
        <h2 class="section-title">管理员管理</h2>
        <p class="section-desc">管理面板管理员账号与角色权限</p>
      </div>
      <div>
        <a-button type="primary" @click="openCreate">
          <template #icon><icon-user-add /></template>
          添加管理员
        </a-button>
      </div>
    </div>

    <div class="custom-card table-card">
      <a-table :data="admins" :loading="loading" :pagination="false" row-key="id">
        <template #columns>
          <a-table-column title="用户名" data-index="username">
            <template #cell="{ record }"><span class="mono-font">{{ record.username }}</span></template>
          </a-table-column>
          <a-table-column title="角色" :width="140">
            <template #cell="{ record }">
              <a-tag :color="record.role === 'admin' ? 'arcoblue' : 'gray'">
                {{ record.role === 'admin' ? '管理员' : '操作员' }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="2FA" :width="120">
            <template #cell="{ record }">
              <a-tag :color="record.totp_enabled ? 'green' : 'gray'">
                {{ record.totp_enabled ? '已开启' : '未开启' }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="操作" :width="200" align="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="openEdit(record)"><template #icon><icon-edit /></template>编辑</a-button>
                <a-popconfirm content="确定删除此管理员？" type="warning" @ok="remove(record)">
                  <a-button type="text" status="danger" size="small"><template #icon><icon-delete /></template>删除</a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <a-drawer :visible="drawerVisible" :title="isEdit ? '编辑管理员' : '添加管理员'" width="480px" :ok-loading="submitLoading" ok-text="保存" @ok="save" @cancel="drawerVisible = false">
      <a-form ref="formRef" :model="form" layout="vertical">
        <a-form-item field="username" label="用户名" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model="form.username" :disabled="isEdit" placeholder="登录用户名" />
        </a-form-item>
        <a-form-item v-if="!isEdit" field="password" label="密码" :rules="[{ required: true, message: '请输入密码' }, { minLength: 8, message: '至少 8 位' }]">
          <a-input-password v-model="form.password" placeholder="至少 8 位" />
        </a-form-item>
        <a-form-item v-else field="password" label="重置密码（留空不修改）">
          <a-input-password v-model="form.password" placeholder="留空则不修改" />
        </a-form-item>
        <a-form-item field="role" label="角色">
          <a-radio-group v-model="form.role" type="button">
            <a-radio value="admin">管理员</a-radio>
            <a-radio value="operator">操作员</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import type { FormInstance } from '@arco-design/web-vue'
import { adminApi } from '@/api'
import type { AdminItem } from '@/types/api'

const admins = ref<AdminItem[]>([])
const loading = ref(false)
const drawerVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance | null>(null)

const emptyForm = () => ({ id: 0, username: '', password: '', role: 'operator' })
const form = reactive(emptyForm())

async function load() {
  loading.value = true
  try {
    const res = await adminApi.getList()
    admins.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  isEdit.value = false
  Object.assign(form, emptyForm())
  drawerVisible.value = true
}

function openEdit(record: AdminItem) {
  isEdit.value = true
  Object.assign(form, { id: record.id, username: record.username, password: '', role: record.role })
  drawerVisible.value = true
}

async function save() {
  const err = await formRef.value?.validate()
  if (err) return
  submitLoading.value = true
  try {
    if (isEdit.value) {
      await adminApi.update(form.id, { role: form.role, password: form.password })
    } else {
      await adminApi.create({ username: form.username, password: form.password, role: form.role })
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

async function remove(record: AdminItem) {
  await adminApi.delete(record.id)
  Message.success('已删除')
  load()
}

onMounted(load)
</script>

<style scoped>
.admins-container { display: flex; flex-direction: column; gap: 16px; }
.page-header { display: flex; align-items: flex-end; justify-content: space-between; }
.section-title { font-size: 18px; font-weight: 600; margin: 0; color: var(--text-main); }
.section-desc { font-size: 12.5px; color: var(--text-muted); margin: 4px 0 0 0; }
.table-card { padding: 12px; }
.mono-font { font-family: monospace; }
</style>
