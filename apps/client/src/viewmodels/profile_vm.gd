class_name ProfileVM
extends ViewModel
## 个人中心 VM：真实用户信息（会话）+ 资产/菜单（演示数据源）。
## 登出调用真实后端 /auth/logout。

func _init() -> void:
	obs("name", App.session.user_name())
	obs("phone", App.session.user_phone())
	obs("level", "Lv.1 新用户")
	obs("coupons_count", 2)
	obs("balance", "12.50")
	obs("points", "320")

func masked_phone() -> String:
	return Fmt.mask_phone(App.session.user_phone())

func open_orders() -> void:
	App.go_shell_tab(2)

func open_coupons() -> void:
	App.open_screen("coupons")

func open_address() -> void:
	App.open_screen("address")

func open_errand() -> void:
	App.open_screen("errand")

func open_review_list() -> void:
	Toast.show_msg("演示环境：暂无评价列表接口")

func contact_service() -> void:
	Toast.show_msg("客服热线：400-800-1234")

func open_help() -> void:
	Toast.show_msg("帮助中心（演示）")

func open_settings() -> void:
	Toast.show_msg("设置（演示）")

func logout() -> void:
	var res: Dictionary = await App.auth.logout()
	if res.ok:
		Toast.show_msg("已退出登录")
	else:
		Toast.show_msg("已退出登录")
	App.replace_screen("login")
