package config

import "testing"

func TestVNCConfigurationDefaultsDisabledAndRequiresExplicitDirectOptIn(t *testing.T) {
	t.Setenv("GI_WEB_VNC_TARGETS", "")
	t.Setenv("GI_WEB_VNC_ALLOW_DIRECT", "")
	root := t.TempDir()
	cfg := Load(root)
	if len(cfg.VNCTargets) != 0 || cfg.VNCAllowDirect {
		t.Fatal("VNC enabled by default")
	}
	t.Setenv("GI_WEB_VNC_TARGETS", `[{"id":"desk","label":"Desktop","host":"127.0.0.1","port":5900,"readOnly":true}]`)
	t.Setenv("GI_WEB_VNC_ALLOW_DIRECT", "1")
	cfg = Load(root)
	if cfg.VNCAllowDirect || len(cfg.VNCTargets) != 1 || cfg.VNCTargets[0].ID != "desk" || !cfg.VNCTargets[0].ReadOnly {
		t.Fatal(cfg.VNCTargets, cfg.VNCAllowDirect)
	}
	t.Setenv("GI_WEB_VNC_ALLOW_DIRECT", "true")
	if !Load(root).VNCAllowDirect {
		t.Fatal("explicit opt-in ignored")
	}
	t.Setenv("GI_WEB_VNC_TARGETS", "{invalid")
	if len(Load(root).VNCTargets) != 0 {
		t.Fatal("invalid targets accepted")
	}
}
