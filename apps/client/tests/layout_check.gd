extends Control
## 布局回归检查（headless）：跑完导航流程后扫描每屏 Label，检测竖排折行（宽度 < 10px 的多字 Label）。
## 运行：godot --headless res://tests/layout_check.tscn

var _fails: int = 0

func _ready() -> void:
	var nav := Navigator.new()
	add_child(nav)
	App.navigator = nav
	_run()

func _run() -> void:
	await get_tree().process_frame
	await get_tree().process_frame
	await get_tree().process_frame

	# 逐屏独立实例化（不依赖导航），保证每屏都能检查
	var seq := [
		["splash", null],
		["login", null],
		["shell", null],
		["home", null],
		["errand", null],
		["orders", null],
		["profile", null],
		["search", null],
		["pay_success", null],
		["order_track", null],
		["review", null],
		["address", null],
		["address_edit", null],
		["coupons", null],
	]
	for s in seq:
		var name: String = s[0]
		var scr: ScreenBase = App.open_screen(name)
		if scr == null:
			print("OPEN_FAIL: " + name)
			_fails += 1
			continue
		# 等布局完成
		for i in 6:
			await get_tree().process_frame
		var cur = App.navigator.current()
		if cur and is_instance_valid(cur):
			_scan(name, cur)
		else:
			print("SCAN_FAIL: " + name)
			_fails += 1
		App.go_back()
		for i in 3:
			await get_tree().process_frame

	# merchant / order_confirm 专项（需要业务 vm 与购物车数据）
	var hv: HomeVM = App.create_vm("home")
	var m: MerchantModel = hv.list("nearby").items()[0]
	var mv: MerchantVM = MerchantVM.new(m)
	mv.add_dish(mv.list("dishes").items()[0])
	App.open_screen("merchant", mv)
	for i in 6:
		await get_tree().process_frame
	var cur_m = App.navigator.current()
	if cur_m and is_instance_valid(cur_m):
		_scan("merchant", cur_m)
	App.go_back()
	for i in 3:
		await get_tree().process_frame
	var ovm: OrderConfirmVM = OrderConfirmVM.for_cart(m, mv.cart)
	App.open_screen("order_confirm", ovm)
	for i in 6:
		await get_tree().process_frame
	var cur_o = App.navigator.current()
	if cur_o and is_instance_valid(cur_o):
		_scan("order_confirm", cur_o)
	App.go_back()
	for i in 3:
		await get_tree().process_frame

	print("LAYOUT CHECK DONE failures=" + str(_fails))
	get_tree().quit(0 if _fails == 0 else 1)

func _scan(tag: String, root: Node) -> void:
	var found := 0
	for l in root.find_children("*", "Label", true, false):
		var w: float = l.size.x
		var t: String = l.text
		if w < 10.0 and t.length() > 1:
			print("VERT[%s] %s autowrap=%d w=%.1f h=%.1f text='%s'" % [tag, l.name, l.autowrap_mode, w, l.size.y, t])
			found += 1
	for b in root.find_children("*", "Button", true, false):
		var w: float = b.size.x
		var t: String = b.text
		if w < 12.0 and t.length() > 1:
			print("VERT_BTN[%s] %s w=%.1f text='%s'" % [tag, b.name, w, t])
			found += 1
	if found == 0:
		print("OK[%s] 无竖排 Label" % tag)
