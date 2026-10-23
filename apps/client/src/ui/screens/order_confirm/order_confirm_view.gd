class_name OrderConfirmView
extends ScreenBase
## 确认订单视图：地址卡、商品明细、费用明细、支付方式、提交按钮。

var _goods_box: VBoxContainer
var _goods_total: Label
var _delivery_fee: Label
var _pack_fee: Label
var _coupon_amount: Label
var _total_label: Label
var _submit_btn: Button

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new("确认订单", func() -> void: App.go_back())
	add_child(nav)

	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = DesignTokens.NAV_H
	sc.offset_bottom = -84.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	var root := Kit.vbox(12)
	root.offset_left = 16.0
	root.offset_right = -16.0
	sc.add_child(root)

	# 地址卡
	var addr_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(addr_card)
	var addr_box := Kit.hbox(10)
	addr_box.offset_left = 16.0
	addr_box.offset_right = -16.0
	addr_box.offset_top = 14.0
	addr_box.offset_bottom = -14.0
	addr_card.add_child(addr_box)
	var loc := Icon.new(Icon.Kind.LOCATION, DesignTokens.PRIMARY, 2.0)
	loc.custom_minimum_size = Vector2(24, 24)
	loc.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	addr_box.add_child(loc)
	var addr_info := Kit.vbox(4)
	addr_info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	addr_box.add_child(addr_info)
	var who := Kit.hbox(8)
	addr_info.add_child(who)
	var n_l := Kit.label_sys(str(vm.get_prop("address_name")), 16, DesignTokens.TEXT, true)
	who.add_child(n_l)
	var p_l := Kit.label_sys(str(vm.get_prop("address_phone")), 13, DesignTokens.TEXT_SEC)
	who.add_child(p_l)
	var d_l := Kit.label_sys(str(vm.get_prop("address_detail")), 13, DesignTokens.TEXT_SEC)
	addr_info.add_child(d_l)
	var arrow := Icon.new(Icon.Kind.CHEVRON_RIGHT, DesignTokens.TEXT_FAINT, 2.0)
	arrow.custom_minimum_size = Vector2(18, 18)
	arrow.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	addr_box.add_child(arrow)

	# 商品卡
	var goods_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(goods_card)
	var g_box := Kit.vbox(0)
	g_box.offset_left = 16.0
	g_box.offset_right = -16.0
	g_box.offset_top = 12.0
	g_box.offset_bottom = -12.0
	goods_card.add_child(g_box)
	var shop_row := Kit.hbox(6)
	g_box.add_child(shop_row)
	var shop_ic := Icon.new(Icon.Kind.BAG, DesignTokens.PRIMARY, 2.0)
	shop_ic.custom_minimum_size = Vector2(18, 18)
	shop_ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	shop_row.add_child(shop_ic)
	var shop_l := Kit.label_sys(str(vm.get_prop("shop_name")), 15, DesignTokens.TEXT, true)
	shop_l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	shop_row.add_child(shop_l)
	g_box.add_child(Kit.spacing(0, 8))
	_goods_box = Kit.vbox(0)
	g_box.add_child(_goods_box)
	g_box.add_child(Kit.spacing(0, 6))
	var g_total_row := Kit.hbox(0)
	g_box.add_child(g_total_row)
	var g_t := Kit.label_sys("商品金额", 13, DesignTokens.TEXT_SEC)
	g_t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	g_total_row.add_child(g_t)
	_goods_total = Kit.label_sys("", 13, DesignTokens.TEXT)
	g_total_row.add_child(_goods_total)

	# 费用卡
	var fee_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(fee_card)
	var f_box := Kit.vbox(8)
	f_box.offset_left = 16.0
	f_box.offset_right = -16.0
	f_box.offset_top = 12.0
	f_box.offset_bottom = -12.0
	fee_card.add_child(f_box)
	f_box.add_child(_build_fee_row("配送费", func(v: float) -> String: return "¥" + Fmt.money(v), "_delivery_fee"))
	f_box.add_child(_build_fee_row("包装费", func(v: float) -> String: return "¥" + Fmt.money(v), "_pack_fee"))
	var coupon_row := _build_fee_row("优惠券", func(v: float) -> String:
		return ("-¥" + Fmt.money(v)) if v > 0 else "无可用优惠券", "_coupon_amount")
	var coupon_hint := Button.new()
	coupon_hint.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	coupon_hint.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	coupon_hint.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	coupon_hint.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	coupon_hint.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	coupon_hint.pressed.connect(func() -> void: vm.choose_coupon())
	coupon_row.add_child(coupon_hint)
	f_box.add_child(coupon_row)
	f_box.add_child(Kit.divider(1.0))
	var total_row := Kit.hbox(0)
	f_box.add_child(total_row)
	var t_l := Kit.label_sys("合计", 15, DesignTokens.TEXT, true)
	t_l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	total_row.add_child(t_l)
	_total_label = Kit.label_sys("", 18, DesignTokens.PRICE, true)
	total_row.add_child(_total_label)

	# 支付方式卡
	var pay_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(pay_card)
	var pay_row := Kit.hbox(0)
	pay_row.offset_left = 16.0
	pay_row.offset_right = -16.0
	pay_row.offset_top = 0.0
	pay_row.offset_bottom = 0.0
	pay_card.add_child(pay_row)
	var pay_ic := Icon.new(Icon.Kind.COIN, DesignTokens.PRIMARY, 2.0)
	pay_ic.custom_minimum_size = Vector2(20, 20)
	pay_ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	pay_row.add_child(pay_ic)
	var pay_l := Kit.label_sys(str(vm.get_prop("pay_method")), 14, DesignTokens.TEXT_SEC)
	pay_l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	pay_row.add_child(pay_l)
	var pay_arrow := Icon.new(Icon.Kind.CHEVRON_RIGHT, DesignTokens.TEXT_FAINT, 2.0)
	pay_arrow.custom_minimum_size = Vector2(16, 16)
	pay_arrow.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	pay_row.add_child(pay_arrow)

	root.add_child(Kit.spacing(0, 6))

	# 底部提交
	var bottom := PanelContainer.new()
	bottom.anchor_left = 0.0
	bottom.anchor_right = 1.0
	bottom.anchor_top = 1.0
	bottom.anchor_bottom = 1.0
	bottom.offset_top = -84.0
	bottom.offset_bottom = 0.0
	bottom.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 0.0))
	add_child(bottom)
	var b_row := Kit.hbox(12)
	b_row.offset_left = 16.0
	b_row.offset_right = -16.0
	bottom.add_child(b_row)
	var total_box := Kit.vbox(0)
	total_box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	total_box.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	b_row.add_child(total_box)
	var total_t := Kit.label_sys("合计", 12, DesignTokens.TEXT_MUTED)
	total_box.add_child(total_t)
	var bottom_total := Kit.label_sys("", 20, DesignTokens.PRICE, true)
	total_box.add_child(bottom_total)
	_submit_btn = Kit.primary_button("提交订单", func() -> void: vm.submit(), 48.0, 24.0, 15)
	b_row.add_child(_submit_btn)
	bottom_total.text = "¥" + Fmt.money(float(vm.get_prop("total")))

