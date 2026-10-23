class_name LoginView
extends ScreenBase
## 登录页视图：渐变头部 + 白色卡片，验证码登录 / 密码登录双 Tab。
## MVVM：输入回写 VM（set_*），状态渲染全部走 hub 绑定。

var _tab_sms: Button
var _tab_pwd: Button
var _sms_box: Control
var _pwd_box: Control
var _phone_input: LineEdit
var _code_input: LineEdit
var _account_input: LineEdit
var _password_input: LineEdit
var _code_btn: CountdownButton
var _login_btn: Button
var _agree_icon: Icon
var _agree_panel: PanelContainer

func build() -> void:
	# ---- 渐变头部 ----
	var header := Kit.gradient_rect(DesignTokens.PRIMARY_LIGHT, DesignTokens.PRIMARY_DARK)
	header.anchor_right = 1.0
	header.offset_bottom = 216.0
	add_child(header)
	var sb := StatusBar.new(Color(1, 1, 1, 0.95))
	sb.set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	add_child(sb)

	# 头部品牌区
	var logo := Control.new()
	logo.custom_minimum_size = Vector2(64, 64)
	logo.mouse_filter = Control.MOUSE_FILTER_IGNORE
	logo.position = Vector2(28, 84)
	logo.draw.connect(func() -> void:
		logo.draw_circle(Vector2(32, 32), 32.0, Color(1, 1, 1, 0.98))
	)
	var logo_icon := Icon.new(Icon.Kind.ERRAND, DesignTokens.PRIMARY)
	logo_icon.custom_minimum_size = Vector2(40, 40)
	logo_icon.set_anchors_preset(Control.PRESET_CENTER)
	logo_icon.anchor_left = 0.5
	logo_icon.anchor_right = 0.5
	logo_icon.anchor_top = 0.5
	logo_icon.anchor_bottom = 0.5
	logo_icon.offset_left = -20.0
	logo_icon.offset_right = 20.0
	logo_icon.offset_top = -20.0
	logo_icon.offset_bottom = 20.0
	logo.add_child(logo_icon)
	add_child(logo)
	var brand := Kit.label_sys("曙光跑腿", 26, Color.WHITE, true)
	brand.position = Vector2(104, 84)
	brand.custom_minimum_size = Vector2(200, 34)
	add_child(brand)
	var slogan := Kit.label_sys("乡镇生活，随叫随到", 12, Color(1, 1, 1, 0.9))
	slogan.position = Vector2(106, 120)
	slogan.custom_minimum_size = Vector2(200, 20)
	add_child(slogan)

	# ---- 白色卡片 ----
	var card := Kit.card(Color.WHITE, 24.0)
	card.anchor_left = 0.0
	card.anchor_right = 1.0
	card.anchor_top = 0.0
	card.anchor_bottom = 1.0
	card.offset_top = 188.0
	card.offset_bottom = 0.0
	add_child(card)

	var inner := Kit.vbox(0)
	inner.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	inner.offset_left = 24.0
	inner.offset_right = -24.0
	inner.offset_top = 20.0
	inner.offset_bottom = -24.0
	card.add_child(inner)

	# 登录方式 Tab
	inner.add_child(_build_tabs())
	inner.add_child(Kit.spacing(0, 18))

	# 验证码模式字段
	_sms_box = Kit.vbox(14)
	_sms_box.visible = true
	_phone_input = _field(_sms_box, Icon.Kind.PHONE, "请输入手机号", false)
	_sms_box.add_child(_build_code_row())
	inner.add_child(_sms_box)

	# 密码模式字段
	_pwd_box = Kit.vbox(14)
	_pwd_box.visible = false
	_account_input = _field(_pwd_box, Icon.Kind.MAIL, "邮箱或手机号", false)
	_password_input = _field(_pwd_box, Icon.Kind.LOCK, "请输入密码（不少于6位）", true)
	inner.add_child(_pwd_box)

	inner.add_child(Kit.spacing(0, 18))

	# 登录按钮
	_login_btn = Kit.primary_button("登录 / 注册", Callable(), 52.0, 26.0, 16)
	_login_btn.pressed.connect(func() -> void: vm.login())
	inner.add_child(_login_btn)

	# 协议行
	inner.add_child(Kit.spacing(0, 12))
	inner.add_child(_build_agree_row())
	inner.add_child(Kit.spacing(0, 14))

	# 其他登录方式
	inner.add_child(_build_other_login())

func _build_tabs() -> Control:
	var seg := PanelContainer.new()
	seg.add_theme_stylebox_override("panel", Kit.style_flat(Color("F5F6F8"), 10.0))
	seg.custom_minimum_size = Vector2(0, 44)
	var row := Kit.hbox(4)
	row.offset_left = 6.0
	row.offset_right = -6.0
	row.offset_top = 6.0
	row.offset_bottom = -6.0
	seg.add_child(row)
	_tab_sms = _tab_button("验证码登录", true)
	_tab_pwd = _tab_button("密码登录", false)
	row.add_child(_tab_sms)
	row.add_child(_tab_pwd)
	return seg

