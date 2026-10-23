class_name CountdownButton
extends Button
## 获取验证码按钮：剩余秒数由 VM 驱动（set_remaining），期间禁用并显示倒计时。

signal requested

var _remaining := 0
var _timer: Timer

func _init() -> void:
	text = "获取验证码"
	add_theme_font_size_override("font_size", 14)
	add_theme_color_override("font_color", DesignTokens.PRIMARY)
	add_theme_color_override("font_hover_color", DesignTokens.PRIMARY)
	add_theme_color_override("font_pressed_color", DesignTokens.PRIMARY)
	add_theme_color_override("font_disabled_color", DesignTokens.TEXT_FAINT)
	add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	pressed.connect(func() -> void: requested.emit())

## VM 驱动剩余秒数；<=0 恢复可点击。
func set_remaining(seconds: int) -> void:
	_remaining = maxi(seconds, 0)
	disabled = _remaining > 0
	text = str(_remaining) + "s 后重发" if _remaining > 0 else "获取验证码"
	if _remaining > 0:
		if _timer == null:
			_timer = Timer.new()
			_timer.timeout.connect(_on_tick)
			add_child(_timer)
		if not _timer.is_stopped():
			_timer.stop()
		_timer.start(1.0)
	elif _timer:
		_timer.stop()

func _on_tick() -> void:
	if _remaining <= 0:
		return
	set_remaining(_remaining - 1)
