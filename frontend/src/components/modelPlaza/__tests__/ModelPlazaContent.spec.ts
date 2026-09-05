import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelPlazaContent from '../ModelPlazaContent.vue'
import PlazaGroupSection from '../PlazaGroupSection.vue'
import type { ModelPlazaGroup, ModelPlazaResponse } from '@/api/modelPlaza'
import { modelSeries } from '../modelSeries'

vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ isAuthenticated: true }) }))

// 只模拟已获授权的响应；测试价格为合成值，不使用经营报价。
function group(id: number, names: string[], overrides: Partial<ModelPlazaGroup> = {}): ModelPlazaGroup {
  return {
    id, name: `分组${id}`, description: '', platform: 'openai',
    subscription_type: 'standard', rate_multiplier: 1,
    peak_rate_enabled: false, peak_start: '', peak_end: '', peak_rate_multiplier: 1,
    is_exclusive: id !== 1, image_rate_independent: false, image_rate_multiplier: 1,
    long_context_pricing_enabled: true,
    models: names.map((name) => ({ name, platform: 'openai', pricing: null, official_pricing: null })),
    ...overrides
  }
}

function response(groups: ModelPlazaGroup[]): ModelPlazaResponse {
  return { description: '', groups }
}

function mountContent(groups = [
  group(1, ['deepseek-v4-pro', 'qwen3.8-max']),
  group(2, ['qwen-plus', 'kimi/kimi-k3', 'ZHIPU/GLM-5.3', 'custom-model'])
]) {
  return mount(ModelPlazaContent, {
    props: { response: response(groups), loading: false },
    global: { stubs: { PlazaGroupSection: true, Icon: true, PlatformIcon: true } }
  })
}

type Wrapper = ReturnType<typeof mountContent>
function button(wrapper: Wrapper, dimension: 'group' | 'series', label: string) {
  const row = wrapper.get(`[role="group"][aria-label="modelPlaza.filters.${dimension}Label"]`)
  const found = row.findAll('button').find((b) => b.text() === label)
  if (!found) throw new Error(`找不到筛选按钮：${label}`)
  return found
}
function shown(wrapper: Wrapper): ModelPlazaGroup[] {
  return wrapper.findAllComponents(PlazaGroupSection).map((s) => s.props('group'))
}
function selected(wrapper: Wrapper, dimension: 'group' | 'series') {
  return wrapper.get(`[role="group"][aria-label="modelPlaza.filters.${dimension}Label"]`)
    .findAll('button[aria-pressed="true"]').map((b) => b.text())
}
const ALL = 'modelPlaza.filters.all'
const SERIES = 'modelPlaza.filters.seriesNames.'

