export interface Snapshot {
  state: GameState;
  rooms: Room[];
  gpus: GPU[];
  gpu_defs: GPUDef[];
  skills: Skill[];
  mercs: Merc[];
  merc_defs: MercDef[];
  log: LogEntry[];
  last_event?: GameEvent;
  modifiers: Modifier[];
  active_research?: ActiveResearch;
  research_tiers: ResearchTier[];
  blueprints: Blueprint[];
  achievements: string[];
  achievement_defs: AchievementDef[];
  mastery_levels: Record<string, number>;
  mastery_tracks: MasteryTrack[];
  legacy_perks: LegacyPerk[];
  legacy: LegacySummary;
  stats: Stats;
  ok: boolean;
}

export interface GameState {
  kitten_name: string;
  btc: number;
  btc_fmt: string;
  tech_point: number;
  research_frags: number;
  reputation: number;
  karma: number;
  current_room: string;
  paused: boolean;
  market_price: number;
  prev_market_price: number;
  market_trend: number;
  lifetime_earned: number;
  lifetime_earned_fmt: string;
  legacy_available: number;
  difficulty: string;
  lang: string;
  room_earn_fmt: string;
  room_bill_fmt: string;
  room_net_fmt: string;
  mining_paused: boolean;
  syndicate_joined: boolean;
  syndicate_can_join: boolean;
  syndicate_contribution: number;
  syndicate_total_dividends: number;
  syndicate_next_payout_sec: number;
  pump_dump_unlocked: boolean;
  pump_dump_cooldown_sec: number;
}

export interface Room {
  id: string;
  name: string;
  flavor: string;
  slots: number;
  unlock_cost: number;
  unlock_cost_fmt: string;
  unlocked: boolean;
  current: boolean;
  gpu_count: number;
  heat: number;
  max_heat: number;
  heat_pct: number;
  heat_delta: number;
  heat_tick_in: number;
  earn_fmt: string;
  bill_fmt: string;
  net_fmt: string;
  defense: Defense;
  background: string;
}

export interface Defense {
  lock: number;
  cctv: number;
  wiring: number;
  cooling: number;
  armor: number;
}

export interface GPU {
  instance_id: number;
  def_id: string;
  blueprint_id?: string;
  name: string;
  status: string;
  room: string;
  upgrade: number;
  oc_level: number;
  hours_left: number;
  earn_fmt: string;
  repairable: boolean;
  ships_at?: number;
  ship_eta_sec?: number;
  ship_total_sec?: number;
}

export interface GPUDef {
  id: string;
  name: string;
  flavor: string;
  tier: string;
  efficiency: number;
  power_draw: number;
  heat_output: number;
  price: number;
  price_fmt: string;
  scrap_fmt: string;
}

export interface Skill {
  id: string;
  lane: string;
  name: string;
  desc: string;
  cost: number;
  prereq: string;
  unlocked: boolean;
}

export interface Merc {
  instance_id: number;
  def_id: string;
  name: string;
  loyalty: number;
  room_id: string;
}

export interface MercDef {
  id: string;
  name: string;
  flavor: string;
  hire_cost_fmt: string;
  wage_fmt: string;
  specialty: string;
}

export interface LogEntry {
  time: number;
  category: string;
  text: string;
}

export interface GameEvent {
  seq: number;
  id: string;
  name: string;
  category: string;
  text: string;
}

export interface Modifier {
  kind: string;
  factor: number;
  expires_at: number;
  seconds_left: number;
}

export interface ActiveResearch {
  tier: number;
  boosts: string[];
  started_at: number;
  duration_sec: number;
  progress: number;
  seconds_left: number;
}

export interface ResearchTier {
  tier: number;
  name: string;
  duration_sec: number;
  frags: number;
  money: number;
  min_lvl: number;
}

export interface Blueprint {
  id: string;
  tier: number;
  boosts: string[];
  created_at: number;
  can_print: boolean;
  print_btc_cost: number;
  print_frag_cost: number;
}

export interface AchievementDef {
  id: string;
  emoji: string;
  name: string;
  desc: string;
  tp_reward: number;
  earned: boolean;
}

export interface MasteryTrack {
  id: string;
  emoji: string;
  name: string;
  desc: string;
  effect: string;
  per_level: number;
  level: number;
  max_level: number;
  next_cost: number;
  maxed: boolean;
}

export interface LegacyPerk {
  id: string;
  name: string;
  desc: string;
  cost: number;
  available: boolean;
  owned: boolean;
}

export interface LegacySummary {
  total_earned: number;
  total_earned_fmt: string;
  total_lp: number;
  spent_lp: number;
  lp_available: number;
  starter_cash: number;
  efficiency_boost: number;
  unlocked_university: boolean;
  carried_tp: number;
}

export interface Stats {
  total_ticks: number;
  total_gpus_bought: number;
  total_gpus_scrapped: number;
  oc_time_t1_sec: number;
  oc_time_t2_sec: number;
  total_wages_paid: number;
  total_wages_paid_fmt: string;
  market_crash_count: number;
  lifetime_earned: number;
  lifetime_earned_fmt: string;
  events_by_category: Record<string, number>;
  market_price_history: number[];
}

export interface ActionRequest {
  action: string;
  id?: string;
  dim?: string;
  instance_id?: number;
  tier?: number;
  boosts?: string[];
  frags?: number;
}

export type TabId =
  | "store"
  | "rooms"
  | "gpus"
  | "defense"
  | "skills"
  | "mercs"
  | "log"
  | "stats";
