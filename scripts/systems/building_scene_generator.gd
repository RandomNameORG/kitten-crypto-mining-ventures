extends RefCounted
class_name BuildingSceneGenerator

static func switch_to(container: Node, building_id: String) -> void:
	if container == null:
		return
	for child in container.get_children():
		if child is CanvasItem:
			(child as CanvasItem).visible = (child.name == building_id)

static func ensure_subtrees(container: Node, building_ids: Array) -> void:
	if container == null:
		return
	var existing: Dictionary = {}
	for child in container.get_children():
		existing[child.name] = true
	for id in building_ids:
		var sid := String(id)
		if sid == "":
			continue
		if not existing.has(sid):
			var node := Node2D.new()
			node.name = sid
			node.visible = false
			container.add_child(node)
