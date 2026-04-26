import type { Snapshot } from "../types";

interface Props {
  snapshot: Snapshot;
}

export function StatsPanel({ snapshot }: Props) {
  const room = snapshot.rooms.find((r) => r.id === snapshot.state.current_room);
  const allGPUs = snapshot.gpus.length;
  const broken = snapshot.gpus.filter((g) => g.status === "broken").length;
  return (
    <>
      <div className="list">
        <article className="row">
          <div className="row-head">
            <span className="row-title">资产</span>
            <span className="tag">{allGPUs}</span>
          </div>
          <div className="facts">
            <span className="fact">broken {broken}</span>
            <span className="fact">TP {snapshot.state.tech_point}</span>
            <span className="fact">frags {snapshot.state.research_frags}</span>
          </div>
        </article>
        {room && (
          <article className="row">
            <div className="row-head">
              <span className="row-title">{room.name}</span>
              <span className="tag">{room.net_fmt}</span>
            </div>
            <div className="facts">
              <span className="fact">earn {room.earn_fmt}</span>
              <span className="fact">bill {room.bill_fmt}</span>
              <span className="fact">heat {room.heat_delta.toFixed(1)}</span>
            </div>
          </article>
        )}
        <article className="row">
          <div className="row-head">
            <span className="row-title">声望</span>
            <span className="tag">{snapshot.state.reputation}</span>
          </div>
          <div className="facts">
            <span className="fact">karma {snapshot.state.karma}</span>
            <span className="fact">life {snapshot.state.lifetime_earned_fmt}</span>
          </div>
        </article>
      </div>
    </>
  );
}
