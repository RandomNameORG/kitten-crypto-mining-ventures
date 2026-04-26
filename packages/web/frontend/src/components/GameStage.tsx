import { useEffect, useRef } from "react";
import sceneMap from "../scene-map.json";
import type { Snapshot } from "../types";

const STAGE = { width: 512, height: 288 };
const DIRECTIONS = ["down", "left", "right", "up"] as const;
const FRAME_COUNT = 4;
const ARRIVAL_DISTANCE = 6;
const CHARACTER_ROOT = "/assets/2d/spritesheet/characters";

type Direction = (typeof DIRECTIONS)[number];
type Mode = "idle" | "walk";
type ActorState =
  | "patrol"
  | "rest"
  | "inspect"
  | "called"
  | "alarm"
  | "happy"
  | "confused"
  | "hot"
  | "critical"
  | "reboot";

interface Sprite {
  image: HTMLImageElement | HTMLCanvasElement;
  bounds: { x: number; y: number; width: number; height: number };
}

interface ScenePoint {
  x: number;
  y: number;
}

interface Actor {
  x: number;
  y: number;
  targetX: number;
  targetY: number;
  direction: Direction;
  frame: number;
  frameClock: number;
  moving: boolean;
  scale: number;
  targetBodyHeight: number;
  state: ActorState;
  stateClock: number;
  decisionClock: number;
  reactionClock: number;
  emote: string;
  emoteClock: number;
  bobClock: number;
  coolingIndex: number;
}

interface Pointer {
  active: boolean;
  x: number;
  y: number;
}

const EVENT_REACTIONS: Record<
  string,
  { state: ActorState; point: string; emote: string; seconds: number }
> = {
  petty_thief: { state: "alarm", point: "door", emote: "!", seconds: 4.2 },
  admin_discovery: { state: "alarm", point: "door", emote: "!", seconds: 4.6 },
  tunnel_heist: { state: "alarm", point: "door", emote: "!!", seconds: 5.0 },
  police_visit: { state: "alarm", point: "door", emote: "!!", seconds: 5.0 },
  government_raid: { state: "alarm", point: "door", emote: "!!", seconds: 5.4 },
  tax_audit: { state: "confused", point: "door", emote: "?", seconds: 4.0 },
  power_outage: { state: "reboot", point: "center", emote: "...", seconds: 3.8 },
  gas_leak: { state: "alarm", point: "door", emote: "!", seconds: 4.2 },
  voltage_dip: { state: "inspect", point: "rig", emote: "⚡", seconds: 3.8 },
  power_surge: { state: "alarm", point: "rig", emote: "⚡!", seconds: 4.4 },
  rain_leak: { state: "alarm", point: "rig", emote: "💧", seconds: 4.0 },
  tech_share: { state: "happy", point: "rest", emote: "💡", seconds: 3.6 },
  extra_delivery: { state: "happy", point: "door", emote: "📦", seconds: 3.8 },
  btc_pump: { state: "happy", point: "rig", emote: "₿", seconds: 4.0 },
  lucky_fish: { state: "happy", point: "rest", emote: "♪", seconds: 4.0 },
  street_dog: { state: "confused", point: "door", emote: "?", seconds: 3.6 },
  group_chat_sos: { state: "confused", point: "rest", emote: "💬", seconds: 3.6 },
};

interface Props {
  snapshot: Snapshot | null;
}

