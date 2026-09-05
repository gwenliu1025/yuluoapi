<template>
  <div class="plaza-pricing-table overflow-x-auto" :style="accentStyle">
    <table class="w-full min-w-[1200px] table-fixed border-collapse text-sm tabular-nums">
      <colgroup>
        <col class="w-[23%]" />
        <col class="w-[11%]" />
        <col class="w-[9%]" />
        <col class="w-[16%]" />
        <col class="w-[11%]" />
        <col class="w-[8%]" />
        <col class="w-[16%]" />
        <col class="w-[6%]" />
      </colgroup>
      <thead>
        <tr
          class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400"
        >
          <th
            rowspan="2"
            class="border-r border-gray-100 py-2.5 pl-5 pr-4 text-left align-middle dark:border-dark-700/60"
          >
            {{ t('modelPlaza.table.model') }}
          </th>
          <th colspan="3" class="pz-bg pt-2 text-center">
            <div class="pz-title border-b pb-2 font-semibold">
              {{ t('modelPlaza.table.paidPrice') }}
              <span class="pz-unit ml-1 normal-case font-normal">{{ t('modelPlaza.table.unitPerMillion') }}</span>
            </div>
          </th>
          <th
            colspan="3"
            class="border-l border-gray-100 pt-2 text-center dark:border-dark-700/60"
          >
            <div class="border-b border-gray-200 pb-2 text-gray-400 dark:border-dark-600 dark:text-dark-500">
              {{ t('modelPlaza.table.officialPrice') }}
              <span class="ml-1 normal-case font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.unitPerMillion') }}</span>
            </div>
          </th>
          <th
            rowspan="2"
            class="border-l border-gray-100 py-2.5 pl-3 pr-5 text-right align-middle dark:border-dark-700/60"
          >
            {{ t('modelPlaza.table.rate') }}
          </th>
        </tr>
        <tr
          class="border-b border-gray-200 text-left text-[11px] font-medium uppercase leading-4 tracking-wide text-gray-400 dark:border-dark-700 dark:text-dark-500"
        >
          <th class="pz-bg px-3 py-2 font-medium">{{ t('modelPlaza.table.input') }}</th>
          <th class="pz-bg px-3 py-2 font-medium">{{ t('modelPlaza.table.output') }}</th>
          <th class="pz-bg px-3 py-2 font-medium">{{ t('modelPlaza.table.cache') }}</th>
          <th class="border-l border-gray-100 px-3 py-2 font-medium dark:border-dark-700/60">
            {{ t('modelPlaza.table.input') }}
          </th>
          <th class="px-3 py-2 font-medium">{{ t('modelPlaza.table.output') }}</th>
          <th class="px-3 py-2 font-medium">{{ t('modelPlaza.table.cache') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="{ model: m, period, pricingMode, key } in rows"
          :key="key"
          class="border-b border-gray-100 transition-colors last:border-b-0 hover:bg-gray-50/70 dark:border-dark-800 dark:hover:bg-dark-800/50"
        >
          <!-- 模型名 + 非 token 计费模式徽章;分时时段行额外标注时段 -->
          <td class="border-r border-gray-100 py-2.5 pl-5 pr-4 align-middle dark:border-dark-700/60">
            <div class="flex flex-wrap items-center gap-1.5">
              <span class="font-medium text-gray-900 dark:text-white">{{ m.name }}</span>
              <span
                v-if="pricingMode"
                class="inline-flex items-center whitespace-nowrap rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
              >
                {{ t(`modelPlaza.table.${pricingMode}`) }}
              </span>
              <!-- 时段徽章紧跟模型名,其余徽章排在后面,空间不足时先换行的是它们 -->
              <span
                v-if="period"
                class="inline-flex items-center whitespace-nowrap rounded-md bg-gray-100 px-1 py-0.5 font-mono text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
                :title="timePricingRowHint(m)"
              >
                <span v-if="m.time_pricing?.weekdays_only" class="mr-1 font-sans">{{
                  t('modelPlaza.table.timePricingWeekdays')
                }}</span>
                {{ formatTimeWindow(period) }}
              </span>
              <span
                v-else-if="timePeriods(m).length"
                class="inline-flex items-center whitespace-nowrap rounded-md bg-gray-100 px-1 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
                :title="t('modelPlaza.table.timePricingStandardHint', { timezone: m.time_pricing?.timezone })"
              >
                {{ t('modelPlaza.table.timePricingStandard') }}
              </span>
              <span
                v-if="platform && m.platform !== platform"
                :class="[
                  'inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium',
                  platformBadgeLightClass(m.platform)
                ]"
              >
                {{ platformLabel(m.platform) }}
              </span>
              <span
                v-if="billingMode(m) !== BILLING_MODE_TOKEN"
                class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
              >
                {{ billingModeLabel(m) }}
              </span>
              <span
                v-if="m.long_context_basis === 'marginal'"
                class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700/70 dark:text-dark-300"
                :title="t('modelPlaza.table.tierHintMarginal')"
              >
                {{ t('modelPlaza.table.marginalBadge') }}
              </span>
            </div>
          </td>

          <!-- token 计费:输入 / 输出 / 缓存(写/读),有阶梯时每档一行;档位标签只放输入列,其余列按行对齐 -->
          <template v-if="billingMode(m) === BILLING_MODE_TOKEN">
            <td class="pz-cell px-3 py-2.5 align-middle font-mono font-semibold text-gray-900 dark:text-gray-50">
              <template v-if="tokenIntervals(m).length">
                <div
                  v-for="(iv, idx) in tokenIntervals(m)"
                  :key="idx"
                  class="tier-price-block flex items-center whitespace-nowrap text-xs leading-5"
                  :style="`--tier-lines: ${cacheTierLineCount(iv)}`"
                >
                  <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500" :title="tierHint(m)">{{ tierLabel(iv) }}</span>
                  {{ paidPerMillion(iv.input_price, period) }}
                </div>
              </template>
              <template v-else>{{ paidPerMillion(m.pricing?.input_price, period) }}</template>
            </td>
            <td class="pz-cell px-3 py-2.5 align-middle font-mono font-semibold text-gray-900 dark:text-gray-50">
              <template v-if="tokenIntervals(m).length">
                <div
                  v-for="(iv, idx) in tokenIntervals(m)"
                  :key="idx"
                  class="tier-price-block flex items-center whitespace-nowrap text-xs leading-5"
                  :style="`--tier-lines: ${cacheTierLineCount(iv)}`"
                  :title="tierHint(m)"
                >
                  {{ paidPerMillion(iv.output_price, period) }}
                </div>
              </template>
              <template v-else>{{ paidPerMillion(m.pricing?.output_price, period) }}</template>
            </td>
            <td class="pz-cell px-3 py-2.5 align-middle">
              <template v-if="hasTierCachePricing(tokenIntervals(m))">
                <div
                  v-for="(iv, idx) in tokenIntervals(m)"
                  :key="idx"
                  class="tier-price-block flex flex-col justify-center font-mono text-xs text-gray-800 dark:text-gray-200"
                  :style="`--tier-lines: ${cacheTierLineCount(iv)}`"
                  :title="tierHint(m)"
                >
                  <div
                    v-if="iv.cache_write_price != null"
                    class="cache-price-line"
                  >
                    <span class="font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWriteShort') }}</span>
                    {{ paidPerMillion(iv.cache_write_price, period) }}
                  </div>
                  <div v-if="iv.cache_write_1h_price != null" class="cache-price-line">
                    <span class="font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWriteShort') }} (1h)</span>
                    {{ paidPerMillion(iv.cache_write_1h_price, period) }}
                  </div>
                  <div v-if="iv.cache_read_price != null" class="cache-price-line">
                    <span class="font-sans font-normal text-gray-400 dark:text-dark-500">{{ t(iv.explicit_cache_read_price != null ? 'modelPlaza.table.cacheReadImplicitShort' : 'modelPlaza.table.cacheReadShort') }}</span>
                    {{ paidPerMillion(iv.cache_read_price, period) }}
                  </div>
                  <div v-if="iv.explicit_cache_read_price != null" class="cache-price-line">
                    <span class="font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheReadExplicitShort') }}</span>
                    {{ paidPerMillion(iv.explicit_cache_read_price, period) }}
                  </div>
                  <span v-if="!hasCacheInterval(iv)" class="cache-price-line text-gray-400 dark:text-dark-500">-</span>
                </div>
              </template>
              <div
                v-else-if="hasCachePricing(m)"
                class="space-y-0.5 font-mono text-xs text-gray-800 dark:text-gray-200"
              >
                <div v-if="m.pricing?.cache_write_price != null" class="cache-price-line">
                  <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWrite') }}</span>
                  {{ paidPerMillion(m.pricing.cache_write_price, period) }}
                </div>
                <div v-if="m.pricing?.cache_write_1h_price != null" class="cache-price-line">
                  <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWrite') }} (1h)</span>
                  {{ paidPerMillion(m.pricing.cache_write_1h_price, period) }}
                </div>
                <div v-if="m.pricing?.cache_read_price != null" class="cache-price-line">
                  <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t(m.pricing?.explicit_cache_read_price != null ? 'modelPlaza.table.cacheReadImplicit' : 'modelPlaza.table.cacheRead') }}</span>
                  {{ paidPerMillion(m.pricing?.cache_read_price, period) }}
                </div>
                <div v-if="m.pricing?.explicit_cache_read_price != null" class="cache-price-line">
                  <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheReadExplicit') }}</span>
                  {{ paidPerMillion(m.pricing.explicit_cache_read_price, period) }}
                </div>
              </div>
              <span v-else class="text-gray-400 dark:text-dark-500">-</span>
            </td>
          </template>

          <!-- 按次 / 按图片计费:实付区整体合并,阶梯芯片或单一按次价 -->
          <template v-else>
            <td colspan="3" class="pz-cell px-3 py-2.5 align-middle">
              <div
                v-if="requestIntervals(m).length"
                class="flex flex-wrap items-center gap-1.5"
              >
                <span
                  v-for="(iv, idx) in requestIntervals(m)"
                  :key="idx"
                  class="inline-flex items-center gap-1 rounded-md bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-800 dark:bg-dark-700/60 dark:text-gray-200"
                >
                  <span class="font-sans text-gray-400 dark:text-dark-500">{{ tierLabel(iv) }}</span>
                  {{ paidRequestPrice(m, iv.per_request_price)
                  }}<span class="font-sans text-gray-400 dark:text-dark-500">{{ perUnitSuffix(m) }}</span>
                </span>
              </div>
              <template v-else-if="m.pricing?.per_request_price != null">
                <span class="font-mono font-semibold text-gray-900 dark:text-gray-50">
                  {{ paidRequestPrice(m, m.pricing.per_request_price) }}
                </span>
                <span class="ml-1 text-xs text-gray-400 dark:text-dark-500">{{ perUnitSuffix(m) }}</span>
              </template>
              <span v-else class="text-gray-400 dark:text-dark-500">-</span>
            </td>
          </template>

          <!-- 官方价格(参考价,不乘倍率;官方有阶梯时每档一行) -->
          <td
            class="border-l border-gray-100 px-3 py-2.5 align-middle font-mono text-xs text-gray-500 dark:border-dark-700/60 dark:text-dark-400"
          >
            <template v-if="officialIntervals(m).length">
              <div
                v-for="(iv, idx) in officialIntervals(m)"
                :key="idx"
                class="tier-price-block flex items-center whitespace-nowrap leading-5"
                :style="`--tier-lines: ${cacheTierLineCount(iv)}`"
              >
                <span class="mr-1 font-sans text-gray-400 dark:text-dark-500" :title="t('modelPlaza.table.tierHint')">{{ tierLabel(iv) }}</span>
                {{ official(iv.input_price, m, period) }}
              </div>
            </template>
            <template v-else>{{ official(m.official_pricing?.input_price, m, period) }}</template>
          </td>
          <td class="px-3 py-2.5 align-middle font-mono text-xs text-gray-500 dark:text-dark-400">
            <template v-if="officialIntervals(m).length">
              <div
                v-for="(iv, idx) in officialIntervals(m)"
                :key="idx"
                class="tier-price-block flex items-center whitespace-nowrap leading-5"
                :style="`--tier-lines: ${cacheTierLineCount(iv)}`"
                :title="t('modelPlaza.table.tierHint')"
              >
                {{ official(iv.output_price, m, period) }}
              </div>
            </template>
            <template v-else>{{ official(m.official_pricing?.output_price, m, period) }}</template>
          </td>
          <td class="px-3 py-2.5 align-middle">
            <template v-if="hasTierCachePricing(officialIntervals(m))">
              <div
                v-for="(iv, idx) in officialIntervals(m)"
                :key="idx"
                class="tier-price-block flex flex-col justify-center font-mono text-xs text-gray-500 dark:text-dark-400"
                :style="`--tier-lines: ${cacheTierLineCount(iv)}`"
                :title="t('modelPlaza.table.tierHint')"
              >
                <div
                  v-if="iv.cache_write_price != null"
                  class="cache-price-line"
                >
                  <span class="font-sans text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWriteShort') }}</span>
                  {{ official(iv.cache_write_price, m, period) }}
                </div>
                <div v-if="iv.cache_write_1h_price != null" class="cache-price-line">
                  <span class="font-sans text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWriteShort') }} (1h)</span>
                  {{ official(iv.cache_write_1h_price, m, period) }}
                </div>
                <div v-if="iv.cache_read_price != null" class="cache-price-line">
                  <span class="font-sans text-gray-400 dark:text-dark-500">{{ t(iv.explicit_cache_read_price != null ? 'modelPlaza.table.cacheReadImplicitShort' : 'modelPlaza.table.cacheReadShort') }}</span>
                  {{ official(iv.cache_read_price, m, period) }}
                </div>
                <div v-if="iv.explicit_cache_read_price != null" class="cache-price-line">
                  <span class="font-sans text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheReadExplicitShort') }}</span>
                  {{ official(iv.explicit_cache_read_price, m, period) }}
                </div>
                <span v-if="!hasCacheInterval(iv)" class="cache-price-line text-gray-400 dark:text-dark-500">-</span>
              </div>
            </template>
            <div
              v-else-if="m.official_pricing && hasOfficialCache(m.official_pricing)"
              class="space-y-0.5 font-mono text-xs text-gray-500 dark:text-dark-400"
            >
              <div v-if="m.official_pricing.cache_write_price != null" class="cache-price-line">
                <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWrite') }}</span>
                {{ official(m.official_pricing.cache_write_price, m, period) }}
              </div>
              <div v-if="m.official_pricing.cache_write_1h_price != null" class="cache-price-line">
                <span class="mr-1 font-sans text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheWrite') }} (1h)</span>
                {{ official(m.official_pricing.cache_write_1h_price, m, period) }}
              </div>
              <div v-if="m.official_pricing.cache_read_price != null" class="cache-price-line">
                <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t(m.official_pricing.explicit_cache_read_price != null ? 'modelPlaza.table.cacheReadImplicit' : 'modelPlaza.table.cacheRead') }}</span>
                {{ official(m.official_pricing.cache_read_price, m, period) }}
              </div>
              <div v-if="m.official_pricing.explicit_cache_read_price != null" class="cache-price-line">
                <span class="mr-1 font-sans font-normal text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cacheReadExplicit') }}</span>
                {{ official(m.official_pricing.explicit_cache_read_price, m, period) }}
              </div>
            </div>
            <span v-else class="text-gray-400 dark:text-dark-500">-</span>
          </td>

          <!-- 折扣倍率(分时时段行展示 生效倍率×时段倍率;生图独立倍率行展示独立倍率;专属倍率划线展示原倍率) -->
          <td
            class="border-l border-gray-100 py-2.5 pl-3 pr-5 text-right align-middle font-mono text-xs dark:border-dark-700/60"
          >
            <span
              v-if="period"
              class="font-bold text-primary-600 dark:text-primary-400"
              :title="t('modelPlaza.table.timePricingRateHint', { rate: effectiveRate, multiplier: period.multiplier })"
              >{{ periodRate(period) }}x</span
            >
            <span
              v-else-if="usesIndependentImageRate(m)"
              class="font-bold text-gray-700 dark:text-gray-300"
              >{{ requestRate(m) }}x</span
            >
            <template v-else-if="hasCustomRate">
              <span class="mr-1 text-gray-400 line-through dark:text-dark-500">{{ rateMultiplier }}x</span>
              <span class="font-bold text-primary-600 dark:text-primary-400">{{ effectiveRate }}x</span>
            </template>
            <span v-else class="font-bold text-gray-700 dark:text-gray-300">{{ effectiveRate }}x</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatScaled } from '@/utils/pricing'
