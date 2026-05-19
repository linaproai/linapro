export interface FileInfo {
  id: number;
  name: string;
  original: string;
  suffix: string;
  scene: string;
  size: number;
  url: string;
  createdBy: number;
  createdByName?: string;
  createdAt: number | null;
  updatedAt: number | null;
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
