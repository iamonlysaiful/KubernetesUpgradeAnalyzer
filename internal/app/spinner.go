package app

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// spinner provides a simple animated progress indicator.
type spinner struct {
	mu      sync.Mutex
	w       io.Writer
	msg     string
	running bool
	done    chan struct{}
	frames  []string
}

// newSpinner creates a spinner that writes to w.
func newSpinner(w io.Writer) *spinner {
	return &spinner{
		w:      w,
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start begins the spinner with the given message.
// If already running, it updates the message.
func (s *spinner) Start(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		// Update message in place
		s.msg = msg
		return
	}

	s.msg = msg
	s.running = true
	s.done = make(chan struct{})

	go s.animate()
}

// Stop halts the spinner and shows a completion mark.
func (s *spinner) Stop(success bool) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.done)
	msg := s.msg
	s.mu.Unlock()

	// Clear the line and show final status
	mark := "✓"
	if !success {
		mark = "✗"
	}
	fmt.Fprintf(s.w, "\r  %s %s\n", mark, msg)
}

// StopWithNext stops the current spinner and immediately starts a new one.
func (s *spinner) StopWithNext(nextMsg string) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		s.Start(nextMsg)
		return
	}
	// Stop current
	s.running = false
	close(s.done)
	prevMsg := s.msg
	s.mu.Unlock()

	// Show completion
	fmt.Fprintf(s.w, "\r  ✓ %s\n", prevMsg)

	// Start next
	s.Start(nextMsg)
}

func (s *spinner) animate() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			if s.running {
				frame := s.frames[i%len(s.frames)]
				fmt.Fprintf(s.w, "\r  %s %s", frame, s.msg)
				i++
			}
			s.mu.Unlock()
		}
	}
}

// progressSpinner wraps a spinner to work with the Progress callback.
type progressSpinner struct {
	s          *spinner
	lastMsg    string
	hasStarted bool
}

func newProgressSpinner(w io.Writer) *progressSpinner {
	return &progressSpinner{s: newSpinner(w)}
}

func (p *progressSpinner) Update(msg string) {
	if p.hasStarted {
		p.s.StopWithNext(msg)
	} else {
		p.s.Start(msg)
		p.hasStarted = true
	}
	p.lastMsg = msg
}

func (p *progressSpinner) Finish(success bool) {
	if p.hasStarted {
		p.s.Stop(success)
	}
}
