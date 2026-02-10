export interface DiskUsage {
  total: number;
  used: number;
  free: number;
  usedPercent: number;
  path: string;
}

export interface CategoryResult {
  id: string;
  name: string;
  description: string;
  size: number;
  fileCount: number;
  color: string;
  selected: boolean;
}

export interface ScanResult {
  categories: CategoryResult[];
  totalSize: number;
  totalFiles: number;
}

export interface CleanResult {
  cleanedSize: number;
  cleanedFiles: number;
  errors: string[];
}

export interface UnusedFile {
  path: string;
  name: string;
  size: number;
  lastAccess: string;
  type: 'document' | 'image' | 'video' | 'archive' | 'other';
}

export interface Config {
  general: {
    startOnBoot: boolean;
    minimizeToTray: boolean;
    scanReminder: boolean;
    reminderInterval: number;
    language: string;
  };
  scan: {
    nodeModules: boolean;
    goModCache: boolean;
    systemTemp: boolean;
    browserCache: boolean;
    customPaths: string[];
  };
  unusedFiles: {
    minDays: number;
    minSize: number;
    excludePaths: string[];
  };
  ui: {
    theme: 'light' | 'dark' | 'system';
  };
}

export type Page = 'dashboard' | 'analyzer' | 'vip' | 'settings';

export interface VIPStatus {
  isVip: boolean;
  isTrialExpired: boolean;
  trialDaysLeft: number;
  trialDays: number;
  firstUseTime: number;
  vipActivatedAt: number;
}
