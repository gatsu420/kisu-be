import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [
    react(),
    {
      name: 'auth-callback-redirect',
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          if (req.url?.startsWith('/auth/v1/callback')) {
            ;(async () => {
              try {
                const backendRes = await fetch(`http://localhost:8080${req.url}`, {
                  headers: { cookie: req.headers.cookie ?? '' },
                })
                const setCookies = backendRes.headers.getSetCookie()
                if (setCookies && setCookies.length > 0) {
                  res.setHeader('Set-Cookie', setCookies)
                }
                res.writeHead(302, { Location: '/' })
                res.end()
              } catch {
                if (!res.headersSent) {
                  res.writeHead(302, { Location: '/login?error=backend_unavailable' })
                }
                res.end()
              }
            })()
            return
          }
          next()
        })
      },
    },
  ],
  server: {
    port: 5173,
    proxy: {
      '/auth/v1/get-permission': 'http://localhost:8080',
      '/answer': 'http://localhost:8080',
    },
  },
})
