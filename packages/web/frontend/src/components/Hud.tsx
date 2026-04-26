import type { GameState, Room } from "../types";

interface Props {
  state: GameState;
  room: Room | null;
}

export function Hud({ state, room }: Props) {
  const heat = room ? `${room.heat.toFixed(0)}° / ${room.max_heat.toFixed(0)}°` : "-";
  const trend = state.market_trend > 0 ? "↑" : state.market_trend < 0 ? "↓" : "→";
  const hot = !!room && room.heat_pct > 0.78;
  const netCls = state.room_net_fmt.trim().startsWith("-") ? "loss" : "gain";
  return (
    <section className="hud">
      <Metric label="BTC" value={state.btc_fmt} primary />
      <Metric label="净" value={state.room_net_fmt} valueCls={netCls} primary />
      <Metric label="温度" value={heat} valueCls={hot ? "hot" : ""} primary />
      <Metric label="产出" value={state.room_earn_fmt} />
      <Metric label="电费" value={state.room_bill_fmt} />
      <Metric label={`市场 ${trend}`} value={state.market_price.toFixed(2)} />
    </section>
  );
}

interface MetricProps {
  label: string;
  value: string;
  valueCls?: string;
  primary?: boolean;
}

function Metric({ label, value, valueCls = "", primary = false }: MetricProps) {
  return (
    <div className={`metric ${primary ? "primary" : ""}`}>
      <span>{label}</span>
      <strong className={valueCls}>{value}</strong>
    </div>
  );
}
