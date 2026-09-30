// unzip is a portable Info-ZIP-compatible extractor for MBWS attachment `unzip`.
//
// Why not ship macOS /usr/bin/unzip: Apple's platform binary is AMFI-restricted and
 // gets killed (exit 137/-1, no stderr) when copied out of /usr/bin into the
// attachment cache. This tool is a pure-Go static binary that runs from any path.
//
// Supported (matches MBWS Video_Cut archive.ts):
//
//	unzip -Z1 <zipPath>
//	unzip -o <zipPath> -d <destDir>
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  unzip -Z1 <zipPath>")
	fmt.Fprintln(os.Stderr, "  unzip -o <zipPath> -d <destDir>")
	os.Exit(2)
}

func parseArgs(argv []string) (mode string, zipPath, destDir string, overwrite bool) {
	if len(argv) == 0 {
		usage()
	}
	if argv[0] == "-Z1" {
		if len(argv) < 2 || argv[1] == "" {
			usage()
		}
		return "list", argv[1], "", false
	}

	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "-o":
			overwrite = true
		case "-d":
			i++
			if i >= len(argv) || argv[i] == "" {
				usage()
			}
			destDir = argv[i]
		default:
			if zipPath == "" {
				zipPath = argv[i]
			} else {
				usage()
			}
		}
	}
	if zipPath == "" || destDir == "" {
		usage()
	}
	return "extract", zipPath, destDir, overwrite
}

func listZip(zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fmt.Println(normalizeName(f.Name))
	}
	return nil
}

func normalizeName(name string) string {
	return strings.ReplaceAll(name, "\\", "/")
}

func safeJoin(destDir, name string) (string, error) {
	name = normalizeName(name)
	if name == "" || strings.Contains(name, "\x00") {
		return "", fmt.Errorf("非法条目名")
	}
	// Clean then ensure still under destDir (block .. traversal).
	clean := filepath.Clean("/" + name)
	if clean == "/" {
		return "", fmt.Errorf("空条目路径")
	}
	rel := strings.TrimPrefix(clean, "/")
	target := filepath.Join(destDir, rel)
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	sep := string(os.PathSeparator)
	if targetAbs != destAbs && !strings.HasPrefix(targetAbs, destAbs+sep) {
		return "", fmt.Errorf("拒绝路径穿越：%s", name)
	}
	return targetAbs, nil
}

func extractZip(zipPath, destDir string, overwrite bool) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		name := normalizeName(f.Name)
		if name == "" {
			continue
		}
		target, err := safeJoin(destDir, name)
		if err != nil {
			return err
		}

		info := f.FileInfo()
		if info.IsDir() || strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if !overwrite {
			if _, err := os.Stat(target); err == nil {
				fmt.Fprintf(os.Stderr, "跳过已存在：%s\n", name)
				continue
			}
		}

		if err := writeFile(f, target); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		fmt.Printf("extracting: %s\n", name)
	}
	return nil
}

func writeFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	mode := f.Mode()
	if mode&0o111 == 0 {
		mode = 0o644
	} else {
		mode = 0o755
	}

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func main() {
	mode, zipPath, destDir, overwrite := parseArgs(os.Args[1:])
	var err error
	switch mode {
	case "list":
		err = listZip(zipPath)
	case "extract":
		err = extractZip(zipPath, destDir, overwrite)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
