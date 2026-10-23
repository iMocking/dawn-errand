class_name HomeView
extends ScreenBase
## 首页视图：渐变头部（定位/问候/搜索）+ 场景入口 + 推荐横滑 + 附近商家。

func build() -> void:
	# ---- 渐变头部 ----
	var header := Kit.gradient_rect(DesignTokens.PRIMARY_LIGHT, DesignTokens.PRIMARY_DARK)
	header.anchor_right = 1.0
	header.offset_bottom = 232.0
	add_child(header)
	var sb := StatusBar.new(Color(1, 1, 1, 0.95))
	sb.set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	add_child(sb)

	# 顶部行：定位 + 通知
	var top_row := Kit.hbox(8)
	top_row.anchor_left = 0.0
	top_row.anchor_right = 1.0
	top_row.offset_left = 20.0
	top_row.offset_right = -20.0
	top_row.offset_top = 52.0
	top_row.offset_bottom = 88.0
	add_child(top_row)
	var loc_pill := PanelContainer.new()
	loc_pill.add_theme_stylebox_override("panel", Kit.style_flat(Color(1, 1, 1, 0.2), 16.0))
	var loc_box := Kit.hbox(5)
	loc_box.offset_left = 12.0
	loc_box.offset_right = -12.0
	loc_box.offset_top = 0.0
	loc_box.offset_bottom = 0.0
	loc_pill.add_child(loc_box)
	var loc_ic := Icon.new(Icon.Kind.LOCATION, Color.WHITE, 1.8)
	loc_ic.custom_minimum_size = Vector2(16, 16)
	loc_ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	loc_box.add_child(loc_ic)
	var loc_l := Kit.label_sys("湖北省黄冈市浠水县", 13, Color.WHITE)
	loc_l.custom_minimum_size = Vector2(0, 32)
	loc_box.add_child(loc_l)
	var loc_arrow := Icon.new(Icon.Kind.ARROW_DOWN, Color.WHITE, 1.8)
	loc_arrow.custom_minimum_size = Vector2(14, 14)
	loc_arrow.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	loc_box.add_child(loc_arrow)
	top_row.add_child(loc_pill)
	var spacer := Control.new()
	spacer.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	top_row.add_child(spacer)
	var bell := Kit.icon_button(Icon.Kind.BELL, Callable(), Color(1, 1, 1, 0.2), Color.WHITE, 36.0)
	bell.pressed.connect(func() -> void: Toast.show_msg("暂无新通知"))
	top_row.add_child(bell)

	# 问候语
	var greet_row := Kit.hbox(6)
	greet_row.anchor_left = 0.0
	greet_row.anchor_right = 1.0
	greet_row.offset_left = 20.0
	greet_row.offset_right = -20.0
	greet_row.offset_top = 98.0
	greet_row.offset_bottom = 132.0
	add_child(greet_row)
	var greet := Kit.label_sys("中午好，", 24, Color.WHITE, true)
	greet_row.add_child(greet)
	var name_l := Kit.label_sys(str(vm.get_prop("username")), 24, Color.WHITE, true)
	greet_row.add_child(name_l)
	var wave := Kit.label_sys("👋", 20, Color.WHITE)
	greet_row.add_child(wave)
	var sub := Kit.label_sys("今天想吃点啥？", 13, Color(1, 1, 1, 0.92))
	sub.anchor_left = 0.0
	sub.anchor_right = 1.0
	sub.offset_left = 20.0
	sub.offset_right = -20.0
	sub.offset_top = 134.0
	sub.offset_bottom = 158.0
	add_child(sub)

	# 搜索栏
	var search_pill := PanelContainer.new()
	search_pill.anchor_left = 0.0
	search_pill.anchor_right = 1.0
	search_pill.offset_left = 20.0
	search_pill.offset_right = -20.0
	search_pill.offset_top = 176.0
	search_pill.offset_bottom = 224.0
	search_pill.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 24.0, Color.TRANSPARENT, 0.0, true))
	add_child(search_pill)
	var search_box := Kit.hbox(8)
	search_box.offset_left = 16.0
	search_box.offset_right = -16.0
	search_pill.add_child(search_box)
	var s_ic := Icon.new(Icon.Kind.SEARCH, DesignTokens.TEXT_MUTED, 1.8)
	s_ic.custom_minimum_size = Vector2(18, 18)
	s_ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	search_box.add_child(s_ic)
	var s_l := Kit.label_sys("搜索商家、商品", 14, DesignTokens.TEXT_MUTED)
	search_box.add_child(s_l)
	var s_btn := Button.new()
	s_btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	s_btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	s_btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	s_btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	s_btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	s_btn.pressed.connect(func() -> void: vm.open_search())
	search_pill.add_child(s_btn)

	# ---- 滚动内容区 ----
	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = 216.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	var root := Kit.vbox(0)
	root.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sc.add_child(root)

	# 场景入口（点个饭 / 跑个腿）
	var scenes := Kit.hbox(12)
	scenes.offset_left = 20.0
	scenes.offset_right = -20.0
	scenes.offset_top = 0.0
	scenes.offset_bottom = 0.0
	root.add_child(scenes)
	scenes.add_child(_scene_card("点个饭", "附近美食 · 30分钟必达", Icon.Kind.RICE, DesignTokens.GRAD_SCENE_EAT, func() -> void: vm.open_search()))
	scenes.add_child(_scene_card("跑个腿", "代取代送 · 随叫随到", Icon.Kind.ERRAND, DesignTokens.GRAD_SCENE_RUN, func() -> void: vm.open_errand()))

	root.add_child(Kit.spacing(0, 22))
	root.add_child(_section_header("推荐"))
	root.add_child(Kit.spacing(0, 10))
	root.add_child(_recommend_scroll())

	root.add_child(Kit.spacing(0, 22))
	root.add_child(_section_header("附近商家"))
	root.add_child(Kit.spacing(0, 8))
	root.add_child(_nearby_list())
	root.add_child(Kit.spacing(0, 24))

