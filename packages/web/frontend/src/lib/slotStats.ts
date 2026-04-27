// slotStats — pure stats over the GPU list for a given room.
// Lifted out of <SlotMeter> so the buy CTA can consume the same numbers
// without forking presentation. Keep this file pure (no React import) so
// it stays unit-testable in vitest.

import type { GPU, Room, Snapshot } from "../types";

export type SlotTone = "ok" | "warn" | "danger";

export interface SlotStats {
  used: number;
  total: number;
  free: number;
  pct: number;
  running: number;
  shipping: number;
  broken: number;
  tone: SlotTone;
  cells: GPU[];
}

export function computeSlotStats(snapshot: Snapshot, room: Room | null): SlotStats | null {
  if (!room) return null;
  const cells = snapshot.gpus.filter((g) => g.room === room.id);
  const running = cells.filter((g) => g.status === "running").length;
  const shipping = cells.filter((g) => g.status === "shipping").length;
  const broken = cells.filter((g) => g.status === "broken").length;
  const used = cells.length;
  const total = room.slots;
  const free = Math.max(0, total - used);
  const pct = total > 0 ? Math.min(1, used / total) : 0;
  const tone: SlotTone = free === 0 ? "danger" : pct >= 0.8 ? "warn" : "ok";
  return { used, total, free, pct, running, shipping, broken, tone, cells };
}
