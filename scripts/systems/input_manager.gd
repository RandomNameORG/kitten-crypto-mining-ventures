extends Node
class_name PlacementInputManager

signal clicked
signal exit

func get_mouse_world_pos(camera: Camera2D) -> Vector2:
	if camera == null:
		return Vector2.ZERO
	return camera.get_global_mouse_position()

func _unhandled_input(event: InputEvent) -> void:
	if event.is_action_pressed("placement_confirm"):
		clicked.emit()
	elif event.is_action_pressed("placement_cancel"):
		exit.emit()
