class_name ReviewVM
extends ViewModel
## 评价 VM：星级 + 标签 + 评语（演示提交）。

var order: OrderModel

func _init(o: OrderModel = null) -> void:
	order = o
	obs("shop", order.shop if order else "村口老陈家常菜")
	obs("stars", 5)
	obs("comment", "")
	list("tags").set_items(App.demo.review_tags())
	list("selected_tags").set_items([])

func set_stars(v: int) -> void:
	set_prop("stars", v)

func set_comment(v: String) -> void:
	set_prop("comment", v)

func toggle_tag(tag: String) -> void:
	var sel: Array = list("selected_tags").items().duplicate()
	if sel.has(tag):
		sel.erase(tag)
	else:
		sel.append(tag)
	list("selected_tags").set_items(sel)

func submit() -> void:
	if int(get_prop("stars")) == 0:
		Toast.show_msg("请先给商家评分")
		return
	Toast.show_msg("感谢您的评价！")
	App.go_back()
