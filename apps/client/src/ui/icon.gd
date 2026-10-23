class_name Icon
extends Control
## 矢量图标：以 24×24 设计网格绘制，按控件尺寸等比缩放。
## 覆盖设计稿中的线框图标（返回/搜索/定位/购物车/微信/支付宝等）。

enum Kind {
	BACK, CHEVRON_RIGHT, SEARCH, LOCATION, ARROW_DOWN, PHONE, MAIL, LOCK,
	CART, PLUS, MINUS, CLOSE, CHECK, STAR, SWAP, HOME, ERRAND, DOC, USER,
	WECHAT, ALIPAY, EDIT, BAG, SHOP, FRUIT, MEAT, RICE, HOT, COUPON, BOX,
	RUNNER, HEADPHONE, HELP, SETTINGS, SCOOTER, CLOCK, MEDAL, COIN, TRASH,
	BELL, MEGAPHONE, FLAG, SAFE,
}

var kind: int = Kind.HOME
var icon_color: Color = Color("1D2129")
var line_width: float = 2.0

func _init(p_kind := Kind.HOME, p_color := Color("1D2129"), p_width := 2.0) -> void:
	kind = p_kind
	icon_color = p_color
	line_width = p_width
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	custom_minimum_size = Vector2(24, 24)

func set_kind(p_kind: int) -> void:
	kind = p_kind
	queue_redraw()

func set_icon_color(c: Color) -> void:
	icon_color = c
	queue_redraw()

