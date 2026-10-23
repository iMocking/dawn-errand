extends Node
## API 客户端（自动加载单例 Api）：统一封装后端返回结构
## { code, message, body, request_id }，自动附加 Bearer Token、
## 401 自动刷新（单飞）与错误映射。业务层一律 await Api.request(...)。

signal session_expired

const METHOD_GET := HTTPClient.METHOD_GET
const METHOD_POST := HTTPClient.METHOD_POST
const METHOD_PUT := HTTPClient.METHOD_PUT
const METHOD_DELETE := HTTPClient.METHOD_DELETE

var base_url: String = ""

func _ready() -> void:
	base_url = ProjectSettings.get_setting("client/api/base_url", "http://127.0.0.1:8080")
	base_url = String(base_url).strip_edges()
	if base_url.ends_with("/"):
		base_url = base_url.substr(0, base_url.length() - 1)

func _gen_rid() -> String:
	return "gd-" + str(Time.get_unix_time_from_system()) + "-" + str(randi() % 100000)

## 请求入口。返回字典：{ok, http_status, code, message, body, request_id}
func request(method: int, path: String, payload: Dictionary = {}, use_auth: bool = true, allow_refresh: bool = true) -> Dictionary:
	var res := await _request_once(method, path, payload, use_auth)
	if not res.ok and res.http_status == 401 and use_auth and allow_refresh and App.session.has_refresh_token():
		var refreshed := await App.session.refresh_tokens()
		if refreshed:
			res = await _request_once(method, path, payload, use_auth)
		else:
			session_expired.emit()
	return res

func _request_once(method: int, path: String, payload: Dictionary, use_auth: bool) -> Dictionary:
	var req := HTTPRequest.new()
	add_child(req)
	req.timeout = 15
	var headers := PackedStringArray([
		"Content-Type: application/json",
		"Accept: application/json",
		"X-Request-ID: " + _gen_rid(),
	])
	if use_auth and App.session.has_access_token():
		headers.append("Authorization: Bearer " + App.session.access_token)
	var body_str := ""
	if not payload.is_empty():
		body_str = JSON.stringify(payload)
	var err := req.request(base_url + path, headers, method, body_str)
	if err != OK:
		req.queue_free()
		return {ok = false, http_status = 0, code = -1, message = "网络请求失败（无法连接服务器）", body = {}, request_id = ""}
	var resp: Array = await req.request_completed
	req.queue_free()
	var status: int = resp[1]
	var resp_body: PackedByteArray = resp[3]
	var json_text := resp_body.get_string_from_utf8()
	return _parse(status, json_text)

func _parse(status: int, json_text: String) -> Dictionary:
	if status == 0:
		return {ok = false, http_status = 0, code = -1, message = "网络请求失败（连接超时或服务器不可达）", body = {}, request_id = ""}
	var data := {}
	var t := json_text.strip_edges()
	if t != "":
		var parsed = JSON.parse_string(t)
		if parsed is Dictionary:
			data = parsed
	if data.has("code"):
		var code := int(data.code)
		if code == 0:
			return {ok = true, http_status = status, code = 0, message = str(data.get("message", "")), body = data.get("body", {}), request_id = str(data.get("request_id", ""))}
		return {ok = false, http_status = status, code = code, message = str(data.get("message", "请求失败")), body = data.get("body", {}), request_id = str(data.get("request_id", ""))}
	return {ok = false, http_status = status, code = -1, message = "服务异常（HTTP " + str(status) + "）", body = {}, request_id = ""}
