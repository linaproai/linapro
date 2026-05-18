/**
 * Download a Blob as a file
 */
export function downloadBlob(data: BlobPart, filename: string, mime?: string) {
  const blob = new Blob([data], {
    type: mime || 'application/octet-stream',
  });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.style.display = 'none';
  link.href = url;
  link.setAttribute('download', filename);
  document.body.append(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
}

/**
 * Download Excel file from an API function
 */
export async function downloadExcel(
  apiFn: (data?: any) => Promise<Blob>,
  filename: string,
  params?: any,
) {
  const data = await apiFn(params);
  downloadBlob(data, `${filename}.xlsx`);
}