func _tab_button(text: String, active: bool) -> Button:
	var b := Button.new()
	b.text = text
	b.custom_minimum_size = Vector2(0, 32)
	b.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	b.add_theme_font_size_override("font_size", 15)
	_set_tab_style(b, active)
	return b

func _set_tab_style(b: Button, active: bool) -> void:
	b.add_theme_color_override("font_color", DesignTokens.TEXT if active else DesignTokens.TEXT_MUTED)
	b.add_theme_color_override("font_hover_color", DesignTokens.TEXT if active else DesignTokens.TEXT_SEC)
	b.add_theme_color_override("font_pressed_color", DesignTokens.TEXT)
	if active:
		b.add_theme_stylebox_override("normal", Kit.style_flat(Color.WHITE, 8.0, Color.TRANSPARENT, 0.0, true))
		b.add_theme_stylebox_override("hover", Kit.style_flat(Color.WHITE, 8.0, Color.TRANSPARENT, 0.0, true))
	else:
		b.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
		b.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	b.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	b.add_theme_stylebox_override("focus", StyleBoxEmpty.new())

func _field(parent: Control, icon_kind: int, placeholder: String, secret: bool) -> LineEdit:
	var row := PanelContainer.new()
	row.custom_minimum_size = Vector2(0, 52)
	row.add_theme_stylebox_override("panel", Kit.style_flat(Color("F5F6F8"), 14.0))
	var box := Kit.hbox(10)
	box.offset_left = 16.0
	box.offset_right = -16.0
	box.offset_top = 0.0
	box.offset_bottom = 0.0
	row.add_child(box)
	var ic := Icon.new(icon_kind, DesignTokens.TEXT_MUTED, 1.8)
	ic.custom_minimum_size = Vector2(20, 20)
	ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	box.add_child(ic)
	var edit := LineEdit.new()
	edit.placeholder_text = placeholder
	edit.secret = secret
	edit.add_theme_font_size_override("font_size", 15)
	edit.add_theme_color_override("font_color", DesignTokens.TEXT)
	edit.add_theme_color_override("font_placeholder_color", DesignTokens.TEXT_FAINT)
	edit.add_theme_color_override("caret_color", DesignTokens.PRIMARY)
	edit.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	edit.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	edit.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	box.add_child(edit)
	parent.add_child(row)
	return edit

func _build_code_row() -> Control:
	var row := PanelContainer.new()
	row.custom_minimum_size = Vector2(0, 52)
	row.add_theme_stylebox_override("panel", Kit.style_flat(Color("F5F6F8"), 14.0))
	var box := Kit.hbox(10)
	box.offset_left = 16.0
	box.offset_right = -6.0
	row.add_child(box)
	var ic := Icon.new(Icon.Kind.CLOCK, DesignTokens.TEXT_MUTED, 1.8)
	ic.custom_minimum_size = Vector2(20, 20)
	ic.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	box.add_child(ic)
	_code_input = LineEdit.new()
	_code_input.placeholder_text = "请输入验证码"
	_code_input.max_length = 6
	_code_input.add_theme_font_size_override("font_size", 15)
	_code_input.add_theme_color_override("font_color", DesignTokens.TEXT)
	_code_input.add_theme_color_override("font_placeholder_color", DesignTokens.TEXT_FAINT)
	_code_input.add_theme_color_override("caret_color", DesignTokens.PRIMARY)
	_code_input.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	_code_input.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	_code_input.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	box.add_child(_code_input)
	_code_btn = CountdownButton.new()
	_code_btn.custom_minimum_size = Vector2(96, 44)
	_code_btn.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	box.add_child(_code_btn)
	return row

func _build_agree_row() -> Control:
	var row := Kit.hbox(8)
	row.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	_agree_panel = PanelContainer.new()
	_agree_panel.custom_minimum_size = Vector2(18, 18)
	_agree_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var ic := Icon.new(Icon.Kind.CHECK, Color.WHITE, 2.2)
	ic.custom_minimum_size = Vector2(12, 12)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -6.0
	ic.offset_right = 6.0
	ic.offset_top = -6.0
	ic.offset_bottom = 6.0
	_agree_panel.add_child(ic)
	_agree_icon = ic
	row.add_child(_agree_panel)
	var tip := Kit.label_sys("我已阅读并同意", 12, DesignTokens.TEXT_MUTED)
	row.add_child(tip)
	var u1 := Kit.label_sys("《用户协议》", 12, DesignTokens.PRIMARY)
	row.add_child(u1)
	var tip2 := Kit.label_sys("与", 12, DesignTokens.TEXT_MUTED)
	row.add_child(tip2)
	var u2 := Kit.label_sys("《隐私政策》", 12, DesignTokens.PRIMARY)
	row.add_child(u2)
	var hit := Button.new()
	hit.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	hit.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	hit.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	hit.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	hit.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	hit.pressed.connect(func() -> void: vm.toggle_agree())
	row.add_child(hit)
	return row

