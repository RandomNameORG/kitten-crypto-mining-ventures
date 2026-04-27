import { useMemo, useRef, useState } from "react";
import { SlotMeter } from "../components/SlotMeter";
import type { Snapshot } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  dispatch: (id: string) => void;
}

const TIER_STRIPE: Record<string, string> = {
  trash: "bg-[oklch(70%_0.018_175/_0.55)]",
  common:
    "bg-[linear-gradient(180deg,var(--color-mint),oklch(70%_0.16_155))] [box-shadow:0_0_12px_oklch(82%_0.16_155/_0.5)]",
  rare: "bg-[linear-gradient(180deg,var(--color-blue),oklch(64%_0.13_240))] [box-shadow:0_0_12px_oklch(76%_0.13_240/_0.5)]",
  epic: "bg-[linear-gradient(180deg,var(--color-gold),oklch(74%_0.15_75))] [box-shadow:0_0_14px_oklch(85%_0.15_85/_0.55)]",
  legendary:
    "bg-[linear-gradient(180deg,var(--color-orange),oklch(60%_0.20_35))] [box-shadow:0_0_16px_oklch(72%_0.18_40/_0.6)]",
};

const TIER_CHIP: Record<string, string> = {
  trash: "text-muted border-muted/30 bg-muted/8",
  common: "text-mint border-mint/45 bg-mint/12",
  rare: "text-blue border-blue/45 bg-blue/12",
  epic: "text-gold border-gold/45 bg-gold/12 [box-shadow:0_0_8px_oklch(85%_0.15_85/_0.25)]",
  legendary: "text-orange border-orange/55 bg-orange/15 [box-shadow:0_0_10px_oklch(72%_0.18_40/_0.35)]",
};

const TIER_BG: Record<string, string> = {
  trash: "bg-panel/65",
  common: "bg-panel/65",
  rare: "bg-[radial-gradient(ellipse_80%_60%_at_80%_0%,oklch(40%_0.10_240/_0.18),transparent_60%),oklch(20%_0.014_200/_0.65)]",
  epic: "bg-[radial-gradient(ellipse_80%_60%_at_80%_0%,oklch(60%_0.18_85/_0.16),transparent_60%),oklch(20%_0.014_200/_0.65)]",
  legendary:
    "bg-[radial-gradient(ellipse_80%_60%_at_80%_0%,oklch(60%_0.18_35/_0.20),transparent_60%),oklch(22%_0.020_30/_0.55)] border-orange/45",
};

const TIER_ART_BG: Record<string, string> = {
  trash:
    "bg-[radial-gradient(circle_at_50%_35%,oklch(82%_0.16_155/_0.12),transparent_65%),oklch(8%_0.008_200)] border-line/55",
  common:
    "bg-[radial-gradient(circle_at_50%_35%,oklch(82%_0.16_155/_0.12),transparent_65%),oklch(8%_0.008_200)] border-mint/45",
  rare: "bg-[radial-gradient(circle_at_50%_35%,oklch(76%_0.13_240/_0.18),transparent_65%),oklch(8%_0.008_200)] border-blue/45",
  epic: "bg-[radial-gradient(circle_at_50%_35%,oklch(85%_0.15_85/_0.20),transparent_65%),oklch(8%_0.008_200)] border-gold/45",
  legendary:
    "bg-[radial-gradient(circle_at_50%_35%,oklch(72%_0.18_40/_0.24),transparent_65%),oklch(10%_0.010_30)] border-orange/55 [box-shadow:inset_0_0_14px_oklch(72%_0.18_40/_0.2)]",
};

