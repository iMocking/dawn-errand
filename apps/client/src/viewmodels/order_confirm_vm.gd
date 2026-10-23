class_name OrderConfirmVM
extends ViewModel
## 确认订单 VM：支持「商家购物车」与「跑腿」两种下单模式；
## 金额明细、地址、支付方式均为数据驱动。

var kind := "cart"                 # "cart" | "errand"
var merchant: MerchantModel
var cart: CartModel
var errand: ErrandModel
var address: AddressModel
var coupon: CouponModel = null
var _use_coupon := false

static func for_cart(m: MerchantModel, c: CartModel, a: AddressModel = null) -> OrderConfirmVM:
	var vm := OrderConfirmVM.new()
	vm.kind = "cart"
	vm.merchant = m
	vm.cart = c
	vm.address = a if a != null else _default_address()
	vm._init_state()
	return vm

static func for_errand(e: ErrandModel, a: AddressModel = null) -> OrderConfirmVM:
	var vm := OrderConfirmVM.new()
	vm.kind = "errand"
	vm.errand = e
	vm.address = a if a != null else _default_address()
	vm._init_state()
	return vm

static func _default_address() -> AddressModel:
	var list: Array = App.demo.addresses()
	return list[0] if list.size() > 0 else AddressModel.new()

func _init() -> void:
	address = _default_address()
	_init_state()

func _init_state() -> void:
	var addr := address if address != null else _default_address()
	obs("mode_title", "确认订单")
	obs("shop_name", _shop_name())
	obs("goods_lines", _goods_lines())
	obs("goods_total", _goods_total())
	obs("delivery_fee", _delivery_fee())
	obs("pack_fee", _pack_fee())
	obs("coupon_amount", 0.0)
	obs("total", _total())
	obs("address_name", addr.name)
	obs("address_phone", Fmt.mask_phone(addr.phone))
	obs("address_detail", addr.province + addr.detail)
	obs("pay_method", "在线支付（演示）")
	obs("loading", false)

func _shop_name() -> String:
	if kind == "errand":
		return "跑腿订单 · " + (errand.type if errand else "帮我买")
	return merchant.name if merchant else "商家"

func _goods_lines() -> Array:
	var lines: Array = []
	if kind == "errand" and errand:
		lines.append({"name": errand.type + "：" + errand.goods_desc, "price": errand.fee, "qty": 1, "kind": "errand"})
	elif cart:
		for line in cart.lines:
			lines.append({"name": line.dish.name, "price": line.dish.price, "qty": int(line.qty), "kind": "dish"})
	return lines

func _goods_total() -> float:
	if kind == "errand":
		return errand.fee if errand else 0.0
	return cart.goods_total() if cart else 0.0

func _delivery_fee() -> float:
	if kind == "errand":
		return 0.0
	return cart.delivery_fee if cart else 0.0

func _pack_fee() -> float:
	if kind == "errand":
		return 0.0
	return cart.pack_fee if cart else 0.0

func _total() -> float:
	return _goods_total() + _delivery_fee() + _pack_fee() - float(get_prop("coupon_amount"))

func apply_coupon(c: CouponModel) -> void:
	if c == null:
		return
	coupon = c
	_use_coupon = true
	set_prop("coupon_amount", c.amount)
	set_prop("total", _total())

func clear_coupon() -> void:
	coupon = null
	_use_coupon = false
	set_prop("coupon_amount", 0.0)
	set_prop("total", _total())

func choose_coupon() -> void:
	var list: Array = App.demo.coupons()
	var usable: Array = []
	for c in list:
		if c.usable:
			usable.append(c)
	if usable.size() == 0:
		Toast.show_msg("暂无可用优惠券")
		return
	apply_coupon(usable[0])
	Toast.show_msg("已使用" + str(usable[0].title))

func submit() -> void:
	if get_prop("loading"):
		return
	set_prop("loading", true)
	await Engine.get_main_loop().create_timer(0.6).timeout
	set_prop("loading", false)
	var no := "SG" + str(Time.get_unix_time_from_system()) + str(randi() % 1000)
	App.open_screen("pay_success", PaySuccessVM.new(_shop_name(), no, _total(), address))
