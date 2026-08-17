package pdfexport

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type ConvertOptions struct {
	InputFilename string
	InputBytes    []byte
}

func ConvertDOCXToPDF(opts ConvertOptions) ([]byte, error) {
	if strings.TrimSpace(opts.InputFilename) == "" {
		return nil, fmt.Errorf("input filename is required")
	}
	if len(opts.InputBytes) == 0 {
		return nil, fmt.Errorf("input bytes are required")
	}

	converter, argsBuilder, err := findConverter()
	if err != nil {
		return nil, err
	}

	tempDir, err := os.MkdirTemp("", "isms-pdf-export-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, filepath.Base(opts.InputFilename))
	if err := os.WriteFile(inputPath, opts.InputBytes, 0644); err != nil {
		return nil, err
	}

	args := argsBuilder(tempDir, inputPath)
	cmd := exec.Command(converter, args...)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pdf conversion failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	pdfPath := replaceExtension(inputPath, ".pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("converted pdf not found: %w", err)
	}
	return pdfBytes, nil
}

func replaceExtension(path, ext string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	return base + ext
}

type converterArgsBuilder func(tempDir, inputPath string) []string

func findConverter() (string, converterArgsBuilder, error) {
	candidates := []struct {
		name    string
		argsFor converterArgsBuilder
	}{
		{
			name: "soffice",
			argsFor: func(tempDir, inputPath string) []string {
				return []string{"--headless", "--convert-to", "pdf", "--outdir", tempDir, inputPath}
			},
		},
		{
			name: "libreoffice",
			argsFor: func(tempDir, inputPath string) []string {
				return []string{"--headless", "--convert-to", "pdf", "--outdir", tempDir, inputPath}
			},
		},
	}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate.name); err == nil {
			return path, candidate.argsFor, nil
		}
	}

	hint := "install LibreOffice and ensure `soffice` is in PATH"
	if runtime.GOOS == "windows" {
		hint = "install LibreOffice and ensure `soffice.exe` is in PATH"
	}
	return "", nil, fmt.Errorf("no DOCX to PDF converter found; %s", hint)
}
