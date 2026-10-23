class_name ErrandView
extends ScreenBase
## 跑腿下单视图：类型切换、起终点、物品描述、费用试算、提交按钮。

var _type_chips: Array = []
var _goods_input: TextEdit
var _submit_btn: Button
var _fee_label: Label

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new("跑腿下单", func() -> void: App.go_back())
	add_child(nav)

	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = DesignTokens.NAV_H
	sc.offset_bottom = -78.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	var root := Kit.vbox(12)
	root.offset_left = 16.0
	root.offset_right = -16.0
	sc.add_child(root)

	# 服务类型
	var type_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(type_card)
	var type_box := Kit.vbox(10)
	type_box.offset_left = 16.0
	type_box.offset_right = -16.0
	type_box.offset_top = 14.0
	type_box.offset_bottom = -14.0
	type_card.add_child(type_box)
	var t_title := Kit.label_sys("选择服务类型", 15, DesignTokens.TEXT, true)
	type_box.add_child(t_title)
	var t_row := Kit.hbox(8)
	for t in vm.list("type_options").items():
		var b := Button.new()
		b.text = str(t)
		b.custom_minimum_size = Vector2(0, 36)
		b.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		b.add_theme_font_size_override("font_size", 14)
		b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
		var tag := str(t)
		b.pressed.connect(func() -> void: vm.set_type(tag))
		t_row.add_child(b)
		_type_chips.append(b)
	type_box.add_child(t_row)

	# 起终点
	var addr_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(addr_card)
	var addr_box := Kit.vbox(0)
	addr_box.offset_left = 16.0
	addr_box.offset_right = -16.0
	addr_card.add_child(addr_box)
	addr_box.add_child(_addr_row("取货地址", Icon.Kind.LOCATION, DesignTokens.PRIMARY, "from_address", "from_contact"))
	addr_box.add_child(Kit.divider(1.0))
	addr_box.add_child(_addr_row("收货地址", Icon.Kind.LOCATION, DesignTokens.SUCCESS, "to_address", "to_contact"))

	# 物品描述
	var goods_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(goods_card)
	var g_box := Kit.vbox(10)
	g_box.offset_left = 16.0
	g_box.offset_right = -16.0
	g_box.offset_top = 14.0
	g_box.offset_bottom = -14.0
	goods_card.add_child(g_box)
	var g_title := Kit.label_sys("买什么 / 送什么", 15, DesignTokens.TEXT, true)
	g_box.add_child(g_title)
	_goods_input = TextEdit.new()
	_goods_input.placeholder_text = "例如：买 2 斤五花肉、1 瓶老抽，送到绿杨乡卫生院"
	_goods_input.custom_minimum_size = Vector2(0, 76)
	_goods_input.add_theme_font_size_override("font_size", 14)
	_goods_input.add_theme_color_override("font_color", DesignTokens.TEXT)
	_goods_input.add_theme_color_override("font_placeholder_color", DesignTokens.TEXT_FAINT)
	_goods_input.add_theme_color_override("caret_color", DesignTokens.PRIMARY)
	_goods_input.add_theme_stylebox_override("normal", Kit.style_flat(Color("F5F6F8"), 10.0))
	_goods_input.add_theme_stylebox_override("focus", Kit.style_flat(Color("F5F6F8"), 10.0, DesignTokens.PRIMARY, 1.0))
	_goods_input.scroll_fit_content_height = true
	g_box.add_child(_goods_input)
	var tag_row := Kit.hbox(8)
	tag_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	g_box.add_child(tag_row)
	for tag in vm.list("quick_tags").items():
		var b := Button.new()
		b.text = str(tag)
		b.custom_minimum_size = Vector2(0, 30)
		b.add_theme_font_size_override("font_size", 12)
		b.add_theme_color_override("font_color", DesignTokens.PRIMARY)
		b.add_theme_color_override("font_hover_color", DesignTokens.PRIMARY)
		b.add_theme_color_override("font_pressed_color", DesignTokens.PRIMARY)
		b.add_theme_stylebox_override("normal", Kit.style_flat(DesignTokens.PRIMARY_BG, 15.0))
		b.add_theme_stylebox_override("hover", Kit.style_flat(DesignTokens.PRIMARY_BG, 15.0))
		b.add_theme_stylebox_override("pressed", Kit.style_flat(DesignTokens.PRIMARY_BG, 15.0))
		b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
		var t := str(tag)
		b.pressed.connect(func() -> void: vm.append_tag(t))
		tag_row.add_child(b)

	# 费用说明
	var fee_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(fee_card)
	var f_box := Kit.vbox(8)
	f_box.offset_left = 16.0
	f_box.offset_right = -16.0
	f_box.offset_top = 14.0
	f_box.offset_bottom = -14.0
	fee_card.add_child(f_box)
	var f_title := Kit.label_sys("费用预估", 15, DesignTokens.TEXT, true)
	f_box.add_child(f_title)
	var f_row := Kit.hbox(0)
	f_box.add_child(f_row)
	var dist := Kit.label_sys("预估距离 " + Fmt.km(5.2) + " · 基础服务费 ", 13, DesignTokens.TEXT_SEC)
	f_row.add_child(dist)
	_fee_label = Kit.label_sys(Fmt.yuan(8.0), 13, DesignTokens.PRICE, true)
	f_row.add_child(_fee_label)
	var f_tip := Kit.label_sys("多退少补，实报实销；商品垫付仅按实际小票结算", 11, DesignTokens.TEXT_MUTED)
	f_box.add_child(f_tip)

	root.add_child(Kit.spacing(0, 8))

	# 底部提交
	var bottom := PanelContainer.new()
	bottom.anchor_left = 0.0
	bottom.anchor_right = 1.0
	bottom.anchor_top = 1.0
	bottom.anchor_bottom = 1.0
	bottom.offset_top = -78.0
	bottom.offset_bottom = 0.0
	bottom.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 0.0))
	add_child(bottom)
	var b_row := Kit.hbox(12)
	b_row.offset_left = 16.0
	b_row.offset_right = -16.0
	b_row.offset_top = 0.0
	b_row.offset_bottom = 0.0
	bottom.add_child(b_row)
	var price_box := Kit.vbox(0)
	price_box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	price_box.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	b_row.add_child(price_box)
	var price_label := Kit.label_sys("服务费", 11, DesignTokens.TEXT_MUTED)
	price_box.add_child(price_label)
	var total_label := Kit.label_sys(Fmt.yuan(8.0), 20, DesignTokens.PRICE, true)
	price_box.add_child(total_label)
	_submit_btn = Kit.primary_button("立即下单", Callable(), 48.0, 24.0, 15)
	_submit_btn.pressed.connect(func() -> void: vm.submit())
	b_row.add_child(_submit_btn)

