class_name MerchantView
extends ScreenBase
## 商家详情视图：渐变横幅、店铺信息、分类、菜品列表、购物车栏与弹层。

var _count_label: Label
var _total_label: Label
var _cart_panel: Control
var _cart_lines: VBoxContainer
var _cart_clear_btn: Button
var _dish_box: VBoxContainer

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new(str(vm.get_prop("name")), func() -> void: App.go_back())
	add_child(nav)

	# 渐变横幅
	var banner := GradientCard.new(DesignTokens.PRIMARY_LIGHT, DesignTokens.PRIMARY_DARK, 0.0, Color.WHITE)
	banner.anchor_left = 0.0
	banner.anchor_right = 1.0
	banner.offset_top = DesignTokens.NAV_H
	banner.offset_bottom = DesignTokens.NAV_H + 120.0
	add_child(banner)
	var shop_ic := Icon.new(vm.merchant.icon, Color.WHITE, 2.0)
	shop_ic.custom_minimum_size = Vector2(44, 44)
	shop_ic.position = Vector2(20, 28)
	banner.add_child(shop_ic)
	var name_l := Kit.label_sys(str(vm.get_prop("name")), 19, Color.WHITE, true)
	name_l.position = Vector2(78, 24)
	name_l.custom_minimum_size = Vector2(280, 30)
	banner.add_child(name_l)
	var meta_l := Kit.label_sys(str(vm.get_prop("meta")), 12, Color(1, 1, 1, 0.92))
	meta_l.position = Vector2(78, 60)
	meta_l.custom_minimum_size = Vector2(280, 22)
	banner.add_child(meta_l)
	var tag_row := Kit.hbox(8)
	tag_row.position = Vector2(78, 88)
	banner.add_child(tag_row)
	for tag in vm.merchant.tags:
		tag_row.add_child(Kit.chip(str(tag), true, Color.WHITE))

	# 分类横滑
	var cat_scroll := ScrollContainer.new()
	cat_scroll.anchor_left = 0.0
	cat_scroll.anchor_right = 1.0
	cat_scroll.offset_top = DesignTokens.NAV_H + 128.0
	cat_scroll.offset_bottom = DesignTokens.NAV_H + 168.0
	cat_scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_AUTO
	add_child(cat_scroll)
	var cat_row := Kit.hbox(8)
	cat_row.offset_left = 16.0
	cat_row.offset_right = -16.0
	cat_scroll.add_child(cat_row)
	for c in vm.list("categories").items():
		cat_row.add_child(Kit.chip(str(c), true))

	# 菜品列表
	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = DesignTokens.NAV_H + 172.0
	sc.offset_bottom = -76.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	_dish_box = Kit.vbox(10)
	_dish_box.offset_left = 16.0
	_dish_box.offset_right = -16.0
	_dish_box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sc.add_child(_dish_box)

	# 底部购物车栏
	var bottom := PanelContainer.new()
	bottom.anchor_left = 0.0
	bottom.anchor_right = 1.0
	bottom.anchor_top = 1.0
	bottom.anchor_bottom = 1.0
	bottom.offset_top = -76.0
	bottom.offset_bottom = 0.0
	bottom.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 0.0))
	add_child(bottom)
	var b_row := Kit.hbox(12)
	b_row.offset_left = 16.0
	b_row.offset_right = -16.0
	bottom.add_child(b_row)
	var cart_btn := Control.new()
	cart_btn.custom_minimum_size = Vector2(52, 52)
	cart_btn.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	b_row.add_child(cart_btn)
	var cart_bg := GradientCard.new(DesignTokens.PRIMARY_LIGHT, DesignTokens.PRIMARY_DARK, 26.0, Color.WHITE)
	cart_bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	cart_btn.add_child(cart_bg)
	var cart_ic := Icon.new(Icon.Kind.CART, Color.WHITE, 2.0)
	cart_ic.custom_minimum_size = Vector2(26, 26)
	cart_ic.set_anchors_preset(Control.PRESET_CENTER)
	cart_ic.anchor_left = 0.5
	cart_ic.anchor_right = 0.5
	cart_ic.anchor_top = 0.5
	cart_ic.anchor_bottom = 0.5
	cart_ic.offset_left = -13.0
	cart_ic.offset_right = 13.0
	cart_ic.offset_top = -13.0
	cart_ic.offset_bottom = 13.0
	cart_btn.add_child(cart_ic)
	_count_label = Kit.label_sys("0", 10, Color.WHITE, true, HORIZONTAL_ALIGNMENT_CENTER)
	_count_label.anchor_left = 1.0
	_count_label.anchor_right = 1.0
	_count_label.offset_left = -14.0
	_count_label.offset_right = 10.0
	_count_label.offset_top = -6.0
	_count_label.offset_bottom = 8.0
	_count_label.add_theme_stylebox_override("normal", Kit.style_flat(DesignTokens.PRICE, 9.0))
	cart_btn.add_child(_count_label)
	var cart_hit := Button.new()
	cart_hit.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	cart_hit.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	cart_hit.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	cart_hit.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	cart_hit.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	cart_hit.pressed.connect(func() -> void: vm.toggle_cart())
	cart_btn.add_child(cart_hit)
	var price_box := Kit.vbox(0)
	price_box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	price_box.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	b_row.add_child(price_box)
	var tip_l := Kit.label_sys("另需配送费 ¥3.00", 11, DesignTokens.TEXT_MUTED)
	price_box.add_child(tip_l)
	_total_label = Kit.label_sys("¥0.00", 19, DesignTokens.PRICE, true)
	price_box.add_child(_total_label)
	var checkout := Kit.primary_button("去结算", func() -> void: vm.go_confirm(), 44.0, 22.0, 15)
	b_row.add_child(checkout)

	# 购物车弹层
	_cart_panel = Control.new()
	_cart_panel.visible = false
	_cart_panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(_cart_panel)
	var dim := ColorRect.new()
	dim.color = Color(0, 0, 0, 0.45)
	dim.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_cart_panel.add_child(dim)
	var dim_hit := Button.new()
	dim_hit.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	dim_hit.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	dim_hit.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	dim_hit.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	dim_hit.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	dim_hit.pressed.connect(func() -> void: vm.close_cart())
	_cart_panel.add_child(dim_hit)
	var sheet := Kit.card(Color.WHITE, 24.0)
	sheet.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
	sheet.anchor_left = 0.0
	sheet.anchor_right = 1.0
	sheet.anchor_top = 1.0
	sheet.anchor_bottom = 1.0
	sheet.offset_top = -360.0
	sheet.offset_bottom = 0.0
	_cart_panel.add_child(sheet)
	var s_head := Kit.hbox(0)
	s_head.set_anchors_preset(Control.PRESET_TOP_WIDE)
	s_head.offset_left = 20.0
	s_head.offset_right = -20.0
	s_head.offset_top = 16.0
	s_head.offset_bottom = 48.0
	sheet.add_child(s_head)
	var s_t := Kit.label_sys("已选商品", 16, DesignTokens.TEXT, true)
	s_t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	s_head.add_child(s_t)
	_cart_clear_btn = Button.new()
	_cart_clear_btn.text = "清空"
	_cart_clear_btn.add_theme_font_size_override("font_size", 13)
	_cart_clear_btn.add_theme_color_override("font_color", DesignTokens.TEXT_MUTED)
	_cart_clear_btn.add_theme_color_override("font_hover_color", DesignTokens.PRICE)
	_cart_clear_btn.add_theme_color_override("font_pressed_color", DesignTokens.PRICE)
	_cart_clear_btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	_cart_clear_btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	_cart_clear_btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	_cart_clear_btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	_cart_clear_btn.pressed.connect(func() -> void: vm.clear_cart())
	s_head.add_child(_cart_clear_btn)
	var s_scroll := ScrollContainer.new()
	s_scroll.set_anchors_preset(Control.PRESET_FULL_RECT)
	s_scroll.offset_top = 52.0
	s_scroll.offset_bottom = -64.0
	sheet.add_child(s_scroll)
	_cart_lines = Kit.vbox(0)
	_cart_lines.offset_left = 20.0
	_cart_lines.offset_right = -20.0
	_cart_lines.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	s_scroll.add_child(_cart_lines)
	var s_btn := Kit.primary_button("去结算", func() -> void: vm.go_confirm(), 46.0, 23.0, 15)
	s_btn.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
	s_btn.anchor_left = 0.0
	s_btn.anchor_right = 1.0
	s_btn.anchor_top = 1.0
	s_btn.anchor_bottom = 1.0
	s_btn.offset_left = 20.0
	s_btn.offset_right = -20.0
	s_btn.offset_top = -54.0
	s_btn.offset_bottom = -10.0
	sheet.add_child(s_btn)

