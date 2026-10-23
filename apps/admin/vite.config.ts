import { defineConfig, loadEnv, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import type { IncomingMessage, ServerResponse } from 'node:http'

/** 生成 ULID 风格 26 字符随机串（与 src/utils/ulid.ts 同口径，仅供 dev mock 中间件使用） */
function mockRequestId(): string {
  const chars = '0123456789ABCDEFGHJKMNPQRSTVWXYZ'
  let out = ''
  for (let i = 0; i < 26; i++) out += chars[Math.floor(Math.random() * chars.length)]
  return out
}

/** 复数转单数（粗略规则：es 结尾去 es；s 结尾去 s），用于 merchant-s / rider-s / order-s 文件名归一 */
function singularize(word: string): string {
  if (word.endsWith('ies')) return word.slice(0, -3) + 'y'
  if (word.endsWith('ses')) return word.slice(0, -2)
  if (word.endsWith('s') && !word.endsWith('ss')) return word.slice(0, -1)
  return word
}

interface MockCtx {
  file: string
  data: unknown
  isIdFallback: boolean
  id?: string
}

/**
 * 根据路径段解析 mock 文件。
 * 规则见 PAGE_DEV_GUIDE.md「mock 中间件契约」：
 *  1) 后缀连字符组合精确命中（首段同时尝试复数→单数归一）
 *  2) 末段像 id 时，去掉末段再命中「…-detail」
 *  3) 未命中返回 null（由中间件回 404）
 */
function resolveMock(segments: string[], mockDir: string): MockCtx | null {
  const combos = (segs: string[], suffix = ''): string[] => {
    const out: string[] = []
    for (let i = 0; i < segs.length; i++) {
      const joined = segs.slice(i).join('-') + suffix
      out.push(joined)
      // 首段单复数归一
      const first = joined.split('-')[0]
      if (first.endsWith('s')) {
        out.push(singularize(first) + joined.slice(first.length))
      }
    }
    return Array.from(new Set(out))
  }

  const tryLoad = (name: string): unknown | null => {
    const p = join(mockDir, `${name}.json`)
    if (existsSync(p)) return JSON.parse(readFileSync(p, 'utf-8'))
    return null
  }

  // 规则 1：精确命中
  for (const name of combos(segments)) {
    const data = tryLoad(name)
    if (data !== null) return { file: name, data, isIdFallback: false }
  }

  // 规则 2：末段像 id（含数字）→ 去掉末段，候选 + -detail
  const last = segments[segments.length - 1] ?? ''
  if (/\d/.test(last) && segments.length >= 2) {
    const rest = segments.slice(0, -1)
    for (const name of combos(rest, '-detail')) {
      const data = tryLoad(name)
      if (data !== null) return { file: name, data, isIdFallback: true, id: last }
    }
  }
  return null
}

function readBody(req: IncomingMessage): Promise<Record<string, unknown>> {
  return new Promise((resolve) => {
    const chunks: Buffer[] = []
    req.on('data', (c: Buffer) => chunks.push(c))
    req.on('end', () => {
      const raw = Buffer.concat(chunks).toString('utf-8')
      if (!raw) return resolve({})
      try {
        resolve(JSON.parse(raw) as Record<string, unknown>)
      } catch {
        resolve({})
      }
    })
  })
}

function isPlainRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
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
          }, 100 + Math.random() * 200)
        }

        // 写操作：统一回显
        if (req.method === 'POST' || req.method === 'PUT' || req.method === 'DELETE') {
          // 登录走 mock/auth.json
          if (req.method === 'POST' && segments[0] === 'auth' && segments[1] === 'login') {
            const authPath = join(mockDir, 'auth.json')
            if (existsSync(authPath)) {
              return send(0, 'ok', JSON.parse(readFileSync(authPath, 'utf-8')))
            }
          }
          return readBody(req).then((body) => send(0, 'ok', body))
        }

        // GET
        const ctx = resolveMock(segments, mockDir)
        if (!ctx) return send(404, 'mock: not found', null, 404)

        const data = ctx.data
        if (Array.isArray(data)) {
          // 列表：分页 / keyword / status
          const page = Number(url.searchParams.get('page') ?? '1')
          const pageSize = Number(url.searchParams.get('pageSize') ?? '10')
          const keyword = (url.searchParams.get('keyword') ?? '').trim().toLowerCase()
          const status = url.searchParams.get('status')

          let list = data.slice()
          if (status) list = list.filter((it) => isPlainRecord(it) && String(it.status) === status)
          if (keyword) {
            list = list.filter((it) => JSON.stringify(it).toLowerCase().includes(keyword))
          }
          const total = list.length
          const start = (page - 1) * pageSize
          const paged = list.slice(start, start + pageSize)
          return send(0, 'ok', { list: paged, total, page, pageSize })
        }

        if (isPlainRecord(data)) {
          // 详情：数组按 id 查找已在 resolveMock 处理；对象直接返回
          if (ctx.isIdFallback) {
            const arr = (data as { list?: unknown[] }).list
            if (Array.isArray(arr) && ctx.id) {
              const hit = arr.find((it) => isPlainRecord(it) && String(it.id) === ctx.id)
              if (hit) return send(0, 'ok', hit)
              return send(404, 'mock: detail not found', null, 404)
            }
          }
          return send(0, 'ok', data)
        }
        return send(0, 'ok', data)
      })
    },
  }
}

export default defineConfig(({ mode }) => {
  // 请求链路开关：
  // - 默认（VITE_USE_MOCK 未设置或非 '1'）：/api/* 由 server.proxy 转发到真实后端（http://localhost:8080）
  // - VITE_USE_MOCK=1：由上方 mockPlugin 中间件接管，返回 ./mock/*.json（无需后端即可开发）
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