func _scene_card(title: String, sub: String, icon_kind: int, colors: Array, cb: Callable) -> Control:
	var card := Control.new()
	card.custom_minimum_size = Vector2(0, 96)
	card.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var grad := GradientCard.new(colors[0], colors[1], 16.0, Color.WHITE)
	grad.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	card.add_child(grad)
	var t := Kit.label_sys(title, 17, Color.WHITE, true)
	t.anchor_left = 0.0
	t.offset_left = 16.0
	t.offset_top = 18.0
	t.offset_right = 120.0
	t.offset_bottom = 46.0
	card.add_child(t)
	var s := Kit.label_sys(sub, 11, Color(1, 1, 1, 0.92))
	s.anchor_left = 0.0
	s.offset_left = 16.0
	s.offset_top = 50.0
	s.offset_right = 160.0
	s.offset_bottom = 72.0
	card.add_child(s)
	var ic := Icon.new(icon_kind, Color(1, 1, 1, 0.95), 2.0)
	ic.custom_minimum_size = Vector2(44, 44)
	ic.anchor_left = 1.0
	ic.anchor_right = 1.0
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -62.0
	ic.offset_right = -18.0
	ic.offset_top = -22.0
	ic.offset_bottom = 22.0
	card.add_child(ic)
	var btn := Button.new()
	btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	if cb.is_valid():
		btn.pressed.connect(cb)
	card.add_child(btn)
	return card

func _section_header(title: String) -> Control:
	var row := Kit.hbox(8)
	row.offset_left = 20.0
	row.offset_right = -20.0
	row.offset_top = 0.0
	row.offset_bottom = 0.0
	var t := Kit.label_sys(title, 17, DesignTokens.TEXT, true)
	row.add_child(t)
	var more := Kit.label_sys("查看全部", 12, DesignTokens.TEXT_MUTED)
	more.size_flags_horizontal = Control.SIZE_SHRINK_END
	row.add_child(more)
	return row

