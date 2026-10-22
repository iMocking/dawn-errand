/**
 * 生成 ULID 风格 26 字符请求 ID（Crockford base32）。
 * 与后端 ULID 口径一致：全大写、无易混字符。
 */
const ALPHABET = '0123456789ABCDEFGHJKMNPQRSTVWXYZ'

export function genRequestId(): string {
  let out = ''
  for (let i = 0; i < 26; i++) {
    out += ALPHABET[Math.floor(Math.random() * ALPHABET.length)]
  }
  return out
}
