class_name Kit
extends RefCounted
## UI 工厂：以设计规范快速构造控件（颜色/字号/圆角统一取 DesignTokens）。

## 圆角纯色面板样式
static func style_flat(bg: Color, radius: float, border_color: Color = Color.TRANSPARENT, border_w: float = 0.0, shadow: bool = false) -> StyleBoxFlat:
	var sb := StyleBoxFlat.new()
	sb.bg_color = bg
	sb.set_corner_radius_all(int(radius))
	if border_w > 0:
		sb.border_color = border_color
		sb.set_border_width_all(int(border_w))
	if shadow:
		sb.shadow_color = Color(0, 0, 0, 0.08)
		sb.shadow_size = 6
	return sb

## 构造 Label。默认不自动换行：中文逐字折行的竖排问题即由全局 WORD_SMART 引起，
## 需要自动换行的文本显式传 wrap=true（否则容器会把 Label 压缩到单字宽）。
static func label(text: String, size: int = DesignTokens.FONT_XL, color: Color = DesignTokens.TEXT, bold: bool = false, align: int = HORIZONTAL_ALIGNMENT_LEFT, wrap: bool = false) -> Label:
	var l := Label.new()
	l.text = text
	l.add_theme_font_size_override("font_size", size)
	l.add_theme_color_override("font_color", color)
	if bold:
		l.add_theme_font_override("font", _bold_font(size))
	l.horizontal_alignment = align
	l.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	if wrap:
		l.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	return l

static func _bold_font(size: int) -> Font:
	var f := SystemFont.new()
	f.font_names = PackedStringArray(["Microsoft YaHei UI", "Microsoft YaHei", "PingFang SC", "Noto Sans CJK SC"])
	f.font_weight = 700
	return f

## 使用系统字体构造 Label（中文字体一致渲染）
static func label_sys(text: String, size: int = DesignTokens.FONT_XL, color: Color = DesignTokens.TEXT, bold: bool = false, align: int = HORIZONTAL_ALIGNMENT_LEFT, wrap: bool = false) -> Label:
	var l := label(text, size, color, bold, align, wrap)
	var f := SystemFont.new()
	f.font_names = PackedStringArray(["Microsoft YaHei UI", "Microsoft YaHei", "PingFang SC", "Noto Sans CJK SC"])
	if bold:
		f.font_weight = 700
	l.add_theme_font_override("font", f)
	return l

static func vbox(spacing: int = 0, margin: Vector2 = Vector2.ZERO) -> VBoxContainer:
	var c := VBoxContainer.new()
	c.add_theme_constant_override("separation", spacing)
	c.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	c.mouse_filter = Control.MOUSE_FILTER_PASS
	return c

static func hbox(spacing: int = 0) -> HBoxContainer:
	var c := HBoxContainer.new()
	c.add_theme_constant_override("separation", spacing)
	c.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	c.mouse_filter = Control.MOUSE_FILTER_PASS
	return c

## 圆角卡片。expand=true 时铺满父容器宽度（VBox 直排卡片必须传 true，
## 否则 PanelContainer 按内容最小宽收缩，内部文字会被压缩成竖排）。
static func card(bg: Color = Color.WHITE, radius: float = DesignTokens.RADIUS_MD, shadow: bool = false, expand: bool = false) -> PanelContainer:
	var p := PanelContainer.new()
	p.add_theme_stylebox_override("panel", style_flat(bg, radius, Color.TRANSPARENT, 0.0, shadow))
	p.mouse_filter = Control.MOUSE_FILTER_PASS
	if expand:
		p.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	return p

## 主按钮（渐变圆角）
static func primary_button(text: String, cb: Callable, height: float = 48.0, radius: float = 24.0, font_size: int = 16) -> Button:
	var b := Button.new()
	b.text = text
	b.custom_minimum_size = Vector2(0, height)
	b.add_theme_font_size_override("font_size", font_size)
	b.add_theme_color_override("font_color", Color.WHITE)
	b.add_theme_color_override("font_hover_color", Color.WHITE)
	b.add_theme_color_override("font_pressed_color", Color.WHITE)
	b.add_theme_color_override("font_disabled_color", Color(1, 1, 1, 0.6))
	b.add_theme_stylebox_override("normal", style_flat(DesignTokens.PRIMARY, radius))
	b.add_theme_stylebox_override("hover", style_flat(DesignTokens.PRIMARY_DARK, radius))
	b.add_theme_stylebox_override("pressed", style_flat(DesignTokens.PRIMARY_DARK, radius))
	b.add_theme_stylebox_override("disabled", style_flat(Color("FFB385"), radius))
	b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	if cb.is_valid():
		b.pressed.connect(cb)
	return b

