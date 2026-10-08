import React from 'react'

const HomePage: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center">
      <h1 className="text-4xl font-bold text-gray-900 mb-4">
        Добро пожаловать в CodeStream
      </h1>
      <p className="text-lg text-gray-600 mb-8 max-w-2xl">
        CodeStream — онлайн-редактор кода с совместным редактированием в реальном времени.
        Создавайте проекты, приглашайте коллег и пишите код вместе.
      </p>
      <div className="flex gap-4">
        <a
          href="/login"
          className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
        >
          Войти
        </a>
        <a
          href="/register"
          className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
        >
          Зарегистрироваться
        </a>
      </div>
    </div>
  )
}

export default HomePage
