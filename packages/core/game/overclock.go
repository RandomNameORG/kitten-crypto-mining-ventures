package game

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

// ocPercent is the human-facing earn boost for each OC level — used in log
// strings. Level 0 prints as 0 (toggling off), 1 as 25, 2 as 50.
var ocPercent = [3]int{0, 25, 50}

// CycleGPUOC rolls the given GPU's overclock level 0 → 1 → 2 → 0. Only works
// on running GPUs — broken/shipping/stolen cards can't be tweaked.
func (s *State) CycleGPUOC(instanceID int) error {
	for _, g := range s.GPUs {
		if g.InstanceID != instanceID {
			continue
		}
		if g.Status != "running" {
			return fmt.Errorf("GPU not running")
		}
		g.OCLevel = (g.OCLevel + 1) % 3
		name := g.DefID
		if g.BlueprintID != "" {
			if bp := s.BlueprintByID(g.BlueprintID); bp != nil {
				name = fmt.Sprintf("MEOWCore v%d", bp.Tier)
			}
		} else if def, ok := data.GPUByID(g.DefID); ok {
			name = def.LocalName()
		}
		s.appendLog("info", i18n.T("log.gpu.oc_set", name, ocPercent[g.OCLevel]))
		return nil
	}
	return fmt.Errorf("no such GPU")
}
