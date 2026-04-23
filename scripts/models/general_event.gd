extends RefCounted
class_name GeneralEvent

var event_probs: float = 0.0
var money_earn: float = 0.0
var volt_earn: float = 0.0

func from_dict(d: Dictionary) -> void:
	event_probs = float(d.get("EventProbs", 0.0))
	money_earn = float(d.get("MoneyEarn", 0.0))
	volt_earn = float(d.get("VoltEarn", 0.0))

func to_dict() -> Dictionary:
	return {
		"EventProbs": event_probs,
		"MoneyEarn": money_earn,
		"VoltEarn": volt_earn,
	}
