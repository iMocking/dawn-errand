class_name SearchVM
extends ViewModel
## 搜索 VM：热门词/历史（演示数据源）+ 按名称检索演示商家。

func _init() -> void:
	obs("keyword", "")
	obs("has_result", false)
	obs("result_count", 0)
	list("hot").set_items(App.demo.hot_keywords())
	list("history").set_items(App.demo.search_history())
	list("results").set_items([])

func set_keyword(v: String) -> void:
	set_prop("keyword", v)
	if v.strip_edges() == "":
		set_prop("has_result", false)
		list("results").set_items([])

func search(kw: String) -> void:
	var key := kw.strip_edges()
	if key == "":
		Toast.show_msg("请输入搜索关键词")
		return
	set_prop("keyword", key)
	var out: Array = []
	for m in App.demo.recommended_merchants() + App.demo.nearby_merchants():
		if str(m.name).contains(key):
			out.append(m)
	list("results").set_items(out)
	set_prop("result_count", out.size())
	set_prop("has_result", true)

func clear_history() -> void:
	list("history").clear()
	Toast.show_msg("搜索历史已清空")

func open_merchant(m: MerchantModel) -> void:
	App.open_screen("merchant", MerchantVM.new(m))
