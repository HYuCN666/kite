<template>
  <a-config-provider>
    <router-view />
  </a-config-provider>
</template>

<script setup lang="ts">
import { useAppStore } from '@/store/app'

// 激活应用状态与主题
useAppStore()
</script>

<style>
#app {
  min-height: 100vh;
  width: 100vw;
  overflow-x: hidden;
}
</style>
