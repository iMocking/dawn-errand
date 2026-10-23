class_name DemoData
extends RefCounted
## 演示数据源：还原设计稿中的静态业务数据。
## 说明：后端当前未提供 C 端商家/订单/地址/优惠券接口（管理端接口需权限），
## 业务数据暂由本层提供；认证类接口（登录/验证码/me/登出）全部走真实后端。
## 后续后端补齐 C 端接口后，在对应 ViewModel 中将 demo 替换为 Api 调用即可。

## ---- 首页：推荐（横滑） ----
func recommended_merchants() -> Array:
	return [
		_m("老陈家常菜", 4.8, 0.0, 326, 25, ["本周热销", "满25免配送费"], 0, Color("FFE8D6"), Color("FFD4B8")),
		_m("王家湾超市", 4.9, 0.0, 512, 15, ["生鲜日配", "烟酒饮料"], 1, Color("E8F5E9"), Color("C8E6C9")),
		_m("乡亲水果店", 4.7, 0.0, 189, 35, ["产地直发"], 2, Color("FCE4EC"), Color("F8BBD0")),
		_m("李记卤味", 4.6, 0.0, 240, 20, ["现卤现卖"], 3, Color("FFF3E0"), Color("FFE0B2")),
	]

## ---- 首页：附近商家（列表） ----
func nearby_merchants() -> Array:
	return [
		_m("村口老陈家常菜", 4.8, 2.1, 326, 25, ["🔥 本周热销", "满25免配送费"], 0, Color("FFE8D6"), Color("FFD4B8")),
		_m("王家湾综合超市", 4.9, 0.8, 512, 15, ["生鲜日配", "烟酒饮料"], 1, Color("E8F5E9"), Color("C8E6C9")),
		_m("乡里乡亲水果店", 4.7, 3.5, 189, 35, ["产地直发"], 2, Color("FCE4EC"), Color("F8BBD0")),
	]

func _m(name: String, score: float, dist: float, sales: int, eta: int, tags: Array, icon: int, c1: Color, c2: Color) -> MerchantModel:
	var m := MerchantModel.new()
	m.name = name
	m.score = score
	m.dist_km = dist
	m.monthly_sales = sales
	m.eta_min = eta
	m.tags = tags
	m.icon = icon
	m.gradient = [c1, c2]
	return m

## ---- 商家详情：分类与菜品 ----
func merchant_categories() -> Array:
	return ["🔥 热销爆款", "招牌炒菜", "家常小炒", "汤类炖菜", "主食米饭"]

func merchant_dishes() -> Array:
	var d := DishModel.new()
	d.name = "农家小炒肉"
	d.desc = "招牌土猪肉现炒，香辣下饭 · 月售86份"
	d.price = 16.8
	d.monthly_sales = 86
	d.icon = 0
	var d2 := DishModel.new()
	d2.name = "红烧土鸡块"
	d2.desc = "散养土鸡慢炖，入味软烂 · 月售64份"
	d2.price = 28.0
	d2.monthly_sales = 64
	d2.icon = 1
	d2.icon_color = Color("FF6B2C")
	var d3 := DishModel.new()
	d3.name = "手撕包菜"
	d3.desc = "本地包菜大火快炒，脆嫩爽口 · 月售52份"
	d3.price = 12.0
	d3.monthly_sales = 52
	d3.icon = 2
	d3.icon_color = Color("52C41A")
	var d4 := DishModel.new()
	d4.name = "西红柿炒蛋"
	d4.desc = "农家土鸡蛋，酸甜可口 · 月售78份"
	d4.price = 10.0
	d4.monthly_sales = 78
	d4.icon = 3
	d4.icon_color = Color("FAAD14")
	var d5 := DishModel.new()
	d5.name = "米饭"
	d5.desc = "本地香米现蒸 · 月售210份"
	d5.price = 2.0
	d5.monthly_sales = 210
	d5.icon = 4
	d5.icon_color = Color("FF8A3D")
	return [d, d2, d3, d4, d5]

