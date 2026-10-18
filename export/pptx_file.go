package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Save beside the destination and replace it only after successful encoding.
// Render errors or failed writes leave existing files intact.
func savePPTX(path string, write func(io.Writer) error) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".animation-pptx-*.pptx")
	if err != nil {
		return fmt.Errorf("create PPTX output: %w", err)
	}
	temporary := f.Name()
	defer os.Remove(temporary)
	if err := write(f); err != nil {
		_ = f.Close()
		return fmt.Errorf("write PPTX: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close PPTX: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("publish PPTX: %w", err)
	}
	return nil
}
