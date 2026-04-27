import { describe, expect, it } from "vitest";
import { computeSlotStats } from "./slotStats";
import type { GPU, Room, Snapshot } from "../types";

function gpu(instance_id: number, room: string, status: string, def_id = "x"): GPU {
  return {
    instance_id,
    def_id,
    name: def_id,
    status,
    room,
    upgrade: 0,
    oc_level: 0,
    hours_left: 1,
    earn_fmt: "0/s",
    repairable: status === "broken",
  };
}

function snap(gpus: GPU[]): Snapshot {
  return {
    state: {} as Snapshot["state"],
    rooms: [],
    gpus,
    gpu_defs: [],
    skills: [],
    mercs: [],
    merc_defs: [],
    log: [],
    modifiers: [],
    research_tiers: [],
    blueprints: [],
    achievements: [],
    achievement_defs: [],
    mastery_levels: {},
    mastery_tracks: [],
    legacy_perks: [],
    legacy: {} as Snapshot["legacy"],
    stats: {} as Snapshot["stats"],
    ok: true,
  };
}

const room: Room = {
  id: "basement",
  name: "Basement",
  flavor: "",
  slots: 4,
  unlock_cost: 0,
  unlock_cost_fmt: "0",
  unlocked: true,
  current: true,
  gpu_count: 0,
  heat: 0,
  max_heat: 100,
  heat_pct: 0,
  heat_delta: 0,
  heat_tick_in: 0,
  earn_fmt: "0/s",
  bill_fmt: "0/s",
  net_fmt: "0/s",
  defense: { lock: 0, cctv: 0, wiring: 0, cooling: 0, armor: 0 },
  background: "",
};

describe("computeSlotStats", () => {
  it("returns null when room is null", () => {
    expect(computeSlotStats(snap([]), null)).toBeNull();
  });

  it("counts running / shipping / broken; remaining are free", () => {
    const s = snap([
      gpu(1, "basement", "running"),
      gpu(2, "basement", "running"),
      gpu(3, "basement", "shipping"),
      gpu(4, "basement", "broken"),
    ]);
    const stats = computeSlotStats(s, room)!;
    expect(stats.running).toBe(2);
    expect(stats.shipping).toBe(1);
    expect(stats.broken).toBe(1);
    expect(stats.used).toBe(4);
    expect(stats.free).toBe(0);
  });

  it("ignores GPUs in other rooms", () => {
    const s = snap([gpu(1, "loft", "running"), gpu(2, "basement", "running")]);
    expect(computeSlotStats(s, room)!.used).toBe(1);
  });

  it("tone: ok < 80%, warn 80-99%, danger at full", () => {
    expect(computeSlotStats(snap([gpu(1, "basement", "running")]), room)!.tone).toBe("ok");
    // warn at 4-slot * 0.8 = 3.2 → 4 slots running 4/4 is danger; need 3 of 4 = 75% (still ok)
    // adjust: pct >= 0.8 needs at least 3.2/4 = 4 → only 4/4 = 1.0 hits, but that's danger first.
    // For warn we need a 5-slot room; reuse room with slots: 5
    const r5 = { ...room, slots: 5 };
    const four = snap([
      gpu(1, "basement", "running"),
      gpu(2, "basement", "running"),
      gpu(3, "basement", "running"),
      gpu(4, "basement", "running"),
    ]);
    expect(computeSlotStats(four, r5)!.tone).toBe("warn");
    const full = snap([
      gpu(1, "basement", "running"),
      gpu(2, "basement", "running"),
      gpu(3, "basement", "running"),
      gpu(4, "basement", "running"),
    ]);
    expect(computeSlotStats(full, room)!.tone).toBe("danger");
  });
});
