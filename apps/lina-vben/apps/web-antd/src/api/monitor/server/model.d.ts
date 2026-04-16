export interface ServerNodeInfo {
  nodeName: string;
  nodeIp: string;
  collectAt: string;
  server: {
    hostname: string;
    os: string;
    arch: string;
    bootTime: string;
    uptime: number;
    startTime: string;
  };
  cpu: {
    cores: number;
    modelName: string;
    usagePercent: number;
  };
  memory: {
    total: number;
    used: number;
    available: number;
    usagePercent: number;
  };
  disks: Array<{
    path: string;
    fsType: string;
    total: number;
    used: number;
    free: number;
    usagePercent: number;
  }>;
  network: {
    bytesSent: number;
    bytesRecv: number;
    sendRate: number;
    recvRate: number;
  };
  goInfo: {
    version: string;
    goroutines: number;
    processCpu: number;
    processMemory: number;
    gcPauseNs: number;
    gfVersion: string;
    serviceUptime: string;
  };
}

export interface ServerMonitorResult {
  nodes: ServerNodeInfo[];
  dbInfo: {
    version: string;
    maxOpenConns: number;
    openConns: number;
    inUse: number;
    idle: number;
  };
}

export interface ServerMonitorParams {
  nodeName?: string;
}
