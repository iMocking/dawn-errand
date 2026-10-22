import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'
import type { AuthUser } from '@/api/auth'
import { hasPermission as checkPerm } from '@/utils/permission'

const TOKEN_KEY = 'admin_token'
const USER_KEY = 'admin_user'
const PERM_KEY = 'admin_permissions'

function readUser(): AuthUser | null {
  const raw = localStorage.getItem(USER_KEY)
  return raw ? (JSON.parse(raw) as AuthUser) : null
}
function readPerms(): string[] {
  const raw = localStorage.getItem(PERM_KEY)
  return raw ? (JSON.parse(raw) as string[]) : []
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref<AuthUser | null>(readUser())
  const permissions = ref<string[]>(readPerms())

  const isLoggedIn = computed(() => !!token.value)

  async function login(account: string, password: string, captcha: string) {
    const res = await authApi.login({ account, password, captcha })
    token.value = res.token
    user.value = res.user
    permissions.value = res.permissions
    localStorage.setItem(TOKEN_KEY, res.token)
    localStorage.setItem(USER_KEY, JSON.stringify(res.user))
    localStorage.setItem(PERM_KEY, JSON.stringify(res.permissions))
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // 忽略登出接口异常，前端仍清理会话
    }
    reset()
  }

  /** 清理会话与持久化（不调用接口） */
  function reset() {
    token.value = ''
    user.value = null
    permissions.value = []
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    localStorage.removeItem(PERM_KEY)
  }

  function hasPerm(code: string | string[] | undefined): boolean {
    return checkPerm(permissions.value, code)
  }

  return { token, user, permissions, isLoggedIn, login, logout, reset, hasPerm }
})