describe('模型广场互斥筛选', () => {
  it('默认全部；切换单组替换列表；再次点击不取消；全部恢复授权分组', async () => {
    const wrapper = mountContent()
    expect(selected(wrapper, 'group')).toEqual([ALL])
    expect(shown(wrapper).map((g) => g.id)).toEqual([1, 2])
    for (const id of [2, 1, 1]) {
      await button(wrapper, 'group', `分组${id}`).trigger('click')
      expect(selected(wrapper, 'group')).toEqual([`分组${id}`])
      expect(shown(wrapper).map((g) => g.id)).toEqual([id])
      expect(button(wrapper, 'group', `分组${id}`).classes()).toContain('text-white')
      expect(button(wrapper, 'group', `分组${3 - id}`).classes()).toContain('bg-white')
      expect(button(wrapper, 'group', `分组${3 - id}`).classes()).not.toContain('chip-tinted')
    }
    await button(wrapper, 'group', ALL).trigger('click')
    expect(selected(wrapper, 'group')).toEqual([ALL])
    expect(shown(wrapper).map((g) => g.id)).toEqual([1, 2])
  })

  it('系列单选与单个分组、搜索取交集，不改变调用名、平台或源响应', async () => {
    const groups = [group(1, ['qwen3.8-max', 'kimi-k3']), group(2, ['qwen-plus', 'deepseek-v4-pro'])]
    const original = JSON.stringify(groups)
    const wrapper = mountContent(groups)
    await button(wrapper, 'series', SERIES + 'qwen').trigger('click')
    expect(selected(wrapper, 'series')).toEqual([SERIES + 'qwen'])
    expect(shown(wrapper).map((g) => g.models.map((m) => m.name))).toEqual([['qwen3.8-max'], ['qwen-plus']])
    await button(wrapper, 'group', '分组2').trigger('click')
    expect(shown(wrapper).map((g) => g.id)).toEqual([2])
    await wrapper.get('input').setValue('PLUS')
    expect(shown(wrapper)[0].models[0]).toEqual(groups[1].models[0])
    await wrapper.get('input').setValue('kimi')
    expect(shown(wrapper)).toEqual([])
    expect(wrapper.text()).toContain('modelPlaza.noSearchResult')
    expect(JSON.stringify(groups)).toBe(original)
  })

  it('无匹配的组合置灰，清除系列后可以正常切换', async () => {
    const wrapper = mountContent()
    await button(wrapper, 'series', SERIES + 'kimi').trigger('click')
    expect(shown(wrapper).map((g) => g.id)).toEqual([2])
    expect(button(wrapper, 'group', '分组1').attributes('disabled')).toBeDefined()
    await button(wrapper, 'group', '分组1').trigger('click')
    expect(selected(wrapper, 'group')).toEqual([ALL])
    await button(wrapper, 'series', ALL).trigger('click')
    await button(wrapper, 'group', '分组1').trigger('click')
    expect(shown(wrapper).map((g) => g.id)).toEqual([1])
  })

  it('不凭常量添加不存在的系列，未知调用名在其他中完整保留', async () => {
    const wrapper = mountContent([group(1, ['tenant/unknown-v1'])])
    expect(wrapper.text()).not.toContain(SERIES + 'kimi')
    expect(wrapper.text()).not.toContain(SERIES + 'qwen')
    await button(wrapper, 'series', SERIES + 'other').trigger('click')
    expect(shown(wrapper)[0].models[0].name).toBe('tenant/unknown-v1')
    expect(shown(wrapper)[0].models[0].platform).toBe('openai')
  })

  it('权限响应更新时移除失效分组和系列，不残留旧模型', async () => {
    const wrapper = mountContent()
    await button(wrapper, 'group', '分组2').trigger('click')
    await button(wrapper, 'series', SERIES + 'kimi').trigger('click')
    await wrapper.setProps({ response: response([group(1, ['qwen-plus'])]) })
    expect(selected(wrapper, 'group')).toEqual([ALL])
    expect(selected(wrapper, 'series')).toEqual([ALL])
    expect(shown(wrapper).map((g) => g.id)).toEqual([1])
    expect(wrapper.text()).not.toContain('分组2')
    expect(wrapper.text()).not.toContain(SERIES + 'kimi')
    await wrapper.setProps({ response: response([]) })
    expect(shown(wrapper)).toEqual([])
    expect(selected(wrapper, 'group')).toEqual([ALL])
  })

  it('加载和错误状态不展示旧分组', async () => {
    const wrapper = mountContent()
    await wrapper.setProps({ loading: true })
    expect(shown(wrapper)).toEqual([])
    await wrapper.setProps({ loading: false, error: true })
    expect(shown(wrapper)).toEqual([])
    expect(wrapper.text()).toContain('modelPlaza.loadFailed')
  })
})

describe('模型系列归类', () => {
  it.each([
    ['deepseek-v4-pro', 'deepseek'], ['DeepSeek/DeepSeek-V4-Flash', 'deepseek'],
    ['qwen3.8-max', 'qwen'], ['QWEN/qwen-plus', 'qwen'],
    ['kimi-k3', 'kimi'], ['kimi/kimi-k3', 'kimi'], [' Kimi/kimi-k2.5 ', 'kimi'],
    ['ZHIPU/GLM-5.3', 'glm'], ['glm-5', 'glm'],
    ['my-qwen-proxy', 'other'], ['qwenish', 'other'], ['gpt-5', 'other'], ['', 'other']
  ])('%s 归入 %s，避免任意包含匹配', (name, expected) => {
    expect(modelSeries(name)).toBe(expected)
  })
})
