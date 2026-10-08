import React, { useEffect, useRef, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import gsap from 'gsap'
import { useAuth } from '../contexts/AuthContext'
import { prefersReducedMotion } from '../utils/animations'

const RegisterPage: React.FC = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const { register } = useAuth()
  const navigate = useNavigate()
  const cardRef = useRef<HTMLDivElement>(null)
  const inputRefs = useRef<(HTMLInputElement | null)[]>([])

  useEffect(() => {
    if (!cardRef.current) return
    if (prefersReducedMotion()) {
      gsap.set(cardRef.current, { opacity: 1, y: 0 })
      return
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        cardRef.current,
        { y: 40, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.6, ease: 'power2.out' }
      )
    }, cardRef)

    return () => ctx.revert()
  }, [])

  const handleInputEnter = (e: React.MouseEvent<HTMLInputElement>) => {
    gsap.to(e.currentTarget, {
      borderColor: '#624EC2',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleInputLeave = (e: React.MouseEvent<HTMLInputElement>) => {
    gsap.to(e.currentTarget, {
      borderColor: '#3A3A3A',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleSubmitEnter = (e: React.MouseEvent<HTMLButtonElement>) => {
    gsap.to(e.currentTarget, { scale: 1.02, duration: 0.2, ease: 'power2.out' })
  }

  const handleSubmitLeave = (e: React.MouseEvent<HTMLButtonElement>) => {
    gsap.to(e.currentTarget, { scale: 1, duration: 0.2, ease: 'power2.out' })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)

    try {
      await register(email, password, displayName)
      navigate('/projects')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <div
        ref={cardRef}
        className="w-full max-w-md bg-[#2A2A2A] rounded-lg shadow-md p-8 border border-[#3A3A3A]"
      >
        <h2 className="text-2xl font-bold text-[#F5F0EA] mb-6 text-center">Register</h2>

        {error && (
          <div className="mb-4 p-3 bg-red-900/20 text-red-400 rounded-md text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="displayName" className="block text-sm font-medium text-[#A0A0A0]">
              Display name
            </label>
            <input
              ref={(el) => { inputRefs.current[0] = el }}
              id="displayName"
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              required
              onMouseEnter={handleInputEnter}
              onMouseLeave={handleInputLeave}
              className="mt-1 block w-full rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
            />
          </div>

          <div>
            <label htmlFor="email" className="block text-sm font-medium text-[#A0A0A0]">
              Email
            </label>
            <input
              ref={(el) => { inputRefs.current[1] = el }}
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              onMouseEnter={handleInputEnter}
              onMouseLeave={handleInputLeave}
              className="mt-1 block w-full rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-[#A0A0A0]">
              Password
            </label>
            <input
              ref={(el) => { inputRefs.current[2] = el }}
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              onMouseEnter={handleInputEnter}
              onMouseLeave={handleInputLeave}
              className="mt-1 block w-full rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
            />
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            onMouseEnter={handleSubmitEnter}
            onMouseLeave={handleSubmitLeave}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-[#624EC2] hover:bg-[#7B68D1] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#624EC2] disabled:opacity-50"
          >
            {isSubmitting ? 'Registering...' : 'Register'}
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-[#A0A0A0]">
          Already have an account?{' '}
          <Link to="/login" className="text-[#624EC2] hover:text-[#7B68D1]">
            Login
          </Link>
        </p>
      </div>
    </div>
  )
}

export default RegisterPage
