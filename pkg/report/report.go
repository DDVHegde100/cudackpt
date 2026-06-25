package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func RenderImage(dir string) (string, error) {
	entries, hdr, err := image.ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return "", err
	}
	meta, _ := image.ReadMeta(filepath.Join(dir, "meta.bin"))
	st, _ := os.Stat(filepath.Join(dir, "device.bin"))
	var b strings.Builder
	writeRow(&b, "field", "value")
	writeRow(&b, "manifest_version", fmt.Sprintf("%d", hdr.Version))
	writeRow(&b, "chunks", fmt.Sprintf("%d", hdr.Count))
	writeRow(&b, "device_bytes", fmt.Sprintf("%d", hdr.TotalBytes))
	if st != nil {
		writeRow(&b, "device_file_bytes", fmt.Sprintf("%d", st.Size()))
	}
	if meta.Pid > 0 {
		writeRow(&b, "source_pid", fmt.Sprintf("%d", meta.Pid))
		writeRow(&b, "gpu_device", fmt.Sprintf("%d", meta.Dev))
		writeRow(&b, "ld_preload", meta.Preload)
		writeRow(&b, "cuda_visible", meta.Visible)
	}
	writeRow(&b, "path", dir)
	b.WriteString("\nchunks\n")
	writeRow(&b, "idx", "ptr", "size", "seq", "crc32c")
	for i, e := range entries {
		if i >= 16 {
			b.WriteString(fmt.Sprintf("... %d more\n", len(entries)-16))
			break
		}
		writeRow(&b, fmt.Sprintf("%d", i), fmt.Sprintf("0x%x", e.Ptr), fmt.Sprintf("%d", e.Size),
			fmt.Sprintf("%d", e.Seq), fmt.Sprintf("%08x", e.CRC32C))
	}
	return b.String(), nil
}

func writeRow(b *strings.Builder, cols ...string) {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c) + 2
		if widths[i] < 14 {
			widths[i] = 14
		}
	}
	b.WriteString("+")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w))
		b.WriteString("+")
	}
	b.WriteString("\n|")
	for i, c := range cols {
		b.WriteString(fmt.Sprintf(" %-*s |", widths[i]-1, c))
	}
	b.WriteString("\n+")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w))
		b.WriteString("+")
	}
	b.WriteString("\n")
}
