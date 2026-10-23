class_name StatusBar
extends Control
## 手机状态栏（44px）：左侧时间，右侧信号/WiFi 示意。

var _color: Color = Color.WHITE

func _init(color: Color = Color.WHITE) -> void:
	_color = color
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	custom_minimum_size = Vector2(0, DesignTokens.STATUS_BAR_H)

func _ready() -> void:
	var row := Kit.hbox(6)
	row.anchor_left = 0.0
	row.anchor_right = 1.0
	row.offset_left = 24.0
	row.offset_right = -24.0
	row.offset_top = 0.0
	row.offset_bottom = DesignTokens.STATUS_BAR_H
	add_child(row)
	var t := Kit.label_sys("9:41", 14, _color, true)
	t.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(t)
	var bars := _bars()
	bars.custom_minimum_size = Vector2(18, 11)
	row.add_child(bars)
	var wifi := _wifi()
	wifi.custom_minimum_size = Vector2(18, 12)
	row.add_child(wifi)

func _bars() -> Control:
	var c := Control.new()
	c.mouse_filter = Control.MOUSE_FILTER_IGNORE
	c.draw.connect(func() -> void:
		for i in 4:
			var h := 3.0 + i * 2.2
			var r := Rect2(Vector2(i * 4.0, 10.5 - h), Vector2(3.0, h))
			c.draw_rect(r, _color)
	)
	c.resized.connect(func() -> void: c.queue_redraw())
	c.ready.connect(func() -> void: c.queue_redraw())
	return c

func _wifi() -> Control:
	var c := Control.new()
	c.mouse_filter = Control.MOUSE_FILTER_IGNORE
	c.draw.connect(func() -> void:
		var w := c.size.x
		var pts := PackedVector2Array()
		pts.append(Vector2(w * 0.5, 10.0))
		pts.append(Vector2(0.0, 3.0))
		pts.append(Vector2(w, 3.0))
		pts.append(Vector2(w * 0.5, 10.0))
		c.draw_polyline(pts, _color, 1.8, true)
		c.draw_rect(Rect2(Vector2(w * 0.42, 8.0), Vector2(w * 0.16, 3.5)), _color)
	)
	c.resized.connect(func() -> void: c.queue_redraw())
	c.ready.connect(func() -> void: c.queue_redraw())
	return c
