class_name ErrandVM
extends ViewModel
## 跑腿下单 VM：服务类型、起终点、物品描述、费用试算，提交进入确认订单。

var _fee_map := {"帮我买": 8.0, "帮我送": 6.0, "帮我取": 6.0, "帮我办": 10.0}

func _init() -> void:
	obs("type", "帮我买")
	obs("from_address", "王家湾村三组26号")
	obs("from_contact", "王建国 138****8888")
	obs("to_address", "绿杨乡卫生院住院部 2楼3床")
	obs("to_contact", "王秀兰 137****6666")
	obs("goods_desc", "")
	obs("distance", 5.2)
	obs("fee", _fee_map["帮我买"])
	list("type_options").set_items(["帮我买", "帮我送", "帮我取", "帮我办"])
	list("quick_tags").set_items(App.demo.errand_tags())

func set_type(t: String) -> void:
	set_prop("type", t)
	set_prop("fee", _fee_map.get(t, 8.0))

func set_goods(v: String) -> void:
	set_prop("goods_desc", v)

func append_tag(tag: String) -> void:
	var cur := str(get_prop("goods_desc"))
	if cur != "" and not cur.contains(tag):
		set_prop("goods_desc", cur + "、" + tag)

func submit() -> void:
	var desc := str(get_prop("goods_desc")).strip_edges()
	if desc == "":
		Toast.show_msg("请填写需要买/送/取的内容")
		return
	var e := ErrandModel.new()
	e.type = str(get_prop("type"))
	e.from_address = str(get_prop("from_address"))
	e.to_address = str(get_prop("to_address"))
	e.goods_desc = desc
	e.distance_km = float(get_prop("distance"))
	e.fee = float(get_prop("fee"))
	App.open_screen("order_confirm", OrderConfirmVM.for_errand(e))
