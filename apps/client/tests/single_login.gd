extends Control
## 单屏验证（login 协议行）：直接打开登录页停留 30 秒。

func _ready() -> void:
	var nav := Navigator.new()
	add_child(nav)
	App.navigator = nav
	_run()

func _run() -> void:
	App.session.clear()
	await get_tree().create_timer(0.5).timeout
	App.open_screen("splash")
	await get_tree().create_timer(1.2).timeout
	App.open_screen("login")
	print("SHOT: single_login")
	await get_tree().create_timer(30.0).timeout
	get_tree().quit(0)
