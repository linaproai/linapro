// This file implements pure-Go image and text watermark rendering.

package water

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"lina-core/pkg/bizerr"
)

// drawWatermark renders the configured image and text watermark.
func drawWatermark(input []byte, cfg watermarkConfig) ([]byte, error) {
	base, err := decodeSupportedImage(input)
	if err != nil {
		return nil, err
	}
	bounds := base.Bounds()
	canvas := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(canvas, canvas.Bounds(), base, bounds.Min, draw.Src)

	if strings.TrimSpace(cfg.Image) != "" || strings.TrimSpace(cfg.Base64) != "" {
		if err := drawImageWatermark(canvas, cfg); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(cfg.Text) != "" {
		if err := drawTextWatermark(canvas, cfg); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterImageEncodeFailed)
	}
	return buf.Bytes(), nil
}

// drawImageWatermark overlays a full-size image watermark.
func drawImageWatermark(dst *image.RGBA, cfg watermarkConfig) error {
	watermarkBytes, err := loadWatermarkImageBytes(cfg)
	if err != nil {
		return err
	}
	if len(watermarkBytes) == 0 {
		return nil
	}
	img, err := decodeSupportedImage(watermarkBytes)
	if err != nil {
		return err
	}
	overlay := image.NewRGBA(dst.Bounds())
	xdraw.ApproxBiLinear.Scale(overlay, overlay.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	blendImage(dst, overlay, cfg.Opacity)
	return nil
}

// drawTextWatermark overlays multi-line text.
func drawTextWatermark(dst *image.RGBA, cfg watermarkConfig) error {
	face := loadFontFace(cfg.Font, cfg.FontSize)
	textColor := parseHexColor(cfg.Color, cfg.Opacity)
	lines := splitTextLines(cfg.Text)
	if len(lines) == 0 {
		return nil
	}

	lineHeight := fontHeight(face)
	if lineHeight <= 0 {
		lineHeight = cfg.FontSize
	}
	lineSpacing := int(math.Ceil(float64(lineHeight) * 1.4))
	totalHeight := lineSpacing*(len(lines)-1) + lineHeight

	for idx, line := range lines {
		lineWidth := textWidth(face, line)
		x, y := textPosition(dst.Bounds(), cfg, lineWidth, totalHeight, idx, lineSpacing, lineHeight)
		drawer := &font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(textColor),
			Face: face,
			Dot:  fixed.P(x, y),
		}
		drawer.DrawString(line)
	}
	return nil
}

// loadWatermarkImageBytes loads watermark image bytes from data URL or file path.
func loadWatermarkImageBytes(cfg watermarkConfig) ([]byte, error) {
	if strings.TrimSpace(cfg.Base64) != "" {
		return decodeImageDataURL(cfg.Base64)
	}
	if strings.TrimSpace(cfg.Image) == "" {
		return nil, nil
	}
	data, err := os.ReadFile(cfg.Image)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterImageDecodeFailed)
	}
	return data, nil
}

// blendImage overlays src onto dst using the configured opacity.
func blendImage(dst *image.RGBA, src *image.RGBA, opacity float64) {
	if opacity <= 0 {
		return
	}
	if opacity > 1 {
		opacity = 1
	}
	bounds := dst.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dr, dg, db, da := dst.At(x, y).RGBA()
			sr, sg, sb, sa := src.At(x, y).RGBA()
			alpha := (float64(sa) / 65535.0) * opacity
			inv := 1 - alpha
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(((float64(sr)/257.0)*alpha + (float64(dr)/257.0)*inv)),
				G: uint8(((float64(sg)/257.0)*alpha + (float64(dg)/257.0)*inv)),
				B: uint8(((float64(sb)/257.0)*alpha + (float64(db)/257.0)*inv)),
				A: uint8(math.Max(float64(da)/257.0, 255*alpha)),
			})
		}
	}
}

// loadFontFace returns a truetype face when available and falls back to basicfont.
func loadFontFace(configuredPath string, size int) font.Face {
	if size <= 0 {
		size = defaultFontSize
	}
	paths := candidateFontPaths(configuredPath)
	for _, path := range paths {
		face, err := loadFontFaceFromPath(path, size)
		if err == nil && face != nil {
			return face
		}
	}
	return basicfont.Face7x13
}

