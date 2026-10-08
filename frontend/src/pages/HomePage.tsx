import React from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

const HomePage: React.FC = () => {
  const { isAuthenticated, user, logout } = useAuth()

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
      <h1 className="text-4xl font-bold text-gray-900 mb-4">
        Добро пожаловать в CodeStream
      </h1>
      <p className="text-lg text-gray-600 mb-8 max-w-2xl">
        CodeStream — онлайн-редактор кода с совместным редактированием в реальном времени.
        Создавайте проекты, приглашайте коллег и пишите код вместе.
      </p>

      {isAuthenticated ? (
        <div className="flex flex-col items-center gap-4">
          <p className="text-gray-700">
            Вы вошли как <span className="font-medium">{user?.email}</span>
          </p>
          <div className="flex gap-4">
            <Link
              to="/projects"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
            >
              Мои проекты
            </Link>
            <button
              onClick={() => logout()}
              className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
            >
              Выйти
            </button>
          </div>
        </div>
      ) : (
        <div className="flex gap-4">
          <Link
            to="/login"
            className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
          >
            Войти
          </Link>
          <Link
            to="/register"
            className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
          >
            Зарегистрироваться
          </Link>
        </div>
      )}
    </div>
  )
}

export default HomePage
