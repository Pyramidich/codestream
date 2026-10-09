import React, { Suspense } from 'react'
import { HashRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import Layout from './components/Layout'
import HomePage from './pages/HomePage'
import ProtectedRoute from './components/ProtectedRoute'

const LoginPage = React.lazy(() => import('./pages/LoginPage'))
const RegisterPage = React.lazy(() => import('./pages/RegisterPage'))
const ProjectsPage = React.lazy(() => import('./pages/ProjectsPage'))
const ProjectPage = React.lazy(() => import('./pages/ProjectPage'))
const ProjectSettings = React.lazy(() => import('./pages/ProjectSettings'))
const EditorPage = React.lazy(() => import('./pages/EditorPage'))

function App() {
  return (
    <AuthProvider>
      <HashRouter>
        <Layout>
          <Suspense fallback={<div className="p-4 text-gray-600">Loading...</div>}>
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route element={<ProtectedRoute />}>
                <Route path="/projects" element={<ProjectsPage />} />
                <Route path="/projects/:id" element={<ProjectPage />} />
                <Route path="/projects/:id/settings" element={<ProjectSettings />} />
                <Route
                  path="/projects/:projectId/files/:fileId"
                  element={<EditorPage />}
                />
              </Route>
            </Routes>
          </Suspense>
        </Layout>
      </HashRouter>
    </AuthProvider>
  )
}

export default App
