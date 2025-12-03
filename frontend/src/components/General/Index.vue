<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Dialogs } from '@wailsio/runtime'
import ListItem from '../Setting/ListRow.vue'
import LanguageSwitcher from '../Setting/LanguageSwitcher.vue'
import ThemeSetting from '../Setting/ThemeSetting.vue'
import { fetchAppSettings, saveAppSettings, type AppSettings } from '../../services/appSettings'
<<<<<<< HEAD
import {
  fetchConfigImportStatus,
  fetchConfigImportStatusForFile,
  importFromCcSwitch,
  importFromCustomFile,
  type ConfigImportResult,
  type ConfigImportStatus,
} from '../../services/configImport'
import { showToast } from '../../utils/toast'
import BaseButton from '../common/BaseButton.vue'

const router = useRouter()
const { t } = useI18n()
const heatmapEnabled = ref(true)
const homeTitleVisible = ref(true)
const autoStartEnabled = ref(false)
=======
import { checkUpdate, downloadUpdate, restartApp, getUpdateState, setAutoCheckEnabled, type UpdateState } from '../../services/update'
import { fetchCurrentVersion } from '../../services/version'
import { getBlacklistSettings, updateBlacklistSettings, getLevelBlacklistEnabled, setLevelBlacklistEnabled, getBlacklistEnabled, setBlacklistEnabled, type BlacklistSettings } from '../../services/settings'

const router = useRouter()
// 从 localStorage 读取缓存值作为初始值，避免加载时的视觉闪烁
const getCachedValue = (key: string, defaultValue: boolean): boolean => {
  const cached = localStorage.getItem(`app-settings-${key}`)
  return cached !== null ? cached === 'true' : defaultValue
}
const heatmapEnabled = ref(getCachedValue('heatmap', true))
const homeTitleVisible = ref(getCachedValue('homeTitle', true))
const autoStartEnabled = ref(getCachedValue('autoStart', false))
const autoUpdateEnabled = ref(getCachedValue('autoUpdate', true))
>>>>>>> rogers/main
const settingsLoading = ref(true)
const saveBusy = ref(false)
const importStatus = ref<ConfigImportStatus | null>(null)
const customImportStatus = ref<ConfigImportStatus | null>(null)
const importBusy = ref(false)

// 更新相关状态
const updateState = ref<UpdateState | null>(null)
const checking = ref(false)
const downloading = ref(false)
const appVersion = ref('')

// 拉黑配置相关状态
const blacklistEnabled = ref(true)  // 拉黑功能总开关
const blacklistThreshold = ref(3)
const blacklistDuration = ref(30)
const levelBlacklistEnabled = ref(false)
const blacklistLoading = ref(false)
const blacklistSaving = ref(false)

const goBack = () => {
  router.push('/')
}

const loadAppSettings = async () => {
  settingsLoading.value = true
  try {
    const data = await fetchAppSettings()
    heatmapEnabled.value = data?.show_heatmap ?? true
    homeTitleVisible.value = data?.show_home_title ?? true
    autoStartEnabled.value = data?.auto_start ?? false
    autoUpdateEnabled.value = data?.auto_update ?? true

    // 缓存到 localStorage，下次打开时直接显示正确状态
    localStorage.setItem('app-settings-heatmap', String(heatmapEnabled.value))
    localStorage.setItem('app-settings-homeTitle', String(homeTitleVisible.value))
    localStorage.setItem('app-settings-autoStart', String(autoStartEnabled.value))
    localStorage.setItem('app-settings-autoUpdate', String(autoUpdateEnabled.value))
  } catch (error) {
    console.error('failed to load app settings', error)
    heatmapEnabled.value = true
    homeTitleVisible.value = true
    autoStartEnabled.value = false
    autoUpdateEnabled.value = true
  } finally {
    settingsLoading.value = false
  }
}

const persistAppSettings = async () => {
  if (settingsLoading.value || saveBusy.value) return
  saveBusy.value = true
  try {
    const payload: AppSettings = {
      show_heatmap: heatmapEnabled.value,
      show_home_title: homeTitleVisible.value,
      auto_start: autoStartEnabled.value,
      auto_update: autoUpdateEnabled.value,
    }
    await saveAppSettings(payload)

    // 同步自动更新设置到 UpdateService
    await setAutoCheckEnabled(autoUpdateEnabled.value)

    // 更新缓存
    localStorage.setItem('app-settings-heatmap', String(heatmapEnabled.value))
    localStorage.setItem('app-settings-homeTitle', String(homeTitleVisible.value))
    localStorage.setItem('app-settings-autoStart', String(autoStartEnabled.value))
    localStorage.setItem('app-settings-autoUpdate', String(autoUpdateEnabled.value))

    window.dispatchEvent(new CustomEvent('app-settings-updated'))
  } catch (error) {
    console.error('failed to save app settings', error)
  } finally {
    saveBusy.value = false
  }
}

