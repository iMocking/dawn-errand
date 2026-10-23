extends CanvasLayer
## 轻提示 Toast（自动加载单例 Toast）：顶部悬浮 2 秒自动消失。

var _panel: PanelContainer
var _label: Label
var _anim: Tween

func _ready() -> void:
	layer = 200
	var holder := Control.new()
	holder.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	holder.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(holder)

	_panel = PanelContainer.new()
	_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_panel.visible = false
	var sb := StyleBoxFlat.new()
	sb.bg_color = Color(0.05, 0.06, 0.09, 0.82)
	sb.corner_radius_top_left = 12
	sb.corner_radius_top_right = 12
	sb.corner_radius_bottom_left = 12
	sb.corner_radius_bottom_right = 12
	sb.content_margin_left = 18.0
	sb.content_margin_right = 18.0
	sb.content_margin_top = 10.0
	sb.content_margin_bottom = 10.0
	_panel.add_theme_stylebox_override("panel", sb)
	_label = Label.new()
	_label.add_theme_font_size_override("font_size", 13)
	_label.add_theme_color_override("font_color", Color.WHITE)
	_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_label.custom_minimum_size = Vector2(0, 0)
	_panel.add_child(_label)
	# 顶部居中（避开状态栏）
	var anchor := Control.new()
	anchor.set_anchors_preset(Control.PRESET_CENTER_TOP)
	anchor.anchor_left = 0.0
	anchor.anchor_right = 1.0
	anchor.offset_top = 68.0
	anchor.mouse_filter = Control.MOUSE_FILTER_IGNORE
	anchor.add_child(_panel)
	holder.add_child(anchor)

func show_msg(text: String) -> void:
	if _anim:
		_anim.kill()
	_label.text = text
	_panel.visible = true
	_panel.modulate.a = 1.0
	# 重新测量尺寸
	_panel.reset_size()
	_anim = create_tween()
	_anim.tween_interval(1.8)
	_anim.tween_property(_panel, "modulate:a", 0.0, 0.35)
	_anim.tween_callback(func() -> void: _panel.visible = false)

func show_error(res: Dictionary) -> void:
	var msg := str(res.get("message", "操作失败"))
	var rid := str(res.get("request_id", ""))
	if rid != "" and rid != "":
		msg += "（" + rid + "）"
	show_msg(msg)