import { platformAccentColor, platformBadgeLightClass, platformLabel } from '@/utils/platformColors'
import {
  BILLING_MODE_TOKEN,
  BILLING_MODE_IMAGE,
  type BillingMode
} from '@/constants/channel'
import type { PlazaModel, PlazaTimePricingPeriod } from '@/api/modelPlaza'
import type { UserPricingInterval } from '@/api/channels'

const props = defineProps<{
  models: PlazaModel[]
  /** 分组平台;实付分区底色随平台着色,未知平台回退品牌青。 */
  platform?: string
  /** 分组默认倍率。 */
  rateMultiplier: number
  /** 用户专属倍率;与默认不同,实付价按此计算并划线展示原倍率。 */
  userRateMultiplier?: number | null
  /** 生图独立倍率:true 时图片计费模型的实付倍率取 imageRateMultiplier,不取分组/专属倍率。 */
  imageRateIndependent?: boolean
  imageRateMultiplier?: number | null
  /**
   * 高峰窗口描述(含倍率与服务器时区标注),空串/缺省 = 分组未启用高峰。
   * 表格所有价格均为不含高峰因子的口径,该窗口仅用于分时时段行的 tooltip 披露:
   * 与高峰重叠的部分实付还会再乘高峰倍率。
   */
  peakWindow?: string
  peakRateMultiplier?: number | null
}>()

