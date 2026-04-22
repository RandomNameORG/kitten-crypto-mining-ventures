extends Node

signal log_requested(message: String, fade_duration: float)

var _messages: Dictionary = {}

func _ready() -> void:
	_messages.clear()
	for entry in DataManager.pop_logs:
		if typeof(entry) != TYPE_DICTIONARY:
			continue
		var lt: int = int(entry.get("LogType", -1))
		var msg: String = String(entry.get("Message", ""))
		if lt >= 0:
			_messages[lt] = msg

func get_message(log_type: int) -> String:
	return String(_messages.get(log_type, ""))

func show(log_type: int, fade_duration: float = 1.0) -> void:
	var msg := get_message(log_type)
	log_requested.emit(msg, fade_duration)