export function GameStage({ snapshot }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const snapshotRef = useRef<Snapshot | null>(snapshot);
  snapshotRef.current = snapshot;

  useEffect(() => {
    const canvasEl = canvasRef.current;
    if (!canvasEl) return;
    const ctxEl = canvasEl.getContext("2d");
    if (!ctxEl) return;
    const canvas: HTMLCanvasElement = canvasEl;
    const ctx: CanvasRenderingContext2D = ctxEl;

    const roomImages = new Map<string, HTMLImageElement>();
    const sprites = new Map<string, Sprite>();
    const actor: Actor = {
      x: sceneMap.actor.spawn.x,
      y: sceneMap.actor.spawn.y,
      targetX: sceneMap.actor.spawn.x,
      targetY: sceneMap.actor.spawn.y,
      direction: "right",
      frame: 0,
      frameClock: 0,
      moving: false,
      scale: 0.72,
      targetBodyHeight: 82,
      state: "patrol",
      stateClock: 0,
      decisionClock: 0,
      reactionClock: 0,
      emote: "",
      emoteClock: 0,
      bobClock: 0,
      coolingIndex: 0,
    };
    const pointer: Pointer = { active: false, x: actor.x, y: actor.y };
    let lastEventKey = "";
    let raf = 0;
    let cancelled = false;
    let previousTime = performance.now();

    function currentRoom() {
      const snap = snapshotRef.current;
      return snap?.rooms.find((r) => r.id === snap.state.current_room) ?? null;
    }

    function currentWalkPolygon(): Array<[number, number]> {
      const room = currentRoom();
      const map = sceneMap.room_walk_polygons as Record<string, number[][]>;
      const raw = (room && map[room.id]) || (sceneMap.actor.walk_polygon as number[][]);
      return raw.map((p) => [p[0], p[1]]);
    }

    function catPoint(name: string): ScenePoint {
      const room = currentRoom();
      const rooms = sceneMap.cat_points.rooms as Record<string, Record<string, number[]>>;
      const defaults = sceneMap.cat_points.default as Record<string, number[]>;
      const point = (room && rooms[room.id]?.[name]) || defaults[name];
      if (point) return { x: point[0], y: point[1] };
      return { x: sceneMap.actor.spawn.x, y: sceneMap.actor.spawn.y };
    }

    function clamp(value: number, min: number, max: number) {
      return Math.min(max, Math.max(min, value));
    }

    function polygonBounds(points: Array<[number, number]>) {
      let minX = Infinity;
      let maxX = -Infinity;
      let minY = Infinity;
      let maxY = -Infinity;
      for (const [x, y] of points) {
        if (x < minX) minX = x;
        if (x > maxX) maxX = x;
        if (y < minY) minY = y;
        if (y > maxY) maxY = y;
      }
      return { minX, maxX, minY, maxY };
    }

    function verticalRangeAtX(points: Array<[number, number]>, x: number) {
      const ys: number[] = [];
      for (let i = 0; i < points.length; i += 1) {
        const [x1, y1] = points[i];
        const [x2, y2] = points[(i + 1) % points.length];
        const lo = Math.min(x1, x2);
        const hi = Math.max(x1, x2);
        if (x < lo || x > hi) continue;
        if (x1 === x2) {
          ys.push(y1, y2);
          continue;
        }
        const t = (x - x1) / (x2 - x1);
        if (t >= 0 && t <= 1) ys.push(y1 + (y2 - y1) * t);
      }
      if (ys.length < 2) {
        const b = polygonBounds(points);
        return { min: b.minY, max: b.maxY };
      }
      ys.sort((a, b) => a - b);
      return { min: ys[0], max: ys[ys.length - 1] };
    }

    function clampPointToGround(x: number, y: number): ScenePoint {
      const points = currentWalkPolygon();
      const b = polygonBounds(points);
      const groundX = clamp(x, b.minX, b.maxX);
      const range = verticalRangeAtX(points, groundX);
      return { x: groundX, y: clamp(y, range.min, range.max) };
    }

    function randomGroundPoint(): ScenePoint {
      const points = currentWalkPolygon();
      const b = polygonBounds(points);
      for (let attempt = 0; attempt < 12; attempt += 1) {
        const x = b.minX + Math.random() * (b.maxX - b.minX);
        const range = verticalRangeAtX(points, x);
        if (range.max - range.min < 8) continue;
        return {
          x,
          y: range.min + 6 + Math.random() * Math.max(1, range.max - range.min - 12),
        };
      }
      return catPoint("center");
    }

    function setActorTarget(point: ScenePoint, state: ActorState = actor.state, emote = "") {
      const grounded = clampPointToGround(point.x, point.y);
      actor.targetX = grounded.x;
      actor.targetY = grounded.y;
      actor.state = state;
      if (emote) showEmote(emote, 1.8);
    }

    function showEmote(emote: string, seconds = 1.8) {
      actor.emote = emote;
      actor.emoteClock = Math.max(actor.emoteClock, seconds);
    }

    function currentCatMood(roomHeatPct: number, miningPaused: boolean, paused: boolean): ActorState {
      if (paused || miningPaused) return "reboot";
      if (roomHeatPct >= 0.95) return "critical";
      if (roomHeatPct >= 0.8) return "hot";
      return "patrol";
    }

    function distanceToTarget() {
      return Math.hypot(actor.targetX - actor.x, actor.targetY - actor.y);
    }

    function coolingPoint(fan: ScenePoint, mood: ActorState): ScenePoint {
      const offsets =
        mood === "critical"
          ? [[-20, 0], [16, 4], [-4, 10]]
          : [[-16, 0], [16, 3], [0, 9]];
      const [dx, dy] = offsets[actor.coolingIndex % offsets.length];
      actor.coolingIndex += 1;
      return { x: fan.x + dx, y: fan.y + dy };
    }

    function actorSpeed() {
      switch (actor.state) {
        case "alarm":
          return 118;
        case "critical":
          return 74;
        case "hot":
          return 58;
        case "happy":
          return 82;
        case "called":
          return 76;
        case "inspect":
          return 52;
        case "rest":
          return distanceToTarget() > 5 ? 42 : 0;
        case "reboot":
          return 0;
        default:
          return 46;
      }
    }

    function handleCatAI(dt: number) {
      const snap = snapshotRef.current;
      if (!snap) return;
      const room = currentRoom();
      if (!room) return;

      actor.stateClock += dt;
      actor.decisionClock -= dt;
      actor.reactionClock = Math.max(0, actor.reactionClock - dt);
      actor.emoteClock = Math.max(0, actor.emoteClock - dt);

      const event = snap.last_event;
      const eventKey = event ? `${event.seq || 0}:${event.id}` : "";
      if (event && eventKey !== lastEventKey) {
        lastEventKey = eventKey;
        const reaction =
          EVENT_REACTIONS[event.id] ||
          (event.category === "crisis"
            ? { state: "alarm" as ActorState, point: "door", emote: "!!", seconds: 5.0 }
            : event.category === "threat"
              ? { state: "alarm" as ActorState, point: "rig", emote: "!", seconds: 4.0 }
              : event.category === "opportunity"
                ? { state: "happy" as ActorState, point: "rig", emote: "♪", seconds: 3.6 }
                : event.category === "social"
                  ? { state: "confused" as ActorState, point: "door", emote: "?", seconds: 3.4 }
                  : { state: "inspect" as ActorState, point: "center", emote: "...", seconds: 2.8 });
        actor.reactionClock = reaction.seconds;
        actor.decisionClock = reaction.seconds;
        setActorTarget(catPoint(reaction.point), reaction.state, reaction.emote);
        return;
      }

      const mood = currentCatMood(room.heat_pct, snap.state.mining_paused, snap.state.paused);
      if (mood === "reboot") {
        actor.state = "reboot";
        actor.targetX = actor.x;
        actor.targetY = actor.y;
        actor.moving = false;
        if (actor.emoteClock <= 0) showEmote("...", 1.5);
        return;
      }
      if ((mood === "hot" || mood === "critical") && actor.reactionClock <= 0) {
        const fan = catPoint("fan");
        const nearFan = Math.hypot(actor.x - fan.x, actor.y - fan.y) < 26;
        const arrived = distanceToTarget() <= ARRIVAL_DISTANCE;
        if (actor.state !== mood || (!nearFan && actor.decisionClock <= 0)) {
          setActorTarget(coolingPoint(fan, mood), mood, mood === "critical" ? "!!" : "热");
          actor.decisionClock = mood === "critical" ? 2.8 : 3.8;
        } else if (arrived && actor.decisionClock <= 0) {
          const next = coolingPoint(fan, mood);
          if (Math.hypot(next.x - actor.x, next.y - actor.y) > 12) {
            setActorTarget(next, mood, mood === "critical" ? "!!" : "热");
          } else {
            actor.direction = "down";
            if (actor.emoteClock <= 0) {
              showEmote(mood === "critical" ? "!!" : "热", mood === "critical" ? 1.3 : 1.7);
            }
          }
          actor.decisionClock = mood === "critical" ? 2.6 : 3.6;
        } else if (arrived) {
          actor.targetX = actor.x;
          actor.targetY = actor.y;
          actor.direction = "down";
          if (actor.emoteClock <= 0) {
            showEmote(mood === "critical" ? "!!" : "热", mood === "critical" ? 1.3 : 1.7);
          }
        }
        return;
      }
      if (actor.reactionClock > 0) return;
      if (actor.decisionClock > 0) return;

      const gpusInRoom = (snap.gpus || []).filter((g) => g.room === snap.state.current_room);
      const roll = Math.random();
      if (gpusInRoom.length && roll < 0.38) {
        const rig = catPoint("rig");
        setActorTarget(
          { x: rig.x + (Math.random() - 0.5) * 42, y: rig.y },
          "inspect",
          "...",
        );
        actor.decisionClock = 2.2 + Math.random() * 2.0;
      } else if (roll < 0.56) {
        setActorTarget(catPoint("rest"), "rest");
        actor.decisionClock = 1.4 + Math.random() * 2.4;
      } else {
        setActorTarget(randomGroundPoint(), "patrol");
        actor.decisionClock = 1.8 + Math.random() * 3.0;
      }
    }

    function movementVector() {
      const toX = actor.targetX - actor.x;
      const toY = actor.targetY - actor.y;
      const distance = Math.hypot(toX, toY);
      if (distance <= ARRIVAL_DISTANCE) {
        actor.targetX = actor.x;
        actor.targetY = actor.y;
        pointer.active = false;
        return { dx: 0, dy: 0 };
      }
      return { dx: toX / distance, dy: toY / distance };
    }

    function setDirection(dx: number, dy: number) {
      if (Math.abs(dx) > Math.abs(dy)) actor.direction = dx < 0 ? "left" : "right";
      else if (Math.abs(dy) > 0.01) actor.direction = dy < 0 ? "up" : "down";
    }

    function updateActor(dt: number) {
      handleCatAI(dt);
      const { dx, dy } = movementVector();
      actor.moving = Math.abs(dx) + Math.abs(dy) > 0;
      if (actor.moving) {
        setDirection(dx, dy);
        actor.x += dx * actorSpeed() * dt;
        actor.y += dy * actorSpeed() * dt;
      }
      const grounded = clampPointToGround(actor.x, actor.y);
      actor.x = grounded.x;
      actor.y = grounded.y;
      const frameDuration = actor.moving ? 0.12 : 0.24;
      actor.frameClock += dt;
      if (actor.frameClock >= frameDuration) {
        actor.frameClock = 0;
        actor.frame = (actor.frame + 1) % FRAME_COUNT;
      }
      actor.bobClock += dt;
    }

    function drawPlayer() {
      const mode: Mode = actor.moving ? "walk" : "idle";
      const sprite = sprites.get(`${mode}:${actor.direction}:${actor.frame + 1}`);
      ctx.save();
      ctx.globalAlpha = 0.32;
      ctx.fillStyle = "#030607";
      ctx.beginPath();
      ctx.ellipse(actor.x, actor.y + 1, 15 * actor.scale, 4 * actor.scale, 0, 0, Math.PI * 2);
      ctx.fill();
      ctx.restore();
      if (!sprite) return;
      const { image, bounds } = sprite;
      const visibleHeight = actor.targetBodyHeight * actor.scale;
      const drawScale = visibleHeight / bounds.height;
      const w = bounds.width * drawScale;
      const h = bounds.height * drawScale;
      ctx.drawImage(
        image,
        bounds.x,
        bounds.y,
        bounds.width,
        bounds.height,
        Math.round(actor.x - w / 2),
        Math.round(actor.y - h),
        Math.round(w),
        Math.round(h),
      );
    }

    function drawEmote() {
      if (!actor.emote || actor.emoteClock <= 0) return;
      const alpha = Math.min(1, actor.emoteClock / 0.35);
      const text = actor.emote;
      const x = Math.round(actor.x + 18);
      const y = Math.round(actor.y - actor.targetBodyHeight * actor.scale - 12);
      ctx.save();
      ctx.globalAlpha = alpha;
      ctx.font = "bold 11px monospace";
      const width = Math.max(18, ctx.measureText(text).width + 12);
      ctx.fillStyle = "rgba(7, 12, 14, 0.86)";
      ctx.strokeStyle =
        actor.state === "alarm" || actor.state === "critical" ? "#ff814d" : "#70e39a";
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.roundRect(x, y - 16, width, 18, 5);
      ctx.fill();
      ctx.stroke();
      ctx.fillStyle =
        actor.state === "alarm" || actor.state === "critical" ? "#ffc857" : "#dfffe9";
      ctx.fillText(text, x + 6, y - 3);
      ctx.restore();
    }

    function draw() {
      ctx.clearRect(0, 0, STAGE.width, STAGE.height);
      const snap = snapshotRef.current;
      const room = currentRoom();
      const bg = room ? roomImages.get(room.id) : null;
      if (bg) {
        ctx.drawImage(bg, 0, 0, STAGE.width, STAGE.height);
      } else {
        ctx.fillStyle = "#071013";
        ctx.fillRect(0, 0, STAGE.width, STAGE.height);
      }
      if (pointer.active) {
        ctx.save();
        ctx.strokeStyle = "#70e39a";
        ctx.beginPath();
        ctx.arc(pointer.x, pointer.y, 5, 0, Math.PI * 2);
        ctx.moveTo(pointer.x - 8, pointer.y);
        ctx.lineTo(pointer.x + 8, pointer.y);
        ctx.moveTo(pointer.x, pointer.y - 8);
        ctx.lineTo(pointer.x, pointer.y + 8);
        ctx.stroke();
        ctx.restore();
      }
      drawPlayer();
      drawEmote();
      if (snap?.state.paused || snap?.state.mining_paused) {
        ctx.save();
        ctx.fillStyle = "rgba(5, 8, 10, 0.42)";
        ctx.fillRect(0, 0, STAGE.width, STAGE.height);
        ctx.fillStyle = "#ffc857";
        ctx.font = "12px monospace";
        ctx.fillText(snap.state.paused ? "PAUSED" : "REBOOTING", 18, 28);
        ctx.restore();
      }
    }

    function loop(now: number) {
      if (cancelled) return;
      const dt = Math.min(0.05, (now - previousTime) / 1000);
      previousTime = now;

      const snap = snapshotRef.current;
      const room = snap?.rooms.find((r) => r.id === snap.state.current_room);
      if (room && !roomImages.has(room.id)) {
        const img = new Image();
        img.onload = () => roomImages.set(room.id, img);
        img.src = room.background;
      }

      updateActor(dt);
      draw();
      raf = requestAnimationFrame(loop);
    }

    function loadImage(src: string): Promise<HTMLImageElement> {
      return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve(img);
        img.onerror = () => reject(new Error(`asset ${src}`));
        img.src = src;
      });
    }

    function extractSpriteBody(image: HTMLImageElement): Sprite {
      const scratch = document.createElement("canvas");
      scratch.width = image.width;
      scratch.height = image.height;
      const sCtx = scratch.getContext("2d", { willReadFrequently: true });
      if (!sCtx) return { image, bounds: { x: 0, y: 0, width: image.width, height: image.height } };
      sCtx.drawImage(image, 0, 0);
      const frame = sCtx.getImageData(0, 0, image.width, image.height);
      const pixels = frame.data;
      const total = image.width * image.height;
      const visited = new Uint8Array(total);
      let best: { component: number[]; bounds: { x: number; y: number; width: number; height: number } } | null = null;

      for (let start = 0; start < total; start += 1) {
        if (visited[start] || pixels[start * 4 + 3] < 8) continue;
        const stack = [start];
        const component: number[] = [];
        visited[start] = 1;
        let left = image.width;
        let right = 0;
        let top = image.height;
        let bottom = 0;

        while (stack.length) {
          const index = stack.pop()!;
          component.push(index);
          const px = index % image.width;
          const py = Math.floor(index / image.width);
          if (px < left) left = px;
          if (px + 1 > right) right = px + 1;
          if (py < top) top = py;
          if (py + 1 > bottom) bottom = py + 1;
          for (let ny = py - 1; ny <= py + 1; ny += 1) {
            for (let nx = px - 1; nx <= px + 1; nx += 1) {
              if (nx < 0 || ny < 0 || nx >= image.width || ny >= image.height) continue;
              const next = ny * image.width + nx;
              if (visited[next] || pixels[next * 4 + 3] < 8) continue;
              visited[next] = 1;
              stack.push(next);
            }
          }
        }
        if (!best || component.length > best.component.length) {
          best = {
            component,
            bounds: { x: left, y: top, width: right - left, height: bottom - top },
          };
        }
      }

      if (!best) {
        return { image, bounds: { x: 0, y: 0, width: image.width, height: image.height } };
      }
      const keep = new Uint8Array(total);
      for (const index of best.component) keep[index] = 1;
      for (let index = 0; index < total; index += 1) {
        if (!keep[index]) pixels[index * 4 + 3] = 0;
      }
      sCtx.putImageData(frame, 0, 0);
      return { image: scratch, bounds: best.bounds };
    }

    async function preloadSprites() {
      const tasks: Promise<void>[] = [];
      for (const mode of ["idle", "walk"] as const) {
        for (const direction of DIRECTIONS) {
          for (let frame = 1; frame <= FRAME_COUNT; frame += 1) {
            const key = `${mode}:${direction}:${frame}`;
            const src = `${CHARACTER_ROOT}/player/${mode}/${direction}-${frame}.png`;
            tasks.push(
              loadImage(src)
                .then((img) => {
                  sprites.set(key, extractSpriteBody(img));
                })
                .catch(() => {
                  /* sprite missing — keep going */
                }),
            );
          }
        }
      }
      await Promise.all(tasks);
    }

    function canvasPoint(event: PointerEvent): ScenePoint {
      const rect = canvas.getBoundingClientRect();
      return {
        x: ((event.clientX - rect.left) / rect.width) * STAGE.width,
        y: ((event.clientY - rect.top) / rect.height) * STAGE.height,
      };
    }

    const onPointerDown = (event: PointerEvent) => {
      const p = canvasPoint(event);
      const grounded = clampPointToGround(p.x, p.y);
      pointer.x = grounded.x;
      pointer.y = grounded.y;
      pointer.active = true;
      setActorTarget(grounded, "called", "?");
      actor.decisionClock = 2.4;
      canvas.setPointerCapture(event.pointerId);
    };

    const onPointerMove = (event: PointerEvent) => {
      if (!pointer.active || event.buttons === 0) return;
      const p = canvasPoint(event);
      const grounded = clampPointToGround(p.x, p.y);
      pointer.x = grounded.x;
      pointer.y = grounded.y;
      setActorTarget(grounded, "called");
      actor.decisionClock = 2.4;
    };

    canvas.addEventListener("pointerdown", onPointerDown);
    canvas.addEventListener("pointermove", onPointerMove);

    preloadSprites().finally(() => {
      previousTime = performance.now();
      raf = requestAnimationFrame(loop);
    });

    return () => {
      cancelled = true;
      cancelAnimationFrame(raf);
      canvas.removeEventListener("pointerdown", onPointerDown);
      canvas.removeEventListener("pointermove", onPointerMove);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      id="stage"
      width={STAGE.width}
      height={STAGE.height}
      aria-label="2D game stage"
    />
  );
}
