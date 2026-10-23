class_name ReviewView
extends ScreenBase
## 评价视图：商家、星级、标签、评语、提交。

var _stars: StarRating
var _tag_buttons: Array = []
var _tag_row: HBoxContainer
var _comment: TextEdit
var _submit: Button

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new("评价订单", func() -> void: App.go_back())
	add_child(nav)

	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = DesignTokens.NAV_H
	sc.offset_bottom = -88.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	var root := Kit.vbox(12)
	root.offset_left = 16.0
	root.offset_right = -16.0
	sc.add_child(root)

	# 商家卡
	var shop_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(shop_card)
	var s_row := Kit.hbox(10)
	s_row.offset_left = 16.0
	s_row.offset_right = -16.0
	s_row.offset_top = 0.0
	s_row.offset_bottom = 0.0
	shop_card.add_child(s_row)
	var icon_box := PanelContainer.new()
	icon_box.custom_minimum_size = Vector2(44, 44)
	icon_box.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	icon_box.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY_BG, 22.0))
	var s_ic := Icon.new(Icon.Kind.BAG, DesignTokens.PRIMARY, 2.0)
	s_ic.custom_minimum_size = Vector2(24, 24)
	s_ic.set_anchors_preset(Control.PRESET_CENTER)
	s_ic.anchor_left = 0.5
	s_ic.anchor_right = 0.5
	s_ic.anchor_top = 0.5
	s_ic.anchor_bottom = 0.5
	s_ic.offset_left = -12.0
	s_ic.offset_right = 12.0
	s_ic.offset_top = -12.0
	s_ic.offset_bottom = 12.0
	icon_box.add_child(s_ic)
	s_row.add_child(icon_box)
	var shop_l := Kit.label_sys(str(vm.get_prop("shop")), 16, DesignTokens.TEXT, true)
	s_row.add_child(shop_l)

	# 评分卡
	var star_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(star_card)
	var star_box := Kit.vbox(12)
	star_box.offset_left = 16.0
	star_box.offset_right = -16.0
	star_box.offset_top = 16.0
	star_box.offset_bottom = -16.0
	star_card.add_child(star_box)
	var star_t := Kit.label_sys("服务评分", 15, DesignTokens.TEXT, true)
	star_box.add_child(star_t)
	_stars = StarRating.new(5, 36.0)
	star_box.add_child(_stars)

	# 标签卡
	var tag_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(tag_card)
	var tag_box := Kit.vbox(12)
	tag_box.offset_left = 16.0
	tag_box.offset_right = -16.0
	tag_box.offset_top = 16.0
	tag_box.offset_bottom = -16.0
	tag_card.add_child(tag_box)
	var tag_t := Kit.label_sys("标签评价（可多选）", 15, DesignTokens.TEXT, true)
	tag_box.add_child(tag_t)
	_tag_row = Kit.hbox(8)
	_tag_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	tag_box.add_child(_tag_row)

	# 评语卡
	var cmt_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(cmt_card)
	var cmt_box := Kit.vbox(10)
	cmt_box.offset_left = 16.0
	cmt_box.offset_right = -16.0
	cmt_box.offset_top = 14.0
	cmt_box.offset_bottom = -14.0
	cmt_card.add_child(cmt_box)
	var cmt_t := Kit.label_sys("评价内容", 15, DesignTokens.TEXT, true)
	cmt_box.add_child(cmt_t)
	_comment = TextEdit.new()
	_comment.placeholder_text = "说说菜品和服务吧（选填）"
	_comment.custom_minimum_size = Vector2(0, 100)
	_comment.add_theme_font_size_override("font_size", 14)
	_comment.add_theme_color_override("font_color", DesignTokens.TEXT)
	_comment.add_theme_color_override("font_placeholder_color", DesignTokens.TEXT_FAINT)
	_comment.add_theme_color_override("caret_color", DesignTokens.PRIMARY)
	_comment.add_theme_stylebox_override("normal", Kit.style_flat(Color("F5F6F8"), 10.0))
	_comment.add_theme_stylebox_override("focus", Kit.style_flat(Color("F5F6F8"), 10.0, DesignTokens.PRIMARY, 1.0))
	cmt_box.add_child(_comment)

	root.add_child(Kit.spacing(0, 4))

	# 提交
	var bottom := PanelContainer.new()
	bottom.anchor_left = 0.0
	bottom.anchor_right = 1.0
	bottom.anchor_top = 1.0
	bottom.anchor_bottom = 1.0
	bottom.offset_top = -88.0
	bottom.offset_bottom = 0.0
	bottom.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 0.0))
	add_child(bottom)
	var b_row := Kit.hbox(12)
	b_row.offset_left = 20.0
	b_row.offset_right = -20.0
	bottom.add_child(b_row)
	_submit = Kit.primary_button("提交评价", func() -> void: vm.submit(), 48.0, 24.0, 15)
	b_row.add_child(_submit)

func bind() -> void:
	_stars.changed.connect(func(v: int) -> void: vm.set_stars(v))
	_comment.text_changed.connect(func() -> void: vm.set_comment(_comment.text))
	hub.bind_list(vm, "tags", func(items: Array) -> void:
		for child in _tag_row.get_children():
			child.queue_free()
		_tag_buttons.clear()
		for tag in items:
			var b := Button.new()
			b.text = str(tag)
			b.custom_minimum_size = Vector2(0, 34)
			b.add_theme_font_size_override("font_size", 13)
			b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
			var t := str(tag)
			b.pressed.connect(func() -> void:
				vm.toggle_tag(t)
				_refresh_tag_styles()
			)
			_tag_row.add_child(b)
			_tag_buttons.append(b)
	)
	hub.bind_list(vm, "selected_tags", func(_items: Array) -> void:
		_refresh_tag_styles()
	)

func _refresh_tag_styles() -> void:
	var sel: Array = vm.list("selected_tags").items()
	for i in _tag_buttons.size():
		var b: Button = _tag_buttons[i]
		var active := sel.has(b.text)
		b.add_theme_color_override("font_color", Color.WHITE if active else DesignTokens.TEXT_SEC)
		b.add_theme_color_override("font_hover_color", Color.WHITE if active else DesignTokens.PRIMARY)
		b.add_theme_color_override("font_pressed_color", DesignTokens.PRIMARY)
		b.add_theme_stylebox_override("normal", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 17.0))
		b.add_theme_stylebox_override("hover", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 17.0))
		b.add_theme_stylebox_override("pressed", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 17.0))
