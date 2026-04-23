# TileSets

TileSet `.tres` resources must be authored interactively in the Godot editor.

The TileMap + TileSet atlas configuration (tile shapes, collisions, autotile rules, terrain) requires a human to lay out tiles visually — it cannot be generated programmatically as part of the Unity→Godot migration.

Open the Godot editor, create a new TileSet resource here, and import the floor/wall PNGs from `../art/` as atlas sources.
