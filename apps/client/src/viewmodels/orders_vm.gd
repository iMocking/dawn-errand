class_name OrdersVM
extends ViewModel
## 我的订单 VM：筛选（全部/进行中/已完成）+ 订单列表（演示数据源）+ 操作分发。
## 后端暂无 C 端订单接口，数据来自 DemoData；动作跳转（追踪/评价）为真实导航。

func _init() -> void:
	obs("filter", "全部")
	list("filters").set_items(["全部", "进行中", "已完成"])
	list("orders").set_items(App.demo.orders())

func set_filter(f: String) -> void:
	set_prop("filter", f)
	_refresh()

func _refresh() -> void:
	var f := str(get_prop("filter"))
	var src := App.demo.orders()
	var out: Array = []
	for o in src:
		if f == "全部":
			out.append(o)
		elif f == "进行中" and o.status_type == OrderModel.StatusType.ONGOING:
			out.append(o)
		elif f == "已完成" and o.status_type == OrderModel.StatusType.DONE:
			out.append(o)
	list("orders").set_items(out)

func order_action(order: OrderModel, action: String) -> void:
	match action:
		"查看配送":
			App.open_screen("order_track", OrderTrackVM.new(order))
		"评价":
			App.open_screen("review", ReviewVM.new(order))
		"取消订单":
			Toast.show_msg("演示环境：订单取消失败（后端未开放取消接口）")
		"再来一单":
			Toast.show_msg("演示环境：已为您重新下单")
		_:
			Toast.show_msg("该操作暂未开放")
