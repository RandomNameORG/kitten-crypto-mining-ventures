import { useEffect, useMemo, useState } from "react";
import { ActionBar, ActionButton } from "../components/ActionButton";
import { SlotMeter } from "../components/SlotMeter";
import type { ActionRequest, GPU, Snapshot } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  dispatch: (payload: ActionRequest) => void;
}

const SHIP_WINDOW_GUESS = 180;
const CAT_FACES = ["🐱", "🐈", "🐈‍⬛", "😺", "😻"];

export function GPUsPanel({ snapshot, dispatch }: Props) {
  const room = useMemo(
    () => snapshot.rooms.find((r) => r.id === snapshot.state.current_room) ?? null,
    [snapshot.rooms, snapshot.state.current_room],
  );
  const inRoom = useMemo(
    () => snapshot.gpus.filter((g) => g.room === snapshot.state.current_room),
    [snapshot.gpus, snapshot.state.current_room],
  );
  const shipping = inRoom.filter((g) => g.status === "shipping");
  const installed = inRoom.filter((g) => g.status !== "shipping");

  return (
    <>
      <SlotMeter snapshot={snapshot} room={room} />
      {shipping.length > 0 && (
        <section className="relative grid gap-2 mb-3 px-3 py-3 rounded-lg border border-blue/45 bg-[oklch(22%_0.06_240/_0.32)] [box-shadow:inset_0_1px_0_oklch(100%_0_0/_0.05),0_8px_24px_oklch(0%_0_0/_0.32)] overflow-hidden">
          <div
            aria-hidden
            className="pointer-events-none absolute left-3 right-3 top-0 h-0.5 [background:linear-gradient(90deg,transparent,oklch(76%_0.13_240/_0.6),transparent)] [background-size:60%_100%] [background-repeat:no-repeat] animate-[shipBeam_2.4s_linear_infinite]"
          />
          <header className="flex items-center justify-between gap-2">
            <span className="inline-flex items-center gap-2 text-blue text-[12px] font-semibold uppercase tracking-wider">
              <span className="text-base animate-[shipBob_1.4s_var(--ease-out)_infinite]">
                🐱📦
              </span>
              在途订单
              <span className="text-muted/80 text-[10px] font-mono normal-case tracking-normal">
                小猫快递服务中
              </span>
            </span>
            <span className="px-1.5 py-px text-[10px] font-bold text-bg bg-blue rounded-sm tabular-nums font-mono min-w-[18px] text-center">
              {shipping.length}
            </span>
          </header>
          <div className="grid gap-2">
            {shipping.map((gpu) => (
              <ShipCard key={gpu.instance_id} gpu={gpu} />
            ))}
          </div>
        </section>
      )}
      {installed.length === 0 && shipping.length === 0 ? (
        <div className="text-muted text-xs px-3 py-6 text-center">
          🐾 空槽位 · 去商店买一张
        </div>
      ) : installed.length === 0 ? null : (
        <div className="grid gap-2">
          {installed.map((gpu) => (
            <article
              key={gpu.instance_id}
              className={`grid grid-cols-[68px_minmax(0,1fr)] items-start gap-2.5 p-2.5 border rounded-md transition-[transform,border-color,box-shadow] duration-200 ${
                gpu.status === "broken"
                  ? "border-red/45 bg-[oklch(28%_0.14_25/_0.18)]"
                  : "border-line/40 bg-panel/50 hover:-translate-y-px hover:border-line/70 hover:[box-shadow:0_8px_20px_oklch(0%_0_0/_0.35)]"
              }`}
            >
              <div className="relative grid place-items-center">
                <img
                  className="w-[60px] h-[60px] object-contain border border-line/50 rounded-md [image-rendering:pixelated] [background:radial-gradient(circle_at_50%_34%,oklch(82%_0.16_155/_0.10),transparent_58%),oklch(8%_0.008_200)]"
                  src={gpuIconSrc(gpu.def_id || "scrap")}
                  alt=""
                  loading="lazy"
                />
                <span className="absolute -bottom-1 px-1 text-[9px] text-muted/80 bg-bg/80 border border-line/40 rounded-sm font-mono">
                  #{gpu.instance_id}
                </span>
              </div>
              <div className="min-w-0 grid gap-1.5">
                <div className="flex items-center justify-between gap-2">
                  <span className="text-ink text-[13px] font-semibold truncate">
                    {gpu.name}
                  </span>
                  <span
                    className={`flex-none px-1.5 py-0.5 text-[10px] font-mono uppercase tracking-wider rounded-sm border ${
                      gpu.status === "broken"
                        ? "text-red border-red/40 bg-red/10 animate-[pulseGlow_1.6s_var(--ease-out)_infinite]"
                        : "text-mint border-mint/40 bg-mint/10"
                    }`}
                  >
                    {gpu.status === "broken" ? "损坏" : "运行"}
                  </span>
                </div>
                <div className="flex flex-wrap gap-1 font-mono">
                  <Chip>L{gpu.upgrade}</Chip>
                  <Chip>OC {gpu.oc_level}</Chip>
                  <Chip className="text-mint border-mint/40 bg-mint/10 font-semibold">
                    {gpu.earn_fmt}
                  </Chip>
                  <Chip className="text-blue border-blue/35 bg-blue/8">
                    {gpu.hours_left.toFixed(1)}h
                  </Chip>
                </div>
                <ActionBar>
                  <ActionButton
                    label="升级"
                    icon="升"
                    intent="primary"
                    onClick={() => dispatch({ action: "upgrade_gpu", instance_id: gpu.instance_id })}
                  />
                  <ActionButton
                    label="超频"
                    icon="频"
                    intent="accent"
                    onClick={() => dispatch({ action: "cycle_oc", instance_id: gpu.instance_id })}
                  />
                  <ActionButton
                    label={gpu.repairable ? "维修" : "正常"}
                    icon="修"
                    intent="warn"
                    disabled={!gpu.repairable}
                    onClick={() => dispatch({ action: "repair_gpu", instance_id: gpu.instance_id })}
                  />
                  <ActionButton
                    label="拆解"
                    icon="拆"
                    intent="danger"
                    onClick={() => dispatch({ action: "scrap_gpu", instance_id: gpu.instance_id })}
                  />
                </ActionBar>
              </div>
            </article>
          ))}
        </div>
      )}
    </>
  );
}

