package hub

import (
	"github.com/ClashrAuto/Clashr/hub/executor"
	"github.com/ClashrAuto/Clashr/hub/route"
)

// Parse call at the beginning of clash
func Parse() error {
	cfg, err := executor.Parse()
	if err != nil {
		return err
	}

	if cfg.General.ExternalUI != "" {
		route.SetUIPath(cfg.General.ExternalUI)
	}

	if cfg.General.ExternalController != "" {
		go route.Start(cfg.General.ExternalController, cfg.General.Secret)
	}

	executor.ApplyConfig(cfg, true)
	return nil
}