<<<<<<< HEAD
onMounted(() => {
  void loadAppSettings()
  void loadImportStatus()
=======
const loadUpdateState = async () => {
  try {
    updateState.value = await getUpdateState()
  } catch (error) {
    console.error('failed to load update state', error)
  }
}

const checkUpdateManually = async () => {
  checking.value = true
  try {
    const info = await checkUpdate()
    await loadUpdateState()

    if (!info.available) {
      alert('已是最新版本')
    } else {
      // 发现新版本，提示用户并开始下载
      const confirmed = confirm(`发现新版本 ${info.version}，是否立即下载？`)
      if (confirmed) {
        downloading.value = true
        checking.value = false
        try {
          await downloadUpdate()
          await loadUpdateState()

          // 下载完成，提示重启
          const restart = confirm('新版本已下载完成，是否立即重启应用？')
          if (restart) {
            await restartApp()
          }
        } catch (downloadError) {
          console.error('download failed', downloadError)
          alert('下载失败，请稍后重试')
        } finally {
          downloading.value = false
        }
      }
    }
  } catch (error) {
    console.error('check update failed', error)
    alert('检查更新失败，请检查网络连接')
  } finally {
    checking.value = false
  }
}

const downloadAndInstall = async () => {
  downloading.value = true
  try {
    await downloadUpdate()
    await loadUpdateState()

    // 弹窗确认重启
    const confirmed = confirm('新版本已下载完成，是否立即重启应用？')
    if (confirmed) {
      await restartApp()
    }
  } catch (error) {
    console.error('download failed', error)
    alert('下载失败，请稍后重试')
  } finally {
    downloading.value = false
  }
}

// 当更新已下载完成时，直接安装并重启（无需再次下载）
const installAndRestart = async () => {
  const confirmed = confirm('是否立即安装更新并重启应用？')
  if (confirmed) {
    try {
      await restartApp()
    } catch (error) {
      console.error('restart failed', error)
      alert('重启失败，请手动重启应用')
    }
  }
}

const formatLastCheckTime = (timeStr?: string) => {
  if (!timeStr) return '从未检查'

  const checkTime = new Date(timeStr)
  const now = new Date()
  const diffMs = now.getTime() - checkTime.getTime()
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))

  if (diffHours < 1) {
    return '刚刚'
  } else if (diffHours < 24) {
    return `${diffHours} 小时前`
  } else {
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays} 天前`
  }
}

// 加载拉黑配置
const loadBlacklistSettings = async () => {
  blacklistLoading.value = true
  try {
    const settings = await getBlacklistSettings()
    blacklistThreshold.value = settings.failureThreshold
    blacklistDuration.value = settings.durationMinutes

    // 加载拉黑功能总开关
    const enabled = await getBlacklistEnabled()
    blacklistEnabled.value = enabled

    // 加载等级拉黑开关状态
    const levelEnabled = await getLevelBlacklistEnabled()
    levelBlacklistEnabled.value = levelEnabled
  } catch (error) {
    console.error('failed to load blacklist settings', error)
    // 使用默认值
    blacklistEnabled.value = true
    blacklistThreshold.value = 3
    blacklistDuration.value = 30
    levelBlacklistEnabled.value = false
  } finally {
    blacklistLoading.value = false
  }
}

// 保存拉黑配置
const saveBlacklistSettings = async () => {
  if (blacklistLoading.value || blacklistSaving.value) return
  blacklistSaving.value = true
  try {
    await updateBlacklistSettings(blacklistThreshold.value, blacklistDuration.value)
    alert('拉黑配置已保存')
  } catch (error) {
    console.error('failed to save blacklist settings', error)
    alert('保存失败：' + (error as Error).message)
  } finally {
    blacklistSaving.value = false
  }
}

// 切换拉黑功能总开关
const toggleBlacklist = async () => {
  if (blacklistLoading.value || blacklistSaving.value) return
  blacklistSaving.value = true
  try {
    await setBlacklistEnabled(blacklistEnabled.value)
  } catch (error) {
    console.error('failed to toggle blacklist', error)
    // 回滚状态
    blacklistEnabled.value = !blacklistEnabled.value
    alert('切换失败：' + (error as Error).message)
  } finally {
    blacklistSaving.value = false
  }
}

// 切换等级拉黑开关
const toggleLevelBlacklist = async () => {
  if (blacklistLoading.value || blacklistSaving.value) return
  blacklistSaving.value = true
  try {
    await setLevelBlacklistEnabled(levelBlacklistEnabled.value)
  } catch (error) {
    console.error('failed to toggle level blacklist', error)
    // 回滚状态
    levelBlacklistEnabled.value = !levelBlacklistEnabled.value
    alert('切换失败：' + (error as Error).message)
  } finally {
    blacklistSaving.value = false
  }
}

onMounted(async () => {
  await loadAppSettings()

  // 加载当前版本号
  try {
    appVersion.value = await fetchCurrentVersion()
  } catch (error) {
    console.error('failed to load app version', error)
  }

  // 加载更新状态
  await loadUpdateState()

  // 加载拉黑配置
  await loadBlacklistSettings()
>>>>>>> rogers/main
})

const loadImportStatus = async () => {
  try {
    importStatus.value = await fetchConfigImportStatus()
  } catch (error) {
    console.error('failed to load cc-switch import status', error)
    importStatus.value = null
  }
}

const activeImportStatus = computed(() => customImportStatus.value ?? importStatus.value)
const hasCustomSelection = computed(() => Boolean(customImportStatus.value))
const shouldShowDefaultMissingHint = computed(() => {
  if (hasCustomSelection.value) return false
  const status = importStatus.value
  if (!status) return false
  return !status.config_exists
})
const pendingProviders = computed(() => activeImportStatus.value?.pending_provider_count ?? 0)
const pendingServers = computed(() => activeImportStatus.value?.pending_mcp_count ?? 0)
const configPath = computed(() => activeImportStatus.value?.config_path ?? '')
const canImportDefault = computed(() => {
  const status = importStatus.value
  if (!status) return false
  return Boolean(status.pending_providers || status.pending_mcp)
})
const canImportCustom = computed(() => {
  const status = customImportStatus.value
  if (!status) return false
  return Boolean(status.pending_providers || status.pending_mcp)
})
const canImportActive = computed(() =>
  hasCustomSelection.value ? canImportCustom.value : canImportDefault.value,
)
const showImportRow = computed(() => Boolean(importStatus.value) || hasCustomSelection.value)
const importPathLabel = computed(() => {
  if (!configPath.value) return ''
  return t('components.general.import.path', { path: configPath.value })
})
const importDetailLabel = computed(() => {
  if (shouldShowDefaultMissingHint.value) {
    return t('components.general.import.missingDefault')
  }
  if (!activeImportStatus.value) {
    return t('components.general.import.noFile')
  }
  const detail = canImportActive.value
    ? t('components.general.import.detail', {
        providers: pendingProviders.value,
        servers: pendingServers.value,
      })
    : t('components.general.import.synced')
  if (!importPathLabel.value) return detail
  return `${importPathLabel.value} · ${detail}`
})
const importButtonText = computed(() => {
  if (importBusy.value) {
    return t('components.general.import.importing')
  }
  if (hasCustomSelection.value) {
    return t('components.general.import.confirm')
  }
  if (shouldShowDefaultMissingHint.value || canImportDefault.value) {
    return t('components.general.import.cta')
  }
  return t('components.general.import.syncedButton')
})
const primaryButtonDisabled = computed(() => importBusy.value || !canImportActive.value)
const secondaryButtonLabel = computed(() =>
  hasCustomSelection.value
    ? t('components.general.import.clear')
    : t('components.general.import.upload'),
)
const secondaryButtonVariant = computed(() => 'outline' as const)

const processImportResult = async (result?: ConfigImportResult | null) => {
  if (!result) return
  if (hasCustomSelection.value && result.status?.config_path === customImportStatus.value?.config_path) {
    customImportStatus.value = result.status
  } else {
    importStatus.value = result.status
  }
  const importedProviders = result.imported_providers ?? 0
  const importedServers = result.imported_mcp ?? 0
  if (importedProviders > 0 || importedServers > 0) {
    showToast(
      t('components.main.importConfig.success', {
        providers: importedProviders,
        servers: importedServers,
      })
    )
  } else if (result.status?.config_exists) {
    showToast(t('components.main.importConfig.empty'))
  }
  await loadImportStatus()
}

const handleImportClick = async () => {
  if (importBusy.value || !importStatus.value || !canImportDefault.value) return
  importBusy.value = true
  try {
    const result = await importFromCcSwitch()
    await processImportResult(result)
  } catch (error) {
    console.error('failed to import cc-switch config', error)
    showToast(t('components.main.importConfig.error'), 'error')
  } finally {
    importBusy.value = false
  }
}

const handleConfirmCustomImport = async () => {
  const path = customImportStatus.value?.config_path
  if (!path || importBusy.value || !canImportCustom.value) return
  importBusy.value = true
  try {
    const result = await importFromCustomFile(path)
    await processImportResult(result)
  } catch (error) {
    console.error('failed to import custom cc-switch config', error)
    showToast(t('components.main.importConfig.error'), 'error')
  } finally {
    importBusy.value = false
  }
}

const handlePrimaryImport = async () => {
  if (hasCustomSelection.value) {
    await handleConfirmCustomImport()
  } else {
    await handleImportClick()
  }
}

const handleUploadClick = async () => {
  if (importBusy.value) return
  let selectedPath = ''
  try {
    const selection = await Dialogs.OpenFile({
      Title: t('components.general.import.uploadTitle'),
      CanChooseFiles: true,
      CanChooseDirectories: false,
      AllowsOtherFiletypes: false,
      Filters: [
        {
          DisplayName: 'JSON (*.json)',
          Pattern: '*.json',
        },
      ],
      AllowsMultipleSelection: false,
    })
    selectedPath = Array.isArray(selection) ? selection[0] : selection
    if (!selectedPath) return
    const status = await fetchConfigImportStatusForFile(selectedPath)
    customImportStatus.value = status
  } catch (error) {
    console.error('failed to load custom cc-switch config status', error)
    showToast(t('components.general.import.loadError'), 'error')
  }
}

const clearCustomSelection = () => {
  customImportStatus.value = null
}

const handleSecondaryImportAction = async () => {
  if (hasCustomSelection.value) {
    clearCustomSelection()
  } else {
    await handleUploadClick()
  }
}
</script>

<template>
  <div class="main-shell general-shell">
    <div class="global-actions">
      <p class="global-eyebrow">{{ $t('components.general.title.application') }}</p>
      <button class="ghost-icon" :aria-label="$t('components.general.buttons.back')" @click="goBack">
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
    </div>

    <div class="general-page">
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
          <ListItem :label="$t('components.general.label.homeTitle')">
            <label class="mac-switch">
              <input
                type="checkbox"
                :disabled="settingsLoading || saveBusy"
                v-model="homeTitleVisible"
                @change="persistAppSettings"
              />
              <span></span>
            </label>
          </ListItem>
          <ListItem :label="$t('components.general.label.autoStart')">
            <label class="mac-switch">
              <input
                type="checkbox"
                :disabled="settingsLoading || saveBusy"
                v-model="autoStartEnabled"
                @change="persistAppSettings"
              />
              <span></span>
            </label>
          </ListItem>
          <ListItem
            v-if="showImportRow"
            :label="$t('components.general.import.label')"
            :sub-label="importDetailLabel"
          >
            <div class="import-actions">
              <BaseButton
                size="sm"
                variant="outline"
                type="button"
                :disabled="primaryButtonDisabled"
                @click="handlePrimaryImport"
              >
                {{ importButtonText }}
              </BaseButton>
              <BaseButton
                size="sm"
                :variant="secondaryButtonVariant"
                type="button"
                :disabled="importBusy"
                @click="handleSecondaryImportAction"
              >
                {{ secondaryButtonLabel }}
              </BaseButton>
              <BaseButton
                v-if="hasCustomSelection"
                size="sm"
                variant="outline"
                type="button"
                :disabled="importBusy"
                @click="handleUploadClick"
              >
                {{ $t('components.general.import.reupload') }}
              </BaseButton>
            </div>
          </ListItem>

        </div>
      </section>

      <section>
        <h2 class="mac-section-title">{{ $t('components.general.title.blacklist') }}</h2>
        <div class="mac-panel">
          <ListItem :label="$t('components.general.label.enableBlacklist')">
            <div class="toggle-with-hint">
              <label class="mac-switch">
                <input
                  type="checkbox"
                  :disabled="blacklistLoading || blacklistSaving"
                  v-model="blacklistEnabled"
                  @change="toggleBlacklist"
                />
                <span></span>
              </label>
              <span class="hint-text">{{ $t('components.general.label.enableBlacklistHint') }}</span>
            </div>
          </ListItem>
          <ListItem :label="$t('components.general.label.enableLevelBlacklist')">
            <div class="toggle-with-hint">
              <label class="mac-switch">
                <input
                  type="checkbox"
                  :disabled="blacklistLoading || blacklistSaving"
                  v-model="levelBlacklistEnabled"
                  @change="toggleLevelBlacklist"
                />
                <span></span>
              </label>
              <span class="hint-text">{{ $t('components.general.label.enableLevelBlacklistHint') }}</span>
            </div>
          </ListItem>
          <ListItem :label="$t('components.general.label.blacklistThreshold')">
            <select
              v-model.number="blacklistThreshold"
              :disabled="blacklistLoading || blacklistSaving"
              class="mac-select">
              <option :value="1">1 {{ $t('components.general.label.times') }}</option>
              <option :value="2">2 {{ $t('components.general.label.times') }}</option>
              <option :value="3">3 {{ $t('components.general.label.times') }}</option>
              <option :value="4">4 {{ $t('components.general.label.times') }}</option>
              <option :value="5">5 {{ $t('components.general.label.times') }}</option>
              <option :value="6">6 {{ $t('components.general.label.times') }}</option>
              <option :value="7">7 {{ $t('components.general.label.times') }}</option>
              <option :value="8">8 {{ $t('components.general.label.times') }}</option>
              <option :value="9">9 {{ $t('components.general.label.times') }}</option>
            </select>
          </ListItem>
          <ListItem :label="$t('components.general.label.blacklistDuration')">
            <select
              v-model.number="blacklistDuration"
              :disabled="blacklistLoading || blacklistSaving"
              class="mac-select">
              <option :value="5">5 {{ $t('components.general.label.minutes') }}</option>
              <option :value="15">15 {{ $t('components.general.label.minutes') }}</option>
              <option :value="30">30 {{ $t('components.general.label.minutes') }}</option>
              <option :value="60">60 {{ $t('components.general.label.minutes') }}</option>
            </select>
          </ListItem>
          <ListItem :label="$t('components.general.label.saveBlacklist')">
            <button
              @click="saveBlacklistSettings"
              :disabled="blacklistLoading || blacklistSaving"
              class="primary-btn">
              {{ blacklistSaving ? $t('components.general.label.saving') : $t('components.general.label.save') }}
            </button>
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

      <section>
        <h2 class="mac-section-title">{{ $t('components.general.title.update') }}</h2>
        <div class="mac-panel">
          <ListItem :label="$t('components.general.label.autoUpdate')">
            <label class="mac-switch">
              <input
                type="checkbox"
                :disabled="settingsLoading || saveBusy"
                v-model="autoUpdateEnabled"
                @change="persistAppSettings"
              />
              <span></span>
            </label>
          </ListItem>

          <ListItem :label="$t('components.general.label.lastCheck')">
            <span class="info-text">{{ formatLastCheckTime(updateState?.last_check_time) }}</span>
            <span v-if="updateState && updateState.consecutive_failures > 0" class="warning-badge">
              ⚠️ {{ $t('components.general.update.checkFailed', { count: updateState.consecutive_failures }) }}
            </span>
          </ListItem>

          <ListItem :label="$t('components.general.label.currentVersion')">
            <span class="version-text">{{ appVersion }}</span>
          </ListItem>

          <ListItem
            v-if="updateState?.latest_known_version && updateState.latest_known_version !== appVersion"
            :label="$t('components.general.label.latestVersion')">
            <span class="version-text highlight">{{ updateState.latest_known_version }} 🆕</span>
          </ListItem>

          <ListItem :label="$t('components.general.label.checkNow')">
            <button
              @click="checkUpdateManually"
              :disabled="checking"
              class="action-btn">
              {{ checking ? $t('components.general.update.checking') : $t('components.general.update.checkNow') }}
            </button>
          </ListItem>

          <ListItem
            v-if="updateState?.update_ready"
            :label="$t('components.general.label.manualUpdate')">
            <button
              @click="installAndRestart"
              class="primary-btn">
              {{ $t('components.general.update.installAndRestart') }}
            </button>
          </ListItem>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
<<<<<<< HEAD
.import-actions {
  display: flex;
  gap: 0.35rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.import-actions .btn {
  min-width: 56px;
  padding: 0.3rem 0.75rem;
  font-size: 0.7rem;
}

.import-actions .btn-outline,
.import-actions .btn-ghost {
  padding-inline: 0.75rem;
=======
.toggle-with-hint {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.hint-text {
  font-size: 11px;
  color: var(--mac-text-secondary);
  line-height: 1.4;
  max-width: 320px;
  text-align: right;
  white-space: nowrap;
}

:global(.dark) .hint-text {
  color: rgba(255, 255, 255, 0.5);
>>>>>>> rogers/main
}
</style>
