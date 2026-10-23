class_name SearchView
extends ScreenBase
## 搜索视图：顶部搜索栏、热门/历史（初始）、结果列表（搜索后）。

var _input: LineEdit
var _hot_row: Control
var _history_box: VBoxContainer
var _result_box: VBoxContainer
var _initial_panel: Control
var _result_panel: Control
var _result_count: Label

func build() -> void:
	var bg := ColorRect.new()
	bg.color = Color.WHITE
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	# 顶部搜索栏
	var top := Control.new()
	top.set_anchors_preset(Control.PRESET_TOP_WIDE)
	top.anchor_left = 0.0
	top.anchor_right = 1.0
	top.offset_top = DesignTokens.STATUS_BAR_H
	top.offset_bottom = DesignTokens.STATUS_BAR_H + 52.0
	add_child(top)
	var back := Kit.icon_button(Icon.Kind.BACK, func() -> void: App.go_back(), Color.WHITE, DesignTokens.TEXT, 36.0)
	back.position = Vector2(10, 8)
	top.add_child(back)
	var pill := PanelContainer.new()
	pill.set_anchors_preset(Control.PRESET_LEFT_WIDE)
	pill.anchor_left = 0.0
	pill.anchor_right = 1.0
	pill.offset_left = 56.0
	pill.offset_right = -66.0
	pill.offset_top = 4.0
	pill.offset_bottom = 48.0
	pill.add_theme_stylebox_override("panel", Kit.style_flat(Color("F5F6F8"), 22.0))
	top.add_child(pill)
	var pill_box := Kit.hbox(8)
	pill_box.offset_left = 12.0
	pill_box.offset_right = -12.0
	pill.add_child(pill_box)
	var s_ic := Icon.new(Icon.Kind.SEARCH, DesignTokens.TEXT_MUTED, 1.8)
	s_ic.custom_minimum_size = Vector2(18, 18)
	s_ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	pill_box.add_child(s_ic)
	_input = LineEdit.new()
	_input.placeholder_text = "搜索商家、商品"
	_input.add_theme_font_size_override("font_size", 15)
	_input.add_theme_color_override("font_color", DesignTokens.TEXT)
	_input.add_theme_color_override("font_placeholder_color", DesignTokens.TEXT_FAINT)
	_input.add_theme_color_override("caret_color", DesignTokens.PRIMARY)
	_input.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	_input.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	_input.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	pill_box.add_child(_input)
	var go := Button.new()
	go.set_anchors_preset(Control.PRESET_RIGHT_WIDE)
	go.anchor_left = 1.0
	go.anchor_right = 1.0
	go.offset_left = -62.0
	go.offset_right = -16.0
	go.offset_top = 0.0
	go.offset_bottom = 52.0
	go.text = "搜索"
	go.add_theme_font_size_override("font_size", 15)
	go.add_theme_color_override("font_color", DesignTokens.PRIMARY)
	go.add_theme_color_override("font_hover_color", DesignTokens.PRIMARY)
	go.add_theme_color_override("font_pressed_color", DesignTokens.PRIMARY)
	go.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	go.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	go.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	go.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	go.pressed.connect(func() -> void: vm.search(_input.text))
	top.add_child(go)

	# 初始面板（热门 + 历史）
	_initial_panel = Control.new()
	_initial_panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_initial_panel.offset_top = DesignTokens.STATUS_BAR_H + 64.0
	add_child(_initial_panel)
	var root := Kit.vbox(0)
	root.offset_left = 20.0
	root.offset_right = -20.0
	root.offset_top = 0.0
	_initial_panel.add_child(root)
	var hot_t := Kit.label_sys("热门搜索", 15, DesignTokens.TEXT, true)
	root.add_child(hot_t)
	root.add_child(Kit.spacing(0, 12))
	_hot_row = Kit.hbox(10)
	_hot_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_hot_row.custom_minimum_size = Vector2(0, 40)
	root.add_child(_hot_row)
	root.add_child(Kit.spacing(0, 24))
	var his_t := Kit.label_sys("搜索历史", 15, DesignTokens.TEXT, true)
	root.add_child(his_t)
	root.add_child(Kit.spacing(0, 6))
	_history_box = Kit.vbox(0)
	root.add_child(_history_box)

	# 结果面板
	_result_panel = Control.new()
	_result_panel.visible = false
	_result_panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_result_panel.offset_top = DesignTokens.STATUS_BAR_H + 64.0
	add_child(_result_panel)
	var r_root := Kit.vbox(12)
	r_root.offset_left = 16.0
	r_root.offset_right = -16.0
	_result_panel.add_child(r_root)
	_result_count = Kit.label_sys("", 13, DesignTokens.TEXT_MUTED)
	r_root.add_child(_result_count)
	_result_box = Kit.vbox(12)
	r_root.add_child(_result_box)

