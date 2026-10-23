class_name ShellView
extends ScreenBase
## 主界面外壳视图：内容区 + 底部 TabBar；四个 Tab 屏懒加载并缓存。

var _content: Control
var _tab_bar: AppTabBar
var _tab_screens: Dictionary = {}
var _tab_names := ["home", "errand", "orders", "profile"]

func build() -> void:
	_content = Control.new()
	_content.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	_content.offset_bottom = -DesignTokens.TABBAR_H
	add_child(_content)
	_tab_bar = AppTabBar.new()
	_tab_bar.anchor_left = 0.0
	_tab_bar.anchor_right = 1.0
	_tab_bar.anchor_top = 1.0
	_tab_bar.anchor_bottom = 1.0
	_tab_bar.offset_top = -DesignTokens.TABBAR_H
	_tab_bar.offset_bottom = 0.0
	add_child(_tab_bar)

func bind() -> void:
	_tab_bar.tab_selected.connect(func(i: int) -> void: vm.switch_tab(i))
	hub.bind(vm, "tab", _on_tab_changed)

func _on_tab_changed(v: Variant) -> void:
	var idx := int(v)
	_tab_bar.set_active(idx)
	_show_tab(idx)

func _show_tab(idx: int) -> void:
	# 隐藏其他已加载 Tab 屏，避免多屏叠加
	for k in _tab_screens:
		if k != idx:
			var other: Control = _tab_screens[k]
			other.hide()
	if _tab_screens.has(idx):
		var scr: ScreenBase = _tab_screens[idx]
		scr.show()
		scr.on_enter()
		return
	var scn: PackedScene = load(App.SCREENS[_tab_names[idx]])
	var scr2: ScreenBase = scn.instantiate()
	scr2.setup(App.create_vm(_tab_names[idx]))
	_content.add_child(scr2)
	_tab_screens[idx] = scr2
	scr2.on_enter()

## 切换到指定 Tab（供其他页面跳转，如"查看订单"回订单 Tab）。
func switch_to(index: int) -> void:
	vm.switch_tab(index)
