package effect_test

import (
	"fmt"

	"github.com/MatusOllah/resona/aio"
	"github.com/MatusOllah/resona/audio"
	"github.com/MatusOllah/resona/effect"
)

func Example() {
	r := audio.NewReader([]float32{0.1, 0.2, 0.3})

	gain := effect.NewGain(1.0)
	effectReader := effect.Reader(r, gain)

	samples, err := aio.ReadAll(effectReader)
	if err != nil {
		panic(err)
	}

	fmt.Println(samples)
	// Output:
	// [0.2 0.4 0.6]
}