func _addr_row(label: String, icon_kind: int, icon_color: Color, addr_prop: String, contact_prop: String) -> Control:
	var row := Kit.hbox(10)
	row.custom_minimum_size = Vector2(0, 78)
	var dot := PanelContainer.new()
	dot.custom_minimum_size = Vector2(26, 26)
	dot.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	dot.add_theme_stylebox_override("panel", Kit.style_flat(icon_color, 13.0))
	var ic := Icon.new(icon_kind, Color.WHITE, 1.8)
	ic.custom_minimum_size = Vector2(14, 14)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -7.0
	ic.offset_right = 7.0
	ic.offset_top = -7.0
	ic.offset_bottom = 7.0
	dot.add_child(ic)
	row.add_child(dot)
	var info := Kit.vbox(4)
	info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	info.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	row.add_child(info)
	var l := Kit.label_sys(label, 12, DesignTokens.TEXT_MUTED)
	info.add_child(l)
	var a := Kit.label_sys(str(vm.get_prop(addr_prop)), 15, DesignTokens.TEXT, true)
	info.add_child(a)
	var c := Kit.label_sys(str(vm.get_prop(contact_prop)), 11, DesignTokens.TEXT_MUTED)
	info.add_child(c)
	return row

func bind() -> void:
	hub.bind(vm, "type", func(v) -> void:
		var idx := 0
		var opts: Array = vm.list("type_options").items()
		for i in opts.size():
			if str(opts[i]) == str(v):
				idx = i
				break
		for i in _type_chips.size():
			var b: Button = _type_chips[i]
			var active := i == idx
			b.add_theme_color_override("font_color", Color.WHITE if active else DesignTokens.TEXT_SEC)
			b.add_theme_stylebox_override("normal", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 10.0))
			b.add_theme_stylebox_override("hover", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 10.0))
			b.add_theme_stylebox_override("pressed", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 10.0))
	)
	hub.bind(vm, "fee", func(v) -> void:
		_fee_label.text = Fmt.yuan(float(v))
	)
	hub.bind(vm, "goods_desc", func(v) -> void:
		pass
	)
	_goods_input.text_changed.connect(func() -> void: vm.set_goods(_goods_input.text))