export function StorePanel({ snapshot, dispatch }: Props) {
  const [pending, setPending] = useState<string | null>(null);
  const lastClickAt = useRef(0);

  const room = useMemo(
    () => snapshot.rooms.find((r) => r.id === snapshot.state.current_room) ?? null,
    [snapshot.rooms, snapshot.state.current_room],
  );
  const usedSlots = room
    ? snapshot.gpus.filter((g) => g.room === room.id).length
    : 0;
  const slotsFull = !!room && usedSlots >= room.slots;

  const ownedByDef = useMemo(() => {
    const map = new Map<string, { running: number; shipping: number; broken: number }>();
    for (const g of snapshot.gpus) {
      const k = g.def_id;
      if (!k) continue;
      const cur = map.get(k) ?? { running: 0, shipping: 0, broken: 0 };
      if (g.status === "running") cur.running += 1;
      else if (g.status === "shipping") cur.shipping += 1;
      else if (g.status === "broken") cur.broken += 1;
      map.set(k, cur);
    }
    return map;
  }, [snapshot.gpus]);

  const handleBuy = (id: string) => {
    const now = Date.now();
    if (now - lastClickAt.current < 350) return;
    lastClickAt.current = now;
    setPending(id);
    dispatch(id);
    window.setTimeout(() => setPending(null), 600);
  };

  return (
    <div className="grid gap-3">
      <SlotMeter snapshot={snapshot} room={room} />
      {slotsFull && (
        <div
          className="mb-2 px-3 py-1.5 text-[11px] text-orange border border-orange/45 rounded-md bg-[oklch(28%_0.10_40/_0.22)] uppercase tracking-wider font-mono"
          role="status"
        >
          🐈‍⬛ 机位已满 · 拆解或换房间后再下单
        </div>
      )}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {snapshot.gpu_defs.map((def) => {
          const tier = def.tier;
          const canAfford = snapshot.state.btc >= def.price;
          const inFlight = pending === def.id;
          const owned = ownedByDef.get(def.id);
          const totalOwned =
            (owned?.running ?? 0) + (owned?.shipping ?? 0) + (owned?.broken ?? 0);
          const disabled = !canAfford || inFlight || slotsFull;
          const label = inFlight
            ? "下单中…"
            : slotsFull
              ? "机位已满"
              : !canAfford
                ? "余额不足"
                : "立即购买";
          return (
            <article
              key={def.id}
              className={`relative grid gap-2.5 px-3 pt-3 pb-3 rounded-xl border border-line/45 overflow-hidden transition-[transform,border-color,box-shadow] duration-200 ${TIER_BG[tier] ?? TIER_BG.common} ${
                disabled
                  ? "opacity-80"
                  : "hover:-translate-y-0.5 hover:border-line hover:[box-shadow:0_18px_36px_oklch(0%_0_0_/_0.45),0_0_0_1px_oklch(82%_0.16_155/_0.18)]"
              }`}
            >
              <span
                aria-hidden
                className={`absolute inset-y-0 left-0 w-1 rounded-l ${TIER_STRIPE[tier] ?? TIER_STRIPE.common}`}
              />
              <header className="flex items-start justify-between gap-2 pl-1.5">
                <div className="min-w-0 flex items-center gap-2">
                  <span className="text-ink text-[14px] font-bold tracking-tight truncate">
                    {def.name}
                  </span>
                  <span
                    className={`flex-none px-1.5 py-0.5 text-[9px] uppercase tracking-[0.16em] font-bold font-mono rounded-sm border ${TIER_CHIP[tier] ?? TIER_CHIP.common}`}
                  >
                    {tier}
                  </span>
                </div>
                {totalOwned > 0 && (
                  <div
                    className="inline-flex items-center gap-1 px-1.5 py-0.5 text-[10px] text-mint border border-mint/45 bg-mint/10 rounded-md tabular-nums font-mono"
                    title="已拥有"
                  >
                    <span className="w-1.5 h-1.5 rounded-full bg-mint [box-shadow:0_0_6px_oklch(82%_0.16_155/_0.6)]" />
                    ×{totalOwned}
                  </div>
                )}
              </header>
              <div className="grid grid-cols-[88px_minmax(0,1fr)] gap-3 pl-1.5 items-center">
                <div
                  className={`relative w-[88px] h-[88px] rounded-lg border grid place-items-center overflow-hidden ${TIER_ART_BG[tier] ?? TIER_ART_BG.common}`}
                >
                  <img
                    className="w-full h-full object-contain [image-rendering:pixelated]"
                    src={gpuIconSrc(def.id)}
                    alt=""
                    loading="lazy"
                  />
                  {tier === "legendary" && (
                    <div
                      aria-hidden
                      className="pointer-events-none absolute inset-0 [background:linear-gradient(115deg,transparent_35%,oklch(100%_0_0/_0.18)_48%,transparent_60%)] [background-size:240%_100%] animate-[gpuShine_3.6s_linear_infinite]"
                    />
                  )}
                </div>
                <div className="min-w-0 grid gap-1.5">
                  <p className="m-0 text-muted text-[11px] leading-snug line-clamp-2">
                    {def.flavor}
                  </p>
                  <div className="grid grid-cols-3 gap-1">
                    <Stat label="效率" value={def.efficiency.toFixed(4)} />
                    <Stat label="功耗" value={def.power_draw.toFixed(2)} />
                    <Stat label="热量" value={def.heat_output.toFixed(2)} />
                  </div>
                </div>
              </div>
              <footer className="flex items-center justify-between gap-2 pl-1.5">
                <div className="grid leading-none">
                  <span className="text-[9px] text-muted/80 uppercase tracking-[0.18em] font-mono">
                    价格
                  </span>
                  <span
                    className={`mt-0.5 text-[18px] font-bold tabular-nums font-mono ${
                      tier === "legendary" ? "text-orange [text-shadow:0_0_14px_oklch(72%_0.18_40/_0.45)]" : "text-gold [text-shadow:0_0_12px_oklch(85%_0.15_85/_0.35)]"
                    }`}
                  >
                    {def.price_fmt}
                  </span>
                </div>
                <button
                  type="button"
                  className={`px-3.5 py-2 text-[12px] font-semibold tracking-wider uppercase rounded-md border transition-[transform,box-shadow,background-color] duration-150 ${
                    disabled
                      ? "text-muted bg-panel-2 border-line/55 cursor-not-allowed"
                      : "text-bg bg-mint border-mint [box-shadow:0_8px_18px_oklch(82%_0.16_155/_0.18),inset_0_-2px_0_oklch(0%_0_0/_0.2)] hover:-translate-y-px hover:[box-shadow:0_12px_24px_oklch(82%_0.16_155/_0.32),inset_0_-2px_0_oklch(0%_0_0/_0.2)]"
                  } ${inFlight ? "animate-[buyPulse_0.9s_var(--ease-out)_infinite]" : ""}`}
                  disabled={disabled}
                  onClick={() => handleBuy(def.id)}
                >
                  {label}
                </button>
              </footer>
              {owned && owned.shipping > 0 && (
                <div
                  aria-hidden
                  className="absolute top-2 right-2 inline-flex items-center gap-1 px-1.5 py-0.5 text-[10px] text-blue border border-blue/45 bg-bg/80 rounded-md font-mono backdrop-blur-md"
                >
                  <span className="animate-[shipBob_1.3s_var(--ease-out)_infinite]">🐈</span>
                  <span>📦 ×{owned.shipping}</span>
                </div>
              )}
            </article>
          );
        })}
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-px px-1.5 py-1 rounded-md border border-line/40 bg-bg/35">
      <span className="text-[9px] text-muted/80 uppercase tracking-wider font-mono">
        {label}
      </span>
      <span className="text-ink text-[12px] font-semibold tabular-nums leading-none font-mono">
        {value}
      </span>
    </div>
  );
}
