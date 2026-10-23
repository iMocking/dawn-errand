class_name AuthService
extends RefCounted
## 认证服务：对接后端 /api/auth 接口（真实 HTTP）。
## 支持：邮件/密码登录、手机号+验证码登录（含发送验证码）、me、登出。

## 邮件/密码登录：POST /api/auth/login {account, password}
## account 可为邮箱或手机号（后端同时支持）。
func login_password(account: String, password: String) -> Dictionary:
	return await Api.request(Api.METHOD_POST, "/api/auth/login", {
		"account": account,
		"password": password,
	})

## 发送手机验证码：POST /api/auth/sms/send {phone}
## dev 环境后端回显 dev_code 便于联调。
func send_sms_code(phone: String) -> Dictionary:
	return await Api.request(Api.METHOD_POST, "/api/auth/sms/send", {
		"phone": phone,
	}, false)

## 手机号+验证码登录：POST /api/auth/sms/login {phone, code}
## 手机号未注册时后端自动创建账号（自动注册）。
func login_sms(phone: String, code: String) -> Dictionary:
	return await Api.request(Api.METHOD_POST, "/api/auth/sms/login", {
		"phone": phone,
		"code": code,
	}, false)

## 当前用户信息：GET /api/auth/me
func fetch_me() -> Dictionary:
	return await Api.request(Api.METHOD_GET, "/api/auth/me")

## 登出：POST /api/auth/logout（后端升级会话版本使全部令牌失效）。
func logout() -> Dictionary:
	var res: Dictionary = await Api.request(Api.METHOD_POST, "/api/auth/logout", {
		"refresh_token": App.session.refresh_token,
	}, true, false)
	App.session.clear()
	return res

## 登录成功后统一写入会话。
func commit_login(res: Dictionary) -> bool:
	if not res.ok:
		return false
	App.session.apply_login(res.body if res.body is Dictionary else {})
	return true