func _dish_row(d: DishModel) -> Control:
	var card := Kit.card(Color.WHITE, 12.0, false, true)
	var row := Kit.hbox(12)
	row.offset_left = 14.0
	row.offset_right = -14.0
	row.offset_top = 0.0
	row.offset_bottom = 0.0
	card.add_child(row)
	var icon_box := PanelContainer.new()
	icon_box.custom_minimum_size = Vector2(56, 56)
	icon_box.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	icon_box.add_theme_stylebox_override("panel", Kit.style_flat(d.icon_color, 12.0))
	var ic := Icon.new(d.icon, Color.WHITE, 2.0)
	ic.custom_minimum_size = Vector2(28, 28)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -14.0
	ic.offset_right = 14.0
	ic.offset_top = -14.0
	ic.offset_bottom = 14.0
	icon_box.add_child(ic)
	row.add_child(icon_box)
	var info := Kit.vbox(5)
	info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	info.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	row.add_child(info)
	var n := Kit.label_sys(d.name, 15, DesignTokens.TEXT, true)
	info.add_child(n)
	var desc := Kit.label_sys(d.desc, 12, DesignTokens.TEXT_MUTED)
	desc.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	info.add_child(desc)
	var price_row := Kit.hbox(8)
	info.add_child(price_row)
	var price := Kit.label_sys("¥" + d.price_text(), 16, DesignTokens.PRICE, true)
	price_row.add_child(price)
	var sales := Kit.label_sys(d.sales_text(), 11, DesignTokens.TEXT_FAINT)
	price_row.add_child(sales)
	var add_btn := Button.new()
	add_btn.text = "＋"
	add_btn.custom_minimum_size = Vector2(30, 30)
	add_btn.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	add_btn.add_theme_font_size_override("font_size", 18)
	add_btn.add_theme_color_override("font_color", Color.WHITE)
	add_btn.add_theme_color_override("font_hover_color", Color.WHITE)
	add_btn.add_theme_color_override("font_pressed_color", Color.WHITE)
	add_btn.add_theme_stylebox_override("normal", Kit.style_flat(DesignTokens.PRIMARY, 15.0))
	add_btn.add_theme_stylebox_override("hover", Kit.style_flat(DesignTokens.PRIMARY_DARK, 15.0))
	add_btn.add_theme_stylebox_override("pressed", Kit.style_flat(DesignTokens.PRIMARY_DARK, 15.0))
	add_btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	var dish := d
	add_btn.pressed.connect(func() -> void: vm.add_dish(dish))
	row.add_child(add_btn)
	return card

