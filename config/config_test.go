package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func TestGetHomeRegion(t *testing.T) {
	falseVal := false
	trueVal := true
	regions := []identity.RegionSubscription{
		{RegionName: common.String("us-ashburn-1"), IsHomeRegion: &falseVal},
		{RegionName: common.String("us-phoenix-1"), IsHomeRegion: &trueVal},
	}

	if got := getHomeRegion(regions); got != "us-phoenix-1" {
		t.Fatalf("getHomeRegion()=%q, want %q", got, "us-phoenix-1")
	}
}

func TestGetHomeRegionNoMatch(t *testing.T) {
	falseVal := false
	regions := []identity.RegionSubscription{
		{RegionName: common.String("us-ashburn-1"), IsHomeRegion: &falseVal},
	}
	if got := getHomeRegion(regions); got != "" {
		t.Fatalf("getHomeRegion()=%q, want empty", got)
	}
}

func TestGetconfigReadsToolkitConfigYaml(t *testing.T) {
	dir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir error: %v", err)
	}

	content := []byte("configPath: ~/.oci/config\nprofileName: TEST\nuseinstanceprincipal: true\nSUPPORT_CSI_NUMBER: \"123\"\n")
	if err := os.WriteFile(filepath.Join(dir, "toolkit-config.yaml"), content, 0644); err != nil {
		t.Fatalf("write config error: %v", err)
	}

	cfg, err := Getconfig()
	if err != nil {
		t.Fatalf("Getconfig error: %v", err)
	}
	if cfg.ConfigPath != "~/.oci/config" || cfg.ProfileName != "TEST" || !cfg.UseInstancePrincipal || cfg.CSI != "123" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestGetconfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir error: %v", err)
	}

	if err := os.WriteFile("toolkit-config.yaml", []byte("configPath: ["), 0644); err != nil {
		t.Fatalf("write config error: %v", err)
	}

	_, err = Getconfig()
	if err == nil {
		t.Fatal("expected yaml parse error")
	}
}
