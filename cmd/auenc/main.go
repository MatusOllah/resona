package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MatusOllah/resona/aio"
	"github.com/MatusOllah/resona/codec"
	"github.com/MatusOllah/resona/codec/au"
	_ "github.com/MatusOllah/resona/codec/flac"
	_ "github.com/MatusOllah/resona/codec/mp3"
	_ "github.com/MatusOllah/resona/codec/oggvorbis"
	_ "github.com/MatusOllah/resona/codec/wav"
)

var (
	formatFlag = flag.String("format", "int16", "AU format (see codec/au/formats.go for options)")
	outputFlag = flag.String("output", "out.au", "Output file name")
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: auenc [options] <input file>\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	input := flag.Arg(0)
	output := *outputFlag

	var format uint32
	switch *formatFlag {
	case "ulaw":
		format = 1 // G.711 μ-law 8-bit
	case "int8":
		format = 2 // Linear PCM 8-bit integer
	case "int16":
		format = 3 // Linear PCM 16-bit integer
	case "int24":
		format = 4 // Linear PCM 24-bit integer
	case "int32":
		format = 5 // Linear PCM 32-bit integer
	case "float32":
		format = 6 // Linear PCM 32-bit float
	case "float64":
		format = 7 // Linear PCM 64-bit float
	case "alaw":
		format = 27 // G.711 A-law 8-bit
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *formatFlag)
		os.Exit(1)
	}

	inFile, err := os.Open(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer inFile.Close()

	dec, name, err := codec.Decode(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding input file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Input: %s, %v, %d channels\n", name, dec.Format().SampleRate, dec.Format().NumChannels)
	fmt.Fprintf(os.Stderr, "\t%s\n", input)

	outFile, err := os.Create(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	enc, err := au.NewEncoder(outFile, dec.Format(), format, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating encoder: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Output: %s\n", *formatFlag)
	fmt.Fprintf(os.Stderr, "\t%s\n\n", output)

	fmt.Fprintln(os.Stderr, "Processing...")
	if _, err := aio.Copy(enc, dec); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding: %v\n", err)
		os.Exit(1)
	}
	if err := enc.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing encoder: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Done!\n")
}
