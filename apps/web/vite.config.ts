import { defineConfig, loadEnv, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import type { IncomingMessage, ServerResponse } from 'node:http'

/** 生成 ULID 风格 26 字符随机串（与后端口径一致，仅供 dev mock 中间件使用） */
function mockRequestId(): string {
  const chars = '0123456789ABCDEFGHJKMNPQRSTVWXYZ'
  let out = ''
  for (let i = 0; i < 26; i++) out += chars[Math.floor(Math.random() * chars.length)]
  return out
}

/**
 * 官网 mock 中间件契约：
 *   GET /api/site/common   → mock/common.json   公共数据（导航 / 页脚）
 *   GET /api/site/home     → mock/home.json     首页数据
 *   GET /api/site/download → mock/download.json 下载中心
 *   GET /api/site/faq      → mock/faq.json      常见问题
 *   GET /api/site/contact  → mock/contact.json  商务合作
 * 统一返回 { message, body, request_id, code }，code === 0 表示成功。
 */
function resolveMock(pathname: string, mockDir: string): unknown | null {
  const segments = pathname.split('/').filter(Boolean)
  if (segments[0] !== 'site') return null
  const name = segments[1]
  if (!name) return null
  const p = join(mockDir, `${name}.json`)
  if (!existsSync(p)) return null
  return JSON.parse(readFileSync(p, 'utf-8'))
}

function mockPlugin(): Plugin {
  const mockDir = fileURLToPath(new URL('./mock', import.meta.url))
  return {
    name: 'dev-mock',
    configureServer(server) {
      server.middlewares.use('/api', (req: IncomingMessage, res: ServerResponse, next: () => void) => {
        if (!req.url) return next()
        const url = new URL(req.url, 'http://localhost')
        const pathname = decodeURIComponent(url.pathname).replace(/^\/api/, '')
        const segments = pathname.split('/').filter(Boolean)
        if (segments.length === 0) return next()

        const requestId = (req.headers['x-request-id'] as string) || mockRequestId()
        const send = (code: number, message: string, body: unknown, status = 200) => {
          setTimeout(() => {
            res.statusCode = status
            res.setHeader('Content-Type', 'application/json; charset=utf-8')
            res.end(JSON.stringify({ message, body, request_id: requestId, code }))
          }, 80 + Math.random() * 160)
        }

        if (req.method !== 'GET') return send(405, 'mock: method not allowed', null, 405)

        const data = resolveMock(pathname, mockDir)
        if (data === null) return send(404, 'mock: not found', null, 404)
        return send(0, 'ok', data)
      })
    },
  }
}

export default defineConfig(({ mode }) => {
  // 官网数据链路开关：
  // - 默认（VITE_USE_MOCK 未设置或非 '1'）：/api/* 由 server.proxy 转发到真实后端（http://localhost:8080）
  // - VITE_USE_MOCK=1：由 mockPlugin 中间件接管，返回 ./mock/*.json（无需后端即可开发）
  const useMock = loadEnv(mode, process.cwd(), '').VITE_USE_MOCK === '1'

  return {
    plugins: useMock ? [vue(), mockPlugin()] : [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      open: true,
      host: '0.0.0.0',
      cors: true,
      port: 5173,
      proxy: useMock
        ? undefined
        : {
            '/api': {
              target: 'http://localhost:8080',
              changeOrigin: true,
            },
          },
    },
    build: {
      outDir: './dist',
      emptyOutDir: true,
      rollupOptions: {
        output: {
          entryFileNames: 'js/[name].js',
          chunkFileNames: 'js/[name]-[hash].js',
          assetFileNames: 'css/[name].[ext]',
        },
      },
    },
  }
})