const { t } = useI18n()

/** 实付分区只从平台拿一个主色,浅底/标题/下划线全部由 scoped CSS 用 color-mix 派生。 */
const accentStyle = computed(() => ({ '--plaza-accent': platformAccentColor(props.platform ?? '') }))

const PER_MILLION = 1_000_000

/**
 * 展示顺序:
 * 1. token 计费的排在前,按图/按次计费的沉到末尾——它们的官方 token 价与实付的按张/按次价不同量纲,混排无意义;
 * 2. 组内按官方输出价从高到低,无官方价的排最后;
 * 3. 同价按名称降序(新版本号在前,如 gpt-5.6 先于 gpt-5.5)。
 */
const sortedModels = computed(() => {
  return [...props.models].sort((a, b) => {
    const ta = billingMode(a) === BILLING_MODE_TOKEN
    const tb = billingMode(b) === BILLING_MODE_TOKEN
    if (ta !== tb) return ta ? -1 : 1
    const pa = a.official_pricing?.output_price ?? null
    const pb = b.official_pricing?.output_price ?? null
    if (pa != null && pb != null && pa !== pb) return pb - pa
    if (pa != null && pb == null) return -1
    if (pa == null && pb != null) return 1
    return b.name.localeCompare(a.name)
  })
})

const effectiveRate = computed(() => props.userRateMultiplier ?? props.rateMultiplier)
const hasCustomRate = computed(
  () => props.userRateMultiplier != null && props.userRateMultiplier !== props.rateMultiplier
)

