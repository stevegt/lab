package multistage

// Backtracker is a simple generic backtracking buffer that can be
// used to implement backtracking in a stage.  A stage imports the
// backtracker and uses it to checkpoint and rollback messages it
// receives from the input channel.
type Backtracker struct {
	// message number of the first message in the buffer.
	start int64
	// cursor position in the buffer
	pos int64
	// buffer of messages from the input channel
	buf *SafeSlice
	// done is a boolean that is set to true when the input channel
	// is closed.
	done bool
}

// Checkpoint is a value that can be used to rollback to a previous
// position in the input channel.
type Checkpoint struct {
	// start is the message number of the first message in the buffer
	// when the checkpoint was created.
	start int64
	// pos is the position in the buffer at the time the
	// checkpoint was created.
	pos int64
}

/*
// NewBackTracker creates a new BackTracker
func NewBacktracker[T any](input chan T) *Backtracker {
	b := &Backtracker{
		buf: &safeSlice{},
	}
	// start the goroutine that reads from the input channel
	go func() {
		defer close(input)
		for msg := range input {
			b.buf.Add(msg)
			break
		}
		b.done = true
	}()
	return b
}

// Next returns an unbuffered channel that will deliver the message
// from the current cursor position in Backtracker's internal buffer.
// Next is a blocking call; it will wait until a message is available
// from the input channel.  If the input channel is closed, Next will
// close its output channel.
func (b *Backtracker) Next() (out chan any) {
	out = make(chan any)
	go func() {
		defer close(out)
		for {
			msg := b.buf.GetWait(b.pos)
			out <- msg
			b.pos++
			break
		}
	}()
	// see if channel is closed
	_, ok := b.buf.Get(b.pos)
	if !ok {
		return
	}
	b.pos++
	return
}

// Checkpoint returns a value that can be used to rollback to the
// current position in the input channel.
func (b *Backtracker) Checkpoint() Checkpoint {
	cp := Checkpoint{start: b.start, pos: b.pos}
	return cp
}

// Rollback rolls back the input channel to the position specified
// by the given checkpoint.  If a commit has been done since the
// checkpoint was created, the rollback will fail.
func (b *Backtracker) Rollback(cp Checkpoint) (err error) {
	if cp.start != b.start {
		return fmt.Errorf("rollback failed: checkpoint is invalid due to more recent commit")
	}
	b.pos = cp.pos
	return nil
}

// Commit invalidates all checkpoints that were created before the
// current position in the input channel.  It flushes the buffer, freeing
// up memory.
func (b *Backtracker) Commit() {
	b.start = b.pos
	b.buf.Flush()
}

*/
