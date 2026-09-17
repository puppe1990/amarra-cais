package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
)

// Generated-file manifest (#169): generators record content hashes so destroy
// can warn before deleting files the user has since modified. Only whole-new
// file outputs are tracked; patched shared files (store.go, routes.go) change
// legitimately between generations and are excluded.

const generatedManifestRel = ".cais-generated.json"

func readGeneratedManifest(dir string) map[string]string {
	out := map[string]string{}
	raw, err := os.ReadFile(filepath.Join(dir, generatedManifestRel))
	if err != nil {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func writeGeneratedManifest(dir string, entries map[string]string) error {
	raw, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, generatedManifestRel), append(raw, '\n'), 0o644)
}

// recordGeneratedFiles hashes each existing rel path into the manifest,
// merging with entries from previous generations.
func recordGeneratedFiles(dir string, rels []string) error {
	entries := readGeneratedManifest(dir)
	for _, rel := range rels {
		sum, err := hashFile(filepath.Join(dir, rel))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		entries[rel] = sum
	}
	return writeGeneratedManifest(dir, entries)
}

// dropManifestEntries removes destroyed files from the manifest.
func dropManifestEntries(dir string, rels []string) error {
	entries := readGeneratedManifest(dir)
	changed := false
	for _, rel := range rels {
		if _, ok := entries[rel]; ok {
			delete(entries, rel)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return writeGeneratedManifest(dir, entries)
}

// manifestHas reports whether rel was recorded as generated.
func manifestHas(dir, rel string) bool {
	_, ok := readGeneratedManifest(dir)[rel]
	return ok
}

// fileDiffersFromManifest reports whether rel is not tracked or no longer
// matches the recorded hash. Untracked (and unreadable) files count as
// differing so destroy fails closed (#102).
func fileDiffersFromManifest(dir, rel string) bool {
	want, ok := readGeneratedManifest(dir)[rel]
	if !ok || want == "" {
		return true
	}
	got, err := hashFile(filepath.Join(dir, rel))
	if err != nil {
		return true
	}
	return got != want
}

// fileRels lists the rel paths of a generator's file map.
func fileRels(files map[string]string) []string {
	rels := make([]string, 0, len(files))
	for rel := range files {
		rels = append(rels, rel)
	}
	return rels
}

// recordScaffoldTree hashes every file written by `amarra-cais new` (#102).
// Generated apps start fully tracked, so destroy can tell scaffold files from
// hand-written ones without requiring --force.
func recordScaffoldTree(dir string) error {
	var rels []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == generatedManifestRel {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rels = append(rels, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return err
	}
	return recordGeneratedFiles(dir, rels)
}

func hashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
