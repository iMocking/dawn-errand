class_name OrderTrackView
extends ScreenBase
## 订单追踪视图：地图区域、配送状态卡、进度节点、骑手卡、操作按钮。

var _steps_box: VBoxContainer

func build() -> void:
	var bg := ColorRect.new()
	bg.color = Color.WHITE
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	var nav := NavBar.new("订单追踪", func() -> void: App.go_back())
	add_child(nav)

	# 地图区域（示意）
	var map := Control.new()
	map.anchor_left = 0.0
	map.anchor_right = 1.0
	map.anchor_top = 0.0
	map.offset_top = DesignTokens.NAV_H
	map.offset_bottom = 296.0
	add_child(map)
	var map_bg := ColorRect.new()
	map_bg.color = Color("EAF4EC")
	map_bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	map.add_child(map_bg)
	var road_color := Color("D8E9DC")
	map.draw.connect(func() -> void:
		map.draw_line(Vector2(0, 150), Vector2(375, 90), road_color, 26.0)
		map.draw_line(Vector2(60, 260), Vector2(330, 60), road_color, 20.0)
		map.draw_line(Vector2(40, 60), Vector2(300, 240), road_color, 14.0)
		map.draw_line(Vector2(120, 20), Vector2(200, 280), road_color, 10.0)
	)
	# 商家 / 目的地 / 骑手 标记
	var shop_pin := _pin(DesignTokens.PRIMARY, Icon.Kind.BAG)
	shop_pin.position = Vector2(60, 210)
	map.add_child(shop_pin)
	var dest_pin := _pin(DesignTokens.SUCCESS, Icon.Kind.LOCATION)
	dest_pin.position = Vector2(280, 80)
	map.add_child(dest_pin)
	var rider_dot := _rider()
	rider_dot.position = Vector2(170, 130)
	map.add_child(rider_dot)

	# 状态卡
	var status_card := Kit.card(Color.WHITE, 14.0, true, true)
	status_card.anchor_left = 0.0
	status_card.anchor_right = 1.0
	status_card.offset_left = 20.0
	status_card.offset_right = -20.0
	status_card.offset_top = 250.0
	status_card.offset_bottom = 330.0
	add_child(status_card)
	var s_box := Kit.hbox(0)
	s_box.offset_left = 16.0
	s_box.offset_right = -16.0
	status_card.add_child(s_box)
	var pulse := PanelContainer.new()
	pulse.custom_minimum_size = Vector2(40, 40)
	pulse.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	pulse.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY_BG, 20.0))
	var p_ic := Icon.new(Icon.Kind.SCOOTER, DesignTokens.PRIMARY, 2.0)
	p_ic.custom_minimum_size = Vector2(22, 22)
	p_ic.set_anchors_preset(Control.PRESET_CENTER)
	p_ic.anchor_left = 0.5
	p_ic.anchor_right = 0.5
	p_ic.anchor_top = 0.5
	p_ic.anchor_bottom = 0.5
	p_ic.offset_left = -11.0
	p_ic.offset_right = 11.0
	p_ic.offset_top = -11.0
	p_ic.offset_bottom = 11.0
	pulse.add_child(p_ic)
	s_box.add_child(pulse)
	var s_info := Kit.vbox(4)
	s_info.offset_left = 12.0
	s_info.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	s_box.add_child(s_info)
	var s_t := Kit.label_sys(str(vm.get_prop("status")), 17, DesignTokens.TEXT, true)
	s_info.add_child(s_t)
	var s_s := Kit.label_sys(str(vm.get_prop("shop_name")) + " · " + str(vm.get_prop("eta")), 12, DesignTokens.TEXT_MUTED)
	s_info.add_child(s_s)

	# 进度节点
	var steps_card := Kit.card(Color.WHITE, 14.0, false, true)
	steps_card.anchor_left = 0.0
	steps_card.anchor_right = 1.0
	steps_card.offset_left = 20.0
	steps_card.offset_right = -20.0
	steps_card.offset_top = 342.0
	steps_card.offset_bottom = 470.0
	add_child(steps_card)
	_steps_box = Kit.vbox(0)
	_steps_box.offset_left = 16.0
	_steps_box.offset_right = -16.0
	_steps_box.offset_top = 10.0
	_steps_box.offset_bottom = -10.0
	steps_card.add_child(_steps_box)

	# 骑手卡
	var rider_card := Kit.card(Color.WHITE, 14.0, false, true)
	rider_card.anchor_left = 0.0
	rider_card.anchor_right = 1.0
	rider_card.offset_left = 20.0
	rider_card.offset_right = -20.0
	rider_card.offset_top = 482.0
	rider_card.offset_bottom = 580.0
	add_child(rider_card)
	var r_box := Kit.hbox(12)
	r_box.offset_left = 16.0
	r_box.offset_right = -16.0
	rider_card.add_child(r_box)
	var avatar := PanelContainer.new()
	avatar.custom_minimum_size = Vector2(48, 48)
	avatar.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	avatar.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY, 24.0))
	var a_ic := Icon.new(Icon.Kind.RUNNER, Color.WHITE, 2.0)
	a_ic.custom_minimum_size = Vector2(26, 26)
	a_ic.set_anchors_preset(Control.PRESET_CENTER)
	a_ic.anchor_left = 0.5
	a_ic.anchor_right = 0.5
	a_ic.anchor_top = 0.5
	a_ic.anchor_bottom = 0.5
	a_ic.offset_left = -13.0
	a_ic.offset_right = 13.0
	a_ic.offset_top = -13.0
	a_ic.offset_bottom = 13.0
	avatar.add_child(a_ic)
	r_box.add_child(avatar)
	var r_info := Kit.vbox(4)
	r_info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	r_info.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	r_box.add_child(r_info)
	var r_name := Kit.label_sys(str(vm.get_prop("rider_name")) + " · " + str(vm.get_prop("rider_plate")), 15, DesignTokens.TEXT, true)
	r_info.add_child(r_name)
	var r_phone := Kit.label_sys(str(vm.get_prop("rider_phone")), 13, DesignTokens.TEXT_MUTED)
	r_info.add_child(r_phone)
	var call := Kit.icon_button(Icon.Kind.PHONE, func() -> void: vm.call_rider(), DesignTokens.PRIMARY_BG, DesignTokens.PRIMARY, 40.0)
	call.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	r_box.add_child(call)

	# 底部操作
	var bottom := PanelContainer.new()
	bottom.anchor_left = 0.0
	bottom.anchor_right = 1.0
	bottom.anchor_top = 1.0
	bottom.anchor_bottom = 1.0
	bottom.offset_top = -64.0
	bottom.offset_bottom = 0.0
	bottom.add_theme_stylebox_override("panel", Kit.style_flat(Color.WHITE, 0.0))
	add_child(bottom)
	var b_row := Kit.hbox(12)
	b_row.offset_left = 20.0
	b_row.offset_right = -20.0
	bottom.add_child(b_row)
	var service := Kit.outline_button("联系客服", func() -> void: vm.contact_service(), 44.0, 14)
	service.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	b_row.add_child(service)
	var confirm := Kit.primary_button("确认送达", func() -> void: Toast.show_msg("感谢确认！订单已完成"), 44.0, 22.0, 14)
	confirm.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	b_row.add_child(confirm)