## 描边按钮
static func outline_button(text: String, cb: Callable, height: float = 34.0, font_size: int = 13, color: Color = DesignTokens.TEXT_SEC, radius: float = 16.0) -> Button:
	var b := Button.new()
	b.text = text
	b.custom_minimum_size = Vector2(0, height)
	b.add_theme_font_size_override("font_size", font_size)
	b.add_theme_color_override("font_color", color)
	b.add_theme_color_override("font_hover_color", color)
	b.add_theme_color_override("font_pressed_color", color)
	var sb := style_flat(Color.WHITE, radius, DesignTokens.BORDER, 1.0)
	b.add_theme_stylebox_override("normal", sb)
	b.add_theme_stylebox_override("hover", sb)
	b.add_theme_stylebox_override("pressed", sb)
	b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	if cb.is_valid():
		b.pressed.connect(cb)
	return b

static func divider(height: float = 1.0, color: Color = DesignTokens.DIVIDER) -> ColorRect:
	var r := ColorRect.new()
	r.color = color
	r.custom_minimum_size = Vector2(0, height)
	r.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return r

static func spacing(w: float = 0.0, h: float = 0.0) -> Control:
	var c := Control.new()
	c.custom_minimum_size = Vector2(w, h)
	c.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return c

## 渐变背景矩形（垂直渐变）
static func gradient_rect(top: Color, bottom: Color) -> TextureRect:
	var tex := GradientTexture2D.new()
	var g := Gradient.new()
	g.set_color(0, top)
	g.set_color(1, bottom)
	tex.gradient = g
	tex.fill_from = Vector2(0, 0)
	tex.fill_to = Vector2(0, 1)
	var tr := TextureRect.new()
	tr.texture = tex
	tr.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_COVERED
	tr.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return tr

## 标签 chip（小圆角胶囊）
static func chip(text: String, active: bool = false, active_color: Color = DesignTokens.PRIMARY) -> PanelContainer:
	var p := PanelContainer.new()
	var bg := DesignTokens.PRIMARY_BG if active else DesignTokens.BG
	var fg := active_color if active else DesignTokens.TEXT_SEC
	var sb := style_flat(bg, 16.0)
	sb.content_margin_left = 12.0
	sb.content_margin_right = 12.0
	sb.content_margin_top = 5.0
	sb.content_margin_bottom = 5.0
	p.add_theme_stylebox_override("panel", sb)
	var l := label_sys(text, DesignTokens.FONT_MD, fg)
	p.add_child(l)
	p.mouse_filter = Control.MOUSE_FILTER_IGNORE
	return p

static func scroll(container: Control) -> ScrollContainer:
	var sc := ScrollContainer.new()
	sc.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	sc.add_child(container)
	return sc

static func set_full_rect(c: Control) -> void:
	c.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

## 图标按钮（圆形底）
static func icon_button(icon_kind: int, cb: Callable, bg: Color = Color.WHITE, icon_color: Color = DesignTokens.TEXT, size: float = 36.0) -> Button:
	var b := Button.new()
	b.custom_minimum_size = Vector2(size, size)
	b.add_theme_stylebox_override("normal", style_flat(bg, size / 2.0))
	b.add_theme_stylebox_override("hover", style_flat(bg, size / 2.0))
	b.add_theme_stylebox_override("pressed", style_flat(bg.darkened(0.06), size / 2.0))
	b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	var ic := Icon.new(icon_kind, icon_color)
	ic.custom_minimum_size = Vector2(size * 0.5, size * 0.5)
	ic.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	b.add_child(ic)
	if cb.is_valid():
		b.pressed.connect(cb)
	return b
