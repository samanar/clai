package man

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	// Matches "foo.1", "bar.3p", "baz.5x"
	manFileRe = regexp.MustCompile(`\.(\d[\w]*)$`)
	// Matches "foo.1.gz", "bar.3p.gz"
	manGzFileRe = regexp.MustCompile(`\.(\d[\w]*)\.gz$`)
)

type Chunk struct {
	Path    string // full path to the man file
	Section string // e.g. "1", "5", "3p"
	Index   int    // 0-based chunk index within the file
	Data    []byte // chunk data (not reused)
	Err     error  // non-nil if an error occurred reading this file
}

type Man struct {
	RootPaths []string
	ManFiles  []string
}

// is it really better using named returns ?? seems so weird why everybody loves this ?
func NewMan() (man Man, err error) {
	man = Man{}
	err = man.getRootPaths()
	if err != nil {
		return
	}
	err = man.getManFiles()
	if err != nil {
		return
	}
	return
}

func (m *Man) getRootPaths() error {
	foundPaths := make([]string, 0, 16)

	// 1) Respect MANPATH if set
	if mp := strings.TrimSpace(os.Getenv("MANPATH")); mp != "" {
		foundPaths = append(foundPaths, SplitColon(mp)...)
		fmt.Println(mp)
	}

	if out, err := exec.Command("manpath", "-q").Output(); err == nil {
		if s := strings.TrimSpace(string(out)); s != "" {
			foundPaths = append(foundPaths, SplitColon(s)...)
		}
	}

	if out, err := exec.Command("man", "--path").Output(); err == nil {
		if s := strings.TrimSpace(string(out)); s != "" {
			foundPaths = append(foundPaths, SplitColon(s)...)
		}
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS Intel and Apple Silicon (Homebrew locations differ)
		foundPaths = append(foundPaths,
			"/opt/homebrew/share/man", // Apple Silicon Homebrew
			"/usr/local/share/man",    // Intel Homebrew
			"/usr/share/man",
		)
	case "linux":
		foundPaths = append(foundPaths,
			"/usr/local/share/man",
			"/usr/share/man",
		)
	default:
		// Generic POSIX-ish fallback
		foundPaths = append(foundPaths,
			"/usr/local/share/man",
			"/usr/share/man",
		)
	}

	foundPaths = ValidatePaths(foundPaths)
	if len(foundPaths) == 0 {
		return errors.New("no valid man paths discovered (checked env, manpath, man --path, and fallbacks)")
	}
	m.RootPaths = foundPaths
	return nil
}

func (m *Man) getManFiles() error {
	if len(m.RootPaths) == 0 {
		return errors.New("no man root paths found")
	}
	var files []string
	seen := make(map[string]struct{})

	for _, root := range m.RootPaths {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fs.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			name := d.Name()
			if manGzFileRe.MatchString(name) || manFileRe.MatchString(name) {
				if _, ok := seen[path]; !ok {
					seen[path] = struct{}{}
					files = append(files, path)
				}
			}
			return nil
		})
	}
	m.ManFiles = files
	return nil
}

func Split(s string, separator string) []string {
	parts := strings.Split(s, separator)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func SplitColon(s string) []string {
	return Split(s, ":")
}

func ValidatePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	valid := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Expand ~ if present
		if strings.HasPrefix(p, "~") {
			if home, _ := os.UserHomeDir(); home != "" {
				p = filepath.Join(home, strings.TrimPrefix(p, "~"))
			}
		}
		ap := p
		if !filepath.IsAbs(ap) {
			if abs, err := filepath.Abs(ap); err == nil {
				ap = abs
			}
		}
		if _, ok := seen[ap]; ok {
			continue
		}
		fi, err := os.Stat(ap)
		if err == nil && fi.IsDir() {
			seen[ap] = struct{}{}
			valid = append(valid, ap)
		}
	}
	return valid
}

func ReadManFileInChunks(path string, chunkSize int) <-chan Chunk {
	out := make(chan Chunk, 1)
	go func() {
		defer close(out)

		f, err := os.Open(path)
		if err != nil {
			out <- Chunk{Path: path, Err: err}
			return
		}
		defer f.Close()

		var r io.Reader = f
		section := extractSection(path)

		// gzip support (standard on most distros for man pages)
		if strings.HasSuffix(path, ".gz") {
			gzr, err := gzip.NewReader(f)
			if err != nil {
				out <- Chunk{Path: path, Section: section, Err: err}
				return
			}
			defer gzr.Close()
			r = gzr
		}

		// Buffered reader helps under different filesystems
		br := bufio.NewReader(r)

		idx := 0
		buf := make([]byte, chunkSize)

		for {
			n, readErr := io.ReadFull(br, buf)
			if readErr != nil {
				// io.ReadFull returns an error at EOF; handle partial tail
				if errors.Is(readErr, io.ErrUnexpectedEOF) || errors.Is(readErr, io.EOF) {
					if n > 0 {
						cp := make([]byte, n)
						copy(cp, buf[:n])
						out <- Chunk{Path: path, Section: section, Index: idx, Data: cp}
					}
					break
				}
				// Real error midstream
				out <- Chunk{Path: path, Section: section, Index: idx, Err: readErr}
				break
			}
			cp := make([]byte, n)
			copy(cp, buf[:n])
			out <- Chunk{Path: path, Section: section, Index: idx, Data: cp}
			idx++
		}
	}()
	return out
}

func extractSection(path string) string {
	base := filepath.Base(path)
	// Try .N.gz first
	if m := manGzFileRe.FindStringSubmatch(base); m != nil {
		return m[1]
	}
	// Then .N
	if m := manFileRe.FindStringSubmatch(base); m != nil {
		return m[1]
	}
	return ""
}
