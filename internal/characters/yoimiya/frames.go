package yoimiya

import "github.com/genshinsim/gsim/pkg/core"

func (c *char) ActionFrames(a core.ActionType, p map[string]int) int {
	switch a {
	case core.ActionAttack:
		f := 0
		switch c.NormalCounter {
		//TODO: need to add atkspd mod
		case 0:
			f = 29
		case 1:
			f = 25
		case 2:
			f = 31
		case 3:
			f = 44
		case 4:
			f = 29
		}
		f = int(float64(f) / (1 + c.Stats[core.AtkSpd]))
		return f
	case core.ActionAim:
		return 94
	case core.ActionSkill:
		return 19 //should be 82
	case core.ActionBurst:
		return 129 //ok
	default:
		c.Log.Warnw("unknown action", "event", core.LogActionEvent, "frame", c.Sim.Frame(), "action", a)
		return 0
	}
}