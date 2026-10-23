class_name AddressEditView
extends ScreenBase
## 地址编辑视图：联系人/手机号/详细地址/标签/默认开关 + 保存。

var _name_input: LineEdit
var _phone_input: LineEdit
var _detail_input: TextEdit
var _landmark_input: LineEdit
var _tag_buttons: Array = []
var _default_panel: PanelContainer
var _default_check: Icon
var _save_btn: Button

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new(str(vm.get_prop("title")), func() -> void: App.go_back())
	add_child(nav)

	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = DesignTokens.NAV_H
	sc.offset_bottom = -90.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	var root := Kit.vbox(12)
	root.offset_left = 16.0
	root.offset_right = -16.0
	sc.add_child(root)

	# 表单卡
	var form := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(form)
	var f_box := Kit.vbox(0)
	f_box.offset_left = 16.0
	f_box.offset_right = -16.0
	form.add_child(f_box)
	f_box.add_child(_row_input("联系人", "请输入收货人姓名", "_name", _name_input, false))
	f_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	f_box.add_child(_row_input("手机号", "请输入收货人手机号", "_phone", _phone_input, false))
	f_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	f_box.add_child(_row_input("所在地区", "湖北省黄冈市浠水县", "_province", null, true))
	f_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	var detail_row := Kit.vbox(6)
	detail_row.custom_minimum_size = Vector2(0, 110)
	f_box.add_child(detail_row)
	var d_l := Kit.label_sys("详细地址", 14, DesignTokens.TEXT)
	detail_row.add_child(d_l)
	_detail_input = TextEdit.new()
	_detail_input.placeholder_text = "街道、门牌号、楼层等"
	_detail_input.custom_minimum_size = Vector2(0, 64)
	_detail_input.add_theme_font_size_override("font_size", 14)
	_detail_input.add_theme_color_override("font_color", DesignTokens.TEXT)
	_detail_input.add_theme_color_override("font_placeholder_color", DesignTokens.TEXT_FAINT)
	_detail_input.add_theme_color_override("caret_color", DesignTokens.PRIMARY)
	_detail_input.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	_detail_input.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	detail_row.add_child(_detail_input)
	f_box.add_child(Kit.divider(1.0, DesignTokens.DIVIDER))
	f_box.add_child(_row_input("标志建筑", "选填：如 门口有棵樟树", "_landmark", _landmark_input, false))

	# 标签卡
	var tag_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(tag_card)
	var tag_box := Kit.vbox(10)
	tag_box.offset_left = 16.0
	tag_box.offset_right = -16.0
	tag_box.offset_top = 14.0
	tag_box.offset_bottom = -14.0
	tag_card.add_child(tag_box)
	var tag_t := Kit.label_sys("地址标签", 15, DesignTokens.TEXT, true)
	tag_box.add_child(tag_t)
	var tag_row := Kit.hbox(8)
	tag_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	tag_box.add_child(tag_row)
	for tag in vm.list("tags").items():
		var b := Button.new()
		b.text = str(tag)
		b.custom_minimum_size = Vector2(0, 34)
		b.add_theme_font_size_override("font_size", 13)
		b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
		var t := str(tag)
		b.pressed.connect(func() -> void:
			vm.set_tag(t)
			_refresh_tag_styles()
		)
		tag_row.add_child(b)
		_tag_buttons.append(b)

	# 默认开关卡
	var def_card := Kit.card(Color.WHITE, 14.0, false, true)
	root.add_child(def_card)
	var def_row := Kit.hbox(0)
	def_row.offset_left = 16.0
	def_row.offset_right = -16.0
	def_row.custom_minimum_size = Vector2(0, 50)
	def_card.add_child(def_row)
	var def_l := Kit.label_sys("设为默认地址", 15, DesignTokens.TEXT)
	def_l.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	def_row.add_child(def_l)
	_default_panel = PanelContainer.new()
	_default_panel.custom_minimum_size = Vector2(22, 22)
	_default_panel.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	def_row.add_child(_default_panel)
	_default_check = Icon.new(Icon.Kind.CHECK, Color.WHITE, 2.0)
	_default_check.custom_minimum_size = Vector2(14, 14)
	_default_check.set_anchors_preset(Control.PRESET_CENTER)
	_default_check.anchor_left = 0.5
	_default_check.anchor_right = 0.5
	_default_check.anchor_top = 0.5
	_default_check.anchor_bottom = 0.5
	_default_check.offset_left = -7.0
	_default_check.offset_right = 7.0
	_default_check.offset_top = -7.0
	_default_check.offset_bottom = 7.0
	_default_panel.add_child(_default_check)
	var def_btn := Button.new()
	def_btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	def_btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	def_btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	def_btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	def_btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	def_btn.pressed.connect(func() -> void: vm.toggle_default())
	def_card.add_child(def_btn)

	root.add_child(Kit.spacing(0, 4))

	# 保存
	var bottom := PanelContainer.new()
	bottom.anchor_left = 0.0
	bottom.anchor_right = 1.0
	bottom.anchor_top = 1.0
	bottom.anchor_bottom = 1.0
	bottom.offset_top = -90.0
	bottom.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 0.0))
	add_child(bottom)
	var b_row := Kit.hbox(12)
	b_row.offset_left = 20.0
	b_row.offset_right = -20.0
	bottom.add_child(b_row)
	_save_btn = Kit.primary_button("保存地址", func() -> void: vm.save(), 48.0, 24.0, 15)
	b_row.add_child(_save_btn)

