import { useMemo, useState } from "react";
import { useNow } from "../lib/useNow";
import type { GPU } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  ships: GPU[];
}

const CAP = 8;

export function ShipStrip({ ships }: Props) {
  const [open, setOpen] = useState(false);
  const now = useNow();

  const enriched = useMemo(() => {
    return ships
      .map((g) => {
        const total = g.ship_total_sec ?? 0;
        // Trust `ships_at` (server-authored absolute deadline) so the local
        // clock keeps the ETA decrementing during a poll stall. Without
        // ship_total_sec we fall back to "always full" — old saves will
        // see the bar pinned at 100% which is honest, not wrong.
        const eta = g.ships_at ? Math.max(0, g.ships_at - now) : 0;
        const progress = total > 0 ? Math.max(0, Math.min(1, 1 - eta / total)) : 1;
        return { gpu: g, eta, total, progress };
      })
      .sort((a, b) => a.eta - b.eta);
  }, [ships, now]);

  if (enriched.length === 0) return null;
  const nearestEta = enriched[0].eta;
  const visible = open ? enriched.slice(0, CAP) : [];
  const overflow = enriched.length - CAP;

  return (
    <section
      className="mb-3 rounded-lg border border-blue/45 bg-[oklch(22%_0.06_240/_0.32)] overflow-hidden"
      data-component="ship-strip"
    >
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="w-full flex items-center justify-between gap-3 px-3 h-8 text-left text-blue text-[12px] font-semibold tracking-wider uppercase hover:bg-blue/8 transition-[background-color] duration-150"
        aria-expanded={open}
      >
        <span className="inline-flex items-center gap-2">
          <span className="text-base leading-none [font-variant-emoji:emoji]">🐈</span>
          <span>在途 {enriched.length}</span>
          <span className="text-muted/80 normal-case tracking-normal text-[11px] font-mono">
            最近 ETA {nearestEta > 0 ? `${nearestEta}s` : "即将抵达"}
          </span>
        </span>
        <span
          aria-hidden
          className={`text-muted/80 text-[10px] transition-transform duration-200 ${open ? "rotate-180" : ""}`}
        >
          ▾
        </span>
      </button>
      {open && (
        <div className="grid gap-2 px-3 py-2 border-t border-blue/30">
          {visible.map(({ gpu, eta, progress }) => (
            <ShipRow key={gpu.instance_id} gpu={gpu} eta={eta} progress={progress} />
          ))}
          {overflow > 0 && (
            <div className="text-[10px] text-muted/80 font-mono uppercase tracking-wider text-center">
              +{overflow} more
            </div>
          )}
        </div>
      )}
    </section>
  );
}

function ShipRow({ gpu, eta, progress }: { gpu: GPU; eta: number; progress: number }) {
  const etaLabel = eta > 0 ? `${eta}s` : "即将抵达";
  return (
    <div className="grid grid-cols-[36px_minmax(0,1fr)] gap-2.5 items-center">
      <div className="w-9 h-9 rounded-md border border-line/50 bg-bg/60 grid place-items-center overflow-hidden">
        <img
          className="w-full h-full object-contain [image-rendering:pixelated]"
          src={gpuIconSrc(gpu.def_id || "scrap")}
          alt=""
          loading="lazy"
        />
      </div>
      <div className="grid gap-1 min-w-0">
        <div className="flex items-baseline justify-between gap-2">
          <span className="text-ink text-[12px] font-semibold truncate">{gpu.name}</span>
          <span className="text-blue text-[11px] font-bold tabular-nums font-mono">
            {etaLabel}
          </span>
        </div>
        <ShipBar progress={progress} />
      </div>
    </div>
  );
}

function ShipBar({ progress }: { progress: number }) {
  // GPU-only path: transform: translateX, not transition: left.
  // The cat sits on the leading edge of the fill and rides it across.
  const pct = Math.max(0, Math.min(1, progress)) * 100;
  return (
    <div className="relative h-3 rounded-full overflow-hidden bg-bg/70 border border-line/50">
      <div
        className="absolute inset-y-0 left-0 right-0 origin-left rounded-full [background:linear-gradient(90deg,oklch(76%_0.13_240/_0.7),oklch(82%_0.16_155/_0.85))] [box-shadow:0_0_8px_oklch(76%_0.13_240/_0.5)] [transition:transform_700ms_cubic-bezier(0.32,0.72,0,1)] [will-change:transform]"
        style={{ transform: `scaleX(${pct / 100})` }}
      />
      <span
        aria-hidden
        className="absolute top-1/2 left-0 -translate-y-1/2 -translate-x-1/2 text-[12px] leading-none [transition:transform_700ms_cubic-bezier(0.32,0.72,0,1)] [will-change:transform] [filter:drop-shadow(0_0_4px_oklch(76%_0.13_240/_0.7))] [font-variant-emoji:emoji] cat-marker"
        style={{ transform: `translate(${pct}%, -50%)` }}
      >
        🐈
      </span>
    </div>
  );
}