function billingMode(m: PlazaModel): BillingMode {
  return (m.pricing?.billing_mode || BILLING_MODE_TOKEN) as BillingMode
}

function billingModeLabel(m: PlazaModel): string {
  return billingMode(m) === BILLING_MODE_IMAGE
    ? t('modelPlaza.table.perImage')
    : t('modelPlaza.table.perRequest')
}

/** 价格统一保底 2 位小数,更长的有效小数原样保留。 */
const MIN_DECIMALS = 2

/** 表格行:每个模型一行标准价;配置了分时倍率的模型再按时段各加一行。 */
interface PlazaRow {
  model: PlazaModel
  period: PlazaTimePricingPeriod | null
  pricingMode: 'nonThinking' | 'thinking' | null
  key: string
}

const rows = computed<PlazaRow[]>(() =>
  sortedModels.value.flatMap((source) => {
    const variants: Array<{ model: PlazaModel; pricingMode: PlazaRow['pricingMode'] }> =
      source.thinking_pricing != null
        ? [
            { model: source, pricingMode: 'nonThinking' },
            {
              model: {
                ...source,
                pricing: source.thinking_pricing,
                official_pricing: source.official_thinking_pricing ?? source.official_pricing
              },
              pricingMode: 'thinking'
            }
          ]
        : [{ model: source, pricingMode: null }]

    return variants.flatMap(({ model, pricingMode }) => {
      const modeKey = pricingMode ?? 'single'
      const base: PlazaRow = {
        model,
        period: null,
        pricingMode,
        key: `${model.platform}:${model.name}:${modeKey}`
      }
      const periodRows = timePeriods(model).map<PlazaRow>((period, idx) => ({
        model,
        period,
        pricingMode,
        key: `${model.platform}:${model.name}:${modeKey}:${idx}`
      }))
      return [base, ...periodRows]
    })
  })
)

