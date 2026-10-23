class_name Fmt
extends RefCounted
## 展示格式化工具（金额/脱敏/时间）。

## 金额：1234.5 -> "1234.50"（两位小数，接收 Variant 避免静态调用类型转换缺陷）
static func money(v: Variant) -> String:
	return String.num(float(v), 2)

## 带 ¥ 前缀的金额
static func yuan(v: Variant) -> String:
	return "¥" + money(v)

## 手机号脱敏：13888888888 -> 138****8888
static func mask_phone(p: String) -> String:
	if p.length() < 7:
		return p
	return p.substr(0, 3) + "****" + p.right(4)

## 邮箱脱敏：a***@example.com
static func mask_email(e: String) -> String:
	var at := e.find("@")
	if at <= 1:
		return e
	return e.substr(0, 1) + "***" + e.substr(at)

## int64 秒（UTC）-> "10-22 15:10" 本地时间
static func fmt_ts(ts: int) -> String:
	if ts <= 0:
		return "-"
	var dt := Time.get_datetime_dict_from_unix_time(ts)
	return "%02d-%02d %02d:%02d" % [dt.month, dt.day, dt.hour, dt.minute]

## 距离
static func km(km: float) -> String:
	return String.num(km, 1) + "km"
