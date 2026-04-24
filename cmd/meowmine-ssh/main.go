// meowmine-ssh runs the game as an SSH server using charmbracelet/wish.
// Connect with: ssh -p 23234 <host>
//
// Each SSH user gets their own save file under ~/.meowmine/ssh_saves/<key>.json
// keyed by a SHA-256 of the client's public key. A connection without a
// pubkey plays with a session-only save that won't persist.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/ui"
)

func main() {
	host := flag.String("host", "0.0.0.0", "bind address")
	port := flag.Int("port", 23234, "SSH listen port")
	hostKey := flag.String("hostkey", "", "path to SSH host key (generated if missing)")
	flag.Parse()

	if *hostKey == "" {
		*hostKey = filepath.Join(sshSaveRoot(), "host_key")
	}

	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(*host, fmt.Sprintf("%d", *port))),
		wish.WithHostKeyPath(*hostKey),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("wish: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("meowmine-ssh listening on %s:%d", *host, *port)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-done
	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Printf("shutdown: %v", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	savePath := sessionSavePath(s)

	// Load or create state for this session.
	var state *game.State
	if b, err := os.ReadFile(savePath); err == nil {
		st, derr := game.LoadFrom(b)
		if derr == nil {
			state = st
		}
	}
	if state == nil {
		state = game.NewState(pickName(s))
	} else {
		// Offline catch-up.
		now := time.Now().Unix()
		if gap := now - state.LastTickUnix; gap > 8*3600 {
			state.LastTickUnix = now - 8*3600
			state.LastBillUnix = now - 8*3600
			state.LastWagesUnix = now - 8*3600
		}
		state.Tick(now)
	}

	return newSSHApp(state, savePath), []tea.ProgramOption{tea.WithAltScreen()}
}

// sshApp wraps ui.App so it persists to the per-session save path every few ticks.
type sshApp struct {
	inner    ui.App
	state    *game.State
	savePath string
	counter  int
}

func newSSHApp(s *game.State, savePath string) sshApp {
	inner := ui.NewApp(s)
	inner.SavePathOverride = savePath
	return sshApp{inner: inner, state: s, savePath: savePath}
}

func (m sshApp) Init() tea.Cmd { return m.inner.Init() }

func (m sshApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.inner.Update(msg)
	if n, ok := next.(ui.App); ok {
		m.inner = n
	}
	// Persist every ~10 updates (≈10 seconds of ticks) to the SSH-keyed path.
	m.counter++
	if m.counter%10 == 0 {
		_ = m.state.SaveAs(m.savePath)
	}
	return m, cmd
}

func (m sshApp) View() string { return m.inner.View() }

// sessionSavePath returns the path where this SSH session's save lives.
func sessionSavePath(s ssh.Session) string {
	root := sshSaveRoot()
	_ = os.MkdirAll(root, 0o755)
	key := s.PublicKey()
	var id string
	if key != nil {
		sum := sha256.Sum256(key.Marshal())
		id = hex.EncodeToString(sum[:])[:16]
	} else {
		// Anonymous user — use a per-session random ID that won't persist usefully.
		id = fmt.Sprintf("anon_%x", time.Now().UnixNano())
	}
	return filepath.Join(root, id+".json")
}

func sshSaveRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".meowmine_ssh"
	}
	return filepath.Join(home, ".meowmine", "ssh_saves")
}

func pickName(s ssh.Session) string {
	if u := s.User(); u != "" {
		return u
	}
	return "Anonymous"
}
