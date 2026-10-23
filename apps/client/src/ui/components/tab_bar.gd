class_name AppTabBar
extends Control
## 底部 TabBar（64px）：首页/跑腿/订单/我的。

signal tab_selected(index: int)

var _items: Array = []
var _active := 0

func _init() -> void:
	custom_minimum_size = Vector2(0, DesignTokens.TABBAR_H)
	mouse_filter = Control.MOUSE_FILTER_PASS
	var bg := ColorRect.new()
	bg.color = Color.WHITE
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	bg.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(bg)
	var line := ColorRect.new()
	line.color = DesignTokens.BORDER
	line.offset_left = 0.0
	line.offset_right = 0.0
	line.offset_top = 0.0
	line.offset_bottom = 1.0
	line.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(line)
	var defs := [
		{"label": "首页", "icon": Icon.Kind.HOME},
		{"label": "跑腿", "icon": Icon.Kind.ERRAND},
		{"label": "订单", "icon": Icon.Kind.DOC},
		{"label": "我的", "icon": Icon.Kind.USER},
	]
	for i in defs.size():
		_items.append(_make_item(i, defs[i]))

func _make_item(index: int, def: Dictionary) -> Control:
	var box := Control.new()
	box.mouse_filter = Control.MOUSE_FILTER_PASS
	box.anchor_left = index / 4.0
	box.anchor_right = (index + 1) / 4.0
	box.offset_bottom = DesignTokens.TABBAR_H
	box.offset_top = 0.0
	var ic := Icon.new(int(def.icon), DesignTokens.TEXT_MUTED)
	ic.custom_minimum_size = Vector2(24, 24)
	ic.set_anchors_preset(Control.PRESET_CENTER_TOP)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.offset_left = -12.0
	ic.offset_right = 12.0
	ic.offset_top = 8.0
	ic.offset_bottom = 32.0
	box.add_child(ic)
	var l := Kit.label_sys(str(def.label), DesignTokens.FONT_XS, DesignTokens.TEXT_MUTED)
	l.set_anchors_preset(Control.PRESET_CENTER_TOP)
	l.anchor_left = 0.5
	l.anchor_right = 0.5
	l.offset_left = -50.0
	l.offset_right = 50.0
	l.offset_top = 34.0
	l.offset_bottom = 52.0
	l.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	box.add_child(l)
	var btn := Button.new()
	btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	var idx := index
	btn.pressed.connect(func() -> void:
		set_active(idx)
		tab_selected.emit(idx)
	)
	box.add_child(btn)
	add_child(box)
	return box

func set_active(index: int) -> void:
	_active = index
	for i in _items.size():
		var box: Control = _items[i]
		var ic: Icon = box.get_child(0)
		var l: Label = box.get_child(1)
		if i == index:
			ic.set_icon_color(DesignTokens.PRIMARY)
			l.add_theme_color_override("font_color", DesignTokens.PRIMARY)
		else:
			ic.set_icon_color(DesignTokens.TEXT_MUTED)
			l.add_theme_color_override("font_color", DesignTokens.TEXT_MUTED)