func _pin(color: Color, kind: int) -> Control:
	var c := Control.new()
	c.custom_minimum_size = Vector2(36, 36)
	c.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var col = color
	c.draw.connect(func() -> void:
		c.draw_circle(Vector2(18, 18), 18.0, Color(1, 1, 1, 0.85))
		c.draw_circle(Vector2(18, 18), 14.0, col)
	)
	var ic := Icon.new(kind, Color.WHITE, 1.8)
	ic.custom_minimum_size = Vector2(16, 16)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -8.0
	ic.offset_right = 8.0
	ic.offset_top = -8.0
	ic.offset_bottom = 8.0
	c.add_child(ic)
	return c

func _rider() -> Control:
	var c := Control.new()
	c.custom_minimum_size = Vector2(44, 44)
	c.mouse_filter = Control.MOUSE_FILTER_IGNORE
	c.draw.connect(func() -> void:
		c.draw_circle(Vector2(22, 22), 22.0, Color(1, 1, 1, 0.9))
		c.draw_circle(Vector2(22, 22), 17.0, DesignTokens.PRIMARY)
	)
	var ic := Icon.new(Icon.Kind.SCOOTER, Color.WHITE, 2.0)
	ic.custom_minimum_size = Vector2(20, 20)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -10.0
	ic.offset_right = 10.0
	ic.offset_top = -10.0
	ic.offset_bottom = 10.0
	c.add_child(ic)
	return c

func _step_row(text: String, active: bool, done: bool) -> Control:
	var row := Kit.hbox(10)
	row.custom_minimum_size = Vector2(0, 40)
	var dot := PanelContainer.new()
	dot.custom_minimum_size = Vector2(22, 22)
	dot.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	if done:
		dot.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.SUCCESS, 11.0))
	elif active:
		dot.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY, 11.0))
	else:
		dot.add_theme_stylebox_override("panel", Kit.style_flat(Color("E5E6EB"), 11.0))
	var dot_ic := Icon.new(Icon.Kind.CHECK if done else (Icon.Kind.SCOOTER if active else Icon.Kind.MINUS), Color.WHITE, 1.6)
	dot_ic.custom_minimum_size = Vector2(12, 12)
	dot_ic.set_anchors_preset(Control.PRESET_CENTER)
	dot_ic.anchor_left = 0.5
	dot_ic.anchor_right = 0.5
	dot_ic.anchor_top = 0.5
	dot_ic.anchor_bottom = 0.5
	dot_ic.offset_left = -6.0
	dot_ic.offset_right = 6.0
	dot_ic.offset_top = -6.0
	dot_ic.offset_bottom = 6.0
	dot.add_child(dot_ic)
	row.add_child(dot)
	var l := Kit.label_sys(text, 13, DesignTokens.TEXT if (done or active) else DesignTokens.TEXT_FAINT, done or active)
	row.add_child(l)
	return row

func bind() -> void:
	hub.bind_list(vm, "steps", func(items: Array) -> void:
		for child in _steps_box.get_children():
			child.queue_free()
		var idx := int(vm.get_prop("step_index"))
		for i in items.size():
			_steps_box.add_child(_step_row(str(items[i]), i == idx, i < idx))
	)
