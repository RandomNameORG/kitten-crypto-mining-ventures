extends RefCounted
class_name ResourceRef

var path: String = ""

func from_dict(d: Dictionary) -> void:
	path = String(d.get("Path", ""))

func to_dict() -> Dictionary:
	return {"Path": path}