func _draw() -> void:
	var s: float = minf(size.x, size.y) / 24.0
	if s <= 0:
		return
	var w: float = line_width * s
	var col := icon_color
	match kind:
		Kind.BACK:
			_line([[15, 4], [9, 12], [15, 20]], s, w, col)
		Kind.CHEVRON_RIGHT:
			_line([[9, 6], [15, 12], [9, 18]], s, w, col)
		Kind.ARROW_DOWN:
			_line([[6, 9], [12, 15], [18, 9]], s, w, col)
		Kind.SEARCH:
			_circle([11, 11], 8, s, w, col)
			_line([[21, 21], [16.5, 16.5]], s, w, col)
		Kind.LOCATION:
			_line([[3, 10], [12, 23], [21, 10]], s, w, col)
			_line([[3, 10], [21, 10]], s, w, col)
			_circle([12, 10], 3, s, w, col)
		Kind.PHONE:
			_line([[5, 3], [8, 3], [9.7, 4.7], [10.4, 7.5], [9.1, 9.1], [12, 12], [14.9, 14.9], [16.5, 13.6], [19.3, 14.3], [21, 16], [21, 19]], s, w, col)
		Kind.MAIL:
			_line([[3, 5], [21, 5], [21, 19], [3, 19], [3, 5]], s, w, col)
			_line([[3, 6], [12, 13], [21, 6]], s, w, col)
		Kind.LOCK:
			_rect([5, 11, 19, 20], s, w, col)
			_circle([12, 7], 4, s, w, col)
			_dot(12, 15, 1.6, s, col)
		Kind.CART:
			_dot(9, 21, 1.2, s, col)
			_dot(20, 21, 1.2, s, col)
			_line([[1, 1], [5, 1], [7.68, 14.39], [17.4, 14.39], [23, 6], [6, 6]], s, w, col)
		Kind.PLUS:
			_line([[12, 5], [12, 19]], s, w, col)
			_line([[5, 12], [19, 12]], s, w, col)
		Kind.MINUS:
			_line([[5, 12], [19, 12]], s, w, col)
		Kind.CLOSE:
			_line([[6, 6], [18, 18]], s, w, col)
			_line([[18, 6], [6, 18]], s, w, col)
		Kind.CHECK:
			_line([[5, 12], [10, 17], [19, 7]], s, w, col)
		Kind.STAR:
			_star(s, col)
		Kind.SWAP:
			_line([[17, 1], [21, 5], [17, 9]], s, w, col)
			_line([[3, 11], [21, 11]], s, w, col)
			_line([[7, 23], [3, 19], [7, 15]], s, w, col)
			_line([[3, 13], [21, 13]], s, w, col)
		Kind.HOME:
			_line([[3, 9], [12, 2], [21, 9]], s, w, col)
			_line([[5, 9], [5, 20], [19, 20], [19, 9]], s, w, col)
		Kind.ERRAND:
			_dot(13, 4, 2, s, col)
			_line([[4, 22], [8, 18], [11, 20], [16, 15], [20, 18]], s, w, col)
			_line([[11, 20], [10, 12], [14, 10], [18, 13], [20, 13]], s, w, col)
		Kind.DOC:
			_line([[6, 2], [14, 2], [20, 8], [20, 22], [6, 22]], s, w, col)
			_line([[14, 2], [14, 8], [20, 8]], s, w, col)
		Kind.USER:
			_circle([12, 7], 4, s, w, col)
			_line([[4, 21], [4, 19], [8, 15], [16, 15], [20, 19], [20, 21]], s, w, col)
		Kind.EDIT:
			_line([[11, 4], [4, 4], [2, 6], [2, 20], [4, 22], [18, 22], [20, 20], [20, 13]], s, w, col)
			_line([[18.5, 2.5], [21.5, 5.5], [12, 15], [8, 16], [9, 12]], s, w, col)
		Kind.BAG:
			_line([[5, 11], [5, 8], [7, 6], [17, 6], [19, 8], [19, 11]], s, w, col)
			_line([[4, 11], [5, 20], [19, 20], [20, 11]], s, w, col)
			_line([[3, 11], [21, 11]], s, w, col)
		Kind.SHOP:
			_line([[3, 9], [4.5, 4], [19.5, 4], [21, 9]], s, w, col)
			_line([[4, 9], [4, 20], [5, 21], [19, 21], [20, 20], [20, 9]], s, w, col)
			_line([[9, 21], [9, 13], [15, 13], [15, 21]], s, w, col)
		Kind.FRUIT:
			_line([[12, 21], [12, 9]], s, w, col)
			_arc([7, 4.5], 3.5, s, w, col, PI, TAU)
			_arc([17, 4.5], 3.5, s, w, col, PI, TAU)
			_line([[3.5, 4.5], [20.5, 4.5]], s, w, col)
		Kind.MEAT:
			_circle([12, 12], 9, s, w, col)
			_circle([8, 14], 2.2, s, w, col)
			_circle([15, 9], 2.2, s, w, col)
		Kind.RICE:
			_line([[5, 11], [5, 8], [7, 6], [17, 6], [19, 8], [19, 11]], s, w, col)
			_line([[4, 11], [5, 20], [19, 20], [20, 11]], s, w, col)
			_line([[3, 11], [21, 11]], s, w, col)
			_line([[9, 16], [10, 17]], s, w, col)
			_line([[12, 15], [13, 16]], s, w, col)
			_line([[15, 16], [16, 17]], s, w, col)
		Kind.HOT:
			_circle([12, 12], 9, s, w, col)
			_line([[12, 7], [12, 12], [15, 14]], s, w, col)
		Kind.COUPON:
			_line([[2, 6], [2, 18], [5, 18], [5, 14], [7, 14], [7, 18], [22, 18], [22, 6], [7, 6], [7, 10], [5, 10], [5, 6]], s, w, col)
			_dot(12, 12, 1.5, s, col)
		Kind.BOX:
			_line([[3, 9], [12, 3], [21, 9]], s, w, col)
			_line([[3, 9], [3, 20], [21, 20], [21, 9]], s, w, col)
			_line([[3, 9], [12, 13], [21, 9]], s, w, col)
			_line([[12, 13], [12, 20]], s, w, col)
		Kind.RUNNER:
			_circle([5.5, 17.5], 3.5, s, w, col)
			_circle([18.5, 17.5], 3.5, s, w, col)
			_line([[12, 17.5], [12, 14], [9, 11], [13, 8], [15, 11], [17, 11]], s, w, col)
		Kind.HEADPHONE:
			_arc([12, 14], 8, s, w, col, PI, TAU)
			_line([[4, 14], [4, 18], [6, 20], [9, 20], [9, 14]], s, w, col)
			_line([[20, 14], [20, 18], [18, 20], [15, 20], [15, 14]], s, w, col)
		Kind.HELP:
			_circle([12, 12], 10, s, w, col)
			_arc([12, 9], 3, s, w, col, 0.5 * PI, 1.6 * PI)
			_line([[12, 13], [12, 14.5]], s, w, col)
			_dot(12, 17, 1.2, s, col)
		Kind.SETTINGS:
			_circle([12, 12], 3, s, w, col)
			for i in 8:
				var a := i * TAU / 8.0
				var outer := Vector2(12 + cos(a) * 10.5, 12 + sin(a) * 10.5)
				var inner := Vector2(12 + cos(a) * 7.5, 12 + sin(a) * 7.5)
				draw_line(Vector2(outer.x * s, outer.y * s), Vector2(inner.x * s, inner.y * s), col, w)
		Kind.SCOOTER:
			_circle([5.5, 17.5], 3.5, s, w, col)
			_circle([18.5, 17.5], 3.5, s, w, col)
			_arc([14.5, 6], 2.5, s, w, col, -PI, PI)
			_line([[12, 17.5], [12, 14], [9, 11], [13, 8], [15, 11], [17, 11]], s, w, col)
		Kind.CLOCK:
			_circle([12, 12], 9, s, w, col)
			_line([[12, 7], [12, 12], [15, 14]], s, w, col)
		Kind.MEDAL:
			_star(s, col)
		Kind.COIN:
			_circle([12, 12], 9, s, w, col)
			_line([[12, 7], [12, 12], [15, 14]], s, w, col)
			_line([[8, 19], [16, 5]], s, w, col)
		Kind.TRASH:
			_line([[3, 6], [21, 6]], s, w, col)
			_line([[8, 6], [10, 2], [14, 2], [16, 6]], s, w, col)
			_line([[5, 6], [6, 22], [18, 22], [19, 6]], s, w, col)
		Kind.BELL:
			_arc([12, 9], 6, s, w, col, 0.1 * PI, 0.9 * PI)
			_line([[6, 10], [6, 14], [4, 17], [20, 17], [18, 14], [18, 10]], s, w, col)
			_dot(12, 20, 1.2, s, col)
		Kind.MEGAPHONE:
			_poly([[4, 9], [4, 15], [10, 15], [15, 20], [15, 9], [10, 14]], s, col)
			_line([[17, 8], [20, 11], [17, 14]], s, w, col)
		Kind.FLAG:
			_line([[5, 3], [5, 21]], s, w, col)
			_line([[5, 3], [19, 6], [5, 10]], s, w, col)
		Kind.SAFE:
			_line([[12, 2], [12, 22]], s, w, col)
			_line([[4, 5], [20, 5], [20, 19], [4, 19], [4, 5]], s, w, col)
		Kind.WECHAT:
			_poly([[8.5, 3], [4.9, 3], [2, 5.5], [2, 8.6], [4.5, 10.6], [4.2, 12.5], [6.8, 11.3], [8, 11.3]], s, col)
			_poly([[22, 13.7], [18.4, 13.7], [16.4, 15.6], [16.4, 18.4], [18.8, 20.2], [18.4, 22.1], [21.3, 20.6], [22, 20.6]], s, col)
			_dot(6.5, 7.6, 0.9, s, Color.WHITE)
			_dot(10.5, 7.6, 0.9, s, Color.WHITE)
			_dot(14.7, 16.2, 0.9, s, Color.WHITE)
			_dot(17.7, 16.2, 0.9, s, Color.WHITE)
		Kind.ALIPAY:
			_rect([3, 3, 21, 21], s, w, col)
			_line([[7, 9], [17, 9]], s, w, col)
			_line([[12, 6], [12, 9]], s, w, col)
			_line([[8, 15], [16, 15]], s, w, col)
			_arc([10, 16], 3, s, w, col, 0, PI)

