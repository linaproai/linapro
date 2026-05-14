// This file implements image base64 and PNG data URL helpers.

package water

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"lina-core/pkg/bizerr"
)

// decodeImageDataURL decodes base64 image text with or without a data URL prefix.
func decodeImageDataURL(value string) ([]byte, error) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return nil, bizerr.NewCode(CodeWaterImageRequired)
	}
	if strings.HasPrefix(clean, "data:image/") {
		comma := strings.Index(clean, ",")
		if comma < 0 {
			return nil, bizerr.NewCode(CodeWaterImageBase64Invalid)
		}
		clean = clean[comma+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(clean)
	}
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterImageBase64Invalid)
	}
	return decoded, nil
}

// encodePNGDataURL encodes PNG bytes as a data URL.
func encodePNGDataURL(img []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(img)
}

// ensurePNGDataURL converts image bytes to PNG data URL.
func ensurePNGDataURL(input []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageDecodeFailed)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageEncodeFailed)
	}
	return encodePNGDataURL(buf.Bytes()), nil
}

// decodeSupportedImage decodes PNG or JPEG image bytes.
func decodeSupportedImage(input []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
	if err == nil {
		return img, nil
	}
	if img, jpegErr := jpeg.Decode(bytes.NewReader(input)); jpegErr == nil {
		return img, nil
	}
	if img, pngErr := png.Decode(bytes.NewReader(input)); pngErr == nil {
		return img, nil
	}
	return nil, bizerr.WrapCode(err, CodeWaterImageDecodeFailed)
}