func _row_input(label: String, placeholder: String, _member: String, edit: LineEdit, readonly: bool) -> Control:
	var row := Kit.hbox(12)
	row.custom_minimum_size = Vector2(0, 52)
	var l := Kit.label_sys(label, 14, DesignTokens.TEXT)
	l.custom_minimum_size = Vector2(72, 0)
	row.add_child(l)
	var e := LineEdit.new()
	e.placeholder_text = placeholder
	e.editable = not readonly
	e.add_theme_font_size_override("font_size", 14)
	e.add_theme_color_override("font_color", DesignTokens.TEXT if not readonly else DesignTokens.TEXT_MUTED)
	e.add_theme_color_override("font_placeholder_color", DesignTokens.TEXT_FAINT)
	e.add_theme_color_override("caret_color", DesignTokens.PRIMARY)
	e.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	e.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	e.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(e)
	match _member:
		"_name":
			_name_input = e
		"_phone":
			_phone_input = e
		"_landmark":
			_landmark_input = e
	return row

func _refresh_tag_styles() -> void:
	var cur := str(vm.get_prop("tag"))
	for b in _tag_buttons:
		var active: bool = b.text == cur
		b.add_theme_color_override("font_color", Color.WHITE if active else DesignTokens.TEXT_SEC)
		b.add_theme_color_override("font_hover_color", Color.WHITE if active else DesignTokens.PRIMARY)
		b.add_theme_color_override("font_pressed_color", DesignTokens.PRIMARY)
		b.add_theme_stylebox_override("normal", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 17.0))
		b.add_theme_stylebox_override("hover", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 17.0))
		b.add_theme_stylebox_override("pressed", Kit.style_flat(DesignTokens.PRIMARY if active else Color("F5F6F8"), 17.0))

func bind() -> void:
	_name_input.text = str(vm.get_prop("name"))
	_phone_input.text = str(vm.get_prop("phone"))
	_detail_input.text = str(vm.get_prop("detail"))
	_landmark_input.text = str(vm.get_prop("landmark"))
	_name_input.text_changed.connect(func(t: String) -> void: vm.set_name(t))
	_phone_input.text_changed.connect(func(t: String) -> void: vm.set_phone(t))
	_detail_input.text_changed.connect(func() -> void: vm.set_detail(_detail_input.text))
	_landmark_input.text_changed.connect(func(t: String) -> void: vm.set_landmark(t))
	hub.bind(vm, "tag", func(_v) -> void: _refresh_tag_styles())
	hub.bind(vm, "is_default", func(v) -> void:
		var on: bool = v
		_default_panel.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY if on else DesignTokens.TEXT_FAINT, 6.0))
		_default_check.visible = on
	)
	_refresh_tag_styles()
