import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { BrowserRouter } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext.tsx'
import { Toaster } from 'react-hot-toast'
import { Analytics } from '@vercel/analytics/react'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <Toaster
          position='bottom-right'
          toastOptions={{
            style: {
              background: '#fff7ed',
              color: '#c2410c',
              border: '1px solid #fed7aa',
              borderRadius: '12px',
            },
            success: {
              style: {
                color: '#15803d',
                background: '#f0fdf4',
                border: '1px solid #bbf7d0',
              }
            }
          }}
        />
        <App />
        <Analytics />
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>,
)
