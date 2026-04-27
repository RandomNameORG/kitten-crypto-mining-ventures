import type { GameState, Room } from "../types";

interface Props {
  state: GameState;
  room: Room | null;
}

export function Hud({ state, room }: Props) {
  const heat = room ? `${room.heat.toFixed(0)}°` : "-";
  const heatMax = room ? `/ ${room.max_heat.toFixed(0)}°` : "";
  const trend = state.market_trend > 0 ? "↑" : state.market_trend < 0 ? "↓" : "→";
  const trendCls = state.market_trend > 0 ? "gain" : state.market_trend < 0 ? "loss" : "";
  const hot = !!room && room.heat_pct > 0.78;
  const netLoss = state.room_net_fmt.trim().startsWith("-");

  return (
    <section className="hud">
      <Card label="BTC" value={state.btc_fmt} delta={`累计 ${state.lifetime_earned_fmt}`} variant="primary" />
      <Card
        label="净收益 / 秒"
        value={state.room_net_fmt}
        delta={`${state.room_earn_fmt} − ${state.room_bill_fmt}`}
        variant={netLoss ? "loss" : "gain"}
      />
      <Card
        label="温度"
        value={`${heat} ${heatMax}`}
        delta={room ? `下次 +Δ ${room.heat_tick_in}s` : ""}
        variant="heat"
        valueCls={hot ? "hot" : ""}
      />
      <Card
        label="市场"
        value={`${state.market_price.toFixed(2)} ${trend}`}
        delta={`产出 ${state.room_earn_fmt}`}
        variant="gold"
        valueCls={trendCls}
      />
      <Card
        label="电费 / 秒"
        value={state.room_bill_fmt}
        delta={room ? `${room.gpu_count}/${room.slots} 槽位` : ""}
      />
      <Card
        label="声望 · 业力"
        value={`${state.reputation} · ${state.karma}`}
        delta={state.syndicate_joined ? "Syndicate ✓" : "未加入工会"}
      />
    </section>
  );
}

interface CardProps {
  label: string;
  value: string;
  delta?: string;
  variant?: "primary" | "gain" | "loss" | "heat" | "gold";
  valueCls?: string;
}

function Card({ label, value, delta, variant, valueCls = "" }: CardProps) {
  return (
    <div className={`metric ${variant ?? ""}`}>
      <span className="label">{label}</span>
      <strong className={`value ${valueCls}`}>{value}</strong>
      {delta ? <span className={`delta ${variant === "gain" ? "gain" : variant === "loss" ? "loss" : ""}`}>{delta}</span> : null}
    </div>
  );
}
