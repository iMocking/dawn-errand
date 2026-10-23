class_name UserModel
extends RefCounted
## 用户模型（登录/me 返回）。

var id: String = ""
var name: String = ""
var avatar: String = ""
var role: String = "user"
var email: String = ""
var phone: String = ""

static func from_dict(d: Dictionary) -> UserModel:
	var u := UserModel.new()
	u.id = str(d.get("id", ""))
	u.name = str(d.get("name", ""))
	u.avatar = str(d.get("avatar", ""))
	u.role = str(d.get("role", "user"))
	u.email = str(d.get("email", ""))
	u.phone = str(d.get("phone", ""))
	return u

func masked_phone() -> String:
	return Fmt.mask_phone(phone)
