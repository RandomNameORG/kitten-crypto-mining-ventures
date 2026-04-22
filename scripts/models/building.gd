extends RefCounted
class_name Building

var id: String = ""
var name: String = ""
var grid_size: int = 0
var volt_per_second: int = 0
var money_per_second: int = 0
var max_volt: int = 0
var max_card_num: int = 0
var probability_of_being_attacked: float = 0.0
var heat_dissipation_level: int = 0
var location_of_the_building: int = 0
var decorations: Array = []
var cats: Array = []
var alts: Array = []
var events: Array = []

func from_dict(d: Dictionary) -> void:
	id = String(d.get("Id", ""))
	name = String(d.get("Name", ""))
	grid_size = int(d.get("GridSize", 0))
	volt_per_second = int(d.get("VoltPerSecond", 0))
	money_per_second = int(d.get("MoneyPerSecond", 0))
	max_volt = int(d.get("MaxVolt", 0))
	max_card_num = int(d.get("MaxCardNum", 0))
	probability_of_being_attacked = float(d.get("ProbabilityOfBeingAttacked", 0.0))
	heat_dissipation_level = int(d.get("HeatDissipationLevel", 0))
	location_of_the_building = int(d.get("LocationOfTheBuilding", 0))
	decorations = []
	cats = []
	alts = []
	events = []

func to_dict() -> Dictionary:
	return {
		"Id": id,
		"Name": name,
		"GridSize": grid_size,
		"VoltPerSecond": volt_per_second,
		"MoneyPerSecond": money_per_second,
		"MaxVolt": max_volt,
		"MaxCardNum": max_card_num,
		"ProbabilityOfBeingAttacked": probability_of_being_attacked,
		"HeatDissipationLevel": heat_dissipation_level,
		"LocationOfTheBuilding": location_of_the_building,
		"Decorations": decorations,
		"Cats": cats,
		"Alts": alts,
		"Events": events,
	}
