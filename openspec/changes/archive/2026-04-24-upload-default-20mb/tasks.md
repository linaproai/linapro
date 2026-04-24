## 1. Align default-value sources

- [x] 1.1 Change the host initialization default for `sys.upload.maxSize` to 20 MB and update any related manifest or derived artifacts
- [x] 1.2 Update the upload config template and backend static fallback default to 20 MB so the current 10 MB / 16 MB split disappears

## 2. Sync upload-chain validation

- [x] 2.1 Update file-upload validation, request-body size protection, and friendly error-message logic or assertions so the default baseline is consistently 20 MB
- [x] 2.2 Update the affected backend automated tests to cover both the default 20 MB case and runtime override cases

## 3. Verification

- [x] 3.1 Run the affected backend tests and any required initialization checks to confirm the initial value, runtime enforcement, and error messages all use a 20 MB baseline
- [x] 3.2 If the build flow produces embedded or packaged manifest artifacts, verify that those artifacts also use the updated 20 MB default
