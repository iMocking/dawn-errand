class_name NavBar
extends Control
## 顶部导航条（44px）：返回 + 居中标题 + 可选右侧操作。

signal back_pressed
signal right_pressed

func _init(title: String, on_back: Callable = Callable(), right_text: String = "", on_right: Callable = Callable()) -> void:
	custom_minimum_size = Vector2(0, DesignTokens.NAV_H)
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	var bg := ColorRect.new()
	bg.color = Color.WHITE
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	bg.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(bg)
	var line := ColorRect.new()
	line.color = DesignTokens.DIVIDER
	line.anchor_top = 1.0
	line.anchor_bottom = 1.0
	line.offset_top = -1.0
	line.offset_bottom = 0.0
	line.offset_left = 0.0
	line.offset_right = 0.0
	line.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(line)
	var title_l := Kit.label_sys(title, DesignTokens.FONT_H1, DesignTokens.TEXT, true)
	title_l.set_anchors_preset(Control.PRESET_CENTER_TOP)
	title_l.anchor_left = 0.5
	title_l.anchor_right = 0.5
	title_l.offset_left = -100.0
	title_l.offset_right = 100.0
	title_l.offset_top = 0.0
	title_l.offset_bottom = DesignTokens.NAV_H
	title_l.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	add_child(title_l)
	if on_back.is_valid():
		var back := Kit.icon_button(Icon.Kind.BACK, func() -> void: back_pressed.emit(), Color.WHITE, DesignTokens.TEXT, 36.0)
		back.anchor_left = 0.0
		back.offset_left = 16.0
		back.offset_top = 4.0
		back.offset_bottom = 40.0
		add_child(back)
	if right_text != "":
		var r := Button.new()
		r.text = right_text
		r.add_theme_font_size_override("font_size", 15)
		r.add_theme_color_override("font_color", DesignTokens.PRIMARY)
		r.add_theme_color_override("font_hover_color", DesignTokens.PRIMARY)
		r.add_theme_color_override("font_pressed_color", DesignTokens.PRIMARY)
		r.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
		r.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
		r.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
		r.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
		r.anchor_right = 1.0
		r.offset_left = -80.0
		r.offset_right = -16.0
		r.offset_top = 0.0
		r.offset_bottom = DesignTokens.NAV_H
		r.pressed.connect(func() -> void: right_pressed.emit())
		add_child(r)

func _process(_delta: float) -> void:
	pass
