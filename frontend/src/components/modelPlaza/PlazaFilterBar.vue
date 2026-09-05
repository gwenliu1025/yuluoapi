<template>
  <div class="space-y-3">
    <!-- 一级:平台 -->
    <div class="flex items-start gap-2">
      <span class="w-16 shrink-0 pt-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.platformLabel') }}
      </span>
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="p in ['all', ...platforms]"
          :key="`platform-${p}`"
          type="button"
          class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40 disabled:grayscale"
          :class="p === 'all' ? chipClass(platform === 'all') : platform === p ? 'chip-tinted-active' : 'chip-tinted'"
          :style="p === 'all' ? undefined : { '--chip-accent': platformAccentColor(p) }"
          :disabled="p !== 'all' && !platformEnabled(p)"
          @click="$emit('update:platform', p)"
        >
          <PlatformIcon v-if="p !== 'all'" :platform="p as GroupPlatform" size="xs" />
          {{ p === 'all' ? t('modelPlaza.filters.all') : p }}
        </button>
      </div>
    </div>

    <!-- 分组与“全部”互斥；未选项保持中性底色，不用平台色暗示选中。 -->
    <div class="flex items-start gap-2">
      <span class="w-16 shrink-0 pt-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.groupLabel') }}
      </span>
      <div class="flex flex-wrap items-center gap-2" role="group" :aria-label="t('modelPlaza.filters.groupLabel')">
        <button
          type="button"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition"
          :class="chipClass(groupId === 'all')"
          :aria-pressed="groupId === 'all'"
          @click="$emit('update:groupId', 'all')"
        >
          {{ t('modelPlaza.filters.all') }}
        </button>
        <button
          v-for="g in groups"
          :key="`group-${g.id}`"
          type="button"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40 disabled:grayscale"
          :class="chipClass(groupId === g.id)"
          :aria-pressed="groupId === g.id"
          :disabled="!groupEnabled(g)"
          @click="$emit('update:groupId', g.id)"
        >
          {{ g.name }}
        </button>
      </div>
    </div>

    <!-- 系列独立于 OpenAI 等接入平台，并与当前分组组合筛选。 -->
    <div class="flex items-start gap-2">
      <span class="w-16 shrink-0 pt-2 text-xs font-semibold text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.seriesLabel') }}
      </span>
      <div class="flex flex-wrap items-center gap-2" role="group" :aria-label="t('modelPlaza.filters.seriesLabel')">
        <button
          type="button"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition"
          :class="chipClass(series === 'all')"
          :aria-pressed="series === 'all'"
          @click="$emit('update:series', 'all')"
        >
          {{ t('modelPlaza.filters.all') }}
        </button>
        <button
          v-for="s in seriesOptions"
          :key="s"
          type="button"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40"
          :class="chipClass(series === s)"
          :aria-pressed="series === s"
          :title="t(`modelPlaza.filters.seriesDescriptions.${s}`)"
          :disabled="!seriesEnabled(s)"
          @click="$emit('update:series', s)"
        >
          {{ t(`modelPlaza.filters.seriesNames.${s}`) }}
        </button>
      </div>
    </div>

    <!-- 倍率(当前组合下不存在的置灰) -->
    <div class="flex items-start gap-2">
      <span class="w-16 shrink-0 pt-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.rateLabel') }}
      </span>
      <div class="flex flex-wrap items-center gap-2">
        <button
          type="button"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition"
          :class="chipClass(rate === 'all')"
          @click="$emit('update:rate', 'all')"
        >
          {{ t('modelPlaza.filters.all') }}
        </button>
        <button
          v-for="r in rates"
          :key="`rate-${r}`"
          type="button"
          class="rounded-lg px-3 py-1.5 font-mono text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40 disabled:grayscale"
          :class="chipClass(rate === r)"
          :disabled="!rateEnabled(r)"
          @click="$emit('update:rate', r)"
        >
          {{ r }}x
        </button>
      </div>
    </div>

    <!-- 四级:模型名搜索(纯前端过滤) -->
    <div class="flex flex-wrap items-start gap-2">
      <span class="w-16 shrink-0 pt-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">
        {{ t('modelPlaza.filters.modelLabel') }}
      </span>
      <div class="relative w-full sm:w-72">
        <Icon
          name="search"
          size="sm"
          class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-dark-500"
        />
        <input
          :value="search"
          type="text"
          :placeholder="t('modelPlaza.filters.searchPlaceholder')"
          class="input rounded-lg py-1.5 pl-9 pr-9"
          @input="$emit('update:search', ($event.target as HTMLInputElement).value)"
        />
        <button
          v-if="search"
          type="button"
          class="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 transition-colors hover:text-gray-600 dark:text-dark-500 dark:hover:text-gray-300"
          @click="$emit('update:search', '')"
        >
          <Icon name="x" size="xs" class="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { platformAccentColor } from '@/utils/platformColors'
