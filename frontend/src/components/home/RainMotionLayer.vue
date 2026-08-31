<template>
  <div
    ref="layerRef"
    data-testid="rain-layer"
    :data-reduced-motion="String(reducedMotion)"
    :data-motion-paused="String(motionPaused)"
    :data-active-drops="String(activeDropCount)"
    :data-impact-ripple-count="String(impactRippleTriggerCount)"
    :data-active-petals="String(activePetalCount)"
    :data-click-ripple-count="String(clickRippleTriggerCount)"
    class="rain-motion-layer"
    aria-hidden="true"
  >
    <template v-if="!reducedMotion">
      <img
        v-for="drop in drops"
        :key="drop.id"
        :ref="(element) => setDropRef(element, drop.id)"
        data-testid="rain-drop"
        :src="rainDropOverlay"
        alt=""
        :class="['rain-drop', { 'mobile-motion-hidden': drop.id >= 2 }]"
        :style="{ left: `${drop.left}%`, animationDelay: `${drop.delay}s` }"
      />
      <span
        v-for="impactRipple in impactRipples"
        :key="`impact-ripple-${impactRipple.id}`"
        :ref="(element) => setImpactRippleRef(element, impactRipple.id)"
        data-testid="impact-ripple"
        class="impact-ripple"
      >
        <img
          v-for="wave in impactRippleWaves"
          :key="wave"
          :src="rippleOverlay"
          alt=""
          data-testid="impact-ripple-wave"
          :class="['impact-ripple-wave', `impact-ripple-wave--${wave}`]"
        />
      </span>
      <span
        v-for="clickRipple in clickRipples"
        :key="`click-ripple-${clickRipple.id}`"
        :ref="(element) => setClickRippleRef(element, clickRipple.id)"
        data-testid="click-ripple"
        class="click-ripple"
      >
        <img
          v-for="wave in clickRippleWaves"
          :key="wave"
          :src="rippleOverlay"
          alt=""
          data-testid="click-ripple-wave"
          :class="['click-ripple-wave', `click-ripple-wave--${wave}`]"
        />
      </span>
      <span
        v-for="petal in petals"
        :key="`petal-${petal.id}`"
        :ref="(element) => setPetalRef(element, petal.id)"
        data-testid="flower-petal"
        data-petal-source="cc0-photo-crop"
        :class="['flower-petal', { 'mobile-motion-hidden': petal.id >= 7 }]"
        :style="petalStyle(petal)"
      >
        <img
          :src="pinkFlowerSource"
          alt=""
          class="flower-petal-photo"
        />
      </span>
    </template>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, type ComponentPublicInstance } from 'vue'
import { gsap } from 'gsap'

import rainDropOverlay from '@/assets/home/rain-drop-overlay.png'
import rippleOverlay from '@/assets/home/ripple-overlay.png'
import pinkFlowerSource from '@/assets/home/source-pink-flower-cc0.webp'

const drops = [
  { id: 0, left: 13, delay: 0.2 },
  { id: 1, left: 36, delay: 2.1 },
  { id: 2, left: 68, delay: 3.8 },
  { id: 3, left: 87, delay: 5.4 },
]

const clickRipples = Array.from({ length: 5 }, (_, id) => ({ id }))
const clickRippleWaves = ['impact', 'primary', 'wake'] as const
const impactRipples = Array.from({ length: 8 }, (_, id) => ({ id }))
const impactRippleWaves = ['primary', 'wake'] as const

