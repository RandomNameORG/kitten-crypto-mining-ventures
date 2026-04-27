// Tier visual scale.
// Single source of truth for how a tier reads at a glance. The point is that
// "legendary feels 5x of trash, not 1.5x" — that takes a SYSTEM (frame +
// typography + motion), not just brighter chrome. Absence of motion on lower
// tiers is itself the signal.

export type Tier = "trash" | "common" | "rare" | "epic" | "legendary";

export interface TierStyle {
  // left stripe inside cards
  stripe: string;
  // small uppercase chip tag
  chip: string;
  // card body background — gated for perf: trash/common stay flat, the heavy
  // radial-gradient washes only paint on the higher tiers where they earn it
  body: string;
  // art-frame border + radial wash inside the card thumb
  art: string;
  // inline TIER_SCALE: frame weight, name typography, motion class
  frame: string;
  name: string;
  motion: string;
  // price color + glow
  priceText: string;
  priceShadow: string;
}

export const TIER_STYLE: Record<Tier, TierStyle> = {
  trash: {
    stripe: "bg-[oklch(70%_0.018_175/_0.55)]",
    chip: "text-muted border-muted/30 bg-muted/8",
    body: "bg-panel/65",
    art: "bg-[oklch(8%_0.008_200)] border-line/55",
    frame: "border",
    name: "text-[12px]",
    motion: "",
    priceText: "text-gold",
    priceShadow: "[text-shadow:0_0_12px_oklch(85%_0.15_85/_0.35)]",
  },
  common: {
    stripe: "bg-[linear-gradient(180deg,var(--color-mint),oklch(70%_0.16_155))] [box-shadow:0_0_12px_oklch(82%_0.16_155/_0.5)]",
    chip: "text-mint border-mint/45 bg-mint/12",
    body: "bg-panel/65",
    art: "bg-[radial-gradient(circle_at_50%_35%,oklch(82%_0.16_155/_0.12),transparent_65%),oklch(8%_0.008_200)] border-mint/45",
    frame: "border",
    name: "text-[13px]",
    motion: "",
    priceText: "text-gold",
    priceShadow: "[text-shadow:0_0_12px_oklch(85%_0.15_85/_0.35)]",
  },
  rare: {
    stripe: "bg-[linear-gradient(180deg,var(--color-blue),oklch(64%_0.13_240))] [box-shadow:0_0_12px_oklch(76%_0.13_240/_0.5)]",
    chip: "text-blue border-blue/45 bg-blue/12",
    body: "bg-[radial-gradient(ellipse_80%_60%_at_80%_0%,oklch(40%_0.10_240/_0.18),transparent_60%),oklch(20%_0.014_200/_0.65)]",
    art: "bg-[radial-gradient(circle_at_50%_35%,oklch(76%_0.13_240/_0.18),transparent_65%),oklch(8%_0.008_200)] border-blue/45",
    frame: "border-2",
    name: "text-[13px]",
    motion: "",
    priceText: "text-gold",
    priceShadow: "[text-shadow:0_0_12px_oklch(85%_0.15_85/_0.35)]",
  },
  epic: {
    stripe: "bg-[linear-gradient(180deg,var(--color-gold),oklch(74%_0.15_75))] [box-shadow:0_0_14px_oklch(85%_0.15_85/_0.55)]",
    chip: "text-gold border-gold/45 bg-gold/12 [box-shadow:0_0_8px_oklch(85%_0.15_85/_0.25)]",
    body: "bg-[radial-gradient(ellipse_80%_60%_at_80%_0%,oklch(60%_0.18_85/_0.16),transparent_60%),oklch(20%_0.014_200/_0.65)]",
    art: "bg-[radial-gradient(circle_at_50%_35%,oklch(85%_0.15_85/_0.20),transparent_65%),oklch(8%_0.008_200)] border-gold/45",
    frame: "border-2",
    name: "text-[14px] text-gold",
    motion: "",
    priceText: "text-gold",
    priceShadow: "[text-shadow:0_0_12px_oklch(85%_0.15_85/_0.35)]",
  },
  legendary: {
    stripe: "bg-[linear-gradient(180deg,var(--color-orange),oklch(60%_0.20_35))] [box-shadow:0_0_16px_oklch(72%_0.18_40/_0.6)]",
    chip: "text-orange border-orange/55 bg-orange/15 [box-shadow:0_0_10px_oklch(72%_0.18_40/_0.35)]",
    body: "bg-[radial-gradient(ellipse_80%_60%_at_80%_0%,oklch(60%_0.18_35/_0.20),transparent_60%),oklch(22%_0.020_30/_0.55)] border-orange/45",
    art: "bg-[radial-gradient(circle_at_50%_35%,oklch(72%_0.18_40/_0.24),transparent_65%),oklch(10%_0.010_30)] border-orange/55 [box-shadow:inset_0_0_14px_oklch(72%_0.18_40/_0.2)]",
    frame: "border-2 [box-shadow:inset_0_0_0_1px_oklch(72%_0.18_40/_0.6)]",
    name: "text-[14px] font-bold text-orange",
    motion: "animate-[gpuLegendaryGlow_4s_ease-in-out_infinite]",
    priceText: "text-orange",
    priceShadow: "[text-shadow:0_0_14px_oklch(72%_0.18_40/_0.45)]",
  },
};

export function tierStyle(tier: string): TierStyle {
  return (TIER_STYLE[tier as Tier] ?? TIER_STYLE.common);
}

const TIER_RANK: Record<Tier, number> = {
  trash: 0,
  common: 1,
  rare: 2,
  epic: 3,
  legendary: 4,
};

export function tierRank(tier: string): number {
  return TIER_RANK[tier as Tier] ?? 1;
}
