import { ActionBar, ActionButton } from "../components/ActionButton";
import { PanelSummary } from "../components/PanelSummary";
import type { Snapshot } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  currentRoomName: string;
  dispatch: (id: string) => void;
}

export function StorePanel({ snapshot, currentRoomName, dispatch }: Props) {
  const room = snapshot.rooms.find((r) => r.id === snapshot.state.current_room);
  return (
    <>
      <h2>显卡商店</h2>
      <PanelSummary
        items={[
          ["余额", snapshot.state.btc_fmt],
          ["当前房间", currentRoomName],
          ["槽位", room ? `${room.gpu_count}/${room.slots}` : "-"],
        ]}
      />
      <div className="list">
        {snapshot.gpu_defs.map((def) => {
          const canBuy = snapshot.state.btc >= def.price;
          return (
            <article key={def.id} className="row item-row">
              <img className="item-icon gpu-icon" src={gpuIconSrc(def.id)} alt="" loading="lazy" />
              <div className="item-content">
                <div className="row-head">
                  <span className="row-title">{def.name}</span>
                  <span className="tag">{def.tier}</span>
                </div>
                <div className="copy">{def.flavor}</div>
                <div className="facts">
                  <span className="fact price">{def.price_fmt}</span>
                  <span className="fact">eff {def.efficiency.toFixed(4)}</span>
                  <span className="fact">heat {def.heat_output.toFixed(2)}</span>
                </div>
                <ActionBar>
                  <ActionButton
                    label={canBuy ? "购买" : "余额不足"}
                    icon="买"
                    intent="primary"
                    disabled={!canBuy}
                    onClick={() => dispatch(def.id)}
                  />
                </ActionBar>
              </div>
            </article>
          );
        })}
      </div>
    </>
  );
}
