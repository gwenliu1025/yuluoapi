import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { nextTick } from 'vue'

import HomeView from '../HomeView.vue'

const { appStore, authStore, localeState } = vi.hoisted(() => ({
  appStore: {
    cachedPublicSettings: {} as Record<string, unknown>,
    siteName: 'Fallback site',
    siteLogo: '',
    docUrl: '',
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn(),
  },
  authStore: {
    isAuthenticated: false,
    isAdmin: false,
    user: null as { email?: string } | null,
    checkAuth: vi.fn(),
  },
  localeState: {
    language: 'zh' as 'zh' | 'en',
  },
}))

const translations = {
  zh: {
    'nav.modelPlaza': '模型广场',
    'home.docs': '文档',
    'home.switchToLight': '切换到浅色模式',
    'home.switchToDark': '切换到深色模式',
    'home.login': '登录',
    'home.dashboard': '控制台',
    'home.goToDashboard': '进入控制台',
    'home.getStarted': '立即开始',
    'home.browseModels': '浏览模型',
    'home.featuredChineseModels': '精选中国模型',
    'home.moreChineseModels': '更多国产模型',
    'home.viewAllModels': '查看全部模型',
  },
  en: {
    'nav.modelPlaza': 'Model Plaza',
    'home.docs': 'Docs',
    'home.switchToLight': 'Switch to Light Mode',
    'home.switchToDark': 'Switch to Dark Mode',
    'home.login': 'Login',
    'home.dashboard': 'Dashboard',
    'home.goToDashboard': 'Go to Dashboard',
    'home.getStarted': 'Get Started',
    'home.browseModels': 'Browse Models',
    'home.featuredChineseModels': 'Featured Chinese Models',
    'home.moreChineseModels': 'More Chinese Models',
    'home.viewAllModels': 'View All Models',
  },
} as const

vi.mock('@/stores', () => ({
  useAppStore: () => appStore,
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => translations[localeState.language][key as keyof typeof translations.zh] ?? key,
    }),
  }
})

function mountHome(settings: Record<string, unknown> = {}) {
  appStore.cachedPublicSettings = {
    site_name: 'Configured Sub2API name',
    site_subtitle: 'Configured subtitle',
    doc_url: 'https://docs.example.test',
    model_plaza_enabled: true,
    ...settings,
  }

  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: { template: '<div data-testid="locale-switcher" />' },
        Icon: { template: '<span data-testid="icon" />' },
        ModelIcon: {
          props: ['model'],
          template: '<span data-testid="model-icon" :data-model="model" />',
        },
      },
    },
  })
}

