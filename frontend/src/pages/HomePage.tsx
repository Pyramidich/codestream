import React, { useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import gsap from 'gsap'
import { prefersReducedMotion } from '../utils/animations'

const HomePage: React.FC = () => {
  const { isAuthenticated, user, logout } = useAuth()
  const titleRef = useRef<HTMLHeadingElement>(null)
  const descRef = useRef<HTMLParagraphElement>(null)
  const buttonsRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (prefersReducedMotion()) {
      gsap.set([titleRef.current, descRef.current, buttonsRef.current], {
        opacity: 1,
        y: 0,
      })
      return
    }

    const ctx = gsap.context(() => {
      const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })
      tl.fromTo(
        titleRef.current,
        { y: 30, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.8 }
      )
        .fromTo(
          descRef.current,
          { y: 20, opacity: 0 },
          { y: 0, opacity: 1, duration: 0.6 },
          '-=0.6'
        )
        .fromTo(
          buttonsRef.current,
          { y: 20, opacity: 0 },
          { y: 0, opacity: 1, duration: 0.6 },
          '-=0.4'
        )
    })

    return () => ctx.revert()
  }, [])

  const handlePrimaryEnter = (e: React.MouseEvent<HTMLElement>) => {
    gsap.to(e.currentTarget, {
      scale: 1.02,
      boxShadow: '0 4px 20px rgba(98, 78, 194, 0.4)',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handlePrimaryLeave = (e: React.MouseEvent<HTMLElement>) => {
    gsap.to(e.currentTarget, {
      scale: 1,
      boxShadow: '0 0 0 rgba(98, 78, 194, 0)',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleSecondaryEnter = (e: React.MouseEvent<HTMLElement>) => {
    gsap.to(e.currentTarget, {
      borderColor: '#624EC2',
      color: '#624EC2',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleSecondaryLeave = (e: React.MouseEvent<HTMLElement>) => {
    gsap.to(e.currentTarget, {
      borderColor: '#3A3A3A',
      color: '#F5F0EA',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
      <h1 ref={titleRef} className="text-4xl font-bold text-[#F5F0EA] mb-4">
        Добро пожаловать в CodeStream
      </h1>
      <p ref={descRef} className="text-lg text-[#A0A0A0] mb-8 max-w-2xl">
        CodeStream — онлайн-редактор кода с совместным редактированием в реальном времени.
        Создавайте проекты, приглашайте коллег и пишите код вместе.
      </p>

      {isAuthenticated ? (
        <div ref={buttonsRef} className="flex flex-col items-center gap-4">
          <p className="text-[#F5F0EA]">
            Вы вошли как <span className="font-medium">{user?.email}</span>
          </p>
          <div className="flex gap-4">
            <Link
              to="/projects"
              onMouseEnter={handlePrimaryEnter}
              onMouseLeave={handlePrimaryLeave}
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-[#624EC2] hover:bg-[#7B68D1]"
            >
              Мои проекты
            </Link>
            <button
              onClick={() => logout()}
              onMouseEnter={handleSecondaryEnter}
              onMouseLeave={handleSecondaryLeave}
              className="inline-flex items-center px-6 py-3 border border-[#3A3A3A] text-base font-medium rounded-md text-[#F5F0EA] bg-transparent hover:bg-[#2A2A2A]"
            >
              Выйти
            </button>
          </div>
        </div>
      ) : (
        <div ref={buttonsRef} className="flex gap-4">
          <Link
            to="/login"
            onMouseEnter={handlePrimaryEnter}
            onMouseLeave={handlePrimaryLeave}
            className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-[#624EC2] hover:bg-[#7B68D1]"
          >
            Войти
          </Link>
          <Link
            to="/register"
            onMouseEnter={handleSecondaryEnter}
            onMouseLeave={handleSecondaryLeave}
            className="inline-flex items-center px-6 py-3 border border-[#3A3A3A] text-base font-medium rounded-md text-[#F5F0EA] bg-transparent hover:bg-[#2A2A2A]"
          >
            Зарегистрироваться
          </Link>
        </div>
      )}
    </div>
  )
}

export default HomePage
