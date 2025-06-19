package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/MatusOllah/resona/codec"
	_ "github.com/MatusOllah/resona/codec/au"
	_ "github.com/MatusOllah/resona/codec/flac"
	_ "github.com/MatusOllah/resona/codec/mp3"
	_ "github.com/MatusOllah/resona/codec/oggvorbis"
	_ "github.com/MatusOllah/resona/codec/wav"
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

	const bufSize = 4096
	buf := make([]float64, bufSize)

	for {
		n, err := dec.ReadSamples(buf)
		if n > 0 {
			out := make([]byte, n*2)
			for i := range n {
				s := int16(buf[i] * 32767)
				binary.LittleEndian.PutUint16(out[i*2:], uint16(s))
			}
			if _, err := os.Stdout.Write(out); err != nil {
				panic(err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading samples: %v\n", err)
			os.Exit(1)
		}
	}
}
