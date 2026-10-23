class_name HomeVM
extends ViewModel
## 首页 VM：问候语（真实用户昵称）、定位、推荐/附近商家（演示数据源）。
## 商家类接口后端暂未开放 C 端，数据来自 DemoData；搜索/跑腿/详情跳转走导航。

func _init() -> void:
	obs("username", App.session.user_name())
	obs("location", "湖北省黄冈市浠水县")
	list("recommend").set_items(App.demo.recommended_merchants())
	list("nearby").set_items(App.demo.nearby_merchants())

func open_search() -> void:
	App.open_screen("search")

func open_merchant(m: MerchantModel) -> void:
	App.open_screen("merchant", MerchantVM.new(m))

func open_errand() -> void:
	App.open_screen("errand")

## 点头像/用户名进入个人中心
func open_profile() -> void:
	App.open_screen("profile")