// loadFontFaceFromPath parses one font file as an opentype face.
func loadFontFaceFromPath(path string, size int) (font.Face, error) {
	if strings.TrimSpace(path) == "" {
		return nil, os.ErrNotExist
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsed, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

// candidateFontPaths returns configured and platform-default font candidates.
func candidateFontPaths(configuredPath string) []string {
	paths := make([]string, 0, 8)
	if strings.TrimSpace(configuredPath) != "" {
		paths = append(paths, strings.TrimSpace(configuredPath))
	}
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
			"/System/Library/Fonts/Supplemental/AppleGothic.ttf",
			"/System/Library/Fonts/SFNS.ttf",
		)
	case "linux":
		paths = append(paths,
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		)
	}
	return paths
}

// splitTextLines returns non-empty text lines.
func splitTextLines(text string) []string {
	rawLines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// fontHeight returns one face line height in pixels.
func fontHeight(face font.Face) int {
	metrics := face.Metrics()
	return (metrics.Ascent + metrics.Descent).Ceil()
}

// textWidth returns rendered string width in pixels.
func textWidth(face font.Face, text string) int {
	drawer := &font.Drawer{Face: face}
	return drawer.MeasureString(text).Ceil()
}

// textPosition calculates the text baseline position.
func textPosition(bounds image.Rectangle, cfg watermarkConfig, width int, totalHeight int, idx int, spacing int, lineHeight int) (int, int) {
	if cfg.PosX > 0 || cfg.PosY > 0 {
		return cfg.PosX, cfg.PosY + idx*spacing + lineHeight
	}
	padding := 16
	x := padding
	yTop := padding
	align := normalizedAlignment(cfg.Align)
	switch align {
	case "left":
		yTop = (bounds.Dy() - totalHeight) / 2
	case "center":
		x = (bounds.Dx() - width) / 2
		yTop = (bounds.Dy() - totalHeight) / 2
	case "right":
		x = bounds.Dx() - width - padding
		yTop = (bounds.Dy() - totalHeight) / 2
	case "top":
		x = (bounds.Dx() - width) / 2
	case "bottom":
		x = (bounds.Dx() - width) / 2
		yTop = bounds.Dy() - totalHeight - padding
	case "topright":
		x = bounds.Dx() - width - padding
	case "bottomleft":
		yTop = bounds.Dy() - totalHeight - padding
	case "bottomright":
		x = bounds.Dx() - width - padding
		yTop = bounds.Dy() - totalHeight - padding
	case "topleft":
	default:
	}
	if x < padding {
		x = padding
	}
	if yTop < padding {
		yTop = padding
	}
	return x, yTop + idx*spacing + lineHeight
}

// normalizedAlignment converts named or hotgo numeric alignment values.
func normalizedAlignment(align watermarkAlignment) string {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(string(align)), "_", ""))
	value = strings.ReplaceAll(value, "-", "")
	switch value {
	case "1", "left":
		return "left"
	case "2", "center", "centre":
		return "center"
	case "3", "right":
		return "right"
	case "4", "top":
		return "top"
	case "5", "bottom":
		return "bottom"
	case "6", "topleft":
		return "topleft"
	case "7", "topright":
		return "topright"
	case "8", "bottomleft":
		return "bottomleft"
	case "9", "bottomright":
		return "bottomright"
	default:
		return "topleft"
	}
}

// parseHexColor parses #RGB or #RRGGBB colors with opacity.
func parseHexColor(value string, opacity float64) color.RGBA {
	trimmed := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(trimmed) == 3 {
		trimmed = string([]byte{trimmed[0], trimmed[0], trimmed[1], trimmed[1], trimmed[2], trimmed[2]})
	}
	if len(trimmed) != 6 {
		trimmed = "ffffff"
	}
	r := parseHexByte(trimmed[0:2])
	g := parseHexByte(trimmed[2:4])
	b := parseHexByte(trimmed[4:6])
	if opacity <= 0 {
		opacity = defaultWatermarkOpacity
	}
	if opacity > 1 {
		opacity = 1
	}
	return color.RGBA{R: r, G: g, B: b, A: uint8(math.Round(opacity * 255))}
}

// parseHexByte parses one two-character hex byte.
func parseHexByte(value string) uint8 {
	parsed, err := strconv.ParseUint(value, 16, 8)
	if err != nil {
		return 255
	}
	return uint8(parsed)
}
