class_name Navigator
extends Control
## 导航器：管理屏幕栈（push/replace/pop），持有每个屏幕的 ViewModel 生命周期。

var _stack: Array[Dictionary] = []

func _init() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

## 压栈打开新屏幕
func push(scene: PackedScene, vm: ViewModel) -> ScreenBase:
	var s: ScreenBase = scene.instantiate()
	s.setup(vm)
	add_child(s)
	_stack.append({"screen": s, "vm": vm})
	s.on_enter()
	return s

## 清空栈并以新屏幕替换（登录成功进主界面 / 登出回登录页）
func replace(scene: PackedScene, vm: ViewModel) -> ScreenBase:
	while _stack.size() > 0:
		var top: Dictionary = _stack.pop_back()
		var scr: ScreenBase = top.get("screen")
		if scr and is_instance_valid(scr):
			scr.on_leave()
			scr.queue_free()
	return push(scene, vm)

## 返回上一屏幕
func pop() -> void:
	if _stack.size() <= 1:
		return
	var top: Dictionary = _stack.pop_back()
	var scr: ScreenBase = top.get("screen")
	if scr and is_instance_valid(scr):
		scr.on_leave()
		scr.queue_free()
	if _stack.size() > 0:
		var prev: Dictionary = _stack.back()
		var prev_scr: ScreenBase = prev.get("screen")
		if prev_scr and is_instance_valid(prev_scr):
			prev_scr.on_enter()

func current() -> ScreenBase:
	if _stack.is_empty():
		return null
	return _stack.back().get("screen")

func stack_size() -> int:
	return _stack.size()
