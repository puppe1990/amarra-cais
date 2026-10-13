package devreload

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// PollDefault is the fingerprint interval: fast enough to feel instant on an
	// edit, slow enough that a big templates tree costs nothing.
	PollDefault = 400 * time.Millisecond
	// SettleDefault waits after the first change so a writer that saves twice in
	// a row (editor atomic rename, Tailwind rewrite) reports the settled file.
	SettleDefault = 150 * time.Millisecond
)

// Config tunes the poller. A zero value uses the defaults above.
type Config struct {
	Poll   time.Duration
	Settle time.Duration
}

type stamp struct {
	mod  int64
	size int64
}

// Watcher fingerprints files (or every file under a directory) and reports the
// first path that changed between two polls.
type Watcher struct {
	targets  []string
	onChange func(path string)
	prev     map[string]stamp
}

// New fingerprints the targets so a later Poll can tell what moved. Targets may
// be files or directories; a missing target is not an error — it starts watching
// when it appears.
func NewWatcher(targets []string, onChange func(path string)) (*Watcher, error) {
	prev, err := fingerprint(targets)
	if err != nil {
		return nil, err
	}
	return &Watcher{targets: targets, onChange: onChange, prev: prev}, nil
}

// Poll reports the first changed path, if any. Deterministic so tests and logs
// do not depend on map ordering.
func (w *Watcher) Poll() (string, bool) {
	next, err := fingerprint(w.targets)
	if err != nil {
		// A file rewritten mid-walk is normal while editing: keep the previous
		// fingerprint and try again on the next tick.
		return "", false
	}
	changed, ok := firstChange(w.prev, next)
	w.prev = next
	return changed, ok
}

// Run polls until ctx is done.
func (w *Watcher) Run(ctx context.Context, cfg Config) {
	poll := cfg.Poll
	if poll <= 0 {
		poll = PollDefault
	}
	settle := cfg.Settle
	if settle <= 0 {
		settle = SettleDefault
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		changed, ok := w.Poll()
		if !ok {
			continue
		}
		// Let the writer finish, then fold the settled state into the
		// fingerprint so the same edit does not fire twice.
		select {
		case <-ctx.Done():
			return
		case <-time.After(settle):
		}
		_, _ = w.Poll()
		w.onChange(changed)
	}
}

func firstChange(prev, next map[string]stamp) (string, bool) {
	paths := make([]string, 0, len(next))
	for path := range next {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if prev[path] != next[path] {
			return path, true
		}
	}
	for path := range prev {
		if _, ok := next[path]; !ok {
			return path, true
		}
	}
	return "", false
}

func fingerprint(targets []string) (map[string]stamp, error) {
	out := make(map[string]stamp, len(targets))
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		if !info.IsDir() {
			out[target] = stamp{mod: info.ModTime().UnixNano(), size: info.Size()}
			continue
		}
		err = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path != target && ignoredName(d.Name()) {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			out[path] = stamp{mod: info.ModTime().UnixNano(), size: info.Size()}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ignoredName skips editor and OS breadcrumbs (.git, .DS_Store, foo.html~,
// .#foo, swap files) — they are noise, not content.
func ignoredName(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasSuffix(name, "~") || strings.HasPrefix(name, "#")
}
