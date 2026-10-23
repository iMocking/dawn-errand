class_name SessionModel
extends RefCounted
## 会话模型：访问/刷新令牌与当前用户。负责本地持久化与 401 自动续期（单飞）。

const SAVE_PATH := "user://session.json"

var access_token := ""
var refresh_token := ""
var user: Dictionary = {}         # {id,name,avatar,role,...}
var permissions: Array = []

signal _refresh_done(ok: bool)

var _refreshing := false

func has_access_token() -> bool:
	return access_token != ""

func has_refresh_token() -> bool:
	return refresh_token != ""

func is_logged_in() -> bool:
	return access_token != "" and not user.is_empty()

func user_name() -> String:
	return str(user.get("name", "用户"))

func user_phone() -> String:
	return str(user.get("phone", ""))

func user_id() -> String:
	return str(user.get("id", ""))

func apply_login(result: Dictionary) -> void:
	access_token = str(result.get("token", ""))
	refresh_token = str(result.get("refresh_token", ""))
	var u = result.get("user", {})
	user = u if u is Dictionary else {}
	var p = result.get("permissions", [])
	permissions = p if p is Array else []
	save()

func apply_me(body: Dictionary) -> void:
	var u = body.get("user", {})
	if u is Dictionary and not u.is_empty():
		user = u
		var p = body.get("permissions", [])
		permissions = p if p is Array else []
		save()

func clear() -> void:
	access_token = ""
	refresh_token = ""
	user = {}
	permissions = []
	var err := DirAccess.remove_absolute(ProjectSettings.globalize_path(SAVE_PATH))
	if err != OK and err != ERR_DOES_NOT_EXIST:
		push_warning("session clear: cannot remove save file")

func save() -> void:
	var f := FileAccess.open(SAVE_PATH, FileAccess.WRITE)
	if f == null:
		push_warning("session save failed: %s" % FileAccess.get_open_error())
		return
	f.store_string(JSON.stringify({
		"access_token": access_token,
		"refresh_token": refresh_token,
		"user": user,
		"permissions": permissions,
	}))
	f.close()

func load() -> void:
	if not FileAccess.file_exists(SAVE_PATH):
		return
	var f := FileAccess.open(SAVE_PATH, FileAccess.READ)
	if f == null:
		return
	var text := f.get_as_text()
	f.close()
	var parsed = JSON.parse_string(text)
	if parsed is Dictionary:
		access_token = str(parsed.get("access_token", ""))
		refresh_token = str(parsed.get("refresh_token", ""))
		var u = parsed.get("user", {})
		user = u if u is Dictionary else {}
		var p = parsed.get("permissions", [])
		permissions = p if p is Array else []

## 401 自动续期：并发请求只发一次 refresh（单飞），刷新成功后重放原请求。
func refresh_tokens() -> bool:
	if not has_refresh_token():
		return false
	if _refreshing:
		var result: bool = await _refresh_done
		return result
	_refreshing = true
	var res: Dictionary = await Api.request(Api.METHOD_POST, "/api/auth/refresh", {"refresh_token": refresh_token}, false, false)
	var ok: bool = res.ok
	if ok and res.body is Dictionary:
		apply_login(res.body)
	_refreshing = false
	_refresh_done.emit(ok)
	return ok
