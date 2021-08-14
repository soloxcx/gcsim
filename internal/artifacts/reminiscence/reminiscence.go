package reminiscence

import (
	"github.com/genshinsim/gsim/pkg/combat"
	"github.com/genshinsim/gsim/pkg/core"
)

func init() {
	combat.RegisterSetFunc("reminiscence of shime", New)
}

func New(c core.Character, s core.Sim, log core.Logger, count int) {
	if count >= 2 {
		m := make([]float64, core.EndStatType)
		m[core.ATKP] = 0.18
		c.AddMod(core.CharStatMod{
			Key: "rem-2pc",
			Amount: func(a core.AttackTag) ([]float64, bool) {
				return m, true
			},
			Expiry: -1,
		})
	}
	if count >= 4 {
		m := make([]float64, core.EndStatType)
		m[core.DmgP] = 0.50
		s.AddEventHook(func(s core.Sim) bool {
			if s.ActiveCharIndex() != c.CharIndex() {
				return false
			}
			if c.CurrentEnergy() > 15 {
				//consume 15 energy, increased normal/charge/plunge dmg by 50%
				c.AddEnergy(-15)
				c.AddMod(core.CharStatMod{
					Key: "rem-4pc",
					Amount: func(ds core.AttackTag) ([]float64, bool) {
						if ds != core.AttackTagNormal && ds != core.AttackTagExtra && ds != core.AttackTagPlunge {
							return nil, false
						}
						return m, true
					},
					Expiry: s.Frame() + 600,
				})

			}
			return false
		}, "rem-4pc", core.PreSkillHook)

	}
	//add flat stat to char
}