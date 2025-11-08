<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ListItem from '../Setting/ListRow.vue'
import LanguageSwitcher from '../Setting/LanguageSwitcher.vue'
import ThemeSetting from '../Setting/ThemeSetting.vue'
import { fetchAppSettings, saveAppSettings, type AppSettings } from '../../services/appSettings'

const router = useRouter()
const heatmapEnabled = ref(true)
const settingsLoading = ref(true)
const saveBusy = ref(false)

const goBack = () => {
  router.push('/')
}

const loadAppSettings = async () => {
  settingsLoading.value = true
  try {
    const data = await fetchAppSettings()
    heatmapEnabled.value = data?.show_heatmap ?? true
  } catch (error) {
    console.error('failed to load app settings', error)
    heatmapEnabled.value = true
  } finally {
    settingsLoading.value = false
  }
}

const persistAppSettings = async () => {
  if (settingsLoading.value || saveBusy.value) return
  saveBusy.value = true
  try {
    const payload: AppSettings = { show_heatmap: heatmapEnabled.value }
    await saveAppSettings(payload)
    window.dispatchEvent(new CustomEvent('app-settings-updated'))
  } catch (error) {
    console.error('failed to save app settings', error)
  } finally {
    saveBusy.value = false
  }
}

onMounted(() => {
  void loadAppSettings()
})
</script>

<template>
  <div class="general-page">
    <header class="general-header">
      <button class="ghost-icon" @click="goBack">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M15 18l-6-6 6-6"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
      <h1>{{ $t('components.general.title.application') }}</h1>
    </header>

    <section>
      <h2 class="mac-section-title">{{ $t('components.general.title.application') }}</h2>
      <div class="mac-panel">
        <ListItem :label="$t('components.general.label.heatmap')">
          <label class="mac-switch">
            <input
              type="checkbox"
              :disabled="settingsLoading || saveBusy"
              v-model="heatmapEnabled"
              @change="persistAppSettings"
            />
            <span></span>
          </label>
        </ListItem>
      </div>
    </section>

    <section>
      <h2 class="mac-section-title">{{ $t('components.general.title.exterior') }}</h2>
      <div class="mac-panel">
        <ListItem :label="$t('components.general.label.language')">
          <LanguageSwitcher />
        </ListItem>
        <ListItem :label="$t('components.general.label.theme')">
          <ThemeSetting />
        </ListItem>
      </div>
    </section>

  </div>
</template>
