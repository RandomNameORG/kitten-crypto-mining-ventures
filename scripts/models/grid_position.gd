extends RefCounted
class_name GridPosition

var x: int = 0
var y: int = 0

func from_dict(d: Dictionary) -> void:
	x = int(d.get("X", 0))
	y = int(d.get("Y", 0))

func to_dict() -> Dictionary:
	return {"X": x, "Y": y}
