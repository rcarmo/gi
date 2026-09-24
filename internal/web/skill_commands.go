package web

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/config"
)

const maxWebSkillBytes = 100 << 10

var webSkillName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

type loadedWebSkill struct {
	name, description, path string
	hash                    [32]byte
}

// Skills are an explicit startup capability, not arbitrary file-read commands.
// All sessions on this instance share this catalogue; no per-agent overrides
// are implied. Paths stay inside the native configured workspace root.
func readWebSkill(rootPath, relative string) ([]byte, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Stat(relative)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxWebSkillBytes {
		return nil, fmt.Errorf("skill must be a regular UTF-8 file up to 100 KiB")
	}
	file, err := openWebSkill(root, relative)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Validate the opened descriptor as well as the path: the entry may have
	// changed since Stat. A replacement FIFO must never block this request.
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxWebSkillBytes {
		return nil, fmt.Errorf("skill must be a regular UTF-8 file up to 100 KiB")
	}
	raw, err := io.ReadAll(io.LimitReader(file, maxWebSkillBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxWebSkillBytes || !utf8.Valid(raw) {
		return nil, fmt.Errorf("skill must be UTF-8 and at most 100 KiB")
	}
	return raw, nil
}
func loadWebSkills(cfg config.RuntimeConfig) map[string]loadedWebSkill {
	loaded := map[string]loadedWebSkill{}
	root, err := filepath.Abs(cfg.WorkspaceRoot)
	if err != nil {
		return loaded
	}
	for _, skill := range cfg.Discovery.Skills {
		name := strings.TrimSpace(skill.Name)
		key := strings.ToLower(name)
		if !webSkillName.MatchString(name) {
			continue
		}
		if _, exists := loaded[key]; exists {
			continue
		}
		path, err := filepath.Abs(skill.Path)
		if err != nil {
			continue
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		raw, err := readWebSkill(root, relative)
		if err != nil {
			continue
		}
		loaded[key] = loadedWebSkill{name: name, description: skill.Description, path: relative, hash: sha256.Sum256(raw)}
	}
	return loaded
}
func (s *Server) skillQuickActions() []map[string]string {
	keys := make([]string, 0, len(s.webSkills))
	for key := range s.webSkills {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	commands := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		skill := s.webSkills[key]
		commands = append(commands, map[string]string{"name": "/skill:" + skill.name, "description": skill.description, "source": "skill"})
	}
	return commands
}
func (s *Server) expandWebSkill(prompt string) (string, map[string]any, error) {
	fields := strings.Fields(prompt)
	if len(fields) == 0 || !strings.HasPrefix(strings.ToLower(fields[0]), "/skill:") {
		return prompt, nil, nil
	}
	key := strings.ToLower(strings.TrimPrefix(strings.ToLower(fields[0]), "/skill:"))
	skill, ok := s.webSkills[key]
	if !ok {
		return "", nil, fmt.Errorf("unknown or unavailable loaded skill: %s", key)
	}
	raw, err := readWebSkill(s.cfg.WorkspaceRoot, skill.path)
	if err != nil {
		return "", nil, fmt.Errorf("loaded skill unavailable; restore file and retry: %s", skill.name)
	}
	if sha256.Sum256(raw) != skill.hash {
		return "", nil, fmt.Errorf("loaded skill changed; restart Gi to load the new version: %s", skill.name)
	}
	args := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(prompt), fields[0]))
	expanded := fmt.Sprintf("Use the loaded workspace skill %s below.\n\n%s", skill.name, string(raw))
	if args != "" {
		expanded += "\n\nUser request:\n" + args
	}
	return expanded, map[string]any{"skill_name": skill.name, "skill_sha256": fmt.Sprintf("%x", skill.hash), "skill_command": prompt}, nil
}