## 折线
func _line(pts: Array, s: float, w: float, col: Color) -> void:
	var arr := PackedVector2Array()
	for p in pts:
		arr.append(Vector2(float(p[0]) * s, float(p[1]) * s))
	draw_polyline(arr, col, w, true)

## 圆（描边或填充）
func _circle(center: Array, r: float, s: float, w: float, col: Color, filled := false) -> void:
	if filled:
		draw_circle(Vector2(float(center[0]) * s, float(center[1]) * s), r * s, col)
	else:
		draw_arc(Vector2(float(center[0]) * s, float(center[1]) * s), r * s, 0, TAU, 32, col, w, true)

## 圆弧
func _arc(center: Array, r: float, s: float, w: float, col: Color, from: float, to: float) -> void:
	draw_arc(Vector2(float(center[0]) * s, float(center[1]) * s), r * s, from, to, 32, col, w, true)

## 实心圆点
func _dot(x: float, y: float, r: float, s: float, col: Color) -> void:
	draw_circle(Vector2(x * s, y * s), r * s, col)

## 描边矩形 [x0,y0,x1,y1]
func _rect(v: Array, s: float, w: float, col: Color) -> void:
	draw_rect(Rect2(Vector2(float(v[0]) * s, float(v[1]) * s), Vector2((float(v[2]) - float(v[0])) * s, (float(v[3]) - float(v[1])) * s)), col, false, w)

## 实心多边形
func _poly(pts: Array, s: float, col: Color) -> void:
	var arr := PackedVector2Array()
	for p in pts:
		arr.append(Vector2(float(p[0]) * s, float(p[1]) * s))
	draw_colored_polygon(arr, col)

## 五角星（填充）
func _star(s: float, col: Color) -> void:
	var pts := PackedVector2Array([
		Vector2(12 * s, 2 * s), Vector2(15.09 * s, 8.26 * s), Vector2(22 * s, 9.27 * s),
		Vector2(17 * s, 14.14 * s), Vector2(18.18 * s, 21.02 * s), Vector2(12 * s, 17.77 * s),
		Vector2(5.82 * s, 21.02 * s), Vector2(7 * s, 14.14 * s), Vector2(2 * s, 9.27 * s),
		Vector2(8.91 * s, 8.26 * s),
	])
	draw_colored_polygon(pts, col)
