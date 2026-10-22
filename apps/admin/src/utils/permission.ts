/**
 * 统一权限判断。
 * - permissions 含 '*' 视为全量；
 * - 其他按精确串匹配，如 'merchant:audit'。
 */
export function hasPermission(permissions: string[], code: string | string[] | undefined): boolean {
  if (!code) return true
  if (permissions.includes('*')) return true
  const list = Array.isArray(code) ? code : [code]
  return list.some((c) => permissions.includes(c))
}
