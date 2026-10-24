/** 生成 ULID 风格 26 字符随机串，作为 X-Request-ID 透传头（与后端口径一致） */
export function genRequestId(): string {
  const chars = '0123456789ABCDEFGHJKMNPQRSTVWXYZ'
  let out = ''
  for (let i = 0; i < 26; i++) out += chars[Math.floor(Math.random() * chars.length)]
  return out
}
