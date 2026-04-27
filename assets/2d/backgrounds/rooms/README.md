# Room Backgrounds

Room background assets are keyed by the runtime room IDs in
`core/data/rooms.json`.

Each room directory contains:

- `background.png`: 512x288 pixel-art background for the room.

The previous placeholder folders used broad art-direction names
(`bedroom_mine`, `garage_workshop`, `ice_server_room`, `neon_warehouse`,
`pirate_harbor`). They did not map cleanly to the current game data, which
now has eight rooms. New room backgrounds should use the room ID directly.
