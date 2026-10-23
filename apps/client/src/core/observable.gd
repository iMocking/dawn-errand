class_name Observable
extends RefCounted
## MVVM 核心：可观察属性（单向数据源）。
## ViewModel 持有 Observable，View 通过 BindHub 订阅 changed 信号完成数据驱动渲染。

signal changed(new_value: Variant)

var _value: Variant

func _init(initial: Variant = null) -> void:
	_value = initial

var value: Variant:
	get:
		return _value
	set(v):
		if _value == v:
			return
		_value = v
		changed.emit(_value)

func get_value() -> Variant:
	return _value
