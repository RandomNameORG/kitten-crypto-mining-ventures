import { ActionBar, ActionButton } from "../components/ActionButton";
import { PanelSummary } from "../components/PanelSummary";
import type { ActionRequest, Snapshot } from "../types";

interface Props {
  snapshot: Snapshot;
  currentRoomName: string;
  dispatch: (payload: ActionRequest) => void;
}

export function RoomsPanel({ snapshot, currentRoomName, dispatch }: Props) {
  return (
    <>
      <h2>房间</h2>
      <PanelSummary
        items={[
          ["当前", currentRoomName],
          ["余额", snapshot.state.btc_fmt],
        ]}
      />
      <div className="list">
        {snapshot.rooms.map((room) => {
          const canAfford = snapshot.state.btc >= room.unlock_cost;
          return (
            <article key={room.id} className="row">
              <div className="row-head">
                <span className="row-title">{room.name}</span>
                <span className="tag">
                  {room.unlocked ? `${room.gpu_count}/${room.slots}` : room.unlock_cost_fmt}
                </span>
              </div>
              <div className="copy">{room.flavor}</div>
              <div className="facts">
                <span className="fact">net {room.net_fmt}</span>
                <span className="fact">heat {room.heat ? room.heat.toFixed(0) : 0}°</span>
                <span className="fact">tick {room.heat_tick_in || 0}s</span>
              </div>
              <ActionBar>
                {room.unlocked ? (
                  <ActionButton
                    label={room.current ? "已在此处" : "进入"}
                    icon="入"
                    intent="primary"
                    disabled={room.current}
                    onClick={() => dispatch({ action: "switch_room", id: room.id })}
                  />
                ) : (
                  <ActionButton
                    label={canAfford ? "解锁" : "余额不足"}
                    icon="解"
                    intent="primary"
                    disabled={!canAfford}
                    onClick={() => dispatch({ action: "unlock_room", id: room.id })}
                  />
                )}
              </ActionBar>
            </article>
          );
        })}
      </div>
    </>
  );
}
