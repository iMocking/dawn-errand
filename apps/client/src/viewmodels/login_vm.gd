class_name LoginVM
extends ViewModel
## 登录页 VM：双模式（手机号+验证码 / 邮件或手机号+密码），
## 对接真实后端认证接口；验证码发送带倒计时，dev 环境回显验证码。

var _phone := "13888888888"
var _code := ""
var _account := ""
var _password := ""
var _agree := true

func _init() -> void:
	obs("mode", "sms")          # "sms" | "password"
	obs("phone", _phone)
	obs("code", _code)
	obs("account", _account)
	obs("password", _password)
	obs("agree", _agree)
	obs("loading", false)
	obs("btn_text", "登录 / 注册")
	obs("countdown", 0)

func set_mode(m: String) -> void:
	set_prop("mode", m)
	set_prop("btn_text", "登录" if m == "password" else "登录 / 注册")

func set_phone(v: String) -> void:
	set_prop("phone", v)

func set_code(v: String) -> void:
	set_prop("code", v)

func set_account(v: String) -> void:
	set_prop("account", v)

func set_password(v: String) -> void:
	set_prop("password", v)

func toggle_agree() -> void:
	set_prop("agree", not get_prop("agree"))

func _valid_phone(p: String) -> bool:
	if p.length() != 11 or not p.is_valid_int():
		return false
	if not p.begins_with("1"):
		return false
	var second := int(p.substr(1, 1))
	return second >= 3 and second <= 9

## 发送验证码
func send_code() -> void:
	if get_prop("loading"):
		return
	var phone := str(get_prop("phone")).strip_edges()
	if not _valid_phone(phone):
		Toast.show_msg("请输入正确的手机号")
		return
	var res: Dictionary = await App.auth.send_sms_code(phone)
	if res.ok:
		_start_countdown(60)
		var dev := str(res.body.get("dev_code", "")) if res.body is Dictionary else ""
		if dev != "":
			Toast.show_msg("验证码已发送（开发模式：" + dev + "）")
		else:
			Toast.show_msg("验证码已发送")
	else:
		Toast.show_error(res)

func _start_countdown(seconds: int) -> void:
	set_prop("countdown", seconds)
	var tree: SceneTree = Engine.get_main_loop()
	for i in range(seconds - 1, 0, -1):
		await tree.create_timer(1.0).timeout
		set_prop("countdown", i)
	set_prop("countdown", 0)

## 登录（按当前模式走对应接口）
func login() -> void:
	if get_prop("loading"):
		return
	if not get_prop("agree"):
		Toast.show_msg("请先阅读并同意《用户协议》与《隐私政策》")
		return
	set_prop("loading", true)
	var res: Dictionary = {}
	var mode := str(get_prop("mode"))
	if mode == "sms":
		var phone := str(get_prop("phone")).strip_edges()
		var code := str(get_prop("code")).strip_edges()
		if not _valid_phone(phone):
			Toast.show_msg("请输入正确的手机号")
			set_prop("loading", false)
			return
		if code.length() != 6:
			Toast.show_msg("请输入 6 位验证码")
			set_prop("loading", false)
			return
		res = await App.auth.login_sms(phone, code)
	else:
		var account := str(get_prop("account")).strip_edges()
		var password := str(get_prop("password"))
		if account == "" or password.length() < 6:
			Toast.show_msg("请输入账号与不少于 6 位的密码")
			set_prop("loading", false)
			return
		res = await App.auth.login_password(account, password)
	set_prop("loading", false)
	if App.auth.commit_login(res):
		App.replace_screen("shell")
	else:
		Toast.show_error(res)

## 第三方登录（设计稿占位，暂未开放）
func login_third(name: String) -> void:
	Toast.show_msg(name + "登录暂未开放，请使用验证码或密码登录")
