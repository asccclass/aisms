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
	if runtime.GOOS == "windows" {
		if powershellPath, err := exec.LookPath("powershell.exe"); err == nil {
			return powershellPath, wordCOMArgs, nil
		}
	}

	if configured := strings.TrimSpace(os.Getenv("SOFFICE_PATH")); configured != "" {
		if fileExists(configured) {
			return configured, libreOfficeArgs, nil
		}
	}

	candidates := []struct {
		name    string
		argsFor converterArgsBuilder
	}{
		{
			name:    "soffice",
			argsFor: libreOfficeArgs,
		},
		{
			name:    "libreoffice",
			argsFor: libreOfficeArgs,
		},
	}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate.name); err == nil {
			return path, candidate.argsFor, nil
		}
	}

	for _, path := range platformSpecificCandidates() {
		if fileExists(path) {
			return path, libreOfficeArgs, nil
		}
	}

	hint := "install LibreOffice and ensure `soffice` is in PATH"
	if runtime.GOOS == "windows" {
		hint = "install Microsoft Word for Word COM export, or install LibreOffice, or set SOFFICE_PATH to soffice.exe, or add soffice.exe to PATH"
	}
	return "", nil, fmt.Errorf("no DOCX to PDF converter found; %s", hint)
}

func libreOfficeArgs(tempDir, inputPath string) []string {
	return []string{"--headless", "--convert-to", "pdf", "--outdir", tempDir, inputPath}
}

func wordCOMArgs(_ string, inputPath string) []string {
	outputPath := replaceExtension(inputPath, ".pdf")
	script := strings.Join([]string{
		"$ErrorActionPreference = 'Stop'",
		"$word = $null",
		"$doc = $null",
		"try {",
		"  $word = New-Object -ComObject Word.Application",
		"  $word.Visible = $false",
		"  $word.DisplayAlerts = 0",
		fmt.Sprintf("  $inputPath = '%s'", escapePowerShellSingleQuoted(inputPath)),
		fmt.Sprintf("  $outputPath = '%s'", escapePowerShellSingleQuoted(outputPath)),
		"  $doc = $word.Documents.Open($inputPath, $false, $true)",
		"  $doc.ExportAsFixedFormat($outputPath, 17)",
		"} finally {",
		"  if ($doc -ne $null) { $doc.Close($false) }",
		"  if ($word -ne $null) { $word.Quit() }",
		"  [System.GC]::Collect()",
		"  [System.GC]::WaitForPendingFinalizers()",
		"}",
	}, "; ")

	return []string{
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-Command", script,
	}
}

func platformSpecificCandidates() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	return []string{
		`C:\Program Files\LibreOffice\program\soffice.exe`,
		`C:\Program Files (x86)\LibreOffice\program\soffice.exe`,
		`C:\Program Files\LibreOffice 25\program\soffice.exe`,
		`C:\Program Files\LibreOffice 24\program\soffice.exe`,
		`C:\Program Files\LibreOffice 7\program\soffice.exe`,
		`C:\Program Files (x86)\LibreOffice 7\program\soffice.exe`,
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
