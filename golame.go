package golame

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const version = "1.0.0"

const (
	// ExitOK : XXX
	ExitOK = iota
	// ExitError : XXX
	ExitError
)

var filenameRegexp = regexp.MustCompile(`(?i)^(\d{2})( -)? (.+)\.wav$`)

var opts Option

// Lame : XXX
type Lame struct {
	Out io.Writer
	Err io.Writer
}

// Run : XXX
func (Lame) Run(args []string) int {
	parser := newOptionParser(&opts)

	args, err := parser.ParseArgs(args)
	if err != nil {
		if ferr, ok := err.(*flags.Error); ok && ferr.Type == flags.ErrHelp {
			return ExitOK
		}
		return ExitError
	}

	if opts.Version {
		fmt.Printf("golame version %s\n", version)
		return ExitOK
	}

	if opts.InputDir == "" {
		// TODO Message
		fmt.Fprintf(os.Stderr, "Directory '%s' is not exists.\n", opts.InputDir)
		return ExitError
	}

	if opts.OutputDir == "" {
		// TODO Message
		fmt.Fprintf(os.Stderr, "'%s' is not a directory.\n", opts.OutputDir)
		return ExitError
	}
	if _, err := os.Stat(opts.OutputDir); err != nil && os.MkdirAll(opts.OutputDir, 0755) != nil {
		// TODO Message
		fmt.Fprintf(os.Stderr, "Cannot create directory '%s'.\n", opts.OutputDir)
		return ExitError
	}

	files, err := findWavFiles(opts.InputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot find target files.\n")
		return ExitError
	}

	fileCnt := len(files)
	fmt.Printf("Start encoding (%d files).\n", fileCnt)
	cnt := 0
	ret := ExitOK
	for _, file := range files {
		src := filepath.Join(opts.InputDir, file)
		destDir := filepath.Dir(filepath.Join(opts.OutputDir, file))
		if os.MkdirAll(destDir, 0755) != nil {
			fmt.Fprintf(os.Stderr, "Cannot create directory '%s'.\n", destDir)
			ret = ExitError
			continue
		}
		if err := convert(src, destDir); err != nil {
			// TODO Message
			fmt.Fprintf(os.Stderr, "Encode error: %s\n", src)
			fmt.Fprintf(os.Stderr, "    Error: %s\n", err)
			ret = ExitError
			continue
		}

		cnt++
		fmt.Printf("(%d/%d) Encoded '%s'\n", cnt, fileCnt, src)
	}

	return ret
}

func findWavFiles(root string) (paths []string, err error) {
	files, err := filepath.Glob(filepath.Join(root, "*", "*", "*.*"))
	if err != nil {
		return
	}

	for _, path := range files {
		filename := filepath.Base(path)
		if !filenameRegexp.MatchString(filename) {
			continue
		}

		relpath, err := filepath.Rel(root, path)
		if err != nil {
			continue
		}

		paths = append(paths, relpath)
	}

	return
}

func convert(srcFile, destDir string) error {
	params := []string{}
	if opts.QualityOption.Higher {
		params = append(params, "-h")
	}
	if opts.QualityOption.Fast {
		params = append(params, "-f")
	}
	if opts.QualityOption.Bitrate > 0 {
		params = append(params, "-b", strconv.Itoa(opts.QualityOption.Bitrate))
	}

	path, filename := filepath.Split(srcFile)
	path, album := filepath.Split(strings.TrimRight(path, `\/`))
	_, artist := filepath.Split(strings.TrimRight(path, `\/`))

	matches := filenameRegexp.FindAllStringSubmatch(filename, -1)
	trackNo, _ := strconv.Atoi(matches[0][1]) // Ignore error
	title := matches[0][3]

	params = append(params, "--tt", title)
	params = append(params, "--ta", artist)
	params = append(params, "--tl", album)
	params = append(params, "--tn", strconv.Itoa(trackNo))

	destFile := filepath.Join(destDir, fmt.Sprintf("%02d %s.mp3", trackNo, title))
	params = append(params, srcFile, destFile)

	return exec.Command("lame", params...).Run()
}
