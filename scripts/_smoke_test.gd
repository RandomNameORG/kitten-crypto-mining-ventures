extends SceneTree

var _start_money: int = -1
var _elapsed: float = 0.0
var _ready_done: bool = false

func _initialize() -> void:
	print("[smoke] initialize")

func _process(delta: float) -> bool:
	if not _ready_done:
		var dm = get_root().get_node_or_null("/root/DataManager")
		if dm == null:
			return false
		var pm = get_root().get_node_or_null("/root/PlayerManager")
		var bm = get_root().get_node_or_null("/root/BuildingManager")
		var gm = get_root().get_node_or_null("/root/GraphicCardManager")
		if dm.player == null:
			push_error("[smoke] player not loaded")
			quit(1)
			return true
		_start_money = dm.player.money
		print("[smoke] player.money = ", _start_money)
		print("[smoke] current_building = ", bm.current_building.name if bm.current_building != null else "null")
		print("[smoke] graphic_cards count = ", gm.cards.size())
		print("[smoke] money_per_second = ", bm.get_money_per_second())
		_ready_done = true

	_elapsed += delta
	if _elapsed >= 2.5:
		var dm = get_root().get_node("/root/DataManager")
		var now: int = dm.player.money
		var delta_money: int = now - _start_money
		print("[smoke] after 2.5s money=", now, " delta=", delta_money)
		if delta_money <= 0:
			push_error("[smoke] FAIL: money did not increase")
			quit(1)
		else:
			print("[smoke] PASS")
			quit(0)
		return true
	return false
