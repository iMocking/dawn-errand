class_name AddressEditVM
extends ViewModel
## 地址编辑 VM：表单状态（真实数据模型 + 演示保存）。

var _addr: AddressModel
var _tags := ["家", "公司", "学校", "医院", "其他"]

func _init(a: AddressModel = null) -> void:
	_addr = a
	obs("title", "编辑地址" if a != null else "新增地址")
	obs("name", a.name if a else "")
	obs("phone", a.phone if a else "")
	obs("province", "湖北省黄冈市浠水县")
	obs("detail", a.detail if a else "")
	obs("landmark", a.landmark if a else "")
	obs("tag", a.tag if a else "家")
	obs("is_default", a.is_default if a else false)
	obs("saving", false)
	list("tags").set_items(_tags)

func set_name(v: String) -> void:
	set_prop("name", v)

func set_phone(v: String) -> void:
	set_prop("phone", v)

func set_detail(v: String) -> void:
	set_prop("detail", v)

func set_landmark(v: String) -> void:
	set_prop("landmark", v)

func set_tag(t: String) -> void:
	set_prop("tag", t)

func toggle_default() -> void:
	set_prop("is_default", not bool(get_prop("is_default")))

func save() -> void:
	if get_prop("saving"):
		return
	if str(get_prop("name")).strip_edges() == "":
		Toast.show_msg("请填写联系人姓名")
		return
	if str(get_prop("phone")).strip_edges() == "":
		Toast.show_msg("请填写联系电话")
		return
	if str(get_prop("detail")).strip_edges() == "":
		Toast.show_msg("请填写详细地址")
		return
	set_prop("saving", true)
	await Engine.get_main_loop().create_timer(0.5).timeout
	set_prop("saving", false)
	Toast.show_msg("地址已保存")
	App.go_back()
