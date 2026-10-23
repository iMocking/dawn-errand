class_name AddressVM
extends ViewModel
## 地址管理 VM：地址列表（演示数据源）、新增/编辑/删除/设默认。

func _init() -> void:
	list("addresses").set_items(App.demo.addresses())

func add_address() -> void:
	App.open_screen("address_edit", AddressEditVM.new())

func edit_address(a: AddressModel) -> void:
	App.open_screen("address_edit", AddressEditVM.new(a))

func set_default(a: AddressModel) -> void:
	for item in list("addresses").items():
		item.is_default = item == a
	list("addresses").set_items(list("addresses").items())
	Toast.show_msg("已设为默认地址")

func delete_address(a: AddressModel) -> void:
	var items: Array = list("addresses").items()
	items.erase(a)
	list("addresses").set_items(items)
	Toast.show_msg("地址已删除（演示）")
