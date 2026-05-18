import type { UploadFile } from 'ant-design-vue';

/** Default image accept extensions */
export const defaultImageAcceptExts = [
  '.jpg',
  '.jpeg',
  '.png',
  '.gif',
  '.webp',
];

/** Default file accept extensions */
export const defaultFileAcceptExts = ['.xlsx', '.csv', '.docx', '.pdf'];

/** Default file preview: open in browser */
export function defaultFilePreview(file: UploadFile) {
  if (file?.url) {
    window.open(file.url);
  }
}
