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

type music struct {
	artist string
	album  string
	title  string
	track  int
	path   string
}

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

	musics, err := findWavFiles(opts.InputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot find target files.\n")
		return ExitError
	}

	fileCnt := len(musics)
	fmt.Printf("Start encoding (%d files).\n", fileCnt)
	cnt := 0
	ret := ExitOK
	for _, music := range musics {
		if err := convert(music); err != nil {
			// TODO Message
			fmt.Fprintf(os.Stderr, "Encode error: %s\n", music.path)
			fmt.Fprintf(os.Stderr, "    Error: %s\n", err)
			ret = ExitError
			continue
		}

		cnt++
		fmt.Printf("(%d/%d) Encoded '%s'\n", cnt, fileCnt, music.path)
	}

	return ret
}

func findWavFiles(root string) (musics []music, err error) {
	files, err := filepath.Glob(filepath.Join(root, "*", "*", "*.*"))
	if err != nil {
		return
	}

	for _, fpath := range files {
		rest, filename := filepath.Split(fpath)
		rest, album := filepath.Split(strings.TrimRight(rest, `\/`))
		_, artist := filepath.Split(strings.TrimRight(rest, `\/`))

		fileparts := filenameRegexp.FindStringSubmatch(filename)
		if len(fileparts) == 0 {
			continue
		}

		relpath, err := filepath.Rel(root, fpath)
		if err != nil {
			continue
		}

		track, _ := strconv.Atoi(fileparts[1])
		musics = append(musics, music{
			artist: artist,
			album:  album,
			title:  fileparts[3],
			track:  track,
			path:   relpath,
		})
	}

	return
}

func convert(music music) error {
	srcFile := filepath.Join(opts.InputDir, music.path)
	destDir := filepath.Dir(filepath.Join(opts.OutputDir, music.path))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create directory '%s'.\n", destDir)
		return err
	}

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

	params = append(params, "--tt", music.title)
	params = append(params, "--ta", music.artist)
	params = append(params, "--tl", music.album)
	params = append(params, "--tn", strconv.Itoa(music.track))

	destFile := filepath.Join(destDir, fmt.Sprintf("%02d %s.mp3", music.track, music.title))
	params = append(params, srcFile, destFile)

	return exec.Command("lame", params...).Run()
}
