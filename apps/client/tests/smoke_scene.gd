extends Control
## 冒烟测试（仅开发验证用）：
## 注入假会话 → 启动页跳主界面 → 切换 4 个 Tab → 逐屏压栈/返回 →
## 登出回登录页。任何脚本/运行时错误会输出到控制台并置失败计数。
## 运行：godot --headless res://tests/smoke_scene.tscn

var _failures: int = 0

func _ready() -> void:
	# 模拟 main.gd：创建导航器
	var nav := Navigator.new()
	add_child(nav)
	App.navigator = nav
	_run()

func _run() -> void:
	await get_tree().create_timer(0.5).timeout

	# 1) 注入假会话（测试后清除）
	App.session.apply_login({
		"token": "smoke-access-token",
		"refresh_token": "smoke-refresh-token",
		"user": {"id": "u-test-001", "name": "测试用户", "phone": "13888888888", "role": "user"},
		"permissions": [],
	})

	# 2) 真实流程：启动页 →（已登录）→ 主界面外壳
	App.open_screen("splash")
	await get_tree().create_timer(2.4).timeout
	_check(App.navigator.current() is ShellView, "启动页已登录跳转主界面外壳")

	# 3) 切换四个 Tab（懒加载各 Tab 屏）
	var shell: ShellView = App.navigator.current()
	shell.switch_to(1)
	await get_tree().create_timer(0.35).timeout
	shell.switch_to(2)
	await get_tree().create_timer(0.35).timeout
	shell.switch_to(3)
	await get_tree().create_timer(0.35).timeout
	shell.switch_to(0)
	await get_tree().create_timer(0.35).timeout
	_check(App.navigator.stack_size() == 1 and App.navigator.current() is ShellView, "Tab 切换（首页/跑腿/订单/我的）无异常")

	# 4) 逐屏压栈（App.create_vm 构造默认 VM）
	for name in ["search", "merchant", "order_confirm", "pay_success", "order_track", "review", "address", "address_edit", "coupons", "errand"]:
		var before := App.navigator.stack_size()
		App.open_screen(name)
		await get_tree().create_timer(0.4).timeout
		_check(App.navigator.stack_size() == before + 1, "压栈 " + name)
		App.go_back()
		await get_tree().create_timer(0.25).timeout
		_check(App.navigator.stack_size() == before, "返回 " + name)

	# 5) 业务跳转：首页商家 → 商家详情 → 确认订单
	var home_vm: HomeVM = App.create_vm("home")
	var merchant: MerchantModel = home_vm.list("nearby").items()[0]
	var mv: MerchantVM = MerchantVM.new(merchant)
	mv.add_dish(mv.list("dishes").items()[0])
	mv.add_dish(mv.list("dishes").items()[0])
	_check(mv.get_prop("cart_count") == 2 and float(mv.get_prop("cart_total")) > 0, "商家购物车数据驱动")
	var ovm: OrderConfirmVM = OrderConfirmVM.for_cart(merchant, mv.cart)
	_check(float(ovm.get_prop("total")) > 0 and (ovm.get_prop("goods_lines") as Array).size() > 0, "确认订单金额明细")
	var evm: ErrandVM = App.create_vm("errand")
	evm.set_goods("2斤五花肉")
	evm.set_type("帮我送")
	_check(str(evm.get_prop("type")) == "帮我送" and float(evm.get_prop("fee")) == 6.0, "跑腿类型/费用联动")

	# 6) 登出 → 回到登录页
	var profile_vm: ProfileVM = App.create_vm("profile")
	_check(str(profile_vm.get_prop("name")) == "测试用户", "个人中心读取真实会话用户")
	App.session.clear()
	App.replace_screen("login")
	await get_tree().create_timer(0.4).timeout
	_check(App.navigator.current() is LoginView, "登出回到登录页")

	print("SMOKE RESULT: failures=" + str(_failures))
	get_tree().quit(0 if _failures == 0 else 1)

func _check(cond: bool, label: String) -> void:
	if cond:
		print("PASS: " + label)
	else:
		_failures += 1
		print("FAIL: " + label)
