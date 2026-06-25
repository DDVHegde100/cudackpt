package control

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func InspectImage(dir string) error {
	metaPath := filepath.Join(dir, "meta.bin")
	if m, err := image.ReadMeta(metaPath); err == nil {
		fmt.Printf("meta pid=%d dev=%d preload=%q visible=%q\n", m.Pid, m.Dev, m.Preload, m.Visible)
	} else {
		fmt.Printf("meta missing or invalid: %v\n", err)
	}
	if dev, err := image.ReadDev(filepath.Join(dir, "dev.bin")); err == nil {
		fmt.Printf("dev=%d\n", dev)
	}
	entries, hdr, err := image.ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return err
	}
	fmt.Printf("manifest count=%d bytes=%d version=%d\n", hdr.Count, hdr.TotalBytes, hdr.Version)
	for i, e := range entries {
		fmt.Printf("  [%d] ptr=0x%x size=%d off=%d seq=%d crc=%08x\n", i, e.Ptr, e.Size, e.Offset, e.Seq, e.CRC32C)
	}
	st, err := os.Stat(filepath.Join(dir, "device.bin"))
	if err == nil {
		fmt.Printf("device.bin size=%d\n", st.Size())
	}
	if st, err := os.Stat(filepath.Join(dir, "criu")); err == nil && st.IsDir() {
		fmt.Println("criu present")
	}
	if b, err := os.ReadFile(filepath.Join(dir, "restored.pid")); err == nil {
		fmt.Printf("restored.pid=%s", string(b))
	}
	if b, err := os.ReadFile(filepath.Join(dir, "restore.err")); err == nil {
		fmt.Printf("restore.err=%s", string(b))
	}
	if b, err := os.ReadFile(filepath.Join(dir, "snapshot.err")); err == nil {
		fmt.Printf("snapshot.err=%s", string(b))
	}
	return nil
}
