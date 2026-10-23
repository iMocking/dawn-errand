extends Control
## 视觉捕获（仅开发验收用）：GUI 运行后逐屏渲染并保存 PNG 到 tests/captures。
## 运行：godot --path apps\client res://tests/visual_capture.tscn（非 headless）

var _failures: int = 0

func _ready() -> void:
	# 模拟 main.gd：创建导航器
	var nav := Navigator.new()
	add_child(nav)
	App.navigator = nav
	_run()

func _run() -> void:
	await get_tree().create_timer(0.6).timeout

	# 启动页（未登录态之前先截：注入会话前直接打开 splash）
	App.open_screen("splash")
	await get_tree().create_timer(0.5).timeout
	_shot("00_splash")

	# 注入假会话 → splash 跳转主界面
	App.session.apply_login({
		"token": "smoke-access-token",
		"refresh_token": "smoke-refresh-token",
		"user": {"id": "u-test-001", "name": "测试用户", "phone": "13888888888", "role": "user"},
		"permissions": [],
	})
	await get_tree().create_timer(2.0).timeout
	if not (App.navigator.current() is ShellView):
		print("WARN: 未进入主界面外壳")

	var shell: ShellView = App.navigator.current()
	await get_tree().create_timer(0.4).timeout
	_shot("01_home")

	shell.switch_to(1)
	await get_tree().create_timer(0.5).timeout
	_shot("02_errand")

	shell.switch_to(2)
	await get_tree().create_timer(0.5).timeout
	_shot("03_orders")

	shell.switch_to(3)
	await get_tree().create_timer(0.5).timeout
	_shot("04_profile")

	shell.switch_to(0)
	await get_tree().create_timer(0.4).timeout

	# 二级屏（带真实业务数据）
	App.open_screen("search")
	await get_tree().create_timer(0.5).timeout
	_shot("05_search")
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	var home_vm: HomeVM = App.create_vm("home")
	var merchant: MerchantModel = home_vm.list("nearby").items()[0]
	var mv: MerchantVM = MerchantVM.new(merchant)
	mv.add_dish(mv.list("dishes").items()[0])
	mv.add_dish(mv.list("dishes").items()[0])
	App.open_screen("merchant", mv)
	await get_tree().create_timer(0.5).timeout
	_shot("06_merchant")

	var ovm: OrderConfirmVM = OrderConfirmVM.for_cart(merchant, mv.cart)
	App.open_screen("order_confirm", ovm)
	await get_tree().create_timer(0.5).timeout
	_shot("07_order_confirm")
	App.go_back()
	await get_tree().create_timer(0.3).timeout
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	App.open_screen("pay_success")
	await get_tree().create_timer(0.5).timeout
	_shot("08_pay_success")
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	App.open_screen("order_track")
	await get_tree().create_timer(0.5).timeout
	_shot("09_order_track")
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	App.open_screen("review")
	await get_tree().create_timer(0.5).timeout
	_shot("10_review")
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	App.open_screen("address")
	await get_tree().create_timer(0.5).timeout
	_shot("11_address")
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	App.open_screen("address_edit")
	await get_tree().create_timer(0.5).timeout
	_shot("12_address_edit")
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	App.open_screen("coupons")
	await get_tree().create_timer(0.5).timeout
	_shot("13_coupons")
	App.go_back()
	await get_tree().create_timer(0.3).timeout

	# 登出 → 登录页
	App.session.clear()
	App.replace_screen("login")
	await get_tree().create_timer(0.5).timeout
	_shot("14_login")

	print("CAPTURE DONE failures=" + str(_failures))
	get_tree().quit(0 if _failures == 0 else 1)

func _shot(name: String) -> void:
	# 打印标记并停留，由外部 PrintWindow 抓取真实窗口画面
	print("SHOT: " + name)
	await get_tree().create_timer(6.0).timeout
