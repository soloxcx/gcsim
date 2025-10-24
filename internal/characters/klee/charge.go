package klee

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	chargeHitmark  = 76
	chargeSnapshot = 29 + 32
)

var (
	boomboosterMult = []float64{1.15, 1.3, 1.5}
	chargeFrames    []int
)

func init() {
	chargeFrames = frames.InitAbilSlice(113)
	chargeFrames[action.ActionAttack] = 59
	chargeFrames[action.ActionCharge] = 59
	chargeFrames[action.ActionSkill] = 59
	chargeFrames[action.ActionBurst] = 59
	chargeFrames[action.ActionDash] = 31
	chargeFrames[action.ActionJump] = 30
	chargeFrames[action.ActionSwap] = 104
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}

	windup := 0
	switch c.Core.Player.CurrentState() {
	case action.NormalAttackState:
		if c.NormalCounter == 1 || c.NormalCounter == 2 {
			windup = 14
		}
	case action.SkillState:
		windup = 14
	}

	c.QueueChargedAttack(windup, travel, false)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeFrames[action.ActionJump] - windup, // earliest cancel
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) QueueChargedAttack(windup int, travel int, coord bool) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Charge",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   180,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// TODO: delay?
	if coord {
		ai.Abil = "Coordinated Charge Attack: Blast"
		c.Core.Log.NewEvent("coordinated CA triggered", glog.LogCharacterEvent, c.Index())
	}

	c.Core.Tasks.Add(func() {
		// apply and clear spark
		snap := c.Snapshot(&ai)
		count := 0
		if c.StatusIsActive(boomboosterNormalKey) {
			count++
		}
		if c.StatusIsActive(boomboosterSkillKey) {
			count++
		}
		if c.StatusIsActive(boomboosterBurstKey) {
			count++
		}
		if c.StatusIsActive(a1SparkKey) {
			snap.Stats[attributes.DmgP] += .50
			// C6 Witchcraft bonus:
			// When Klee uses an Explosive Spark, there is a 50% chance it will not be consumed.
			previous := c.a1CurrentStack
			if c.Base.Cons < 6 || c.witchcraft && c.Core.Rand.Float64() < 0.5 {
				c.a1CurrentStack--
			}
			if c.a1CurrentStack == 0 {
				c.DeleteStatus(a1SparkKey)
			}
			if count > 0 {
				ai.Mult *= boomboosterMult[count-1]
				c.Core.Log.NewEvent("applying boombooster", glog.LogCharacterEvent, c.Index()).
					Write("stacks", count).
					Write("multiplier", boomboosterMult[count-1])
			}
			c.Core.Log.NewEvent("consuming spark", glog.LogCharacterEvent, c.Index()).
				Write("previous", previous).
				Write("new", c.a1CurrentStack)
		}

		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 3),
			(chargeHitmark-chargeSnapshot)+travel,
			c.makeA4CB(),
		)
	}, chargeSnapshot-windup)

	c.c1(chargeHitmark - windup + travel)
}
