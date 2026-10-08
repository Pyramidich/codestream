import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { loader } from '@monaco-editor/react'
import './index.css'
import App from './App.tsx'

loader.config({
  paths: {
    vs: '/node_modules/monaco-editor/min/vs',
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
