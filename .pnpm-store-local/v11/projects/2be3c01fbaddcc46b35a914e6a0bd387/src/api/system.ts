import { http } from '@/utils/request'

/** 系统配置（邮箱 SMTP / 短信服务）API 模块。 */

/** SMTP 加密方式：tls（STARTTLS）/ ssl（SMTPS）/ none（明文） */
export type SmtpEncryption = 'tls' | 'ssl' | 'none'

/** 后端密码掩码常量（internal/application/system/service.go），GET 返回此值表示已配置密码 */
export const SMTP_PASSWORD_MASK = '******'

/** GET /system/config/smtp 返回的 SMTP 配置（password 为掩码或空串；timeout 后端未返回，表单默认 10） */
export interface SmtpConfigDto {
  /** 是否已配置（host 非空） */
  enabled: boolean
  host: string
  port: number
  from_email: string
  encryption: SmtpEncryption
  /** 已配置为 '******' 掩码，未配置为空串 */
  password: string
}

/** PUT /system/config/smtp 与 POST /system/config/smtp/test 请求体 */
export interface SmtpConfigPayload {
  host: string
  port: number
  from_email: string
  /** 留空或 '******' 掩码时，后端保留原密码 */
  password: string
  encryption: SmtpEncryption
  /** 连接超时（秒） */
  timeout: number
}

/** 读取 SMTP 配置（脱敏） */
export function getSmtpConfig() {
  return http.get<SmtpConfigDto>('/system/config/smtp')
}

/** 保存 SMTP 配置 */
export function updateSmtpConfig(payload: SmtpConfigPayload) {
  return http.put<Record<string, never>>('/system/config/smtp', payload)
}

/** 测试 SMTP 连接（不落库） */
export function testSmtpConfig(payload: SmtpConfigPayload) {
  return http.post<Record<string, never>>('/system/config/smtp/test', payload)
}

/**
 * 短信服务配置：后端尚未提供对应接口（GET/PUT /system/config/sms、POST /system/config/sms/test）。
 * 以下类型与函数按与 SMTP 对称的契约预留，后端实现后前端无需改动。
 */
export interface SmsConfigDto {
  enabled: boolean
  provider: string
  access_key: string
  secret_key: string
  sign_name: string
  template_code: string
}

export interface SmsConfigPayload {
  provider: string
  access_key: string
  secret_key: string
  sign_name: string
  template_code: string
}

export function getSmsConfig() {
  return http.get<SmsConfigDto>('/system/config/sms')
}

export function updateSmsConfig(payload: SmsConfigPayload) {
  return http.put<Record<string, never>>('/system/config/sms', payload)
}

export function testSmsConfig(payload: SmsConfigPayload) {
  return http.post<Record<string, never>>('/system/config/sms/test', payload)
}

/** 提现规则配置。 */
export interface WithdrawRuleDto {
  enabled: boolean
  /** 手续费率（0.006 = 0.6%） */
  fee_rate: number
  /** 单笔最低提现金额（元） */
  min_amount: number
  /** 单笔最高提现金额（元） */
  max_amount: number
  /** 每日提现次数上限 */
  daily_limit: number
  /** auto 自动审核 / manual 人工审核 */
  mode: 'auto' | 'manual'
  /** 到账周期（T+N 个工作日） */
  settle_days: number
}

export type WithdrawRulePayload = WithdrawRuleDto

/** 读取提现规则配置 */
export function getWithdrawRule() {
  return http.get<WithdrawRuleDto>('/system/config/withdraw-rule')
}

/** 保存提现规则配置 */
export function updateWithdrawRule(payload: WithdrawRulePayload) {
  return http.put<Record<string, never>>('/system/config/withdraw-rule', payload)
}
