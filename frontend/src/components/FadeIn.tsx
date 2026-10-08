import { useEffect, useRef, type ReactNode } from 'react'
import gsap from 'gsap'
import { prefersReducedMotion } from '../utils/animations'

interface FadeInProps {
  children: ReactNode
  delay?: number
  className?: string
}

export const FadeIn: React.FC<FadeInProps> = ({ children, delay = 0, className }) => {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!ref.current) return
    if (prefersReducedMotion()) {
      gsap.set(ref.current, { opacity: 1, y: 0 })
      return
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        ref.current,
        { opacity: 0, y: 20 },
        { opacity: 1, y: 0, duration: 0.6, delay, ease: 'power2.out' }
      )
    })

    return () => ctx.revert()
  }, [delay])

  return <div ref={ref} className={className}>{children}</div>
}