function Chip({ children, className = "" }: { children: React.ReactNode; className?: string }) {
  return (
    <span
      className={`px-1.5 py-0.5 text-[10px] text-muted bg-bg/40 border border-line/40 rounded-sm ${className}`}
    >
      {children}
    </span>
  );
}

function ShipCard({ gpu }: { gpu: GPU }) {
  const initialEta = typeof gpu.ship_eta_sec === "number" ? gpu.ship_eta_sec : 0;
  const [eta, setEta] = useState(initialEta);
  useEffect(() => {
    setEta(initialEta);
  }, [initialEta, gpu.ships_at]);
  useEffect(() => {
    const id = window.setInterval(() => {
      setEta((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => window.clearInterval(id);
  }, [gpu.instance_id]);

  const progress = Math.max(0, Math.min(1, 1 - eta / SHIP_WINDOW_GUESS));
  const etaLabel = eta > 0 ? `${eta}s` : "🐱 即将抵达";
  const face = CAT_FACES[gpu.instance_id % CAT_FACES.length];

  return (
    <div className="relative grid grid-cols-[44px_minmax(0,1fr)] gap-2.5 px-2.5 pt-3 pb-2 rounded-md border border-blue/35 bg-bg/45 overflow-hidden">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 [background:linear-gradient(90deg,transparent_0%,oklch(76%_0.13_240/_0.08)_50%,transparent_100%)] [transform:translateX(-100%)] animate-[shipSweep_2.2s_linear_infinite]"
      />
      <div className="relative w-11 h-11 rounded-md border border-line/50 bg-bg/60 grid place-items-center overflow-hidden">
        <img
          className="w-full h-full object-contain [image-rendering:pixelated]"
          src={gpuIconSrc(gpu.def_id || "scrap")}
          alt=""
          loading="lazy"
        />
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 [background:linear-gradient(135deg,transparent_30%,oklch(76%_0.13_240/_0.18)_50%,transparent_70%)] [background-size:200%_200%] animate-[shipShimmer_2s_linear_infinite]"
        />
        <span
          aria-hidden
          className="absolute -top-0.5 -right-0.5 text-[10px] leading-none animate-[shipBob_1.3s_var(--ease-out)_infinite] [filter:drop-shadow(0_0_4px_oklch(76%_0.13_240/_0.6))]"
        >
          {face}
        </span>
      </div>
      <div className="min-w-0 grid gap-1 content-center">
        <div className="flex items-baseline justify-between gap-2">
          <span className="text-ink text-[12px] font-semibold truncate">
            {gpu.name}
          </span>
          <span className="text-blue text-[11px] font-bold tabular-nums font-mono">
            {etaLabel}
          </span>
        </div>
        <div
          aria-hidden
          className="relative h-3 rounded-full overflow-hidden bg-bg/70 border border-line/50"
        >
          <div
            className="absolute inset-y-0 left-0 rounded-full transition-[width] duration-700 [background:linear-gradient(90deg,oklch(76%_0.13_240/_0.7),oklch(82%_0.16_155/_0.85))] [box-shadow:0_0_8px_oklch(76%_0.13_240/_0.5)]"
            style={{ width: `${progress * 100}%` }}
          />
          <div className="absolute inset-0 [background:linear-gradient(90deg,transparent,oklch(100%_0_0/_0.18),transparent)] [background-size:40%_100%] [background-repeat:no-repeat] animate-[shipBeam_1.8s_linear_infinite]" />
          <span
            aria-hidden
            className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 text-[11px] leading-none transition-[left] duration-700 [filter:drop-shadow(0_0_4px_oklch(76%_0.13_240/_0.7))]"
            style={{ left: `${Math.max(4, progress * 100)}%` }}
          >
            {face}
          </span>
          <span
            aria-hidden
            className="absolute inset-y-0 right-1 grid place-items-center text-[10px] opacity-70"
          >
            🏠
          </span>
        </div>
        <div className="flex justify-between text-[10px] text-muted/80 uppercase tracking-wider tabular-nums font-mono">
          <span className="inline-flex items-center gap-1">
            <span className="opacity-60">🐾🐾🐾</span>
            #{gpu.instance_id}
          </span>
          <span>运输中</span>
        </div>
      </div>
    </div>
  );
}