const flowerWidth = 1644
const petalCrops = [
  { x: 548, y: 32, width: 510, height: 565, clip: 'polygon(48% 100%, 18% 86%, 4% 55%, 8% 27%, 28% 5%, 55% 0%, 82% 12%, 97% 42%, 87% 75%)' },
  { x: 870, y: 145, width: 515, height: 595, clip: 'polygon(2% 100%, 4% 73%, 18% 42%, 43% 16%, 70% 4%, 94% 20%, 100% 48%, 82% 75%, 46% 94%)' },
  { x: 1080, y: 420, width: 550, height: 610, clip: 'polygon(0% 72%, 10% 43%, 36% 18%, 69% 3%, 94% 16%, 100% 46%, 91% 72%, 58% 93%, 25% 91%)' },
  { x: 930, y: 790, width: 560, height: 650, clip: 'polygon(0% 2%, 30% 8%, 62% 28%, 91% 58%, 96% 84%, 75% 98%, 44% 91%, 20% 66%, 5% 35%)' },
  { x: 600, y: 965, width: 500, height: 585, clip: 'polygon(3% 0%, 28% 8%, 52% 27%, 82% 14%, 97% 37%, 91% 72%, 66% 98%, 36% 96%, 13% 69%)' },
  { x: 250, y: 800, width: 550, height: 630, clip: 'polygon(98% 2%, 92% 35%, 73% 66%, 44% 93%, 17% 98%, 2% 73%, 10% 44%, 36% 17%, 67% 4%)' },
  { x: 45, y: 430, width: 610, height: 625, clip: 'polygon(100% 70%, 71% 92%, 38% 93%, 10% 73%, 0% 44%, 13% 17%, 42% 5%, 76% 20%, 94% 43%)' },
  { x: 300, y: 145, width: 500, height: 585, clip: 'polygon(98% 100%, 63% 91%, 31% 70%, 9% 43%, 14% 17%, 39% 2%, 67% 12%, 85% 41%, 97% 72%)' },
]

const petals = [
  { id: 0, crop: 0, left: 8, size: 42, drift: 115, duration: 14.5, rotation: 310, opacity: 0.68, tint: 'saturate(.74) brightness(1.2)' },
  { id: 1, crop: 1, left: 17, size: 30, drift: -88, duration: 12.2, rotation: -260, opacity: 0.56, tint: 'saturate(.52) brightness(1.42)' },
  { id: 2, crop: 2, left: 29, size: 48, drift: 92, duration: 16.8, rotation: 380, opacity: 0.62, tint: 'saturate(.82) brightness(1.12) hue-rotate(-8deg)' },
  { id: 3, crop: 3, left: 42, size: 35, drift: -122, duration: 13.8, rotation: -340, opacity: 0.58, tint: 'saturate(.58) brightness(1.35)' },
  { id: 4, crop: 4, left: 55, size: 54, drift: 128, duration: 18.2, rotation: 420, opacity: 0.64, tint: 'saturate(.7) brightness(1.24)' },
  { id: 5, crop: 5, left: 67, size: 32, drift: -76, duration: 12.8, rotation: -290, opacity: 0.52, tint: 'saturate(.46) brightness(1.48)' },
  { id: 6, crop: 6, left: 83, size: 45, drift: 108, duration: 15.6, rotation: 350, opacity: 0.62, tint: 'saturate(.78) brightness(1.16) hue-rotate(-6deg)' },
  { id: 7, crop: 7, left: 94, size: 37, drift: -104, duration: 14.1, rotation: -320, opacity: 0.56, tint: 'saturate(.54) brightness(1.4)' },
  { id: 8, crop: 2, left: 23, size: 27, drift: 64, duration: 11.8, rotation: 280, opacity: 0.46, tint: 'saturate(.42) brightness(1.5)' },
  { id: 9, crop: 5, left: 76, size: 51, drift: -136, duration: 17.5, rotation: -390, opacity: 0.6, tint: 'saturate(.72) brightness(1.22)' },
  { id: 10, crop: 0, left: 35, size: 34, drift: -72, duration: 13.2, rotation: 300, opacity: 0.5, tint: 'saturate(.5) brightness(1.44)' },
  { id: 11, crop: 3, left: 88, size: 29, drift: 84, duration: 12.5, rotation: -270, opacity: 0.52, tint: 'saturate(.62) brightness(1.32)' },
]

const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
const reducedMotion = ref(motionQuery.matches)
const motionPaused = ref(false)
const activeDropCount = ref(0)
const activePetalCount = ref(0)
const clickRippleTriggerCount = ref(0)
const impactRippleTriggerCount = ref(0)
const layerRef = ref<HTMLElement | null>(null)
const dropRefs: Array<HTMLElement | null> = []
const impactRippleRefs: Array<HTMLElement | null> = []
const clickRippleRefs: Array<HTMLElement | null> = []
const petalRefs: Array<HTMLElement | null> = []

let responsiveMotion: gsap.MatchMedia | null = null
let visibilityObserver: IntersectionObserver | null = null
let animations: gsap.core.Animation[] = []
let impactRippleAnimations: Array<gsap.core.Animation | null> = []
let clickRippleAnimations: Array<gsap.core.Animation | null> = []
let interactionSurface: HTMLElement | null = null
let nextImpactRippleIndex = 0
let nextClickRippleIndex = 0
let isLayerIntersecting = true

