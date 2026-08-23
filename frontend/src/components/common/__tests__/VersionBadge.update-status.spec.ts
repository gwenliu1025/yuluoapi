import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import VersionBadge from '../VersionBadge.vue'

const mockPerformUpdate = vi.fn()
const mockRestartService = vi.fn()
const mockGetUpdateStatus = vi.fn()
const mockGetRollbackVersions = vi.fn()
const mockRollback = vi.fn()
const mockClearVersionCache = vi.fn()
let mockHasUpdate = true

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      locale: { value: 'zh' },
      t: (key: string) => key
    }
  }),
  useI18n: () => ({ t: (key: string) => key })
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ isAdmin: true }),
  useAppStore: () => ({
    versionLoading: false,
    currentVersion: '0.1.156',
    latestVersion: '0.1.169',
    hasUpdate: mockHasUpdate,
    releaseInfo: { html_url: '#' },
    buildType: 'release',
    fetchVersion: vi.fn(),
    clearVersionCache: mockClearVersionCache
  })
}))

vi.mock('@/api/admin/system', () => ({
  performUpdate: (...args: unknown[]) => mockPerformUpdate(...args),
  restartService: (...args: unknown[]) => mockRestartService(...args),
  getUpdateStatus: (...args: unknown[]) => mockGetUpdateStatus(...args),
  getRollbackVersions: (...args: unknown[]) => mockGetRollbackVersions(...args),
  rollback: (...args: unknown[]) => mockRollback(...args)
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copied: { value: false }, copyToClipboard: vi.fn() })
}))

describe('VersionBadge Docker 更新状态', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockHasUpdate = true
    mockPerformUpdate.mockResolvedValue({
      message: 'prepared',
      need_restart: true,
      update_mode: 'docker_agent'
    })
    mockRestartService.mockResolvedValue({
      message: 'activating',
      update_mode: 'docker_agent',
      status: { state: 'activating' }
    })
    mockGetUpdateStatus.mockResolvedValue({
      state: 'rolled_back',
      current_image: 'ghcr.io/gwenliu1025/sub2api:0.1.156',
      target_image: 'ghcr.io/gwenliu1025/sub2api:0.1.169',
      previous_image: 'ghcr.io/gwenliu1025/sub2api:0.1.156',
      message: '健康检查失败，已自动回滚',
      updated_at: '2026-08-01T00:00:00Z'
    })
    mockGetRollbackVersions.mockResolvedValue({
      versions: [
        {
          version: '0.1.172',
          published_at: '2026-08-01T00:00:00Z',
          html_url: '#'
        }
      ]
    })
    mockRollback.mockResolvedValue({
      message: 'rollback prepared',
      need_restart: true,
      update_mode: 'docker_agent'
    })
  })

  it('重启切换后轮询代理状态并显示自动回滚结果', async () => {
    const wrapper = mount(VersionBadge, {
      global: {
        stubs: { Icon: true }
      }
    })

    await wrapper.get('button').trigger('click')
    const updateButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('version.updateNow'))
    expect(updateButton).toBeDefined()
    await updateButton!.trigger('click')
    await flushPromises()

    const restartButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('version.restartNow'))
    expect(restartButton).toBeDefined()
    await restartButton!.trigger('click')
    await flushPromises()

    expect(mockGetUpdateStatus).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('健康检查失败，已自动回滚')
    expect(mockClearVersionCache).toHaveBeenCalledTimes(1)
  })

  it('Docker 激活请求失败时显示错误且不轮询状态', async () => {
    mockRestartService.mockRejectedValue(new Error('更新代理不可用'))
    const wrapper = mount(VersionBadge, {
      global: {
        stubs: { Icon: true }
      }
    })

    await wrapper.get('button').trigger('click')
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('version.updateNow'))!
      .trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('version.restartNow'))!
      .trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('更新代理不可用')
    expect(mockGetUpdateStatus).not.toHaveBeenCalled()
  })

  it('Docker 回退准备后激活请求失败时显示错误且不进入普通重启倒计时', async () => {
    mockHasUpdate = false
    mockRestartService.mockRejectedValue(new Error('回退代理不可用'))
    const wrapper = mount(VersionBadge, {
      global: {
        stubs: { Icon: true }
      }
    })

    await wrapper.get('button').trigger('click')
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('version.rollback'))!
      .trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('v0.1.172'))!
      .trigger('click')
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('version.rollbackConfirm'))!
      .trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('version.restartNow'))!
      .trigger('click')
    await flushPromises()

    expect(mockRollback).toHaveBeenCalledWith('0.1.172')
    expect(wrapper.text()).toContain('回退代理不可用')
    expect(mockGetUpdateStatus).not.toHaveBeenCalled()
  })
})
