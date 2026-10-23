class_name ProfileView
extends ScreenBase
## 个人中心视图：渐变头部（真实用户）、资产栏、订单入口、菜单与退出登录。

func build() -> void:
	# 渐变头部
	var header := Kit.gradient_rect(DesignTokens.PRIMARY_LIGHT, DesignTokens.PRIMARY_DARK)
	header.anchor_right = 1.0
	header.offset_bottom = 196.0
	add_child(header)
	var sb := StatusBar.new(Color(1, 1, 1, 0.95))
	sb.set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	add_child(sb)

	# 用户信息
	var user_row := Kit.hbox(14)
	user_row.anchor_left = 0.0
	user_row.anchor_right = 1.0
	user_row.offset_left = 24.0
	user_row.offset_right = -24.0
	user_row.offset_top = 84.0
	user_row.offset_bottom = 160.0
	add_child(user_row)
	var avatar := PanelContainer.new()
	avatar.custom_minimum_size = Vector2(62, 62)
	avatar.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 31.0))
	var av_ic := Icon.new(Icon.Kind.USER, DesignTokens.PRIMARY, 2.0)
	av_ic.custom_minimum_size = Vector2(36, 36)
	av_ic.set_anchors_preset(Control.PRESET_CENTER)
	av_ic.anchor_left = 0.5
	av_ic.anchor_right = 0.5
	av_ic.anchor_top = 0.5
	av_ic.anchor_bottom = 0.5
	av_ic.offset_left = -18.0
	av_ic.offset_right = 18.0
	av_ic.offset_top = -18.0
	av_ic.offset_bottom = 18.0
	avatar.add_child(av_ic)
	user_row.add_child(avatar)
	var info := Kit.vbox(6)
	info.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	user_row.add_child(info)
	var name_row := Kit.hbox(8)
	info.add_child(name_row)
	var name_l := Kit.label_sys(str(vm.get_prop("name")), 21, Color.WHITE, true)
	name_row.add_child(name_l)
	var lv := Kit.chip(str(vm.get_prop("level")), true, Color.WHITE)
	lv.add_theme_stylebox_override("panel", Kit.style_flat(Color(1, 1, 1, 0.22), 12.0))
	name_row.add_child(lv)
	var phone_l := Kit.label_sys(vm.masked_phone(), 13, Color(1, 1, 1, 0.92))
	info.add_child(phone_l)

	# 资产栏
	var asset_card := Kit.card(Color.WHITE, 14.0, true, true)
	asset_card.anchor_left = 0.0
	asset_card.anchor_right = 1.0
	asset_card.offset_left = 20.0
	asset_card.offset_right = -20.0
	asset_card.offset_top = 178.0
	asset_card.offset_bottom = 250.0
	add_child(asset_card)
	var asset_row := Kit.hbox(0)
	asset_row.offset_left = 0.0
	asset_row.offset_right = 0.0
	asset_card.add_child(asset_row)
	asset_row.add_child(_asset("优惠券", str(vm.get_prop("coupons_count")) + " 张", DesignTokens.PRIMARY, func() -> void: vm.open_coupons()))
	asset_row.add_child(_asset("余额", "¥" + str(vm.get_prop("balance")), DesignTokens.PRICE, func() -> void: Toast.show_msg("余额充值（演示）")))
	asset_row.add_child(_asset("积分", str(vm.get_prop("points")), DesignTokens.SUCCESS, func() -> void: Toast.show_msg("积分商城（演示）")))

	# 内容区
	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = 258.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	var root := Kit.vbox(12)
	root.offset_left = 16.0
	root.offset_right = -16.0
	sc.add_child(root)

	# 我的订单入口
	var order_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(order_card)
	var o_box := Kit.vbox(0)
	o_box.offset_left = 16.0
	o_box.offset_right = -16.0
	o_box.offset_top = 14.0
	o_box.offset_bottom = -14.0
	order_card.add_child(o_box)
	var o_head := Kit.hbox(0)
	o_box.add_child(o_head)
	var o_t := Kit.label_sys("我的订单", 15, DesignTokens.TEXT, true)
	o_t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	o_head.add_child(o_t)
	var o_all := Kit.label_sys("全部订单", 12, DesignTokens.TEXT_MUTED)
	o_head.add_child(o_all)
	var o_arrow := Icon.new(Icon.Kind.CHEVRON_RIGHT, DesignTokens.TEXT_FAINT, 2.0)
	o_arrow.custom_minimum_size = Vector2(16, 16)
	o_head.add_child(o_arrow)
	o_box.add_child(Kit.spacing(0, 12))
	var o_icons := Kit.hbox(0)
	o_box.add_child(o_icons)
	o_icons.add_child(_order_entry("待付款", Icon.Kind.COIN, func() -> void: vm.open_orders()))
	o_icons.add_child(_order_entry("配送中", Icon.Kind.SCOOTER, func() -> void: vm.open_orders()))
	o_icons.add_child(_order_entry("待评价", Icon.Kind.STAR, func() -> void: vm.open_review_list()))
	o_icons.add_child(_order_entry("退款/售后", Icon.Kind.MEDAL, func() -> void: vm.open_orders()))
	var o_btn := Button.new()
	o_btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	o_btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	o_btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	o_btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	o_btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	o_btn.pressed.connect(func() -> void: vm.open_orders())
	order_card.add_child(o_btn)

	# 功能菜单
	var menu_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(menu_card)
	var m_box := Kit.vbox(0)
	m_box.offset_left = 16.0
	m_box.offset_right = -16.0
	menu_card.add_child(m_box)
	m_box.add_child(_menu_row("收货地址", Icon.Kind.LOCATION, func() -> void: vm.open_address()))
	m_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	m_box.add_child(_menu_row("跑腿下单", Icon.Kind.ERRAND, func() -> void: vm.open_errand()))
	m_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	m_box.add_child(_menu_row("联系客服", Icon.Kind.HEADPHONE, func() -> void: vm.contact_service()))
	m_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	m_box.add_child(_menu_row("帮助中心", Icon.Kind.HELP, func() -> void: vm.open_help()))
	m_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	m_box.add_child(_menu_row("设置", Icon.Kind.SETTINGS, func() -> void: vm.open_settings()))

	# 退出登录
	var logout_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(logout_card)
	var logout := Kit.label_sys("退出登录", 15, DesignTokens.PRICE, false, HORIZONTAL_ALIGNMENT_CENTER)
	logout.custom_minimum_size = Vector2(0, 46)
	logout_card.add_child(logout)
	var lo_btn := Button.new()
	lo_btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	lo_btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	lo_btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	lo_btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	lo_btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	lo_btn.pressed.connect(func() -> void: vm.logout())
	logout_card.add_child(lo_btn)
	root.add_child(Kit.spacing(0, 16))