function resolveElement(element: Element | ComponentPublicInstance | null): HTMLElement | null {
  return element instanceof HTMLElement ? element : null
}

function setDropRef(element: Element | ComponentPublicInstance | null, index: number) {
  dropRefs[index] = resolveElement(element)
}

function setImpactRippleRef(element: Element | ComponentPublicInstance | null, index: number) {
  impactRippleRefs[index] = resolveElement(element)
}

function setClickRippleRef(element: Element | ComponentPublicInstance | null, index: number) {
  clickRippleRefs[index] = resolveElement(element)
}

function setPetalRef(element: Element | ComponentPublicInstance | null, index: number) {
  petalRefs[index] = resolveElement(element)
}

function petalStyle(petal: typeof petals[number]) {
  const crop = petalCrops[petal.crop]
  const cropScale = petal.size / crop.width
  return {
    left: `${petal.left}%`,
    width: `${petal.size}px`,
    height: `${petal.size * crop.height / crop.width}px`,
    filter: petal.tint,
    clipPath: crop.clip,
    '--petal-photo-width': `${flowerWidth / crop.width * 100}%`,
    '--petal-photo-left': `${-crop.x * cropScale}px`,
    '--petal-photo-top': `${-crop.y * cropScale}px`,
  }
}

function setMotionPaused(paused: boolean) {
  motionPaused.value = paused
  animations.forEach((animation) => paused ? animation.pause() : animation.resume())
  impactRippleAnimations.forEach((animation) => paused ? animation?.pause() : animation?.resume())
  clickRippleAnimations.forEach((animation) => paused ? animation?.pause() : animation?.resume())
}

function syncMotionState() {
  setMotionPaused(document.hidden || !isLayerIntersecting)
}

function triggerImpactRipple(x: number, y: number, kind: 'drop' | 'petal') {
  if (reducedMotion.value || motionPaused.value || !layerRef.value) return

  const ripple = impactRippleRefs[nextImpactRippleIndex]
  if (!ripple) return
  const waves = Array.from(ripple.querySelectorAll<HTMLElement>('.impact-ripple-wave'))
  if (waves.length !== impactRippleWaves.length) return

  const animationIndex = nextImpactRippleIndex
  nextImpactRippleIndex = (nextImpactRippleIndex + 1) % impactRippleRefs.length
  impactRippleTriggerCount.value += 1
  impactRippleAnimations[animationIndex]?.kill()

  // 固定节点池模拟落水冲击，避免雨滴/花瓣高频落点反复创建 DOM。
  gsap.set(ripple, { x, y, autoAlpha: 1 })
  ripple.dataset.impactKind = kind
  ripple.dataset.impactAnchor = 'element-bottom-center'
  impactRippleAnimations[animationIndex] = gsap.timeline({
    onComplete: () => {
      gsap.set(ripple, { autoAlpha: 0 })
      impactRippleAnimations[animationIndex] = null
    },
  })
    .fromTo(waves[0],
      { xPercent: -50, yPercent: -50, scaleX: 0.06, scaleY: 0.04, autoAlpha: 0.78 },
      { scaleX: 0.68, scaleY: 0.46, autoAlpha: 0.46, duration: 0.62, ease: 'power1.out' },
    )
    .to(waves[0], {
      scaleX: 1.24,
      scaleY: 0.86,
      autoAlpha: 0,
      duration: 1.48,
      ease: 'sine.out',
    })
    .fromTo(waves[1],
      { xPercent: -50, yPercent: -50, scaleX: 0.04, scaleY: 0.03, autoAlpha: 0 },
      { scaleX: 0.88, scaleY: 0.58, autoAlpha: 0.32, duration: 1.1, ease: 'power1.out' },
      0.16,
    )
    .to(waves[1], {
      scaleX: 1.6,
      scaleY: 1.08,
      autoAlpha: 0,
      duration: 1.18,
      ease: 'sine.out',
    }, 1.1)
}

