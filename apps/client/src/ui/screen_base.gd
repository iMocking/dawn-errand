class_name ScreenBase
extends Control
## 屏幕基类（MVVM 视图层）：
## setup(vm) -> build() 构造控件；bind() 通过 hub 注册数据绑定；
## 视图只读 ViewModel，写操作走命令/回调，绝不反向依赖视图。

var vm: ViewModel
var hub: BindHub

func setup(p_vm: ViewModel) -> void:
	vm = p_vm
	hub = BindHub.new()
	add_child(hub)
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	build()
	bind()

## 构造控件树
func build() -> void:
	pass

## 注册数据绑定
func bind() -> void:
	pass

## 屏幕进入前台（返回时回调）
func on_enter() -> void:
	pass

## 屏幕离开前台（压栈覆盖或销毁前回调）
func on_leave() -> void:
	pass

func _exit_tree() -> void:
	if vm:
		vm.dispose()
		vm = null
