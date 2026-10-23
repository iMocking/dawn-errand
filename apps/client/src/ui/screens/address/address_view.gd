class_name AddressView
extends ScreenBase
## 地址管理视图：地址卡片列表 + 底部新增按钮。

var _list_box: VBoxContainer
var _empty: Control

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new("收货地址", func() -> void: App.go_back(), "新增", func() -> void: vm.add_address())
	add_child(nav)

	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = DesignTokens.NAV_H
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	_list_box = Kit.vbox(12)
	_list_box.offset_left = 16.0
	_list_box.offset_right = -16.0
	_list_box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sc.add_child(_list_box)

	_empty = Control.new()
	_empty.visible = false
	_empty.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(_empty)
	var eb := Kit.vbox(10)
	eb.set_anchors_preset(Control.PRESET_CENTER)
	eb.anchor_left = 0.5
	eb.anchor_right = 0.5
	eb.anchor_top = 0.5
	eb.anchor_bottom = 0.5
	eb.offset_left = -90.0
	eb.offset_right = 90.0
	eb.offset_top = -60.0
	eb.offset_bottom = 60.0
	_empty.add_child(eb)
	var e_ic := Icon.new(Icon.Kind.LOCATION, DesignTokens.TEXT_FAINT, 2.0)
	e_ic.custom_minimum_size = Vector2(44, 44)
	e_ic.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	eb.add_child(e_ic)
	eb.add_child(Kit.label_sys("还没有收货地址", 14, DesignTokens.TEXT_MUTED, false, HORIZONTAL_ALIGNMENT_CENTER))

	# 底部新增
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
	var add_btn := Kit.primary_button("＋ 新增地址", func() -> void: vm.add_address(), 48.0, 24.0, 15)
	b_row.add_child(add_btn)

func _addr_card(a: AddressModel) -> Control:
	var card := Kit.card(Color.WHITE, 14.0, false, true)
	var box := Kit.vbox(0)
	box.offset_left = 16.0
	box.offset_right = -16.0
	box.offset_top = 12.0
	box.offset_bottom = -12.0
	card.add_child(box)
	var head := Kit.hbox(8)
	box.add_child(head)
	var who := Kit.label_sys(a.name + "  " + Fmt.mask_phone(a.phone), 15, DesignTokens.TEXT, true)
	who.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	head.add_child(who)
	if a.is_default:
		head.add_child(Kit.chip("默认", true, DesignTokens.PRIMARY))
	else:
		head.add_child(Kit.chip(a.tag, false, DesignTokens.PRIMARY))
	box.add_child(Kit.spacing(0, 5))
	var detail := Kit.label_sys(a.full_detail(), 13, DesignTokens.TEXT_SEC)
	box.add_child(detail)
	box.add_child(Kit.spacing(0, 6))
	# 操作行
	var ops := Kit.hbox(6)
	box.add_child(ops)
	if not a.is_default:
		var def := Kit.outline_button("设为默认", Callable(), 28.0, 12, DesignTokens.TEXT_MUTED)
		def.pressed.connect(func() -> void: vm.set_default(a))
		ops.add_child(def)
	var edit := Kit.outline_button("编辑", Callable(), 28.0, 12)
	edit.pressed.connect(func() -> void: vm.edit_address(a))
	ops.add_child(edit)
	var del := Kit.outline_button("删除", Callable(), 28.0, 12, DesignTokens.PRICE)
	del.pressed.connect(func() -> void: vm.delete_address(a))
	ops.add_child(del)
	return card

func bind() -> void:
	hub.bind_list(vm, "addresses", func(items: Array) -> void:
		for child in _list_box.get_children():
			child.queue_free()
		if items.size() == 0:
			_empty.visible = true
			_list_box.visible = false
			return
		_empty.visible = false
		_list_box.visible = true
		for a in items:
			_list_box.add_child(_addr_card(a))
	)
