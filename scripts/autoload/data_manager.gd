extends Node

const BuildingModel = preload("res://scripts/models/building.gd")
const GraphicCardModel = preload("res://scripts/models/graphic_card.gd")
const PlayerModel = preload("res://scripts/models/player.gd")

const PLAYER_SAVE_PATH := "user://player.json"

var buildings: Array = []
var graphic_cards: Array = []
var player: RefCounted = null
var pop_logs: Array = []
var data_types: Array = []

func _ready() -> void:
	_load_buildings()
	_load_graphic_cards()
	_load_pop_logs()
	_load_data_types()
	_load_player()

func _notification(what: int) -> void:
	if what == NOTIFICATION_WM_CLOSE_REQUEST or what == NOTIFICATION_PREDELETE:
		_save_player()

func _read_json(res_path: String) -> Variant:
	var f := FileAccess.open(res_path, FileAccess.READ)
	if f == null:
		push_error("Failed to open %s" % res_path)
		return null
	var text := f.get_as_text()
	f.close()
	var parsed = JSON.parse_string(text)
	if parsed == null:
		push_error("Failed to parse JSON at %s" % res_path)
	return parsed

func _load_buildings() -> void:
	buildings = []
	var data = _read_json("res://data/buildings.json")
	if typeof(data) != TYPE_DICTIONARY:
		return
	var arr = data.get("Buildings", [])
	for entry in arr:
		if typeof(entry) != TYPE_DICTIONARY:
			continue
		var b = BuildingModel.new()
		b.from_dict(entry)
		buildings.append(b)

func _load_graphic_cards() -> void:
	graphic_cards = []
	var data = _read_json("res://data/graphiccards.json")
	if typeof(data) != TYPE_DICTIONARY:
		return
	var arr = data.get("GraphicCards", [])
	for entry in arr:
		if typeof(entry) != TYPE_DICTIONARY:
			continue
		var c = GraphicCardModel.new()
		c.from_dict(entry)
		graphic_cards.append(c)

func _load_pop_logs() -> void:
	pop_logs = []
	var data = _read_json("res://data/poplogs.json")
	if typeof(data) != TYPE_DICTIONARY:
		return
	pop_logs = data.get("Logs", [])

func _load_data_types() -> void:
	data_types = []
	var data = _read_json("res://data/datatypes.json")
	if typeof(data) != TYPE_DICTIONARY:
		return
	data_types = data.get("DataTypes", [])

func _load_player() -> void:
	var dict: Variant = null
	if FileAccess.file_exists(PLAYER_SAVE_PATH):
		var f := FileAccess.open(PLAYER_SAVE_PATH, FileAccess.READ)
		if f != null:
			var text := f.get_as_text()
			f.close()
			dict = JSON.parse_string(text)
	if typeof(dict) != TYPE_DICTIONARY:
		dict = _read_json("res://data/player.json")
	player = PlayerModel.new()
	if typeof(dict) == TYPE_DICTIONARY:
		player.from_dict(dict)

func _save_player() -> void:
	if player == null:
		return
	var f := FileAccess.open(PLAYER_SAVE_PATH, FileAccess.WRITE)
	if f == null:
		push_error("Failed to write %s" % PLAYER_SAVE_PATH)
		return
	f.store_string(JSON.stringify(player.to_dict()))
	f.close()