func _cart_line(line: Dictionary) -> Control:
	var row := Kit.hbox(10)
	row.custom_minimum_size = Vector2(0, 46)
	var n := Kit.label_sys(str(line.dish.name), 14, DesignTokens.TEXT)
	n.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(n)
	var minus := Kit.icon_button(Icon.Kind.MINUS, Callable(), Color("F5F6F8"), DesignTokens.TEXT_SEC, 28.0)
	minus.pressed.connect(func() -> void: vm.remove_dish(line.dish))
	row.add_child(minus)
	var qty := Kit.label_sys("×" + str(int(line.qty)), 13, DesignTokens.TEXT)
	row.add_child(qty)
	var plus := Kit.icon_button(Icon.Kind.PLUS, Callable(), Color("F5F6F8"), DesignTokens.PRIMARY, 28.0)
	plus.pressed.connect(func() -> void: vm.add_dish(line.dish))
	row.add_child(plus)
	var price := Kit.label_sys("¥" + Fmt.money(float(line.dish.price) * float(line.qty)), 14, DesignTokens.PRICE, true)
	row.add_child(price)
	return row

func bind() -> void:
	hub.bind(vm, "cart_count", func(v) -> void:
		_count_label.text = str(int(v))
		_count_label.visible = int(v) > 0
	)
	hub.bind(vm, "cart_total", func(v) -> void:
		_total_label.text = "¥" + Fmt.money(float(v))
	)
	hub.bind(vm, "cart_open", func(v) -> void:
		_cart_panel.visible = bool(v)
	)
	hub.bind_list(vm, "dishes", func(items: Array) -> void:
		for child in _dish_box.get_children():
			child.queue_free()
		for d in items:
			_dish_box.add_child(_dish_row(d))
	)
	hub.bind_list(vm, "cart_lines", func(lines: Array) -> void:
		for child in _cart_lines.get_children():
			child.queue_free()
		for line in lines:
			_cart_lines.add_child(_cart_line(line))
	)
