import gsap from 'gsap'

export const hoverScale = (el: HTMLElement, scale = 1.02) => {
  gsap.to(el, { scale, duration: 0.2, ease: 'power2.out' })
}

export const resetScale = (el: HTMLElement) => {
  gsap.to(el, { scale: 1, duration: 0.2, ease: 'power2.out' })
}

export const prefersReducedMotion = () =>
  typeof window !== 'undefined' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches
