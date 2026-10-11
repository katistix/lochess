// Package engine runs a UCI chess engine as a subprocess.
package engine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Engine is a running UCI engine. Calls are serialized; it is safe for concurrent use.
type Engine struct {
	mu    sync.Mutex
	cmd   *exec.Cmd
	stdin io.WriteCloser
	lines chan string // stdout, one line at a time; closed when the process exits
}

// Start launches the engine at path and completes the UCI handshake.
func Start(path string) (*Engine, error) {
	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start engine %s: %w", path, err)
	}
	e := &Engine{cmd: cmd, stdin: stdin, lines: make(chan string, 64)}
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 64*1024), 1024*1024) // `uci` option lists can be long
		for sc.Scan() {
			e.lines <- sc.Text()
		}
		close(e.lines)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.send("uci"); err != nil {
		e.Close()
		return nil, err
	}
	if _, err := e.waitFor(ctx, "uciok"); err != nil {
		e.Close()
		return nil, fmt.Errorf("uci handshake: %w", err)
	}
	return e, nil
}

// SetOption sends `setoption name <name> value <value>`.
func (e *Engine) SetOption(name string, value any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.send(fmt.Sprintf("setoption name %s value %v", name, value))
}

// NewGame resets engine state between games and waits until it is ready.
func (e *Engine) NewGame(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.send("ucinewgame"); err != nil {
		return err
	}
	return e.ready(ctx)
}

// Info is the result of a search. Scores are from the side to move's point of view.
type Info struct {
	BestMove string // UCI; empty when the side to move has no legal moves
	CP       int    // centipawns, valid when !HasMate
	Mate     int    // moves to mate when HasMate; negative if being mated, 0 if already mated
	HasMate  bool
}

// BestMove searches the position and returns the engine's move in UCI notation.
// goArgs is everything after "go", e.g. "movetime 50 depth 5".
func (e *Engine) BestMove(ctx context.Context, fen, goArgs string) (string, error) {
	info, err := e.Search(ctx, fen, goArgs)
	if err != nil {
		return "", err
	}
	if info.BestMove == "" {
		return "", errors.New("engine returned no move")
	}
	return info.BestMove, nil
}

// Search runs `go goArgs` on the position and returns the best move with
// the score from the last reported search depth.
func (e *Engine) Search(ctx context.Context, fen, goArgs string) (Info, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.ready(ctx); err != nil {
		return Info{}, err
	}
	if err := e.send("position fen " + fen); err != nil {
		return Info{}, err
	}
	if err := e.send("go " + goArgs); err != nil {
		return Info{}, err
	}
	var info Info
	for {
		select {
		case line, ok := <-e.lines:
			if !ok {
				return Info{}, errors.New("engine exited")
			}
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			switch fields[0] {
			case "info":
				parseScore(fields, &info)
			case "bestmove":
				if len(fields) > 1 && fields[1] != "(none)" {
					info.BestMove = fields[1]
				}
				return info, nil
			}
		case <-ctx.Done():
			e.send("stop") // best effort; the engine is unusable if it never answers
			return Info{}, fmt.Errorf("search: %w", ctx.Err())
		}
	}
}

// parseScore reads "score cp N" or "score mate N" from an info line, if present.
func parseScore(fields []string, info *Info) {
	for i := 0; i+2 < len(fields); i++ {
		if fields[i] != "score" {
			continue
		}
		n, err := strconv.Atoi(fields[i+2])
		if err != nil {
			return
		}
		switch fields[i+1] {
		case "cp":
			info.CP, info.Mate, info.HasMate = n, 0, false
		case "mate":
			info.CP, info.Mate, info.HasMate = 0, n, true
		}
		return
	}
}

// ErrNotFound means no Fairy-Stockfish binary could be located.
var ErrNotFound = errors.New("fairy-stockfish not found (install with: brew install fairy-stockfish)")

// Close asks the engine to quit and kills it if it does not exit promptly.
func (e *Engine) Close() error {
	e.send("quit")
	e.stdin.Close()
	done := make(chan error, 1)
	go func() { done <- e.cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		e.cmd.Process.Kill()
		return <-done
	}
}

func (e *Engine) ready(ctx context.Context) error {
	if err := e.send("isready"); err != nil {
		return err
	}
	_, err := e.waitFor(ctx, "readyok")
	return err
}

func (e *Engine) send(cmd string) error {
	_, err := io.WriteString(e.stdin, cmd+"\n")
	return err
}

// waitFor discards output until a line starting with prefix arrives.
func (e *Engine) waitFor(ctx context.Context, prefix string) (string, error) {
	for {
		select {
		case line, ok := <-e.lines:
			if !ok {
				return "", errors.New("engine exited")
			}
			if strings.HasPrefix(line, prefix) {
				return line, nil
			}
		case <-ctx.Done():
			return "", fmt.Errorf("waiting for %q: %w", prefix, ctx.Err())
		}
	}
}
