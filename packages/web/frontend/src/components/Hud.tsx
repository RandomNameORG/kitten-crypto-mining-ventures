import type { GameState, Room } from "../types";

interface Props {
  state: GameState;
  room: Room | null;
}

export function Hud({ state, room }: Props) {
  const heat = room ? `${room.heat.toFixed(0)}° / ${room.max_heat.toFixed(0)}°` : "-";
  const trend = state.market_trend > 0 ? "↑" : state.market_trend < 0 ? "↓" : "→";
  const hot = !!room && room.heat_pct > 0.78;
  return (
    <section className="hud">
      <Metric label="BTC" value={state.btc_fmt} />
      <Metric label="净收益" value={state.room_net_fmt} />
      <Metric label="产出" value={state.room_earn_fmt} />
      <Metric label="电费" value={state.room_bill_fmt} />
      <Metric label="温度" value={heat} hot={hot} />
      <Metric label="市场" value={`${state.market_price.toFixed(2)} ${trend}`} />
    </section>
  );
}

function Metric({ label, value, hot = false }: { label: string; value: string; hot?: boolean }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong className={hot ? "hot" : ""}>{value}</strong>
    </div>
  );
}
