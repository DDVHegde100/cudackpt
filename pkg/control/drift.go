package control

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

type DriftEntry struct {
	Ptr    uint64
	Field  string
	Left   string
	Right  string
}

type DriftReport struct {
	Left      string
	Right     string
	Manifest  []DriftEntry
	MetaDrift []DriftEntry
	FlagDrift []DriftEntry
}

func CompareImages(left, right string) (DriftReport, error) {
	rep := DriftReport{Left: left, Right: right}
	le, lh, err := image.ReadManifest(filepath.Join(left, "manifest.bin"))
	if err != nil {
		return rep, err
	}
	re, rh, err := image.ReadManifest(filepath.Join(right, "manifest.bin"))
	if err != nil {
		return rep, err
	}
	if lh.Flags != rh.Flags {
		rep.FlagDrift = append(rep.FlagDrift, DriftEntry{
			Field: "flags", Left: fmt.Sprintf("0x%x", lh.Flags), Right: fmt.Sprintf("0x%x", rh.Flags),
		})
	}
	if lh.Count != rh.Count {
		rep.Manifest = append(rep.Manifest, DriftEntry{
			Field: "count", Left: fmt.Sprintf("%d", lh.Count), Right: fmt.Sprintf("%d", rh.Count),
		})
	}
	lm, lerr := image.ReadMeta(filepath.Join(left, "meta.bin"))
	rm, rerr := image.ReadMeta(filepath.Join(right, "meta.bin"))
	if lerr == nil && rerr == nil {
		if lm.Dev != rm.Dev {
			rep.MetaDrift = append(rep.MetaDrift, DriftEntry{
				Field: "dev", Left: fmt.Sprintf("%d", lm.Dev), Right: fmt.Sprintf("%d", rm.Dev),
			})
		}
		if lm.Visible != rm.Visible {
			rep.MetaDrift = append(rep.MetaDrift, DriftEntry{
				Field: "cuda_visible", Left: lm.Visible, Right: rm.Visible,
			})
		}
	}
	byPtr := make(map[uint64]image.Entry, len(re))
	for _, e := range re {
		byPtr[e.Ptr] = e
	}
	n := len(le)
	if len(re) < n {
		n = len(re)
	}
	for i := 0; i < n; i++ {
		a := le[i]
		b, ok := byPtr[a.Ptr]
		if !ok {
			rep.Manifest = append(rep.Manifest, DriftEntry{
				Ptr: a.Ptr, Field: "missing", Left: fmt.Sprintf("size=%d", a.Size), Right: "absent",
			})
			continue
		}
		if a.Size != b.Size {
			rep.Manifest = append(rep.Manifest, DriftEntry{
				Ptr: a.Ptr, Field: "size", Left: fmt.Sprintf("%d", a.Size), Right: fmt.Sprintf("%d", b.Size),
			})
		}
		if a.CRC32C != b.CRC32C {
			rep.Manifest = append(rep.Manifest, DriftEntry{
				Ptr: a.Ptr, Field: "crc32c", Left: fmt.Sprintf("%08x", a.CRC32C), Right: fmt.Sprintf("%08x", b.CRC32C),
			})
		}
	}
	return rep, nil
}

func FormatDrift(r DriftReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("compare %s vs %s\n", r.Left, r.Right))
	writeDrift := func(title string, rows []DriftEntry) {
		if len(rows) == 0 {
			return
		}
		b.WriteString(title + "\n")
		for _, d := range rows {
			if d.Ptr != 0 {
				b.WriteString(fmt.Sprintf("  ptr=0x%x %s left=%s right=%s\n", d.Ptr, d.Field, d.Left, d.Right))
			} else {
				b.WriteString(fmt.Sprintf("  %s left=%s right=%s\n", d.Field, d.Left, d.Right))
			}
		}
	}
	writeDrift("flags", r.FlagDrift)
	writeDrift("meta", r.MetaDrift)
	writeDrift("manifest", r.Manifest)
	if len(r.FlagDrift)+len(r.MetaDrift)+len(r.Manifest) == 0 {
		b.WriteString("no drift\n")
	}
	return b.String()
}