function triggerElementImpactRipple(element: HTMLElement, kind: 'drop' | 'petal') {
  if (!layerRef.value) return
  const layerBounds = layerRef.value.getBoundingClientRect()
  const elementBounds = element.getBoundingClientRect()
  // 以素材实际可见盒的下缘为落点，避免固定水面线把波纹画在花瓣上方。
  const x = elementBounds.left - layerBounds.left + elementBounds.width * 0.5
  const y = elementBounds.bottom - layerBounds.top
  triggerImpactRipple(x, y, kind)
}

function startMotion() {
  if (reducedMotion.value || !layerRef.value) return

  responsiveMotion?.revert()
  impactRippleAnimations.forEach((animation) => animation?.kill())
  impactRippleAnimations = []
  animations = []
  responsiveMotion = gsap.matchMedia()
  responsiveMotion.add({
    isMobile: '(max-width: 767px)',
    isDesktop: '(min-width: 768px)',
  }, (context) => {
    animations = []
    const isMobile = Boolean(context.conditions?.isMobile)
    const activeDrops = dropRefs.slice(0, isMobile ? 2 : dropRefs.length)
    const activePetals = petalRefs.slice(0, isMobile ? 7 : petalRefs.length)
    activeDropCount.value = activeDrops.filter(Boolean).length
    activePetalCount.value = activePetals.filter(Boolean).length
    impactRippleAnimations = Array.from({ length: impactRipples.length }, () => null)

    // 使用真实可视层高计算水面位置；固定下限 720 会把水面推到容器之外，导致花瓣与水波错位。
    const layerHeight = layerRef.value?.clientHeight || 720
    const waterline = layerHeight * 0.78

    activeDrops.forEach((drop, index) => {
      if (!drop) return
      const dropTimeline = gsap.timeline({
        delay: drops[index].delay,
        repeat: -1,
        repeatDelay: 4.6,
      })
      dropTimeline
        .fromTo(drop,
          { y: -130, autoAlpha: 0 },
          { y: waterline - 18, autoAlpha: 0.52, duration: 3.4 + index * 0.2, ease: 'sine.in' },
        )
        .call(() => triggerElementImpactRipple(drop, 'drop'))
        .to(drop, { y: waterline + 18, autoAlpha: 0, duration: 0.28, ease: 'sine.out' })
      animations.push(dropTimeline)
    })

    activePetals.forEach((petal, index) => {
      if (!petal) return
      const config = petals[index]
      const surfaceDrift = index % 2 === 0 ? 34 : -30
      const surfaceBob = (index % 3 - 1) * 4
      const petalTimeline = gsap.timeline({
        repeat: -1,
        repeatDelay: 0.8 + index % 4 * 0.45,
      })
      petalTimeline
        .fromTo(petal,
          {
            x: 0,
            y: -100 - index % 3 * 65,
            rotationX: index % 2 ? 32 : -24,
            rotationY: index % 3 ? -34 : 38,
            rotationZ: index * 23,
            scale: 1,
            autoAlpha: 0,
          },
          {
            x: config.drift * 0.48,
            y: layerHeight * 0.48,
            rotationX: index % 2 ? -48 : 54,
            rotationY: index % 3 ? 58 : -52,
            rotationZ: config.rotation * 0.48,
            duration: config.duration * 0.59,
            ease: 'none',
          },
        )
        .to(petal, { autoAlpha: config.opacity, duration: 0.75, ease: 'sine.out' }, 0)
        .to(petal, {
          x: config.drift,
          y: waterline,
          rotationX: index % 2 ? 38 : -34,
          rotationY: index % 3 ? -24 : 28,
          rotationZ: config.rotation,
          autoAlpha: config.opacity * 0.84,
          duration: config.duration * 0.29,
          // 两段下落按路程比例分配时长并保持线性速度，避免中段缓动衔接产生停顿。
          ease: 'none',
        })
        .call(() => triggerElementImpactRipple(petal, 'petal'))
        .to(petal, {
          x: config.drift + surfaceDrift * 0.58,
          y: waterline + surfaceBob,
          rotationX: index % 2 ? 20 : -18,
          rotationY: index % 3 ? -12 : 14,
          rotationZ: config.rotation + surfaceDrift * 0.62,
          scale: 0.94,
          autoAlpha: config.opacity * 0.72,
          duration: 1.9,
          ease: 'sine.inOut',
        })
        .to(petal, {
          x: config.drift + surfaceDrift,
          y: waterline + surfaceBob * 0.5,
          rotationX: index % 2 ? 14 : -12,
          rotationY: index % 3 ? -8 : 10,
          rotationZ: config.rotation + surfaceDrift,
          scale: 0.86,
          autoAlpha: 0,
          // 花瓣先随水面漂移，再在继续漂动时缓慢淡出，而不是落水后立即消失。
          duration: 1.3,
          ease: 'sine.inOut',
        })
      // 初始错峰只调整播放头，不把跳过的落水回调误算成一次真实撞击。
      petalTimeline.progress((index * 0.137) % 0.94, true)
      animations.push(petalTimeline)
    })
    syncMotionState()
  }, layerRef.value)
}

