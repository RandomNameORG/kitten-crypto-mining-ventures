import { useMemo } from "react";
import { ActionBar, ActionButton } from "../components/ActionButton";
import { SlotMeter } from "../components/SlotMeter";
import type { ActionRequest, Snapshot } from "../types";
import { gpuIconSrc } from "../util";
import { ShipStrip } from "./_shipStrip";

interface Props {
  snapshot: Snapshot;
  dispatch: (payload: ActionRequest) => void;
}

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
      <ShipStrip ships={shipping} />
      {installed.length === 0 && shipping.length === 0 ? (
        <div className="text-muted text-xs px-3 py-6 text-center">
          空槽位 · 去商店买一张
        </div>
      ) : installed.length === 0 ? null : (
        <div className="grid gap-2">
          {installed.map((gpu) => (
            <article
              key={gpu.instance_id}
              className={`grid grid-cols-[68px_minmax(0,1fr)] items-start gap-2.5 p-2.5 border rounded-md transition-colors duration-200 ${
                gpu.status === "broken"
                  ? "border-red/45 bg-[oklch(28%_0.14_25/_0.18)]"
                  : "border-line/40 bg-panel/50 hover:border-line/70"
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
                  <span className="text-ink text-[13px] font-semibold truncate">{gpu.name}</span>
                  <span
                    className={`flex-none px-1.5 py-0.5 text-[10px] font-mono uppercase tracking-wider rounded-sm border ${
                      gpu.status === "broken"
                        ? "text-red border-red/40 bg-red/10"
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
