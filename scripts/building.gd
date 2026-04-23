extends Node2D

@onready var buildings_container: Node = $Buildings

func _ready() -> void:
	var ids: Array = []
	for b in DataManager.buildings:
		ids.append(b.id)
	BuildingSceneGenerator.ensure_subtrees(buildings_container, ids)
	var current = BuildingManager.current_building
	if current != null:
		BuildingSceneGenerator.switch_to(buildings_container, current.id)
	if BuildingManager.building_changed.is_connected(_on_building_changed):
		return
	BuildingManager.building_changed.connect(_on_building_changed)

func _on_building_changed(id: String) -> void:
	BuildingSceneGenerator.switch_to(buildings_container, id)
