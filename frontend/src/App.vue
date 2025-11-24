<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, ref } from 'vue'
import { Events } from '@wailsio/runtime'
import DeepLinkImportDialog from './components/DeepLinkImportDialog.vue'

const applyTheme = () => {
  const userTheme = localStorage.getItem('theme')
  const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  const isDark = userTheme === 'dark' || (!userTheme && systemPrefersDark)

  document.documentElement.classList.toggle('dark', isDark)
}

// 深度链接对话框状态
const showDeepLinkDialog = ref(false)
const deepLinkURL = ref('')

onMounted(() => {
  applyTheme()

  // 可监听系统主题变化自动更新（可选）
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    applyTheme()
  })

  // 监听深度链接导入事件
  Events.On('deeplink:import', (data: any) => {
    console.log('[DeepLink] 接收事件:', data)
    deepLinkURL.value = data.url
    showDeepLinkDialog.value = true
  })
})

const handleDeepLinkClose = () => {
  showDeepLinkDialog.value = false
  deepLinkURL.value = ''
}

const handleDeepLinkImported = (providerId: string) => {
  console.log('[DeepLink] 导入成功:', providerId)
  // 可以在这里添加导航或通知逻辑
}
</script>

<template>
  <RouterView />

  <!-- 深度链接导入对话框 -->
  <DeepLinkImportDialog
    :show="showDeepLinkDialog"
    :url="deepLinkURL"
    @close="handleDeepLinkClose"
    @imported="handleDeepLinkImported"
  />
</template>
