// useNow — single 1Hz clock for everyone who needs "wall-clock seconds now."
// Hoisted out of per-card <ShipCard> setIntervals so N shipping cards trigger
// 1 timer + 1 re-render, not N timers + N re-renders.
//
// Returns Date.now()/1000 floored to the second so consumers see consistent
// integer ETA values across one paint frame.

import { useEffect, useState } from "react";

export function useNow(): number {
  const [now, setNow] = useState(() => Math.floor(Date.now() / 1000));
  useEffect(() => {
    const id = window.setInterval(() => {
      setNow(Math.floor(Date.now() / 1000));
    }, 1000);
    return () => window.clearInterval(id);
  }, []);
  return now;
}
