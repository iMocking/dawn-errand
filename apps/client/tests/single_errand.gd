extends Control
## 单屏验证（errand 费用区）：注入会话 → shell tab1（跑腿下单）停留 40 秒。

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
	var shell: ShellView = App.navigator.current()
	if shell:
		shell.switch_to(1)
		print("switched to errand tab")
	print("SHOT: single_errand")
	await get_tree().create_timer(40.0).timeout
	get_tree().quit(0)
