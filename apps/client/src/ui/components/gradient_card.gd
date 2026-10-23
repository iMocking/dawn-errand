class_name GradientCard
extends Control
## 渐变圆角卡片：垂直渐变主体 + 背景色圆角（四角覆盖近似裁剪）。

var colors: Array = [Color("FF8A3D"), Color("FF5A1F")]
var radius := 16.0
var corner_color := Color.WHITE

func _init(c1: Color, c2: Color, r := 16.0, corner: Color = Color.WHITE) -> void:
	colors = [c1, c2]
	radius = r
	corner_color = corner
	mouse_filter = Control.MOUSE_FILTER_IGNORE

func set_colors(c1: Color, c2: Color) -> void:
	colors = [c1, c2]
	queue_redraw()

func _draw() -> void:
	var steps := maxi(12, int(size.y / 3.0))
	var r := minf(radius, min(size.x, size.y) / 2.0)
	for i in steps:
		var y0 := size.y * i / steps
		var y1 := size.y * (i + 1) / steps
		var t := (i + 0.5) / steps
		var col: Color = (colors[0] as Color).lerp(colors[1], t)
		var u := (t - 0.5) * 2.0
		var inset := r * (1.0 - sqrt(maxf(0.0, 1.0 - u * u)))
		draw_rect(Rect2(inset, y0, size.x - 2.0 * inset, y1 - y0), col)
	# 四角覆盖背景色
	draw_circle(Vector2(r, r), r, corner_color)
	draw_circle(Vector2(size.x - r, r), r, corner_color)
	draw_circle(Vector2(r, size.y - r), r, corner_color)
	draw_circle(Vector2(size.x - r, size.y - r), r, corner_color)