## ---- 我的订单 ----
func orders() -> Array:
	var o1 := OrderModel.new()
	o1.shop = "村口老陈家常菜"
	o1.status = "配送中"
	o1.status_type = OrderModel.StatusType.ONGOING
	o1.goods_summary = "农家小炒肉、西红柿炒蛋、米饭×2"
	o1.time_text = "今天 12:30 · 预计12:45送达"
	o1.total = 32.3
	o1.icon = 0
	o1.actions = ["取消订单", "查看配送"]
	var o2 := OrderModel.new()
	o2.shop = "王家湾综合超市"
	o2.status = "已完成"
	o2.status_type = OrderModel.StatusType.DONE
	o2.goods_summary = "纯牛奶×2箱、洗衣粉×1、抽纸×3"
	o2.time_text = "昨天 18:20 · 已送达"
	o2.total = 86.5
	o2.icon = 1
	o2.actions = ["再来一单", "评价"]
	var o3 := OrderModel.new()
	o3.shop = "跑腿代办 · 镇中心"
	o3.status = "已完成"
	o3.status_type = OrderModel.StatusType.DONE
	o3.goods_summary = "帮取快递（镇邮政所→王家湾村）"
	o3.time_text = "10月20日 15:10 · 已送达"
	o3.total = 6.0
	o3.icon = 5
	o3.actions = ["再次下单"]
	var o4 := OrderModel.new()
	o4.shop = "乡里乡亲水果店"
	o4.status = "已完成"
	o4.status_type = OrderModel.StatusType.DONE
	o4.goods_summary = "红富士苹果5斤、香蕉1把、葡萄2斤"
	o4.time_text = "10月18日 10:00 · 已送达"
	o4.total = 45.8
	o4.icon = 2
	o4.actions = ["再来一单", "评价"]
	return [o1, o2, o3, o4]

## ---- 地址管理 ----
func addresses() -> Array:
	var a1 := AddressModel.new()
	a1.name = "王建国"
	a1.phone = "13888888888"
	a1.detail = "王家湾村三组26号 · 红砖大铁门，门口有棵樟树"
	a1.tag = "家"
	a1.is_default = true
	var a2 := AddressModel.new()
	a2.name = "王秀兰"
	a2.phone = "13766666666"
	a2.detail = "绿杨乡卫生院住院部 2楼3床"
	a2.tag = "医院"
	var a3 := AddressModel.new()
	a3.name = "王强"
	a3.phone = "13599999999"
	a3.detail = "兰溪中学 高三(2)班宿舍"
	a3.tag = "学校"
	return [a1, a2, a3]

## ---- 我的优惠券 ----
func coupons() -> Array:
	var c1 := CouponModel.new()
	c1.amount = 5.0
	c1.condition = "满20可用"
	c1.title = "新用户立减券"
	c1.desc = "全场通用 · 外卖跑腿均可使用"
	c1.expire_text = "有效期至 2026-10-29"
	c1.usable = true
	var c2 := CouponModel.new()
	c2.amount = 3.0
	c2.condition = "无门槛"
	c2.title = "骑手好评奖励券"
	c2.desc = "完成订单评价后获得"
	c2.expire_text = "有效期至 2026-10-25"
	c2.usable = true
	var c3 := CouponModel.new()
	c3.amount = 8.0
	c3.condition = "满30可用"
	c3.title = "周末专享券"
	c3.desc = "仅限周六周日使用"
	c3.expire_text = "已过期 2026-10-20"
	c3.usable = false
	return [c1, c2, c3]

## ---- 搜索：热门/历史 ----
func hot_keywords() -> Array:
	return ["小炒肉", "大米", "卤味", "超市", "水果", "感冒药", "快递", "烟"]

func search_history() -> Array:
	return ["农家小炒肉", "王家湾超市", "西红柿炒蛋"]

## ---- 评价：标签 ----
func review_tags() -> Array:
	return ["味道好", "分量足", "上菜快", "包装好", "骑手准时", "价格实惠"]

## ---- 跑腿下单：快捷标签 ----
func errand_tags() -> Array:
	return ["品牌商品", "先行垫付", "买前拍照确认"]
