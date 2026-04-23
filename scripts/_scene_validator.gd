extends SceneTree

func _initialize() -> void:
	var scenes := [
		"res://scenes/main_menu.tscn",
		"res://scenes/main.tscn",
		"res://scenes/building.tscn",
		"res://scenes/store.tscn",
		"res://scenes/settings.tscn",
		"res://scenes/ui/store_item_slot.tscn",
		"res://scenes/ui/cell_indicator.tscn",
		"res://scenes/ui/esc_window.tscn",
		"res://scenes/ui/storage_menu.tscn",
	]
	var failures := 0
	for path in scenes:
		var ps: PackedScene = load(path)
		if ps == null:
			print("[validator] FAIL load: ", path)
			failures += 1
			continue
		var inst = ps.instantiate()
		if inst == null:
			print("[validator] FAIL instantiate: ", path)
			failures += 1
			continue
		print("[validator] OK: ", path, " -> ", inst.get_class())
		inst.queue_free()
	if failures > 0:
		print("[validator] ", failures, " failures")
		quit(1)
	else:
		print("[validator] all ", scenes.size(), " scenes load+instantiate cleanly")
		quit(0)

func _process(_delta: float) -> bool:
	return false
