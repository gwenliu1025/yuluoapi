<template>
  <div
    data-testid="yuluo-home"
    class="min-h-screen overflow-hidden bg-primary-50 text-primary-900 transition-colors duration-300 dark:bg-primary-950 dark:text-primary-100"
  >
    <header class="relative z-30 h-[72px] border-b border-primary-900/10 bg-white/75 px-4 backdrop-blur-md sm:px-8 dark:border-white/10 dark:bg-primary-950/80">
      <nav class="mx-auto flex h-full max-w-7xl items-center justify-between gap-4">
        <router-link to="/home" class="brand-wordmark shrink-0 text-xl font-semibold tracking-[0.08em] text-primary-800 dark:text-primary-200 sm:text-2xl">
          雨落 API
        </router-link>

        <div class="flex min-w-0 items-center gap-2 text-sm text-primary-800 dark:text-primary-200 sm:gap-4">
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            data-testid="nav-model-plaza"
            class="hidden rounded-lg px-2 py-2 transition-colors hover:bg-primary-600/10 sm:inline-flex"
          >
            {{ t('nav.modelPlaza') }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex h-10 w-10 items-center justify-center rounded-lg px-2 py-2 transition-colors hover:bg-primary-600/10 sm:h-auto sm:w-auto sm:gap-2"
          >
            <Icon name="book" size="sm" />
            <span class="hidden sm:inline">{{ t('home.docs') }}</span>
          </a>
          <LocaleSwitcher />
          <button
            type="button"
            class="inline-flex h-10 w-10 items-center justify-center rounded-full transition-colors hover:bg-primary-600/10"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="$emit('toggle-theme')"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex min-h-10 shrink-0 items-center justify-center rounded-full bg-primary-800 px-4 py-2 font-medium text-white transition-all hover:-translate-y-0.5 hover:bg-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-gold-deep sm:px-5"
          >
            <template v-if="isAuthenticated">{{ t('home.goToDashboard') }}</template>
            <template v-else>
              <span class="sm:hidden">{{ t('home.login') }}</span>
              <span class="hidden sm:inline">{{ t('home.login') }} / {{ t('home.dashboard') }}</span>
            </template>
          </router-link>
        </div>
      </nav>
    </header>

    <main class="brand-main">
      <section class="brand-hero relative isolate flex items-center justify-center overflow-hidden px-5 py-14 sm:px-8">
        <img
          :src="isDark ? darkBackground : lightBackground"
          alt=""
          class="absolute inset-0 -z-20 h-full w-full object-cover object-center"
          aria-hidden="true"
        />
        <div class="absolute inset-0 -z-10 bg-white/8 dark:bg-primary-950/8" aria-hidden="true"></div>
        <RainMotionLayer />

        <div class="relative z-10 mx-auto flex max-w-5xl flex-col items-center text-center sm:-translate-y-10">
          <h1 class="brand-title text-[clamp(3.7rem,6.6vw,6.7rem)] font-medium leading-none tracking-[0.025em] text-primary-900 dark:text-primary-100">
            雨落 API
          </h1>
          <p class="mt-4 text-xl tracking-[0.12em] text-primary-800 dark:text-primary-200 sm:text-2xl">
            企业级 API 服务网关
          </p>
          <p class="brand-story mt-8 max-w-5xl text-balance text-[clamp(1.2rem,2vw,1.85rem)] leading-[1.8] tracking-[0.04em] text-primary-800 dark:text-primary-200">
            <span class="block">世上没有两场相同的雨，也没有一个模型适合所有问题。</span>
            <span class="block">雨落，让每个问题遇见合适的中国模型。</span>
          </p>

          <div class="mt-8 flex flex-col items-center gap-4 sm:flex-row">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="inline-flex min-h-14 min-w-44 items-center justify-center rounded-full bg-primary-500 px-8 py-3 text-base font-medium text-white shadow-glow transition-all hover:-translate-y-0.5 hover:bg-primary-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-gold-deep"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
            </router-link>
            <router-link
              v-if="showModelPlazaEntry"
              to="/model-plaza"
              data-testid="secondary-cta"
              class="brand-cta-secondary"
            >
              {{ t('home.browseModels') }}
              <Icon name="arrowRight" size="sm" />
            </router-link>
            <a
              v-else-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              data-testid="secondary-cta"
              class="brand-cta-secondary"
            >
              {{ t('home.docs') }}
              <Icon name="arrowRight" size="sm" />
            </a>
          </div>
        </div>
      </section>

      <section class="featured-models bg-white px-4 py-8 dark:bg-primary-950 sm:px-6 sm:py-8" aria-labelledby="featured-models-title">
        <div class="mx-auto max-w-7xl">
          <div class="mb-6 flex items-center gap-4">
            <span class="h-px flex-1 bg-primary-900/15 dark:bg-white/12"></span>
            <span class="brand-gold-dot h-1.5 w-1.5 shrink-0 rounded-full"></span>
            <h2 id="featured-models-title" class="shrink-0 text-lg font-medium tracking-[0.1em] text-primary-800 dark:text-primary-200 sm:text-xl">
              {{ t('home.featuredChineseModels') }}
            </h2>
            <span class="brand-gold-dot h-1.5 w-1.5 shrink-0 rounded-full"></span>
            <span class="h-px flex-1 bg-primary-900/15 dark:bg-white/12"></span>
          </div>

          <div class="grid grid-cols-2 gap-x-4 gap-y-6 md:grid-cols-5 md:divide-x md:divide-primary-900/10 dark:md:divide-white/10">
            <div
              v-for="model in featuredModels"
              :key="model.key"
              data-testid="featured-model"
              class="flex min-h-12 items-center justify-center gap-3 px-3 text-lg font-medium text-primary-900 dark:text-primary-100 sm:text-xl"
            >
              <img
                v-if="model.key === 'kimi'"
                data-testid="kimi-brand-icon"
                :src="kimiBrandIcon"
                alt=""
                class="h-9 w-9 shrink-0"
                :style="isDark ? { filter: 'invert(1) brightness(1.35)' } : undefined"
              />
              <ModelIcon
                v-else
                :model="model.iconModel"
                size="36px"
              />
              <span>{{ model.label }}</span>
            </div>
            <div
              data-testid="all-models-slot"
              class="col-span-2 flex min-h-12 items-center justify-center px-3 md:col-span-1"
            >
              <router-link
                v-if="showModelPlazaEntry"
                to="/model-plaza"
                data-testid="all-models-link"
                class="inline-flex items-center justify-center gap-2 text-sm font-medium text-primary-600 transition-colors hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
              >
                <span>{{ t('home.moreChineseModels') }}</span>
                <span aria-hidden="true">·</span>
                <span>{{ t('home.viewAllModels') }}</span>
                <Icon name="arrowRight" size="sm" />
              </router-link>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { useAppStore } from '@/stores/app'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import darkBackground from '@/assets/home/hero-night-rain-dark.webp'
import lightBackground from '@/assets/home/hero-rain-celadon-light.webp'
import kimiBrandIcon from '@/assets/home/kimi-brand.svg'
import RainMotionLayer from './RainMotionLayer.vue'

const props = defineProps<{
  isAuthenticated: boolean
  dashboardPath: string
  docUrl: string
  isDark: boolean
}>()

defineEmits<{
  'toggle-theme': []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))