func _build_fee_row(label: String, format: Callable, member: String) -> Control:
	var row := Kit.hbox(0)
	var l := Kit.label_sys(label, 13, DesignTokens.TEXT_SEC)
	l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(l)
	var ref := Kit.label_sys("", 13, DesignTokens.TEXT_SEC)
	row.add_child(ref)
	match member:
		"_delivery_fee":
			_delivery_fee = ref
		"_pack_fee":
			_pack_fee = ref
		"_coupon_amount":
			_coupon_amount = ref
	row.set_meta("fmt", format)
	row.set_meta("ref", ref)
	return row

func _goods_line(line: Dictionary) -> Control:
	var row := Kit.hbox(0)
	row.custom_minimum_size = Vector2(0, 32)
	var name := Kit.label_sys(str(line.name), 13, DesignTokens.TEXT)
	name.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(name)
	var price := Kit.label_sys("¥" + Fmt.money(float(line.price)), 13, DesignTokens.TEXT)
	row.add_child(price)
	var qty := Kit.label_sys("×" + str(int(line.qty)), 12, DesignTokens.TEXT_MUTED)
	row.add_child(qty)
	return row

func bind() -> void:
	hub.bind(vm, "goods_lines", func(lines) -> void:
		for child in _goods_box.get_children():
			child.queue_free()
		for line in lines:
			_goods_box.add_child(_goods_line(line))
	)
	hub.bind(vm, "goods_total", func(v) -> void: _goods_total.text = "¥" + Fmt.money(float(v)))
	hub.bind(vm, "delivery_fee", func(v) -> void: _delivery_fee.text = "¥" + Fmt.money(float(v)))
	hub.bind(vm, "pack_fee", func(v) -> void: _pack_fee.text = "¥" + Fmt.money(float(v)))
	hub.bind(vm, "coupon_amount", func(v) -> void: _coupon_amount.text = ("-¥" + Fmt.money(float(v))) if float(v) > 0 else "无可用优惠券")
	hub.bind(vm, "total", func(v) -> void:
		_total_label.text = "¥" + Fmt.money(float(v))
		_submit_btn.text = "提交订单 · ¥" + Fmt.money(float(v))
	)
