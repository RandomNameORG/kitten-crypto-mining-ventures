extends RefCounted
class_name Player

var name: String = ""
var tech_point: int = 0
var money: int = 0
var total_card_num: int = 0
var curr_building_id: String = ""
var curr_building_name: String = ""
var buildings: Array = []

func from_dict(d: Dictionary) -> void:
	name = String(d.get("Name", ""))
	tech_point = int(d.get("TechPoint", 0))
	money = int(d.get("Money", 0))
	total_card_num = int(d.get("TotalCardNum", 0))
	var curr = d.get("CurrBuildingAt", {})
	if typeof(curr) == TYPE_DICTIONARY:
		curr_building_id = String(curr.get("Id", ""))
		curr_building_name = String(curr.get("Name", ""))
	buildings = []
	var refs = d.get("BuildingsRef", [])
	if typeof(refs) == TYPE_ARRAY:
		for r in refs:
			if typeof(r) == TYPE_DICTIONARY:
				buildings.append({
					"id": String(r.get("Id", "")),
					"name": String(r.get("Name", "")),
				})

func to_dict() -> Dictionary:
	var refs: Array = []
	for b in buildings:
		refs.append({"Id": b.get("id", ""), "Name": b.get("name", "")})
	return {
		"Name": name,
		"TechPoint": tech_point,
		"Money": money,
		"TotalCardNum": total_card_num,
		"CurrBuildingAt": {"Id": curr_building_id, "Name": curr_building_name},
		"BuildingsRef": refs,
	}
