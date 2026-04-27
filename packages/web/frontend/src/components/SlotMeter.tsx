import { computeSlotStats } from "../lib/slotStats";
import type { Room, Snapshot } from "../types";

interface Props {
  snapshot: Snapshot;
  room: Room | null;
}

export function SlotMeter({ snapshot, room }: Props) {
  const stats = computeSlotStats(snapshot, room);
  if (!room || !stats) return null;
  const { used, total, free, running, shipping, broken, tone, cells } = stats;

  // No idle animation: tone changes border color only when state changes.
  const wrap =
    tone === "danger"
      ? "border-red/65 bg-[oklch(28%_0.14_25/_0.22)]"
      : tone === "warn"
        ? "border-orange/55 bg-[oklch(28%_0.10_40/_0.18)]"
        : "border-line/55 bg-panel-2/40";
  const countTone =
    tone === "danger" ? "text-red" : tone === "warn" ? "text-gold" : "text-mint";

  return (
    <div
      className={`relative grid gap-2 px-3 py-2.5 mb-3 rounded-lg border [box-shadow:inset_0_1px_0_oklch(100%_0_0_/_0.04)] ${wrap}`}
    >
      <div className="flex items-baseline gap-2 leading-none">
        <span className="text-muted text-[10px] uppercase tracking-[0.18em] font-semibold">
          机位
        </span>
        <span className="flex-1 truncate text-blue text-[12px] font-semibold">{room.name}</span>
        <span className="font-mono text-lg font-bold tabular-nums leading-none">
          <strong className={`${countTone} text-[22px] leading-none`}>{used}</strong>
          <span className="mx-px text-muted/70 text-base font-normal">/</span>
          <span className="text-muted/80">{total}</span>
        </span>
      </div>
      <div className="flex flex-wrap gap-1" aria-hidden>
        {Array.from({ length: total }).map((_, i) => {
          const gpu = cells[i];
          const cls = !gpu
            ? "bg-[oklch(8%_0.008_200/_0.6)] border-line/40"
            : gpu.status === "shipping"
              ? "border-blue/50 bg-blue/35"
              : gpu.status === "broken"
                ? "border-red/50 bg-[oklch(70%_0.20_22/_0.55)]"
                : "border-mint/60 bg-[linear-gradient(180deg,oklch(82%_0.16_155/_0.85),oklch(70%_0.16_155/_0.7))] [box-shadow:0_0_6px_oklch(82%_0.16_155/_0.5)]";
          return (
            <span
              key={i}
              className={`h-2 flex-1 min-w-[14px] rounded-sm border transition-colors duration-200 ${cls}`}
            />
          );
        })}
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-1 font-mono text-[10px] uppercase tracking-wider tabular-nums text-muted/90">
        <span className="inline-flex items-center gap-1 text-mint">
          <span className="w-1.5 h-1.5 rounded-full bg-mint" />
          运行 {running}
        </span>
        <span className="inline-flex items-center gap-1 text-blue">
          <span className="w-1.5 h-1.5 rounded-full bg-blue" />
          运输 {shipping}
        </span>
        <span className="inline-flex items-center gap-1 text-red">
          <span className="w-1.5 h-1.5 rounded-full bg-red" />
          损坏 {broken}
        </span>
        <span className="inline-flex items-center gap-1 text-muted">
          <span className="w-1.5 h-1.5 rounded-full bg-[oklch(70%_0.018_175/_0.5)]" />
          空位 {free}
        </span>
      </div>
    </div>
  );
}
