class_name ReactiveList
extends RefCounted
## MVVM 核心：可观察列表。数据变化时 emit items_changed，View 重建列表项。

signal items_changed

var _items: Array = []

func items() -> Array:
	return _items

func set_items(arr: Array) -> void:
	_items = arr.duplicate()
	items_changed.emit()

func append_item(item: Variant) -> void:
	_items.append(item)
	items_changed.emit()

func remove_at(index: int) -> void:
	if index >= 0 and index < _items.size():
		_items.remove_at(index)
		items_changed.emit()

func clear() -> void:
	_items.clear()
	items_changed.emit()

func size() -> int:
	return _items.size()

func get_item(index: int) -> Variant:
	if index >= 0 and index < _items.size():
		return _items[index]
	return null
