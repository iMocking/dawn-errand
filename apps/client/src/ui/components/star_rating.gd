class_name StarRating
extends Control
## 星级评分（横向 5 星，可点击）。

signal changed(v: int)

var _value := 5
var _size := 32.0

func _init(value := 5, size := 32.0) -> void:
	_value = value
	_size = size
	mouse_filter = Control.MOUSE_FILTER_PASS
	custom_minimum_size = Vector2(size * 5 + 8 * 4, size)

func _ready() -> void:
	_redraw()

func set_value(v: int) -> void:
	_value = clampi(v, 0, 5)
	changed.emit(_value)
	_redraw()

func _redraw() -> void:
	for child in get_children():
		child.queue_free()
	var row := Kit.hbox(8)
	row.mouse_filter = Control.MOUSE_FILTER_PASS
	add_child(row)
	for i in 5:
		var idx := i
		var ic := Icon.new(Icon.Kind.STAR, DesignTokens.STAR if idx < _value else DesignTokens.TEXT_FAINT)
		ic.custom_minimum_size = Vector2(_size, _size)
		row.add_child(ic)
		var btn := Button.new()
		btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
		btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
		btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
		btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
		btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
		btn.pressed.connect(func() -> void: set_value(idx + 1))
		# 覆盖在整个评分条上，按 1/5 宽度分区
		btn.anchor_left = idx / 5.0
		btn.anchor_right = (idx + 1) / 5.0
		btn.anchor_top = 0.0
		btn.anchor_bottom = 1.0
		add_child(btn)
