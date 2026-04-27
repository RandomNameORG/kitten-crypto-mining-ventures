import type { Room } from "../types";

interface Props {
  room: Room | null;
}

export function StageFoot({ room }: Props) {
  if (!room) return <div className="stage-foot" />;
  const slots = room.slots ? room.gpu_count / room.slots : 0;
  const d = room.defense || ({} as Room["defense"]);
  const shield = ((d.lock || 0) + (d.cctv || 0) + (d.armor || 0)) / 24;
  const cooling = ((d.cooling || 0) + (d.wiring || 0)) / 16;
  return (
    <div className="stage-foot">
      <Bar label="槽位" value={`${room.gpu_count}/${room.slots}`} amount={slots} />
      <Bar label="温度" value={`${room.heat.toFixed(0)}°`} amount={room.heat_pct} variant="heat" />
      <Bar label="安防" value={`${Math.round(shield * 100)}%`} amount={shield} />
      <Bar label="维护" value={`${Math.round(cooling * 100)}%`} amount={cooling} />
    </div>
  );
}

function Bar({
  label,
  value,
  amount,
  variant,
}: {
  label: string;
  value: string;
  amount: number;
  variant?: "heat";
}) {
  const pct = `${Math.max(0, Math.min(100, amount * 100)).toFixed(0)}%`;
  return (
    <div className="bar">
      <div className="bar-label">
        <span>{label}</span>
        <span>{value}</span>
      </div>
      <div className="bar-track">
        <div className={`bar-fill ${variant ?? ""}`} style={{ width: pct }} />
      </div>
    </div>
  );
}
