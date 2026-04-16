export interface FileInfo {
  id: number;
  name: string;
  original: string;
  suffix: string;
  scene: string;
  size: number;
  url: string;
  path: string;
  engine: string;
  hash: string;
  createdBy: number;
  createdByName?: string;
  createdAt: string;
  updatedAt: string;
}

export interface FileUsageSceneItem {
  value: string;
  label: string;
}

export interface FileSuffixItem {
  value: string;
  label: string;
}

export interface FileDetail extends FileInfo {
  sceneLabel: string;
}
