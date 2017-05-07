package golame

import (
	"github.com/jessevdk/go-flags"
	"os"
	"path/filepath"
)

// Option : Top level options
type Option struct {
	Version       bool           `short:"v" long:"version" description:"Show version"`
	Input         func(string)   `short:"i" long:"input" description:"Input root directory" default-mask:"./src"`
	Output        func(string)   `short:"o" long:"output" description:"Output root directory" default-mask:"./dest"`
	QualityOption *QualityOption `group:"Quality Options"`

	InputDir  string // Input root directory. Not user option.
	OutputDir string // Output root directory. Not user option.
}

// QualityOption : Quality options
type QualityOption struct {
	Bitrate  int  `short:"b" long:"bitrate" description:"Set the bitrate." value-name:"BITRATE" default-mask:"128 [kbps]"`
	Higher   bool `short:"H" long:"higher" description:"Higher quality, but a little slower. Recommended." default-mask:"ON"`
	Fast     bool `short:"f" long:"fast" description:"Fast mode (lower quality)" optional:"a"`
	Priority int  `short:"p" long:"priority" description:"Sets the process priority.\n0,1 = Low priority\n2   = Normal primarity\n3,4 = High priority" value-name:"PRIORITY"`
}

func newOptionParser(opts *Option) *flags.Parser {
	output := flags.NewNamedParser("golame", flags.Default)
	output.AddGroup("Quality Options", "", &QualityOption{})

	currentDir, err := os.Getwd()
	if err == nil {
		opts.InputDir = filepath.Join(currentDir, "src")
		opts.OutputDir = filepath.Join(currentDir, "dest")
	}

	opts.Input = func(path string) {
		if p, err := os.Stat(path); err != nil || !p.IsDir() {
			opts.InputDir = ""
			return
		}

		opts.InputDir = path
	}

	opts.Output = func(path string) {
		if p, err := os.Stat(path); err == nil && !p.IsDir() {
			opts.OutputDir = ""
			return
		}

		opts.OutputDir = path
	}

	opts.QualityOption = &QualityOption{}
	opts.QualityOption.Bitrate = -1

	parser := flags.NewParser(opts, flags.Default)
	parser.Name = "golame"
	parser.Usage = "[options]"
	return parser
}