func _recommend_scroll() -> Control:
	var hs := ScrollContainer.new()
	hs.custom_minimum_size = Vector2(0, 150)
	hs.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_AUTO
	var row := Kit.hbox(12)
	row.offset_left = 20.0
	row.offset_right = -20.0
	row.offset_top = 0.0
	row.offset_bottom = 0.0
	hs.add_child(row)
	for m in vm.list("recommend").items():
		row.add_child(_rec_card(m))
	return hs

func _rec_card(m: MerchantModel) -> Control:
	var card := Kit.card(Color.WHITE, 12.0, true)
	card.custom_minimum_size = Vector2(132, 138)
	var box := Kit.vbox(0)
	box.offset_left = 0.0
	box.offset_right = 0.0
	box.offset_top = 0.0
	box.offset_bottom = 0.0
	card.add_child(box)
	# 渐变图标区
	var banner := GradientCard.new(m.gradient[0], m.gradient[1], 12.0, Color.WHITE)
	banner.custom_minimum_size = Vector2(132, 66)
	banner.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	box.add_child(banner)
	var ic := Icon.new(m.icon, Color.WHITE, 2.0)
	ic.custom_minimum_size = Vector2(34, 34)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -17.0
	ic.offset_right = 17.0
	ic.offset_top = -17.0
	ic.offset_bottom = 17.0
	banner.add_child(ic)
	var info := Kit.vbox(4)
	info.offset_left = 10.0
	info.offset_right = -10.0
	info.offset_top = 6.0
	info.offset_bottom = -8.0
	box.add_child(info)
	var n := Kit.label_sys(m.name, 14, DesignTokens.TEXT, true)
	info.add_child(n)
	var meta := Kit.label_sys(m.meta_text(), 10, DesignTokens.TEXT_MUTED)
	info.add_child(meta)
	var chips := Kit.hbox(5)
	info.add_child(chips)
	for i in mini(2, m.tags.size()):
		chips.add_child(Kit.chip(str(m.tags[i]), false))
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

func _nearby_list() -> Control:
	var col := Kit.vbox(12)
	col.offset_left = 20.0
	col.offset_right = -20.0
	for m in vm.list("nearby").items():
		col.add_child(_nearby_row(m))
	return col

func _nearby_row(m: MerchantModel) -> Control:
	var card := Kit.card(Color.WHITE, 14.0, false, true)
	card.custom_minimum_size = Vector2(0, 92)
	var row := Kit.hbox(12)
	row.offset_left = 14.0
	row.offset_right = -14.0
	row.offset_top = 0.0
	row.offset_bottom = 0.0
	card.add_child(row)
	var icon_box := PanelContainer.new()
	icon_box.custom_minimum_size = Vector2(58, 58)
	icon_box.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	icon_box.add_theme_stylebox_override("panel", Kit.style_flat(m.gradient[0].lerp(m.gradient[1], 0.35), 12.0))
	var ic := Icon.new(m.icon, Color.WHITE, 2.0)
	ic.custom_minimum_size = Vector2(30, 30)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -15.0
	ic.offset_right = 15.0
	ic.offset_top = -15.0
	ic.offset_bottom = 15.0
	icon_box.add_child(ic)
	row.add_child(icon_box)
	var info := Kit.vbox(5)
	info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	info.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	row.add_child(info)
	var n := Kit.label_sys(m.name, 16, DesignTokens.TEXT, true)
	info.add_child(n)
	var meta := Kit.label_sys("⭐ " + m.score_text() + " · 月售" + str(m.monthly_sales) + "单 · " + Fmt.km(m.dist_km) + " · " + str(m.eta_min) + "分钟送到", 12, DesignTokens.TEXT_MUTED)
	info.add_child(meta)
	var chips := Kit.hbox(5)
	for tag in m.tags:
		chips.add_child(Kit.chip(str(tag), false))
	info.add_child(chips)
	var arrow := Icon.new(Icon.Kind.CHEVRON_RIGHT, DesignTokens.TEXT_FAINT, 2.0)
	arrow.custom_minimum_size = Vector2(18, 18)
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

func bind() -> void:
	pass