func _build_other_login() -> Control:
	var box := Kit.vbox(14)
	box.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	var divider_row := Kit.hbox(10)
	divider_row.custom_minimum_size = Vector2(0, 20)
	divider_row.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var l1 := ColorRect.new()
	l1.color = DesignTokens.DIVIDER
	l1.custom_minimum_size = Vector2(90, 1)
	l1.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	divider_row.add_child(l1)
	var t := Kit.label_sys("其他登录方式", 12, DesignTokens.TEXT_FAINT)
	divider_row.add_child(t)
	var l2 := ColorRect.new()
	l2.color = DesignTokens.DIVIDER
	l2.custom_minimum_size = Vector2(90, 1)
	l2.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	divider_row.add_child(l2)
	box.add_child(divider_row)
	var icons_row := Kit.hbox(32)
	icons_row.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	icons_row.add_child(_third("微信", Color("07C160"), Icon.Kind.WECHAT))
	icons_row.add_child(_third("支付宝", Color("1677FF"), Icon.Kind.ALIPAY))
	box.add_child(icons_row)
	return box

func _third(name: String, color: Color, icon_kind: int) -> Control:
	var box := Kit.vbox(6)
	var circle := Control.new()
	circle.custom_minimum_size = Vector2(48, 48)
	circle.mouse_filter = Control.MOUSE_FILTER_IGNORE
	var c = color
	circle.draw.connect(func() -> void:
		circle.draw_circle(Vector2(24, 24), 24.0, c)
	)
	var ic := Icon.new(icon_kind, Color.WHITE, 1.8)
	ic.custom_minimum_size = Vector2(28, 28)
	ic.set_anchors_preset(Control.PRESET_CENTER)
	ic.anchor_left = 0.5
	ic.anchor_right = 0.5
	ic.anchor_top = 0.5
	ic.anchor_bottom = 0.5
	ic.offset_left = -14.0
	ic.offset_right = 14.0
	ic.offset_top = -14.0
	ic.offset_bottom = 14.0
	circle.add_child(ic)
	box.add_child(circle)
	var l := Kit.label_sys(name, 11, DesignTokens.TEXT_SEC, false, HORIZONTAL_ALIGNMENT_CENTER)
	box.add_child(l)
	var btn := Button.new()
	btn.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	btn.add_theme_stylebox_override("normal", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("hover", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("pressed", StyleBoxEmpty.new())
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())
	var provider := name
	btn.pressed.connect(func() -> void: vm.login_third(provider))
	box.add_child(btn)
	return box

func bind() -> void:
	# 模式切换：显示对应字段组、切换 Tab 样式
	hub.bind(vm, "mode", func(v) -> void:
		var is_sms: bool = v == "sms"
		_sms_box.visible = is_sms
		_pwd_box.visible = not is_sms
		_set_tab_style(_tab_sms, is_sms)
		_set_tab_style(_tab_pwd, not is_sms)
	)
	hub.bind(vm, "btn_text", func(v) -> void: _login_btn.text = str(v))
	hub.bind(vm, "loading", func(v) -> void:
		var busy: bool = v
		_login_btn.disabled = busy
		_login_btn.text = "登录中..." if busy else str(vm.get_prop("btn_text"))
	)
	hub.bind(vm, "countdown", func(v) -> void: _code_btn.set_remaining(int(v)))
	hub.bind(vm, "agree", func(v) -> void:
		var agreed: bool = v
		_agree_panel.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY if agreed else DesignTokens.TEXT_FAINT, 4.0))
		_agree_icon.visible = agreed
	)

	# 输入回写 ViewModel（单向：用户输入 -> VM）
	_phone_input.text_changed.connect(func(t: String) -> void: vm.set_phone(t))
	_code_input.text_changed.connect(func(t: String) -> void: vm.set_code(t))
	_account_input.text_changed.connect(func(t: String) -> void: vm.set_account(t))
	_password_input.text_changed.connect(func(t: String) -> void: vm.set_password(t))
	_code_btn.requested.connect(func() -> void: vm.send_code())

	# Tab 点击
	_tab_sms.pressed.connect(func() -> void: vm.set_mode("sms"))
	_tab_pwd.pressed.connect(func() -> void: vm.set_mode("password"))

	# 初始绑定值落屏
	_phone_input.text = str(vm.get_prop("phone"))
	_agree_panel.add_theme_stylebox_override("panel", Kit.style_flat(DesignTokens.PRIMARY if vm.get_prop("agree") else DesignTokens.TEXT_FAINT, 4.0))
