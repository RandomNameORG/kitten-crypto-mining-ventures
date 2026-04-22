extends RefCounted
class_name GraphicCard

var name: String = ""
var id: String = ""
var is_locked: bool = false
var per_second_earn: int = 0
var price: int = 0
var per_second_lose_volt: int = 0
var quantity: int = 0
var image_source_path: String = ""

func from_dict(d: Dictionary) -> void:
	name = String(d.get("Name", ""))
	id = String(d.get("Id", ""))
	is_locked = bool(d.get("IsLocked", false))
	per_second_earn = int(d.get("PerSecondEarn", 0))
	price = int(d.get("Price", 0))
	per_second_lose_volt = int(d.get("PerSecondLoseVolt", 0))
	quantity = int(d.get("Quantity", 0))
	var img = d.get("ImageSource", {})
	if typeof(img) == TYPE_DICTIONARY:
		image_source_path = String(img.get("Path", ""))

func to_dict() -> Dictionary:
	return {
		"Name": name,
		"Id": id,
		"IsLocked": is_locked,
		"PerSecondEarn": per_second_earn,
		"Price": price,
		"PerSecondLoseVolt": per_second_lose_volt,
		"Quantity": quantity,
		"ImageSource": {"Path": image_source_path},
	}