const modelPlazaRequiresAuth = computed(
  () => appStore.cachedPublicSettings?.model_plaza_require_auth === true,
)
const showModelPlazaEntry = computed(
  () => modelPlazaEnabled.value && (props.isAuthenticated || !modelPlazaRequiresAuth.value),
)

const featuredModels = [
  { key: 'deepseek', label: 'DeepSeek', iconModel: 'deepseek' },
  { key: 'glm', label: '智谱 GLM', iconModel: 'glm-4' },
  { key: 'kimi', label: 'Kimi', iconModel: 'kimi' },
  { key: 'qwen', label: 'Qwen', iconModel: 'qwen' },
]
</script>

<style scoped>
.brand-wordmark,
.brand-title,
.brand-story {
  font-family: 'Noto Serif SC', 'Source Han Serif SC', 'Songti SC', STSong, serif;
}

.brand-main {
  --brand-gold: theme('colors.brand.gold');
  display: grid;
  min-height: calc(100svh - 72px);
  grid-template-rows: minmax(34rem, 1fr) auto;
}

.brand-hero {
  min-height: calc(100svh - 17rem);
}

.featured-models {
  min-height: 12.5rem;
}

.brand-cta-secondary {
  @apply inline-flex min-h-14 min-w-44 items-center justify-center gap-3 rounded-full
    bg-white/45 px-8 py-3 text-base font-medium text-primary-700 backdrop-blur-sm transition-all
    hover:-translate-y-0.5 hover:bg-white/75 focus-visible:outline focus-visible:outline-2
    focus-visible:outline-offset-2 dark:bg-primary-950/45 dark:text-primary-100 dark:hover:bg-primary-950/75;
  border: 1px solid var(--brand-gold);
  outline-color: var(--brand-gold);
}

.brand-gold-dot {
  background-color: var(--brand-gold);
}

@media (max-width: 767px) {
  .brand-main {
    grid-template-rows: minmax(40rem, auto) auto;
  }

  .brand-hero {
    min-height: 40rem;
  }

  .featured-models {
    min-height: auto;
  }
}
</style>