function handleSurfaceClick(event: MouseEvent) {
  if (reducedMotion.value || motionPaused.value || event.button !== 0 || !layerRef.value) return

  const ripple = clickRippleRefs[nextClickRippleIndex]
  if (!ripple) return

  const waves = Array.from(ripple.querySelectorAll<HTMLElement>('.click-ripple-wave'))
  if (waves.length !== clickRippleWaves.length) return

  const bounds = layerRef.value.getBoundingClientRect()
  const x = event.clientX - bounds.left
  const y = event.clientY - bounds.top
  const animationIndex = nextClickRippleIndex
  nextClickRippleIndex = (nextClickRippleIndex + 1) % clickRippleRefs.length
  clickRippleTriggerCount.value += 1

  clickRippleAnimations[animationIndex]?.kill()
  // 复用固定节点池，避免连续点击时反复创建和销毁 DOM。
  gsap.set(ripple, { x, y, autoAlpha: 1 })
  // 依据真实落水照片分三段模拟：冲击环先起、主波随后扩张、尾波最后缓慢消散。
  clickRippleAnimations[animationIndex] = gsap.timeline({
    onComplete: () => {
      gsap.set(ripple, { autoAlpha: 0 })
      clickRippleAnimations[animationIndex] = null
    },
  })
    .fromTo(waves[0],
      {
        xPercent: -50,
        yPercent: -50,
        scaleX: 0.12,
        scaleY: 0.08,
        autoAlpha: 0.96,
      },
      {
        scaleX: 0.72,
        scaleY: 0.5,
        autoAlpha: 0.7,
        duration: 1.05,
        ease: 'power1.out',
      },
    )
    .to(waves[0], {
      scaleX: 1.28,
      scaleY: 0.9,
      autoAlpha: 0,
      duration: 1.75,
      ease: 'sine.out',
    }, 1.05)
    .fromTo(waves[1],
      {
        xPercent: -50,
        yPercent: -50,
        scaleX: 0.08,
        scaleY: 0.06,
        autoAlpha: 0,
      },
      {
        scaleX: 0.98,
        scaleY: 0.7,
        autoAlpha: 0.62,
        duration: 1.8,
        ease: 'power1.out',
      },
      0.18,
    )
    .to(waves[1], {
      scaleX: 1.72,
      scaleY: 1.2,
      autoAlpha: 0,
      duration: 2,
      ease: 'sine.out',
    }, 1.98)
    .fromTo(waves[2],
      {
        xPercent: -50,
        yPercent: -50,
        scaleX: 0.06,
        scaleY: 0.04,
        autoAlpha: 0,
      },
      {
        scaleX: 1.16,
        scaleY: 0.82,
        autoAlpha: 0.42,
        duration: 2.2,
        ease: 'power1.out',
      },
      0.5,
    )
    .to(waves[2], {
      scaleX: 2.05,
      scaleY: 1.42,
      autoAlpha: 0,
      duration: 1.8,
      ease: 'sine.out',
    }, 2.7)
}

function handleMotionPreference(event: MediaQueryListEvent) {
  reducedMotion.value = event.matches
  responsiveMotion?.revert()
  responsiveMotion = null
  animations = []
  impactRippleAnimations.forEach((animation) => animation?.kill())
  impactRippleAnimations = []
  clickRippleAnimations.forEach((animation) => animation?.kill())
  clickRippleAnimations = []
  activeDropCount.value = 0
  activePetalCount.value = 0
  if (!event.matches) nextTick(startMotion)
}

function handleDocumentVisibility() {
  syncMotionState()
}

