// Sort + filter helpers for the Store. Pure functions, vitest-tested.

import { tierRank } from "../components/tier";
import type { GPUDef } from "../types";

export function sortGpuDefs(defs: GPUDef[]): GPUDef[] {
  // Tier ladder first (legendary at top), then efficiency descending so the
  // best-in-tier rises. Stable sort: identical tiers preserve source order.
  return [...defs].sort((a, b) => {
    const tr = tierRank(b.tier) - tierRank(a.tier);
    if (tr !== 0) return tr;
    return b.efficiency - a.efficiency;
  });
}

export function affordableOnly(defs: GPUDef[], btc: number): GPUDef[] {
  return defs.filter((d) => d.price <= btc);
}
