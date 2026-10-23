class_name PaySuccessView
extends ScreenBase
## 支付成功视图：成功图标 + 订单信息卡 + 查看订单/返回首页。

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	# 成功图标区
	var top := Control.new()
	top.anchor_left = 0.0
	top.anchor_right = 1.0
	top.offset_top = DesignTokens.NAV_H + 20.0
	top.offset_bottom = DesignTokens.NAV_H + 180.0
	add_child(top)
	var ring := Control.new()
	ring.custom_minimum_size = Vector2(96, 96)
	ring.set_anchors_preset(Control.PRESET_CENTER)
	ring.anchor_left = 0.5
	ring.anchor_right = 0.5
	ring.anchor_top = 0.0
	ring.anchor_bottom = 1.0
	ring.offset_left = -48.0
	ring.offset_right = 48.0
	ring.offset_top = 0.0
	ring.offset_bottom = -40.0
	top.add_child(ring)
	ring.draw.connect(func() -> void:
		ring.draw_circle(Vector2(48, 48), 48.0, DesignTokens.SUCCESS_BG)
		ring.draw_circle(Vector2(48, 48), 30.0, DesignTokens.SUCCESS)
	)
	var check := Icon.new(Icon.Kind.CHECK, Color.WHITE, 3.0)
	check.custom_minimum_size = Vector2(36, 36)
	check.set_anchors_preset(Control.PRESET_CENTER)
	check.anchor_left = 0.5
	check.anchor_right = 0.5
	check.anchor_top = 0.5
	check.anchor_bottom = 0.5
	check.offset_left = -18.0
	check.offset_right = 18.0
	check.offset_top = -18.0
	check.offset_bottom = 18.0
	ring.add_child(check)
	var title := Kit.label_sys("下单成功", 24, DesignTokens.TEXT, true, HORIZONTAL_ALIGNMENT_CENTER)
	title.set_anchors_preset(Control.PRESET_CENTER_TOP)
	title.anchor_left = 0.0
	title.anchor_right = 1.0
	title.offset_top = 96.0
	title.offset_bottom = 132.0
	top.add_child(title)
	var sub := Kit.label_sys("骑手正在前往商家取货", 13, DesignTokens.TEXT_MUTED, false, HORIZONTAL_ALIGNMENT_CENTER)
	sub.set_anchors_preset(Control.PRESET_CENTER_TOP)
	sub.anchor_left = 0.0
	sub.anchor_right = 1.0
	sub.offset_top = 136.0
	sub.offset_bottom = 162.0
	top.add_child(sub)

	# 订单信息卡
	var card := Kit.card(Color.WHITE, 14.0, false, true)
	card.anchor_left = 0.0
	card.anchor_right = 1.0
	card.offset_left = 20.0
	card.offset_right = -20.0
	card.offset_top = DesignTokens.NAV_H + 196.0
	card.offset_bottom = DesignTokens.NAV_H + 372.0
	add_child(card)
	var box := Kit.vbox(0)
	box.offset_left = 16.0
	box.offset_right = -16.0
	box.offset_top = 14.0
	box.offset_bottom = -14.0
	card.add_child(box)
	box.add_child(_info_row("商家", str(vm.get_prop("shop")), true))
	box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	box.add_child(_info_row("订单编号", str(vm.get_prop("order_no")), false))
	box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	box.add_child(_info_row("实付金额", str(vm.get_prop("total")), false))
	box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	box.add_child(_info_row("预计送达", str(vm.get_prop("eta")), false))
	box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	box.add_child(_info_row("收货人", str(vm.get_prop("address_name")) + " " + str(vm.get_prop("address_detail")), false))

	# 按钮
	var btn_row := Kit.hbox(12)
	btn_row.anchor_left = 0.0
	btn_row.anchor_right = 1.0
	btn_row.anchor_top = 1.0
	btn_row.anchor_bottom = 1.0
	btn_row.offset_left = 20.0
	btn_row.offset_right = -20.0
	btn_row.offset_top = -88.0
	btn_row.offset_bottom = -36.0
	add_child(btn_row)
	var home := Kit.outline_button("返回首页", func() -> void: vm.go_home(), 48.0, 15, DesignTokens.TEXT_SEC, 24.0)
	home.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	btn_row.add_child(home)
	var view := Kit.primary_button("查看订单", func() -> void: vm.go_orders(), 48.0, 24.0, 15)
	view.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	btn_row.add_child(view)

func _info_row(label: String, value: String, bold_value: bool) -> Control:
	var row := Kit.hbox(0)
	row.custom_minimum_size = Vector2(0, 30)
	var l := Kit.label_sys(label, 13, DesignTokens.TEXT_MUTED)
	l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(l)
	var v := Kit.label_sys(value, 13, DesignTokens.TEXT if not bold_value else DesignTokens.PRICE, bold_value, HORIZONTAL_ALIGNMENT_RIGHT)
	v.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	v.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(v)
	return row

func bind() -> void:
	pass
