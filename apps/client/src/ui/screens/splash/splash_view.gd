class_name SplashView
extends ScreenBase
## 启动页视图：紫色→橙色渐变底、Logo、品牌名与口号。

func build() -> void:
	# 渐变背景（紫→橙）
	var tex := GradientTexture2D.new()
	var g := Gradient.new()
	g.offsets = PackedFloat32Array([0.0, 0.45, 1.0])
	g.colors = PackedColorArray([Color("2A1B4A"), Color("7B3F9E"), Color("FF6B2C")])
	tex.gradient = g
	tex.fill_from = Vector2(0, 0)
	tex.fill_to = Vector2(1, 1)
	var bg := TextureRect.new()
	bg.texture = tex
	bg.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_COVERED
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	bg.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(bg)

	# 状态栏
	var sb := StatusBar.new(Color(1, 1, 1, 0.92))
	sb.set_anchors_and_offsets_preset(Control.PRESET_TOP_WIDE)
	add_child(sb)

	# 居中品牌区
	var center := Control.new()
	center.set_anchors_preset(Control.PRESET_CENTER)
	center.anchor_left = 0.0
	center.anchor_right = 1.0
	center.anchor_top = 0.0
	center.anchor_bottom = 1.0
	center.offset_top = -170.0
	center.offset_bottom = 120.0
	center.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(center)

	var box := Kit.vbox(10)
	box.mouse_filter = Control.MOUSE_FILTER_IGNORE
	box.set_anchors_preset(Control.PRESET_CENTER)
	box.anchor_left = 0.5
	box.anchor_right = 0.5
	box.anchor_top = 0.5
	box.anchor_bottom = 0.5
	box.offset_left = -120.0
	box.offset_right = 120.0
	box.offset_top = -120.0
	box.offset_bottom = 120.0
	center.add_child(box)

	# Logo：渐变圆 + 跑腿剪影
	var logo := Control.new()
	logo.custom_minimum_size = Vector2(84, 84)
	logo.mouse_filter = Control.MOUSE_FILTER_IGNORE
	logo.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	box.add_child(logo)
	logo.draw.connect(func() -> void:
		logo.draw_circle(Vector2(42, 42), 38.0, Color(1, 1, 1, 0.18))
		logo.draw_circle(Vector2(42, 42), 30.0, Color(1, 1, 1, 0.9))
	)
	box.add_child(Kit.spacing(0, 14))

	var brand := Kit.label_sys("曙光跑腿", 32, Color.WHITE, true, HORIZONTAL_ALIGNMENT_CENTER)
	box.add_child(brand)
	var en := Kit.label_sys("D A W N   E R R A N D", 12, Color(1, 1, 1, 0.8), false, HORIZONTAL_ALIGNMENT_CENTER)
	box.add_child(en)
	box.add_child(Kit.spacing(0, 6))
	var slogan := Kit.label_sys("乡镇生活，随叫随到", 14, Color(1, 1, 1, 0.9), false, HORIZONTAL_ALIGNMENT_CENTER)
	box.add_child(slogan)

	# 版本号
	var ver := Kit.label_sys("v1.0.0", 11, Color(1, 1, 1, 0.55), false, HORIZONTAL_ALIGNMENT_CENTER)
	ver.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
	ver.anchor_left = 0.0
	ver.anchor_right = 1.0
	ver.offset_top = -56.0
	ver.offset_bottom = -36.0
	add_child(ver)

func bind() -> void:
	pass

func on_enter() -> void:
	vm.start()