describe('HomeView 雨落品牌默认首页', () => {
  beforeEach(() => {
    localeState.language = 'zh'
    authStore.isAuthenticated = false
    authStore.isAdmin = false
    authStore.user = null
    authStore.checkAuth.mockClear()
    appStore.fetchPublicSettings.mockClear()
    localStorage.clear()
    document.documentElement.classList.remove('dark')
    vi.spyOn(window, 'matchMedia').mockImplementation((query: string) => ({
      matches: query.includes('min-width'),
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }) as unknown as MediaQueryList)
  })

  it('只在默认分支展示固定的雨落品牌叙事', () => {
    const wrapper = mountHome()

    const home = wrapper.get('[data-testid="yuluo-home"]')
    expect(home.text()).toContain('雨落 API')
    expect(home.text()).toContain('企业级 API 服务网关')
    expect(home.text()).toContain('世上没有两场相同的雨，也没有一个模型适合所有问题。')
    expect(home.text()).toContain('雨落，让每个问题遇见合适的中国模型。')
    expect(home.text()).not.toContain('Configured Sub2API name')
  })

  it('保留登录控制台路径并提供模型广场入口', () => {
    const wrapper = mountHome()
    const destinations = wrapper.findAllComponents(RouterLinkStub).map((link) => link.props('to'))

    expect(destinations).toContain('/login')
    expect(destinations).toContain('/model-plaza')

    authStore.isAuthenticated = true
    authStore.isAdmin = true
    const adminWrapper = mountHome()
    const adminDestinations = adminWrapper.findAllComponents(RouterLinkStub).map((link) => link.props('to'))
    expect(adminDestinations).toContain('/admin/dashboard')
  })

  it('展示四个带品牌图标的主推模型并指向完整模型目录', () => {
    const wrapper = mountHome()
    const models = wrapper.findAll('[data-testid="featured-model"]').map((item) => item.text())
    const icons = wrapper.findAll('[data-testid="model-icon"]').map((icon) => icon.attributes('data-model'))

    expect(models).toEqual(['DeepSeek', '智谱 GLM', 'Kimi', 'Qwen'])
    expect(icons).toEqual(['deepseek', 'glm-4', 'qwen'])
    expect(wrapper.get('[data-testid="kimi-brand-icon"]').attributes('src')).toContain('kimi-brand.svg')
    expect(wrapper.get('[data-testid="all-models-link"]').text()).toContain('更多国产模型')
    expect(wrapper.get('[data-testid="all-models-link"]').text()).toContain('查看全部模型')
  })

  it('语言切换后本地化操作文案，同时保留固定中文品牌叙事', () => {
    localeState.language = 'en'
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Model Plaza')
    expect(wrapper.text()).toContain('Browse Models')
    expect(wrapper.text()).toContain('Featured Chinese Models')
    expect(wrapper.text()).toContain('More Chinese Models')
    expect(wrapper.text()).toContain('View All Models')
    expect(wrapper.text()).toContain('雨落 API')
    expect(wrapper.text()).toContain('企业级 API 服务网关')
  })

  it('模型广场开关关闭时隐藏广场入口，次要按钮回退到文档链接', () => {
    const wrapper = mountHome({ model_plaza_enabled: false })

    expect(wrapper.findAll('[to="/model-plaza"]')).toHaveLength(0)
    expect(wrapper.find('[data-testid="all-models-link"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="all-models-slot"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('模型广场')

    const secondary = wrapper.get('[data-testid="secondary-cta"]')
    expect(secondary.attributes('href')).toContain('docs.example.test')
    expect(secondary.text()).toContain('文档')
  })

  it('模型广场关闭且未配置文档地址时不渲染次要按钮', () => {
    const wrapper = mountHome({ model_plaza_enabled: false, doc_url: '' })

    expect(wrapper.find('[data-testid="secondary-cta"]').exists()).toBe(false)
  })

  it('模型广场开启时第五列显示更多模型入口，桌面与移动端均保留文档入口', () => {
    const wrapper = mountHome({ model_plaza_enabled: true, doc_url: 'https://docs.example.test' })
    const slot = wrapper.get('[data-testid="all-models-slot"]')

    expect(slot.findComponent(RouterLinkStub).props('to')).toBe('/model-plaza')
    const docs = wrapper.findAll('a').find((link) => link.attributes('href')?.includes('docs.example.test'))
    expect(docs).toBeDefined()
    expect(docs?.classes()).toContain('inline-flex')
    expect(docs?.classes()).not.toContain('hidden')
  })

  it('模型广场要求登录时匿名用户不显示入口但保留第五列', () => {
    const wrapper = mountHome({ model_plaza_enabled: true, model_plaza_require_auth: true })

    expect(wrapper.findAll('[to="/model-plaza"]')).toHaveLength(0)
    expect(wrapper.get('[data-testid="all-models-slot"]').exists()).toBe(true)
  })

  it('移除旧终端、通用功能卡和精选模型区下面的末尾小字行', () => {
    const wrapper = mountHome()

    expect(wrapper.find('.terminal-container').exists()).toBe(false)
    expect(wrapper.find('[data-testid="legacy-feature-grid"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="default-home-footer"]').exists()).toBe(false)
    expect(wrapper.find('footer').exists()).toBe(false)
  })

  it('常规模式只在雨滴或花瓣落水时触发水波，移动端只运行可见节点', async () => {
    const regular = mountHome()
    await flushPromises()
    expect(regular.get('[data-testid="rain-layer"]').attributes('data-reduced-motion')).toBe('false')
    expect(regular.get('[data-testid="rain-layer"]').attributes('data-active-drops')).toBe('4')
    expect(regular.get('[data-testid="rain-layer"]').attributes('data-impact-ripple-count')).toBe('0')
    expect(regular.get('[data-testid="rain-layer"]').attributes('data-click-ripple-count')).toBe('0')
    expect(regular.get('[data-testid="rain-layer"]').attributes('data-active-petals')).toBe('12')
    expect(regular.findAll('[data-testid="rain-drop"]')).toHaveLength(4)
    expect(regular.findAll('[data-testid="impact-ripple"]')).toHaveLength(8)
    expect(regular.findAll('[data-testid="impact-ripple-wave"]')).toHaveLength(16)
    expect(regular.findAll('[data-testid="click-ripple"]')).toHaveLength(5)
    expect(regular.findAll('[data-testid="click-ripple-wave"]')).toHaveLength(15)
    expect(regular.findAll('[data-testid="flower-petal"]')).toHaveLength(12)
    expect(regular.get('[data-testid="flower-petal"]').attributes('data-petal-source')).toBe('cc0-photo-crop')
    expect(regular.findAll('[data-testid="rain-drop"].mobile-motion-hidden')).toHaveLength(2)

    await regular.get('main section').trigger('click', { button: 0, clientX: 320, clientY: 420 })
    await nextTick()
    expect(regular.get('[data-testid="rain-layer"]').attributes('data-click-ripple-count')).toBe('1')
    regular.unmount()

    vi.mocked(window.matchMedia).mockImplementation((query: string) => ({
      matches: query.includes('max-width'),
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }) as unknown as MediaQueryList)

    const mobile = mountHome()
    await flushPromises()
    expect(mobile.get('[data-testid="rain-layer"]').attributes('data-active-drops')).toBe('2')
    expect(mobile.get('[data-testid="rain-layer"]').attributes('data-impact-ripple-count')).toBe('0')
    expect(mobile.get('[data-testid="rain-layer"]').attributes('data-active-petals')).toBe('7')
    mobile.unmount()

    vi.mocked(window.matchMedia).mockImplementation((query: string) => ({
      matches: query.includes('prefers-reduced-motion'),
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }) as unknown as MediaQueryList)

    const reduced = mountHome()
    await flushPromises()
    expect(reduced.get('[data-testid="rain-layer"]').attributes('data-reduced-motion')).toBe('true')
    expect(reduced.get('[data-testid="rain-layer"]').attributes('data-active-drops')).toBe('0')
    expect(reduced.get('[data-testid="rain-layer"]').attributes('data-impact-ripple-count')).toBe('0')
    expect(reduced.findAll('[data-testid="rain-drop"]')).toHaveLength(0)
    expect(reduced.findAll('[data-testid="click-ripple"]')).toHaveLength(0)
    expect(reduced.findAll('[data-testid="click-ripple-wave"]')).toHaveLength(0)
    expect(reduced.findAll('[data-testid="flower-petal"]')).toHaveLength(0)
  })

  it('页面隐藏与离屏状态合并计算，并在卸载时清理观察器', async () => {
    let observerCallback: IntersectionObserverCallback | undefined
    const disconnect = vi.fn()
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(false)

    vi.stubGlobal('IntersectionObserver', class {
      constructor(callback: IntersectionObserverCallback) {
        observerCallback = callback
      }

      observe = vi.fn()
      unobserve = vi.fn()
      disconnect = disconnect
      takeRecords = () => []
      root = null
      rootMargin = '0px'
      thresholds = [0.05]
    })

    const wrapper = mountHome()
    await nextTick()
    const layer = wrapper.get('[data-testid="rain-layer"]')
    expect(layer.attributes('data-motion-paused')).toBe('false')

    observerCallback?.([{ isIntersecting: false } as IntersectionObserverEntry], {} as IntersectionObserver)
    await nextTick()
    expect(layer.attributes('data-motion-paused')).toBe('true')

    hidden.mockReturnValue(true)
    document.dispatchEvent(new Event('visibilitychange'))
    observerCallback?.([{ isIntersecting: true } as IntersectionObserverEntry], {} as IntersectionObserver)
    await nextTick()
    expect(layer.attributes('data-motion-paused')).toBe('true')

    hidden.mockReturnValue(false)
    document.dispatchEvent(new Event('visibilitychange'))
    await nextTick()
    expect(layer.attributes('data-motion-paused')).toBe('false')

    wrapper.unmount()
    expect(disconnect).toHaveBeenCalledOnce()
    hidden.mockRestore()
    vi.unstubAllGlobals()
  })

  it('组件卸载时清理动态偏好监听器', () => {
    const addMotionListener = vi.fn()
    const removeMotionListener = vi.fn()
    vi.mocked(window.matchMedia).mockImplementation((query: string) => ({
      matches: false,
      media: query,
      addEventListener: query.includes('prefers-reduced-motion') ? addMotionListener : vi.fn(),
      removeEventListener: query.includes('prefers-reduced-motion') ? removeMotionListener : vi.fn(),
    }) as unknown as MediaQueryList)

    const wrapper = mountHome()
    const changeHandler = addMotionListener.mock.calls[0]?.[1]
    expect(changeHandler).toBeTypeOf('function')

    wrapper.unmount()
    expect(removeMotionListener).toHaveBeenCalledWith('change', changeHandler)
  })
})