/** 时段行的生效倍率 = 生效倍率 × 时段倍率(去掉浮点噪声)。 */
function periodRate(period: PlazaTimePricingPeriod): number {
  return Math.round(effectiveRate.value * period.multiplier * 1000) / 1000
}

/** 实付价 = 渠道单价 × 生效倍率(时段行再乘时段倍率),按 ¥/1M token 展示。 */
function paidPerMillion(value: number | null | undefined, period: PlazaTimePricingPeriod | null = null): string {
  if (value == null) return '-'
  const rate = period ? periodRate(period) : effectiveRate.value
  return formatScaled(value * rate, PER_MILLION, MIN_DECIMALS)
}

/** 图片计费模型且分组开启生图独立倍率:实付倍率取独立倍率,与计费口径一致。 */
function usesIndependentImageRate(m: PlazaModel): boolean {
  return billingMode(m) === BILLING_MODE_IMAGE && props.imageRateIndependent === true
}

/** 按次/按图片行的生效倍率。 */
function requestRate(m: PlazaModel): number {
  return usesIndependentImageRate(m) ? (props.imageRateMultiplier ?? 1) : effectiveRate.value
}

/** 按次 / 按图片单价(乘该行生效倍率,不换算 1M)。 */
function paidRequestPrice(m: PlazaModel, value: number | null | undefined): string {
  if (value == null) return '-'
  return formatScaled(value * requestRate(m), 1, MIN_DECIMALS)
}