import type { GroupPlatform } from '@/types'
import type { ModelSeries } from './modelSeries'

const props = defineProps<{
  /** 数据中出现的平台(去重排序后)。 */
  platforms: string[]
  /** 仅含服务端已授权分组；系列元数据不引入新的模型或价格。 */
  groups: Array<{ id: number; name: string; platform: string; rate: number; series: ModelSeries[] }>
  /** 全量生效倍率去重升序。 */
  rates: number[]
  platform: string
  groupId: number | 'all'
  rate: number | 'all'
  seriesOptions: ModelSeries[]
  series: ModelSeries | 'all'
  /** 模型名搜索词(纯前端过滤)。 */
  search: string
}>()

defineEmits<{
  'update:platform': [value: string]
  'update:groupId': [value: number | 'all']
  'update:rate': [value: number | 'all']
  'update:series': [value: ModelSeries | 'all']
  'update:search': [value: string]
}>()

const { t } = useI18n()

/**
 * 各维度互为约束：某选项可点，当且仅当其余维度下仍有分组命中。
 * 「全部」永远可点,作为解除本维约束的出口;可点项组合恒有结果,无需选择修正。
 */
function platformEnabled(p: string): boolean {
  return props.groups.some(
    (g) =>
      g.platform === p &&
      (props.groupId === 'all' || g.id === props.groupId) &&
      (props.rate === 'all' || g.rate === props.rate) &&
      (props.series === 'all' || g.series.includes(props.series))
  )
}

function groupEnabled(g: { platform: string; rate: number; series: ModelSeries[] }): boolean {
  return (
    (props.platform === 'all' || g.platform === props.platform) &&
    (props.rate === 'all' || g.rate === props.rate) &&
    (props.series === 'all' || g.series.includes(props.series))
  )
}

function rateEnabled(r: number): boolean {
  return props.groups.some(
    (g) =>
      g.rate === r &&
      (props.platform === 'all' || g.platform === props.platform) &&
      (props.groupId === 'all' || g.id === props.groupId) &&
      (props.series === 'all' || g.series.includes(props.series))
  )
}

function seriesEnabled(s: ModelSeries): boolean {
  return props.groups.some((g) =>
    g.series.includes(s) &&
    (props.platform === 'all' || g.platform === props.platform) &&
    (props.groupId === 'all' || g.id === props.groupId) &&
    (props.rate === 'all' || g.rate === props.rate)
  )
}

function chipClass(active: boolean): string {
  return active
    ? 'bg-gradient-to-r from-primary-500 to-primary-600 text-white shadow-sm shadow-primary-500/30'
    : 'bg-white text-gray-600 ring-1 ring-inset ring-gray-200 enabled:hover:bg-gray-50 enabled:hover:text-gray-900 enabled:hover:ring-gray-300 dark:bg-dark-800/60 dark:text-dark-300 dark:ring-dark-700 dark:enabled:hover:bg-dark-800 dark:enabled:hover:text-white'
}
</script>

<style scoped>
/* 平台 chip 的配色统一从 --chip-accent(平台主色)派生,新增平台无需扩展样式。
   激活态与非激活态在模板上互斥挂载,避免选择器优先级互相覆盖。 */
.chip-tinted {
  color: var(--chip-accent);
  color: color-mix(in srgb, var(--chip-accent) 78%, black);
  background-color: color-mix(in srgb, var(--chip-accent) 9%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--chip-accent) 25%, transparent);
}

.chip-tinted:not(:disabled):hover {
  background-color: color-mix(in srgb, var(--chip-accent) 16%, transparent);
}

.dark .chip-tinted {
  color: color-mix(in srgb, var(--chip-accent) 72%, white);
  background-color: color-mix(in srgb, var(--chip-accent) 12%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--chip-accent) 30%, transparent);
}

.dark .chip-tinted:not(:disabled):hover {
  background-color: color-mix(in srgb, var(--chip-accent) 18%, transparent);
}

.chip-tinted-active {
  color: #fff;
  background-color: var(--chip-accent);
  background-color: color-mix(in srgb, var(--chip-accent) 85%, black);
  box-shadow: 0 1px 2px 0 color-mix(in srgb, var(--chip-accent) 35%, transparent);
}

.chip-tinted-active:not(:disabled):hover {
  background-color: color-mix(in srgb, var(--chip-accent) 75%, black);
}

.dark .chip-tinted-active {
  background-color: color-mix(in srgb, var(--chip-accent) 80%, transparent);
}

.dark .chip-tinted-active:not(:disabled):hover {
  background-color: var(--chip-accent);
}
</style>
