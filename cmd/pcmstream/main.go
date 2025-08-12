package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/MatusOllah/resona/afmt"
	"github.com/MatusOllah/resona/aio"
	"github.com/MatusOllah/resona/codec"
	_ "github.com/MatusOllah/resona/codec/au"
	_ "github.com/MatusOllah/resona/codec/flac"
	_ "github.com/MatusOllah/resona/codec/mp3"
	_ "github.com/MatusOllah/resona/codec/oggvorbis"
	_ "github.com/MatusOllah/resona/codec/svx"
	_ "github.com/MatusOllah/resona/codec/wav"
	"github.com/MatusOllah/resona/encoding/pcm"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <audio file>\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	dec, name, err := codec.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding file: %v\n", err)
		os.Exit(1)
	}

	format := dec.Format()
	fmt.Fprintf(os.Stderr, "Format: %s, %v, %d channels\n", name, format.SampleRate, format.NumChannels)

	e := pcm.NewEncoder(os.Stdout, afmt.SampleFormat{
		BitDepth: 16,
		Encoding: afmt.SampleEncodingInt,
		Endian:   binary.LittleEndian,
	})
	if _, err := aio.Copy(e, dec); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing samples: %v\n", err)
		os.Exit(1)
	}
}