/** 官方参考价只在实际分时行与官方分时时区、工作日规则和时段完全匹配时乘官方倍率。 */
function official(
  value: number | null | undefined,
  m: PlazaModel,
  period: PlazaTimePricingPeriod | null
): string {
  if (value == null) return '-'
  const multiplier = matchingOfficialPeriod(m, period)?.multiplier ?? 1
  return formatScaled(value * multiplier, PER_MILLION, MIN_DECIMALS)
}

function matchingOfficialPeriod(
  m: PlazaModel,
  period: PlazaTimePricingPeriod | null
): PlazaTimePricingPeriod | null {
  if (!period || !m.time_pricing || !m.official_pricing?.time_pricing) return null
  const actual = m.time_pricing
  const reference = m.official_pricing.time_pricing
  if (
    actual.timezone !== reference.timezone ||
    Boolean(actual.weekdays_only) !== Boolean(reference.weekdays_only)
  ) {
    return null
  }
  return (
    reference.periods.find(
      (candidate) =>
        normalizeClock(candidate.start_time) === normalizeClock(period.start_time) &&
        normalizeClock(candidate.end_time) === normalizeClock(period.end_time)
    ) ?? null
  )
}

/** 非 token 计费的单位后缀:按图片 → “/ 张”,按次 → “/ 次”。 */
function perUnitSuffix(m: PlazaModel): string {
  return billingMode(m) === BILLING_MODE_IMAGE
    ? t('modelPlaza.table.perUnitImage')
    : t('modelPlaza.table.perUnitRequest')
}

