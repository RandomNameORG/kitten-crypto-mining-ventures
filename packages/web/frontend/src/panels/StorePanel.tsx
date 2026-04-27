import { useMemo, useRef, useState } from "react";
import { SlotMeter } from "../components/SlotMeter";
import { tierStyle } from "../components/tier";
import { computeSlotStats } from "../lib/slotStats";
import { affordableOnly, sortGpuDefs } from "../lib/sort";
import type { Snapshot } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  dispatch: (id: string) => void;
}

export function StorePanel({ snapshot, dispatch }: Props) {
  const [pending, setPending] = useState<string | null>(null);
  const [affordOnly, setAffordOnly] = useState(false);
  const lastClickAt = useRef(0);

  const room = useMemo(
    () => snapshot.rooms.find((r) => r.id === snapshot.state.current_room) ?? null,
    [snapshot.rooms, snapshot.state.current_room],
  );
  const slot = computeSlotStats(snapshot, room);
  const slotsFull = !!slot && slot.free === 0;

  const visible = useMemo(() => {
    let defs = sortGpuDefs(snapshot.gpu_defs);
    if (affordOnly) defs = affordableOnly(defs, snapshot.state.btc);
    return defs;
  }, [snapshot.gpu_defs, snapshot.state.btc, affordOnly]);

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

      <div className="flex items-center justify-between gap-2">
        <div className="text-[10px] text-muted/80 uppercase tracking-wider font-mono">
          {visible.length} / {snapshot.gpu_defs.length} 显卡
        </div>
        <button
          type="button"
          onClick={() => setAffordOnly((v) => !v)}
          className={`px-2 py-1 rounded-md text-[11px] uppercase tracking-wider font-mono border transition-colors duration-150 ${
            affordOnly
              ? "text-mint border-mint/55 bg-mint/12"
              : "text-muted border-line/45 bg-bg/35 hover:border-line"
          }`}
          aria-pressed={affordOnly}
        >
          只看买得起
        </button>
      </div>

      {visible.length === 0 && affordOnly && (
        <div className="text-muted text-xs px-3 py-6 text-center">
          暂时没有买得起的显卡 · 卖一张拆解或挖一会儿
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {visible.map((def) => {
          const tier = tierStyle(def.tier);
          const canAfford = snapshot.state.btc >= def.price;
          const inFlight = pending === def.id;
          const disabled = !canAfford || inFlight || slotsFull;
          // 6-state CTA: in-flight, full, broke+full, broke, full, normal.
          const buttonLabel = inFlight
            ? "下单中…"
            : slotsFull && !canAfford
              ? "机位已满"
              : slotsFull
                ? "机位已满"
                : !canAfford
                  ? "余额不足"
                  : "购买";

          return (
            <article
              key={def.id}
              className={`relative grid gap-2.5 px-3 pt-3 pb-3 rounded-xl ${tier.frame} border-line/45 overflow-hidden transition-[transform,border-color,box-shadow] duration-200 ${tier.body} ${tier.motion} ${
                disabled
                  ? "opacity-80"
                  : "hover:-translate-y-0.5 hover:border-line hover:[box-shadow:0_18px_36px_oklch(0%_0_0_/_0.45),0_0_0_1px_oklch(82%_0.16_155/_0.18)]"
              }`}
            >
              <span aria-hidden className={`absolute inset-y-0 left-0 w-1 rounded-l ${tier.stripe}`} />

              <header className="flex items-start justify-between gap-2 pl-1.5">
                <div className="min-w-0 flex items-center gap-2">
                  <span className={`tracking-tight truncate font-bold ${tier.name}`}>
                    {def.name}
                  </span>
                  <span
                    className={`flex-none px-1.5 py-0.5 text-[9px] uppercase tracking-[0.16em] font-bold font-mono rounded-sm border ${tier.chip}`}
                  >
                    {def.tier}
                  </span>
                </div>
              </header>

              <div className="grid grid-cols-[88px_minmax(0,1fr)] gap-3 pl-1.5 items-center">
                <div className={`relative w-[88px] h-[88px] rounded-lg border grid place-items-center overflow-hidden ${tier.art}`}>
                  <img
                    className="w-full h-full object-contain [image-rendering:pixelated]"
                    src={gpuIconSrc(def.id)}
                    alt=""
                    loading="lazy"
                  />
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

              <footer className="grid gap-1.5 pl-1.5">
                <div className="flex items-baseline justify-between gap-2 text-[10px] text-muted/80 font-mono uppercase tracking-wider">
                  <span>
                    价格
                    <span className={`ml-1.5 normal-case tracking-normal text-[14px] font-bold tabular-nums ${tier.priceText} ${tier.priceShadow}`}>
                      {def.price_fmt}
                    </span>
                  </span>
                  {slot && (
                    <span className="tabular-nums">
                      占用 {slot.used}/{slot.total}
                    </span>
                  )}
                </div>
                <button
                  type="button"
                  className={`px-3.5 py-2 text-[12px] font-semibold tracking-wider uppercase rounded-md border transition-[transform,box-shadow,background-color] duration-150 ${
                    disabled
                      ? "text-muted bg-panel-2 border-line/55 cursor-not-allowed"
                      : "text-bg bg-mint border-mint [box-shadow:0_8px_18px_oklch(82%_0.16_155/_0.18),inset_0_-2px_0_oklch(0%_0_0/_0.2)] hover:-translate-y-px hover:[box-shadow:0_12px_24px_oklch(82%_0.16_155/_0.32),inset_0_-2px_0_oklch(0%_0_0/_0.2)]"
                  } ${inFlight ? "animate-[buyPulse_0.9s_ease-out_infinite]" : ""}`}
                  disabled={disabled}
                  onClick={() => handleBuy(def.id)}
                >
                  {buttonLabel}
                </button>
              </footer>
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
      <span className="text-[9px] text-muted/80 uppercase tracking-wider font-mono">{label}</span>
      <span className="text-ink text-[12px] font-semibold tabular-nums leading-none font-mono">
        {value}
      </span>
    </div>
  );
}