onMounted(async () => {
  motionQuery.addEventListener?.('change', handleMotionPreference)
  document.addEventListener('visibilitychange', handleDocumentVisibility)
  await nextTick()
  startMotion()

  // 动效层本身不接管指针事件，监听父级可保留按钮与链接的原有交互。
  interactionSurface = layerRef.value?.parentElement ?? null
  interactionSurface?.addEventListener('click', handleSurfaceClick, { passive: true })

  if ('IntersectionObserver' in window && layerRef.value) {
    visibilityObserver = new IntersectionObserver(([entry]) => {
      if (!entry) return
      isLayerIntersecting = entry.isIntersecting
      syncMotionState()
    }, {
      threshold: 0.05,
    })
    visibilityObserver.observe(layerRef.value)
  }
})

onBeforeUnmount(() => {
  motionQuery.removeEventListener?.('change', handleMotionPreference)
  document.removeEventListener('visibilitychange', handleDocumentVisibility)
  interactionSurface?.removeEventListener('click', handleSurfaceClick)
  interactionSurface = null
  visibilityObserver?.disconnect()
  responsiveMotion?.revert()
  impactRippleAnimations.forEach((animation) => animation?.kill())
  impactRippleAnimations = []
  clickRippleAnimations.forEach((animation) => animation?.kill())
  clickRippleAnimations = []
  responsiveMotion = null
  animations = []
  activeDropCount.value = 0
  activePetalCount.value = 0
})
</script>

<style scoped>
.rain-motion-layer {
  pointer-events: none;
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.rain-drop {
  position: absolute;
  top: 0;
  width: clamp(22px, 3.3vw, 54px);
  height: auto;
  opacity: 0;
  will-change: transform, opacity;
}

.impact-ripple {
  position: absolute;
  top: 0;
  left: 0;
  width: 0;
  height: 0;
  opacity: 0;
  visibility: hidden;
}

.impact-ripple-wave {
  position: absolute;
  top: 0;
  left: 0;
  width: clamp(120px, 15vw, 220px);
  max-width: none;
  height: auto;
  opacity: 0;
  visibility: hidden;
  mix-blend-mode: multiply;
  filter: brightness(0.8) contrast(1.75) saturate(1.12) drop-shadow(0 0 9px rgb(42 122 111 / 30%));
  transform-origin: 50% 62%;
  will-change: transform, opacity;
}

.impact-ripple-wave--wake {
  filter: brightness(0.9) contrast(1.46) saturate(1.04) drop-shadow(0 0 7px rgb(54 133 122 / 20%));
}

.click-ripple {
  position: absolute;
  top: 0;
  left: 0;
  width: 0;
  height: 0;
  opacity: 0;
  visibility: hidden;
}

.click-ripple-wave {
  position: absolute;
  top: 0;
  left: 0;
  width: clamp(210px, 28vw, 400px);
  max-width: none;
  height: auto;
  opacity: 0;
  visibility: hidden;
  mix-blend-mode: multiply;
  filter: brightness(0.74) contrast(1.9) saturate(1.28) drop-shadow(0 0 14px rgb(34 111 101 / 42%));
  transform-origin: 50% 62%;
  will-change: transform, opacity;
}

.click-ripple-wave--primary {
  filter: brightness(0.8) contrast(1.72) saturate(1.16) drop-shadow(0 0 12px rgb(42 122 111 / 32%));
}

.click-ripple-wave--wake {
  filter: brightness(0.88) contrast(1.5) saturate(1.06) drop-shadow(0 0 10px rgb(54 133 122 / 22%));
}

.flower-petal {
  position: absolute;
  top: 0;
  display: block;
  overflow: hidden;
  opacity: 0;
  transform-style: preserve-3d;
  will-change: transform, opacity;
}

.flower-petal-photo {
  position: absolute;
  top: var(--petal-photo-top);
  left: var(--petal-photo-left);
  width: var(--petal-photo-width);
  max-width: none;
  height: auto;
  user-select: none;
}

:global(.dark) .impact-ripple-wave {
  mix-blend-mode: screen;
  filter: contrast(1.32) brightness(1.18) saturate(0.84);
}

:global(.dark) .click-ripple-wave {
  mix-blend-mode: screen;
  filter: contrast(1.34) brightness(1.22) saturate(0.86);
}

@media (max-width: 767px) {
  .mobile-motion-hidden {
    display: none;
  }
}
</style>
