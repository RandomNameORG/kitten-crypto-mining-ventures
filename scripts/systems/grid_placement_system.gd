extends Node2D
class_name GridPlacementSystem

const VALID_COLOR := Color(0.2, 1.0, 0.2, 0.8)
const INVALID_COLOR := Color(1.0, 0.2, 0.2, 0.8)

@export var tile_map: TileMap
@export var cell_indicator: Node2D
@export var input_manager_path: NodePath

var _active: bool = false
var selected_card_id: String = ""
var _input_manager: Node

func _ready() -> void:
	if input_manager_path != NodePath(""):
		_input_manager = get_node_or_null(input_manager_path)
	if cell_indicator != null:
		cell_indicator.visible = false

func start_placement(id: int) -> void:
	selected_card_id = str(id)
	_active = true
	if cell_indicator != null:
		cell_indicator.visible = true
	if _input_manager != null:
		if not _input_manager.clicked.is_connected(_on_clicked):
			_input_manager.clicked.connect(_on_clicked)
		if not _input_manager.exit.is_connected(stop_placement):
			_input_manager.exit.connect(stop_placement)

func stop_placement() -> void:
	_active = false
	if cell_indicator != null:
		cell_indicator.visible = false
	if _input_manager != null:
		if _input_manager.clicked.is_connected(_on_clicked):
			_input_manager.clicked.disconnect(_on_clicked)
		if _input_manager.exit.is_connected(stop_placement):
			_input_manager.exit.disconnect(stop_placement)

func _current_cell() -> Vector2i:
	if tile_map == null:
		return Vector2i.ZERO
	var mouse := get_global_mouse_position()
	return tile_map.local_to_map(tile_map.to_local(mouse))

func _is_cell_valid(cell: Vector2i) -> bool:
	var b := BuildingManager.current_building
	if b == null:
		return true
	var cards: Array = BuildingManager._cards_for(b.id)
	for entry in cards:
		if typeof(entry) == TYPE_DICTIONARY:
			var p = entry.get("pos", null)
			if p is Vector2i and p == cell:
				return false
	return true

func _process(_delta: float) -> void:
	if not _active or tile_map == null or cell_indicator == null:
		return
	var cell := _current_cell()
	cell_indicator.global_position = tile_map.to_global(tile_map.map_to_local(cell))
	if cell_indicator is CanvasItem:
		(cell_indicator as CanvasItem).modulate = VALID_COLOR if _is_cell_valid(cell) else INVALID_COLOR

func _on_clicked() -> void:
	if not _active or tile_map == null:
		return
	var cell := _current_cell()
	if not _is_cell_valid(cell):
		return
	var b := BuildingManager.current_building
	if b == null:
		return
	BuildingManager.place_card(b.id, selected_card_id, cell)
