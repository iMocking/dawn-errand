class_name OrdersView
extends ScreenBase
## 我的订单视图：筛选 + 订单卡片列表 + 操作按钮。

var _filter_chips: Array = []
var _list_box: VBoxContainer
var _empty: Control

func build() -> void:
	var bg := ColorRect.new()
	bg.color = DesignTokens.BG
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(bg)
	# 标题
	var title := Kit.label_sys("我的订单", 18, DesignTokens.TEXT, true)
	title.set_anchors_preset(Control.PRESET_TOP_WIDE)
	title.offset_left = 20.0
	title.offset_right = -20.0
	title.offset_top = DesignTokens.STATUS_BAR_H + 6.0
	title.offset_bottom = DesignTokens.STATUS_BAR_H + 40.0
	add_child(title)
	# 筛选
	var filter_row := Kit.hbox(8)
	filter_row.anchor_left = 0.0
	filter_row.anchor_right = 1.0
	filter_row.offset_left = 20.0
	filter_row.offset_right = -20.0
	filter_row.offset_top = DesignTokens.STATUS_BAR_H + 44.0
	filter_row.offset_bottom = DesignTokens.STATUS_BAR_H + 80.0
	add_child(filter_row)
	for f in vm.list("filters").items():
		var b := Button.new()
		b.text = str(f)
		b.custom_minimum_size = Vector2(0, 32)
		b.add_theme_font_size_override("font_size", 14)
		b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
		var label := str(f)
		b.pressed.connect(func() -> void: vm.set_filter(label))
		filter_row.add_child(b)
		_filter_chips.append(b)

	# 列表
	var sc := ScrollContainer.new()
	sc.anchor_left = 0.0
	sc.anchor_right = 1.0
	sc.anchor_top = 0.0
	sc.anchor_bottom = 1.0
	sc.offset_top = DesignTokens.STATUS_BAR_H + 88.0
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	add_child(sc)
	_list_box = Kit.vbox(12)
	_list_box.offset_left = 16.0
	_list_box.offset_right = -16.0
	_list_box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sc.add_child(_list_box)

	# 空态
	_empty = Control.new()
	_empty.visible = false
	_empty.set_anchors_preset(Control.PRESET_CENTER)
	_empty.anchor_left = 0.0
	_empty.anchor_right = 1.0
	_empty.anchor_top = 0.0
	_empty.anchor_bottom = 1.0
	add_child(_empty)
	var empty_box := Kit.vbox(10)
	empty_box.set_anchors_preset(Control.PRESET_CENTER)
	empty_box.anchor_left = 0.5
	empty_box.anchor_right = 0.5
	empty_box.anchor_top = 0.5
	empty_box.anchor_bottom = 0.5
	empty_box.offset_left = -100.0
	empty_box.offset_right = 100.0
	empty_box.offset_top = -60.0
	empty_box.offset_bottom = 60.0
	_empty.add_child(empty_box)
	var e_ic := Icon.new(Icon.Kind.DOC, DesignTokens.TEXT_FAINT, 2.0)
	e_ic.custom_minimum_size = Vector2(48, 48)
	e_ic.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	empty_box.add_child(e_ic)
	empty_box.add_child(Kit.label_sys("暂无相关订单", 14, DesignTokens.TEXT_MUTED, false, HORIZONTAL_ALIGNMENT_CENTER))

func _order_card(o: OrderModel) -> Control:
	var card := Kit.card(Color.WHITE, 14.0, false, true)
	var box := Kit.vbox(0)
	box.offset_left = 14.0
	box.offset_right = -14.0
	box.offset_top = 12.0
	box.offset_bottom = -12.0
	card.add_child(box)
	# 头部：商家 + 状态
	var head := Kit.hbox(0)
	box.add_child(head)
	var shop := Kit.label_sys(o.shop, 15, DesignTokens.TEXT, true)
	shop.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	head.add_child(shop)
	var status := Kit.label_sys(o.status, 13, o.status_color(), true)
	head.add_child(status)
	box.add_child(Kit.spacing(0, 6))
	# 商品摘要
	var goods := Kit.label_sys(o.goods_summary, 13, DesignTokens.TEXT_SEC)
	box.add_child(goods)
	box.add_child(Kit.spacing(0, 4))
	var time := Kit.label_sys(o.time_text, 11, DesignTokens.TEXT_MUTED)
	box.add_child(time)
	box.add_child(Kit.spacing(0, 6))
	# 金额 + 操作
	var foot := Kit.hbox(8)
	box.add_child(foot)
	var total := Kit.label_sys("¥" + Fmt.money(o.total), 16, DesignTokens.PRICE, true)
	total.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	foot.add_child(total)
	for action in o.actions:
		var b := Kit.outline_button(str(action), Callable(), 30.0, 12)
		var act := str(action)
		b.pressed.connect(func() -> void: vm.order_action(o, act))
		foot.add_child(b)
	return card

func bind() -> void:
	hub.bind(vm, "filter", func(v) -> void:
		var opts: Array = vm.list("filters").items()
		var idx := 0
		for i in opts.size():
			if str(opts[i]) == str(v):
				idx = i
				break
		for i in _filter_chips.size():
			var b: Button = _filter_chips[i]
			var active := i == idx
			b.add_theme_color_override("font_color", DesignTokens.PRIMARY if active else DesignTokens.TEXT_SEC)
			b.add_theme_color_override("font_hover_color", DesignTokens.PRIMARY if active else DesignTokens.TEXT_SEC)
			b.add_theme_stylebox_override("normal", Kit.style_flat(Color.WHITE if active else Color("F5F6F8"), 16.0))
			b.add_theme_stylebox_override("hover", Kit.style_flat(Color.WHITE if active else Color("F5F6F8"), 16.0))
			b.add_theme_stylebox_override("pressed", Kit.style_flat(Color.WHITE if active else Color("F5F6F8"), 16.0))
	)
	hub.bind_list(vm, "orders", _render_orders)

func _render_orders(items: Array) -> void:
	for child in _list_box.get_children():
		child.queue_free()
	var orders: Array = items
	if orders.size() == 0:
		_empty.visible = true
		_list_box.visible = false
		return
	_empty.visible = false
	_list_box.visible = true
	for o in orders:
		_list_box.add_child(_order_card(o))