func _asset(name: String, value: String, color: Color, cb: Callable) -> Control:
	var box := Control.new()
	box.custom_minimum_size = Vector2(0, 72)
	box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var v := Kit.label_sys(value, 16, color, true, HORIZONTAL_ALIGNMENT_CENTER)
	v.set_anchors_preset(Control.PRESET_CENTER_TOP)
	v.anchor_left = 0.0
	v.anchor_right = 1.0
	v.offset_top = 12.0
	v.offset_bottom = 40.0
	box.add_child(v)
	var n := Kit.label_sys(name, 12, DesignTokens.TEXT_MUTED, false, HORIZONTAL_ALIGNMENT_CENTER)
	n.set_anchors_preset(Control.PRESET_CENTER_TOP)
	n.anchor_left = 0.0
	n.anchor_right = 1.0
	n.offset_top = 42.0
	n.offset_bottom = 64.0
	box.add_child(n)
	var btn := Button.new()
	btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	if cb.is_valid():
		btn.pressed.connect(cb)
	box.add_child(btn)
	return box

func _order_entry(name: String, icon_kind: int, cb: Callable) -> Control:
	var box := Control.new()
	box.custom_minimum_size = Vector2(0, 72)
	box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var ic := Icon.new(icon_kind, DesignTokens.PRIMARY, 2.0)
	ic.custom_minimum_size = Vector2(26, 26)
	ic.set_anchors_preset(Control.PRESET_CENTER_TOP)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.offset_left = -13.0
	ic.offset_right = 13.0
	ic.offset_top = 6.0
	ic.offset_bottom = 32.0
	box.add_child(ic)
	var n := Kit.label_sys(name, 12, DesignTokens.TEXT_SEC, false, HORIZONTAL_ALIGNMENT_CENTER)
	n.set_anchors_preset(Control.PRESET_CENTER_TOP)
	n.anchor_left = 0.0
	n.anchor_right = 1.0
	n.offset_top = 36.0
	n.offset_bottom = 58.0
	box.add_child(n)
	var btn := Button.new()
	btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	if cb.is_valid():
		btn.pressed.connect(cb)
	box.add_child(btn)
	return box

func _menu_row(name: String, icon_kind: int, cb: Callable) -> Control:
	var row := Control.new()
	row.custom_minimum_size = Vector2(0, 50)
	var ic := Icon.new(icon_kind, DesignTokens.TEXT_SEC, 2.0)
	ic.custom_minimum_size = Vector2(20, 20)
	ic.position = Vector2(0, 15)
	row.add_child(ic)
	var n := Kit.label_sys(name, 14, DesignTokens.TEXT)
	n.position = Vector2(30, 14)
	n.custom_minimum_size = Vector2(200, 22)
	row.add_child(n)
	var arrow := Icon.new(Icon.Kind.CHEVRON_RIGHT, DesignTokens.TEXT_FAINT, 2.0)
	arrow.custom_minimum_size = Vector2(16, 16)
	arrow.anchor_left = 1.0
	arrow.anchor_right = 1.0
	arrow.anchor_top = 0.5
	arrow.anchor_bottom = 0.5
	arrow.offset_left = -22.0
	arrow.offset_right = -4.0
	arrow.offset_top = -8.0
	arrow.offset_bottom = 8.0
	row.add_child(arrow)
	var btn := Button.new()
	btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	if cb.is_valid():
		btn.pressed.connect(cb)
	row.add_child(btn)
	return row

func bind() -> void:
	pass
