package pcm

import "errors"

var ErrInvalidSampleEncoding error = errors.New("pcm: invalid sample encoding")
var ErrInvalidBitDepth error = errors.New("pcm: invalid bit depth")
