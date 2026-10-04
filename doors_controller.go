package backend

import (
	"errors"
	"sync"

	"github.com/EpicCMoment/doors-backend/dxl"
	"github.com/EpicCMoment/doors-backend/module"
	"github.com/EpicCMoment/doors-backend/scripts"
)

// DoorsController is the single entry point of the backend. It wires the
// runner, DXL controller, and module controller together, and is created
// exactly once via Init.
type DoorsController struct {
	cfg Config

	runner  *Runner
	dxl     *dxl.DxlController
	modules *module.ModuleController
}

var (
	doorsOnce    sync.Once
	doorsDefault *DoorsController
	doorsErr     error
)

// Init creates the DoorsController singleton: renders/extracts the DXL
// server script, starts the DOORS process, and creates controllers.
func Init(cfg Config) (*DoorsController, error) {
	doorsOnce.Do(func() {
		doorsErr = nil
		d := &DoorsController{cfg: cfg}
		d.runner = NewRunner(cfg)
		if err := d.runner.Start(); err != nil {
			doorsErr = err
			return
		}
		d.dxl = dxl.NewDxlController(cfg.Host, cfg.Port)
		d.modules = module.NewModuleController(d.dxl)
		for _, name := range scripts.Names() {
			body, err := scripts.Load(name)
			if err != nil {
				doorsErr = err
				return
			}
			d.dxl.MustRegisterTemplate(name, body)
		}
		doorsDefault = d
	})
	return doorsDefault, doorsErr
}

// Default returns the singleton, or an error if Init has not succeeded.
func Default() (*DoorsController, error) {
	if doorsDefault == nil {
		return nil, errors.New("backend not initialized")
	}
	return doorsDefault, nil
}

// Modules exposes the module controller to callers.
func (d *DoorsController) Modules() *module.ModuleController { return d.modules }

// Dxl exposes the DXL controller for advanced use.
func (d *DoorsController) Dxl() *dxl.DxlController { return d.dxl }

// Runner exposes the underlying runner.
func (d *DoorsController) Runner() *Runner { return d.runner }

// Shutdown stops DOORS (graceful via the DXL server first) and tears down.
func (d *DoorsController) Shutdown() error {
	if d.dxl != nil {
		_ = d.dxl.Shutdown()
	}
	if d.runner != nil {
		return d.runner.Stop()
	}
	return nil
}