function hasCachePricing(m: PlazaModel): boolean {
  return m.pricing?.cache_write_price != null || m.pricing?.cache_write_1h_price != null || m.pricing?.cache_read_price != null || m.pricing?.explicit_cache_read_price != null
}

function hasOfficialCache(o: NonNullable<PlazaModel['official_pricing']>): boolean {
  return o.cache_write_price != null || o.cache_read_price != null || o.cache_write_1h_price != null || o.explicit_cache_read_price != null
}

/** 分时倍率时段(后端只给出倍率 ≠ 1 的时段,已升序)。 */
function timePeriods(m: PlazaModel): PlazaTimePricingPeriod[] {
  return m.time_pricing?.periods ?? []
}

/**
 * 时段行 tooltip:仅工作日生效的配置换用带周末回落说明的文案;
 * 分组启用高峰倍率时追加披露——本行价格不含高峰因子,与高峰窗口重叠的部分实付再乘高峰倍率。
 */
function timePricingRowHint(m: PlazaModel): string {
  const key = m.time_pricing?.weekdays_only
    ? 'modelPlaza.table.timePricingRowHintWeekdays'
    : 'modelPlaza.table.timePricingRowHint'
  let hint = t(key, { timezone: m.time_pricing?.timezone })
  if (props.peakWindow) {
    hint += t('modelPlaza.table.timePricingRowHintPeak', {
      window: props.peakWindow,
      multiplier: props.peakRateMultiplier ?? 1
    })
  }
  return hint
}

/** “00:30–08:30”;整分钟的 HH:mm:ss 省略秒。 */
function formatTimeWindow(p: PlazaTimePricingPeriod): string {
  return `${normalizeClock(p.start_time)}–${normalizeClock(p.end_time)}`
}

function normalizeClock(value: string): string {
  return value.replace(/^(\d{2}:\d{2}):00$/, '$1')
}

/** 上下文档位按下限升序展示(后端已升序,此处兜底)。 */
function sortByContext(intervals: UserPricingInterval[]): UserPricingInterval[] {
  return [...intervals].sort((a, b) => a.min_tokens - b.min_tokens)
}

/** token 模式的阶梯定价(内联进输入/输出/缓存列)。 */
function tokenIntervals(m: PlazaModel): UserPricingInterval[] {
  return sortByContext(m.pricing?.intervals ?? [])
}

/** 官方阶梯(后端按目录规则合成,不受分组开关影响)。 */
function officialIntervals(m: PlazaModel): UserPricingInterval[] {
  return sortByContext(m.official_pricing?.intervals ?? [])
}

function hasCacheInterval(iv: UserPricingInterval): boolean {
  return (
    iv.cache_write_price != null ||
    iv.cache_write_1h_price != null ||
    iv.cache_read_price != null ||
    iv.explicit_cache_read_price != null
  )
}

