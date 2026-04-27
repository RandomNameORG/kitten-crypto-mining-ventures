import type { GameEvent } from "../types";

interface Props {
  event: GameEvent | undefined;
}

export function EventBanner({ event }: Props) {
  if (!event) return null;
  return (
    <div className={`event-banner ${event.category}`}>
      <span className="event-banner-tag">{event.category}</span>
      <span className="event-banner-name">{event.name}</span>
      <span className="event-banner-text">{event.text}</span>
    </div>
  );
}
