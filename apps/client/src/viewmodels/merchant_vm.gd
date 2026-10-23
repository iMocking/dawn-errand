class_name MerchantVM
extends ViewModel
## 商家详情 VM：分类、菜品、购物车（数据驱动）。
## 商家/菜品数据来自 DemoData（后端暂无 C 端商家接口）。

var merchant: MerchantModel
var cart := CartModel.new()

func _init(m: MerchantModel = null) -> void:
	var mm := m
	if mm == null:
		var src := App.demo.nearby_merchants()
		mm = src[0] if src.size() > 0 else App.demo.recommended_merchants()[0]
	merchant = mm
	obs("name", mm.name)
	obs("score", mm.score_text())
	obs("meta", "月售" + str(mm.monthly_sales) + "单 · " + Fmt.km(mm.dist_km) + " · " + str(mm.eta_min) + "分钟送到")
	obs("cart_count", 0)
	obs("cart_total", 0.0)
	obs("cart_open", false)
	list("categories").set_items(App.demo.merchant_categories())
	list("dishes").set_items(App.demo.merchant_dishes())

func add_dish(d: DishModel) -> void:
	cart.add(d)
	_sync()

func remove_dish(d: DishModel) -> void:
	cart.remove(d)
	_sync()

func _sync() -> void:
	set_prop("cart_count", cart.count())
	set_prop("cart_total", cart.total())
	list("cart_lines").set_items(cart.lines)

func clear_cart() -> void:
	cart.clear()
	_sync()

func toggle_cart() -> void:
	if cart.count() <= 0:
		Toast.show_msg("购物车还是空的")
		return
	set_prop("cart_open", not bool(get_prop("cart_open")))

func close_cart() -> void:
	set_prop("cart_open", false)

func go_confirm() -> void:
	if cart.count() <= 0:
		Toast.show_msg("请先选购商品")
		return
	set_prop("cart_open", false)
	App.open_screen("order_confirm", OrderConfirmVM.for_cart(merchant, cart))