func bind() -> void:
	hub.bind_list(vm, "hot", func(items: Array) -> void:
		for child in _hot_row.get_children():
			child.queue_free()
		for kw in items:
			var b := Button.new()
			b.text = str(kw)
			b.custom_minimum_size = Vector2(0, 34)
			b.add_theme_font_size_override("font_size", 13)
			b.add_theme_color_override("font_color", DesignTokens.TEXT_SEC)
			b.add_theme_color_override("font_hover_color", DesignTokens.PRIMARY)
			b.add_theme_color_override("font_pressed_color", DesignTokens.PRIMARY)
			b.add_theme_stylebox_override("normal", Kit.style_flat(Color("F5F6F8"), 17.0))
			b.add_theme_stylebox_override("hover", Kit.style_flat(Color("F5F6F8"), 17.0))
			b.add_theme_stylebox_override("pressed", Kit.style_flat(Color("F5F6F8"), 17.0))
			b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
			var key := str(kw)
			b.pressed.connect(func() -> void: vm.search(key))
			_hot_row.add_child(b)
	)
	hub.bind_list(vm, "history", func(items: Array) -> void:
		for child in _history_box.get_children():
			child.queue_free()
		if items.size() == 0:
			_history_box.add_child(Kit.label_sys("暂无搜索历史", 13, DesignTokens.TEXT_FAINT))
		for item in items:
			_history_box.add_child(_history_row(str(item)))
	)
	hub.bind(vm, "has_result", func(v) -> void:
		var has: bool = v
		_initial_panel.visible = not has
		_result_panel.visible = has
	)
	hub.bind(vm, "result_count", func(v) -> void:
		_result_count.text = "为你找到 " + str(int(v)) + " 个相关商家"
	)
	hub.bind_list(vm, "results", func(items: Array) -> void:
		for child in _result_box.get_children():
			child.queue_free()
		for m in items:
			_result_box.add_child(_merchant_row(m))
	)
	_input.text_submitted.connect(func(t: String) -> void: vm.search(t))

func _history_row(text: String) -> Control:
	var row := Control.new()
	row.custom_minimum_size = Vector2(0, 42)
	var l := Kit.label_sys(text, 13, DesignTokens.TEXT_SEC)
	l.position = Vector2(0, 12)
	row.add_child(l)
	var del := Kit.icon_button(Icon.Kind.CLOSE, Callable(), Color.WHITE, DesignTokens.TEXT_FAINT, 24.0)
	del.anchor_left = 1.0
	del.anchor_right = 1.0
	del.anchor_top = 0.5
	del.anchor_bottom = 0.5
	del.offset_left = -30.0
	del.offset_right = -6.0
	del.offset_top = -12.0
	del.offset_bottom = 12.0
	del.pressed.connect(func() -> void: vm.clear_history())
	row.add_child(del)
	return row

func _merchant_row(m: MerchantModel) -> Control:
	var card := Kit.card(Color.WHITE, 14.0, false, true)
	var row := Kit.hbox(12)
	row.offset_left = 14.0
	row.offset_right = -14.0
	row.offset_top = 0.0
	row.offset_bottom = 0.0
	card.add_child(row)
	var icon_box := PanelContainer.new()
	icon_box.custom_minimum_size = Vector2(52, 52)
	icon_box.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	icon_box.add_theme_stylebox_override("panel", Kit.style_flat(m.gradient[0].lerp(m.gradient[1], 0.35), 12.0))
	var ic := Icon.new(m.icon, Color.WHITE, 2.0)
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
	var n := Kit.label_sys(m.name, 15, DesignTokens.TEXT, true)
	info.add_child(n)
	var meta := Kit.label_sys("⭐ " + m.score_text() + " · 月售" + str(m.monthly_sales) + "单 · " + Fmt.km(m.dist_km), 12, DesignTokens.TEXT_MUTED)
	info.add_child(meta)
	var arrow := Icon.new(Icon.Kind.CHEVRON_RIGHT, DesignTokens.TEXT_FAINT, 2.0)
	arrow.custom_minimum_size = Vector2(16, 16)
	arrow.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	row.add_child(arrow)
	var btn := Button.new()
	btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	var merchant := m
	btn.pressed.connect(func() -> void: vm.open_merchant(merchant))
	card.add_child(btn)
	return card
