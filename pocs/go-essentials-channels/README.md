# go-essentials-channels

Channel patterns: a generator, a pipeline stage, fan-in with `select`, closing a channel to signal completion, and `time.After` for timeouts.

### How it works

`src/main.go` builds a `generator` that emits numbers on a channel, chains a `square` stage into a pipeline, merges two channels with a `select` loop that nils out closed channels (fan-in), and races a slow channel against `time.After` to produce a timeout.

### Run

```bash
./run.sh
```

### Output

```
pipeline gen->square: [1 4 9 16]
fanIn: [1 2 3 4 5 6]
withTimeout: timeout
```
