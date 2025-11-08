<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const year = computed(() => new Date().getFullYear())
</script>

<template>
  <section class="space-y-3">
    <h2 class="mac-section-title">{{ t('menus.about') }}</h2>
    <div class="mac-panel about-panel">
      <div>
        <p class="about-title">{{ t('app.title') }}</p>
        <p class="about-subtitle">macOS inspired experience</p>
        <p class="about-meta">© {{ year }} — Built with Wails + Vue</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.about-panel {
  padding: 24px;
}

.about-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 6px;
}

.about-subtitle {
  color: var(--mac-text-secondary);
  margin-bottom: 12px;
}

.about-meta {
  font-size: 0.85rem;
  color: var(--mac-text-secondary);
}
</style>
