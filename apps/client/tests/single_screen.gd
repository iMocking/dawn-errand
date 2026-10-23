extends Control
## 单屏验证：注入会话 → shell 首页停留 40 秒，外部 PrintWindow 固定间隔抓取。

func _ready() -> void:
	var nav := Navigator.new()
	add_child(nav)
	App.navigator = nav
	_run()

func _run() -> void:
	await get_tree().create_timer(0.6).timeout
	App.open_screen("splash")
	await get_tree().create_timer(1.0).timeout
	App.session.apply_login({
		"token": "smoke-access-token",
		"refresh_token": "smoke-refresh-token",
		"user": {"id": "u-test-001", "name": "测试用户", "phone": "13888888888", "role": "user"},
		"permissions": [],
	})
	await get_tree().create_timer(2.0).timeout
	if not (App.navigator.current() is ShellView):
		print("WARN: 未进入主界面外壳")
		return
	print("SHOT: single_home_begin")
	await get_tree().create_timer(45.0).timeout
	print("SHOT: single_home_end")
	get_tree().quit(0)
