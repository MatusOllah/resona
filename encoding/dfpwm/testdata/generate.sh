#!/usr/bin/bash
set -euo pipefail

# Generate golden test data for decoder tests.

rm -f sine.wav sine.dfpwm sine.pcm

# generate 440Hz sine wave, 1 second @ 48kHz
ffmpeg -f lavfi -i "sine=frequency=440:sample_rate=48000:duration=1" -c:a pcm_f32le -f wav -ac 1 ./sine.wav

ffmpeg -i ./sine.wav -c:a dfpwm -f dfpwm ./sine.dfpwm

ffmpeg -i ./sine.dfpwm -ar 48000 -ac 1 -f f32le -acodec pcm_f32le ./sine.pcm

rm sine.wav # remove temporary file
