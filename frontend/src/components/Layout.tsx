import React, { useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import gsap from 'gsap'
import { prefersReducedMotion } from '../utils/animations'

interface LayoutProps {
  children: React.ReactNode
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const { isAuthenticated, user, logout } = useAuth()
  const headerRef = useRef<HTMLElement>(null)
  const logoRef = useRef<HTMLAnchorElement>(null)

  useEffect(() => {
    if (!headerRef.current) return
    if (prefersReducedMotion()) {
      gsap.set(headerRef.current, { opacity: 1, y: 0 })
      return
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        headerRef.current,
        { y: -20, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.6, ease: 'power2.out' }
      )
    })
    return () => ctx.revert()
  }, [])

  const handleLogoEnter = () => {
    if (logoRef.current) {
      gsap.to(logoRef.current, {
        scale: 1.05,
        textShadow: '0 0 12px rgba(98, 78, 194, 0.6)',
        duration: 0.3,
        ease: 'power2.out',
      })
    }
  }

  const handleLogoLeave = () => {
    if (logoRef.current) {
      gsap.to(logoRef.current, {
        scale: 1,
        textShadow: '0 0 0px rgba(98, 78, 194, 0)',
        duration: 0.3,
        ease: 'power2.out',
      })
    }
  }

  return (
    <div className="min-h-screen bg-[#202020] text-[#F5F0EA]">
      <header
        ref={headerRef}
        className="bg-[#202020] border-b border-[#3A3A3A]"
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
          <Link
            ref={logoRef}
            to="/"
            onMouseEnter={handleLogoEnter}
            onMouseLeave={handleLogoLeave}
            className="text-2xl font-bold text-[#624EC2] inline-block"
          >
            CodeStream
          </Link>

          <nav className="flex items-center gap-4">
            {isAuthenticated ? (
              <>
                <span className="text-sm text-[#A0A0A0] hidden sm:inline">
                  {user?.email}
                </span>
                <Link
                  to="/projects"
                  className="relative text-sm font-medium text-[#F5F0EA] hover:text-[#624EC2] transition-colors duration-200 group"
                >
                  Projects
                  <span className="absolute left-0 -bottom-0.5 w-0 h-0.5 bg-[#624EC2] transition-all duration-300 group-hover:w-full" />
                </Link>
                <button
                  onClick={() => logout()}
                  className="text-sm font-medium text-[#F5F0EA] hover:text-[#624EC2] transition-colors duration-200"
                >
                  Logout
                </button>
              </>
            ) : (
              <>
                <Link
                  to="/login"
                  className="relative text-sm font-medium text-[#F5F0EA] hover:text-[#624EC2] transition-colors duration-200 group"
                >
                  Login
                  <span className="absolute left-0 -bottom-0.5 w-0 h-0.5 bg-[#624EC2] transition-all duration-300 group-hover:w-full" />
                </Link>
                <Link
                  to="/register"
                  className="relative text-sm font-medium text-[#F5F0EA] hover:text-[#624EC2] transition-colors duration-200 group"
                >
                  Register
                  <span className="absolute left-0 -bottom-0.5 w-0 h-0.5 bg-[#624EC2] transition-all duration-300 group-hover:w-full" />
                </Link>
              </>
            )}
          </nav>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>

      <footer className="mt-auto border-t border-[#3A3A3A] bg-[#202020]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 text-center text-sm text-[#A0A0A0]">
          © CodeStream
        </div>
      </footer>
    </div>
  )
}

export default Layout
