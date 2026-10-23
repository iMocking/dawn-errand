class_name CouponsVM
extends ViewModel
## 我的优惠券 VM：券列表（演示数据源）。

func _init() -> void:
	obs("usable_count", 0)
	_apply()
	list("coupons").set_items(App.demo.coupons())

func _apply() -> void:
	var n := 0
	for c in list("coupons").items():
		if c.usable:
			n += 1
	set_prop("usable_count", n)

func use_tip(c: CouponModel) -> void:
	if c.usable:
		Toast.show_msg("下单时可自动抵扣 ¥" + Fmt.money(c.amount))
	else:
		Toast.show_msg("该券已过期，无法使用")
