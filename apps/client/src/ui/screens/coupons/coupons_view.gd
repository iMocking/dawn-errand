class_name CouponsView
extends ScreenBase
## 我的优惠券视图：券卡片列表（可用/已过期）。

var _list_box: VBoxContainer

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new("我的优惠券", func() -> void: App.go_back())
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

func _coupon_card(c: CouponModel) -> Control:
	var card := Kit.card(Color.WHITE, 14.0, false, true)
	var row := Kit.hbox(0)
	row.offset_left = 0.0
	row.offset_right = 0.0
	row.offset_top = 0.0
	row.offset_bottom = 0.0
	card.add_child(row)
	# 左侧金额区
	var left := PanelContainer.new()
	left.custom_minimum_size = Vector2(104, 0)
	left.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY if c.usable else Color("C9CDD4"), 0.0))
	var lb := Kit.vbox(0)
	lb.offset_left = 0.0
	lb.offset_right = 0.0
	lb.offset_top = 16.0
	lb.offset_bottom = -16.0
	left.add_child(lb)
	var amount_row := Kit.hbox(0)
	lb.add_child(amount_row)
	var yuan := Kit.label_sys("¥", 15, Color.WHITE, true)
	amount_row.add_child(yuan)
	var amount := Kit.label_sys(c.amount_text(), 28, Color.WHITE, true)
	amount_row.add_child(amount)
	var cond := Kit.label_sys(c.condition, 11, Color(1, 1, 1, 0.9), false, HORIZONTAL_ALIGNMENT_CENTER)
	lb.add_child(cond)
	row.add_child(left)
	# 右侧信息
	var right := Kit.vbox(5)
	right.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	right.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	right.offset_left = 16.0
	right.offset_right = -12.0
	row.add_child(right)
	var title := Kit.label_sys(c.title, 15, DesignTokens.TEXT, true)
	right.add_child(title)
	var desc := Kit.label_sys(c.desc, 12, DesignTokens.TEXT_MUTED)
	right.add_child(desc)
	var expire := Kit.label_sys(c.expire_text, 11, DesignTokens.TEXT_FAINT)
	right.add_child(expire)
	# 去使用/已过期
	var state := Kit.outline_button("去使用" if c.usable else "已过期", Callable(), 28.0, 12, DesignTokens.PRIMARY if c.usable else DesignTokens.TEXT_FAINT)
	state.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	state.offset_left = 6.0
	state.offset_right = -14.0
	row.add_child(state)
	var hit := Button.new()
	hit.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	hit.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	hit.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	hit.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	hit.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	var coupon := c
	hit.pressed.connect(func() -> void: vm.use_tip(coupon))
	card.add_child(hit)
	return card

func bind() -> void:
	hub.bind_list(vm, "coupons", func(items: Array) -> void:
		for child in _list_box.get_children():
			child.queue_free()
		if items.size() == 0:
			_list_box.add_child(Kit.label_sys("暂无优惠券", 13, DesignTokens.TEXT_FAINT, false, HORIZONTAL_ALIGNMENT_CENTER))
			return
		for c in items:
			_list_box.add_child(_coupon_card(c))
	)