/** 同一档的输入、输出与缓存块共用缓存明细行数，避免相邻档位纵向错位。 */
function cacheTierLineCount(iv: UserPricingInterval): number {
  const lines = Number(iv.cache_write_price != null) +
    Number(iv.cache_write_1h_price != null) +
    Number(iv.cache_read_price != null) +
    Number(iv.explicit_cache_read_price != null)
  return Math.max(lines, 1)
}

/** 任一档带缓存价才按档渲染缓存列；否则沿用平价缓存明细。 */
function hasTierCachePricing(intervals: UserPricingInterval[]): boolean {
  return intervals.some(hasCacheInterval)
}

/** 档位说明:整单按档计价,或(平台旧规则)仅超出部分按档计价。 */
function tierHint(m: PlazaModel): string {
  return m.long_context_basis === 'marginal'
    ? t('modelPlaza.table.tierHintMarginal')
    : t('modelPlaza.table.tierHint')
}


/** 按次/按图模式的阶梯定价(仅保留配了按次价的档位)。 */
function requestIntervals(m: PlazaModel): UserPricingInterval[] {
  return (m.pricing?.intervals ?? []).filter((iv) => iv.per_request_price != null)
}

/**
 * 档位标签:优先后端/管理员给出的 tier_label,否则按区间生成统一形态——
 * 有上限为「≤上限」,末档为「>下限」;档位升序排列,相邻的 ≤100K / ≤200K 即表示 (100K,200K]。
 */
function tierLabel(iv: UserPricingInterval): string {
  if (iv.tier_label) return iv.tier_label
  const { min_tokens: min, max_tokens: max } = iv
  return max == null ? `>${formatTokenCount(min)}` : `≤${formatTokenCount(max)}`
}

function formatTokenCount(n: number): string {
  if (n >= 1_000_000) return `${trimZero(n / 1_000_000)}M`
  if (n >= 1_000) return `${trimZero(n / 1_000)}K`
  return String(n)
}

function trimZero(n: number): string {
  // 保留精确阈值，避免将尚未到达 32K 的低价档误标为包含 32K。
  return String(n)
}
</script>

<style scoped>
/* 实付分区配色统一从 --plaza-accent(平台主色)派生,新增平台无需扩展样式 */
.plaza-pricing-table {
  --pz-title: color-mix(in srgb, var(--plaza-accent) 88%, black);
  --pz-bg: color-mix(in srgb, var(--plaza-accent) 7%, transparent);
  --pz-bg-hover: color-mix(in srgb, var(--plaza-accent) 13%, transparent);
}

.dark .plaza-pricing-table {
  --pz-title: color-mix(in srgb, var(--plaza-accent) 70%, white);
  --pz-bg: color-mix(in srgb, var(--plaza-accent) 6%, transparent);
  --pz-bg-hover: color-mix(in srgb, var(--plaza-accent) 10%, transparent);
}

.pz-bg,
.pz-cell {
  background-color: var(--pz-bg);
}

.pz-cell {
  transition: background-color 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

tbody tr:hover .pz-cell {
  background-color: var(--pz-bg-hover);
}

.pz-title {
  /* color-mix 不可用的老浏览器回退为平台原色 */
  color: var(--plaza-accent);
  color: var(--pz-title);
  border-color: color-mix(in srgb, var(--pz-title) 30%, transparent);
}

.pz-unit {
  color: color-mix(in srgb, var(--pz-title) 62%, transparent);
}

/* 各计价列以同一缓存明细行数撑开档位，缓存标签与金额保持为不可拆整体。 */
.tier-price-block {
  min-height: calc(var(--tier-lines, 1) * 1.25rem);
}

.cache-price-line {
  /* flex 统一混合字体的实际行框，避免浏览器缩放时缓存块比同档输入/输出更高。 */
  display: flex;
  align-items: center;
  gap: 0.25rem;
  white-space: nowrap;
  line-height: 1.25rem;
}
</style>
