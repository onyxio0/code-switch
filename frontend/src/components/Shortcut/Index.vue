<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import ListItem from '../Setting/ListRow.vue'
import ShortcutInput from '../Setting/ShortcutInput.vue'
import { parseShortcutToHotkey, formatHotkeyStringmac } from '../../utils/hotkeyUtils'
import { UpHotkey, GetHotkeys } from '../../../bindings/codeswitch/services/suistore'
import { showToast } from '../../utils/toast'

const { t } = useI18n()

const openShortcut = ref('')
const openSetting = ref('')

const isInitialized = ref(false)

const saveHotkey = async (slot: number, shortcut: string) => {
  if (!shortcut) return
  try {
    const parsed = parseShortcutToHotkey(shortcut)
    if (!parsed) {
      showToast(t('components.shortcut.messages.invalid'), 'error')
      return
    }
    await UpHotkey(slot, parsed.key, parsed.modifier)
    showToast(t('components.shortcut.messages.saved'))
  } catch (error: any) {
    showToast(`${t('components.shortcut.messages.failed')}: ${error.message}`, 'error')
  }
}

watch(openShortcut, async (value) => {
  if (!isInitialized.value) return
  await saveHotkey(1, value)
}, { flush: 'post' })

watch(openSetting, async (value) => {
  if (!isInitialized.value) return
  await saveHotkey(2, value)
}, { flush: 'post' })

const loadHotkeys = async () => {
  const entry = await GetHotkeys()
  if (entry && entry.length > 0) {
    openShortcut.value = formatHotkeyStringmac(entry[0].keycode, entry[0].modifiers)
    if (entry[1]) {
      openSetting.value = formatHotkeyStringmac(entry[1].keycode, entry[1].modifiers)
    }
    await nextTick()
    isInitialized.value = true
  }
}

onMounted(async () => {
  await loadHotkeys()
})
</script>

<template>
  <section>
    <h2 class="mac-section-title">{{ $t('components.shortcut.title') }}</h2>
    <p class="mac-section-description">{{ $t('components.shortcut.helper') }}</p>
    <div class="mac-panel">
      <ListItem
        :label="$t('components.shortcut.rows.openMain')"
        :subLabel="$t('components.shortcut.descriptions.openMain')"
      >
        <ShortcutInput v-model:modelValue="openShortcut" />
      </ListItem>
      <ListItem
        :label="$t('components.shortcut.rows.openPreferences')"
        :subLabel="$t('components.shortcut.descriptions.openPreferences')"
      >
        <ShortcutInput v-model:modelValue="openSetting" />
      </ListItem>
    </div>
  </section>
</template>
