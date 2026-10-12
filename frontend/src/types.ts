export interface KnownFolder {
  key: string;
  name: string;
  path: string;
  icon: string;
}

export interface Category {
  key: string;
  name: string;
  icon: string;
  known: string;
  cluster: string;
  exts: string[];
}

export interface Move {
  from: string;
  to: string;
  size: number;
  category: string;
  knownKey: string;
}

export interface RunRecord {
  id: string;
  timestamp: string;
  source: string;
  moves: Move[];
  bytes: number;
  undone: boolean;
}

export interface ScheduleStatus {
  enabled: boolean;
  times: string[];
  registered: string[];
  instant: boolean;
  watchActive: boolean;
  isAdmin: boolean;
  detail: string;
}

export interface CustomRule {
  ext: string;
  target: string;
}

export interface PlannedMove {
  from: string;
  to: string;
  name: string;
  size: number;
  modified: number;
  category: string;
  icon: string;
  knownKey: string;
  kind: string;
  err: string;
}

export interface GroupStat {
  category: string;
  icon: string;
  folder: string;
  count: number;
  bytes: number;
}

export interface Report {
  source: string;
  apply: boolean;
  scanned: number;
  skipped: number;
  skippedYoung: number;
  moved: number;
  moves: PlannedMove[];
  bytes: number;
  groups: GroupStat[];
  errors: string[];
}

export interface DashboardData {
  sources: string[];
  knownFolders: KnownFolder[];
  categories: Category[];
  schedule: ScheduleStatus;
  lastRun: RunRecord | null;
  totalBytes: number;
  runCount: number;
  configPath: string;
  downloadsCount: number;
}

export interface AppConfig {
  sources: string[];
  customRules: CustomRule[];
  times: string[];
  scheduleTime: string;
  enabled: boolean;
  instant: boolean;
  instantDelay: number;
  theme: string;
  minCluster: number;
  minAgeSeconds: number;
}

export interface AboutInfo {
  name: string;
  version: string;
  author: string;
  configPath: string;
  exePath: string;
}
