class_name ViewModel
extends RefCounted
## MVVM 核心：ViewModel 基类。维护一组 Observable 属性，
## View 只通过 BindHub 订阅属性变化，ViewModel 绝不直接操作节点。

var _props: Dictionary = {}   # name -> Observable
var _lists: Dictionary = {}   # name -> ReactiveList

func obs(name: String, initial: Variant = null) -> Observable:
	if not _props.has(name):
		_props[name] = Observable.new(initial)
	elif initial != null:
		# 已存在的属性：显式传入新初值时更新（支持 VM 二次初始化覆盖）
		_props[name].value = initial
	return _props[name]

func list(name: String) -> ReactiveList:
	if not _lists.has(name):
		_lists[name] = ReactiveList.new()
	return _lists[name]

func set_prop(name: String, v: Variant) -> void:
	obs(name).value = v

func get_prop(name: String) -> Variant:
	return obs(name).value

func has_prop(name: String) -> bool:
	return _props.has(name)

## 视图卸载时由 View 调用，释放订阅，避免悬挂引用。
func dispose() -> void:
	_props.clear()
	_lists.clear()
