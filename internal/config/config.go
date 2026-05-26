package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/poouo/VoiceCast/internal/validate"
	"github.com/poouo/VoiceCast/pkg/brand"
)

type Config struct {
	LastTargetIP   string `json:"last_target_ip"`
	LastTargetPort int    `json:"last_target_port"`
	ListenPort     int    `json:"listen_port"`
	SampleRate     int    `json:"sample_rate"`
	Channels       int    `json:"channels"`
	FrameMillis    int    `json:"frame_millis"`
	Codec          string `json:"codec"`
	AutoListen     bool   `json:"auto_listen"`
}

func Default() Config {
	return Config{
		LastTargetPort: brand.DefaultPort,
		ListenPort:     brand.DefaultPort,
		SampleRate:     brand.SampleRate,
		Channels:       brand.Channels,
		FrameMillis:    brand.FrameMillis,
		Codec:          brand.CodecPCM16LE,
	}
}

func Load() (Config, string, error) {
	path, err := Path()
	if err != nil {
		return Default(), "", err
	}
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, path, Save(cfg)
		}
		return cfg, path, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, path, err
	}
	Normalize(&cfg)
	return cfg, path, nil
}

func Save(cfg Config) error {
	Normalize(&cfg)
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, brand.AppName, "config.json"), nil
}

func Normalize(cfg *Config) {
	if cfg.LastTargetPort == 0 {
		cfg.LastTargetPort = brand.DefaultPort
	}
	if cfg.ListenPort == 0 {
		cfg.ListenPort = brand.DefaultPort
	}
	if cfg.SampleRate == 0 {
		cfg.SampleRate = brand.SampleRate
	}
	if cfg.Channels == 0 {
		cfg.Channels = brand.Channels
	}
	if cfg.FrameMillis == 0 {
		cfg.FrameMillis = brand.FrameMillis
	}
	if cfg.Codec == "" {
		cfg.Codec = brand.CodecPCM16LE
	}
	if cfg.LastTargetIP != "" && validate.Target(cfg.LastTargetIP, cfg.LastTargetPort) != nil {
		cfg.LastTargetIP = ""
	}
	if validate.ListenPort(cfg.ListenPort) != nil {
		cfg.ListenPort = brand.DefaultPort
	}
}
