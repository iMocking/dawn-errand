class_name BindHub
extends Node
## MVVM 核心：视图绑定中枢。View 在 build 阶段用 bind() 注册绑定，
## 场景退出时自动断开全部信号，杜绝 ViewModel 悬挂引用被释放对象。

var _binds: Array[Callable] = []

## 绑定 Observable：立即执行一次（初始渲染），后续值变化自动重绘。
func bind(vm: ViewModel, prop: String, apply: Callable) -> void:
	var o: Observable = vm.obs(prop)
	o.changed.connect(apply)
	apply.call(o.get_value())
	_binds.append(func() -> void: o.changed.disconnect(apply))

## 绑定 ReactiveList：列表变化时调用 rebuild(列表)。
func bind_list(vm: ViewModel, prop: String, rebuild: Callable) -> void:
	var l: ReactiveList = vm.list(prop)
	l.items_changed.connect(rebuild)
	rebuild.call(l.items())
	_binds.append(func() -> void: l.items_changed.disconnect(rebuild))

## 绑定命令（按钮点击等）：执行 command.execute()。
func bind_command(vm: ViewModel, prop: String, node: BaseButton) -> void:
	var cmd: Command = vm.get_prop(prop)
	node.pressed.connect(func() -> void: cmd.execute())
	_binds.append(func() -> void: node.pressed.disconnect(cmd.execute))

func clear_all() -> void:
	for b in _binds:
		b.call()
	_binds.clear()

func _exit_tree() -> void:
	clear_all()
