package fischl

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a4IcdKey            = "fischl-a4-icd"
	witchcraftAtkKey    = "fischl-witchcraft-atk%"
	witchcraftEmKey     = "fischl-witchcraft-em"
	witchcraftBonusCKey = "fischl-witch-c6bonus"
)

// A1 is not implemented:
// TODO: When Fischl hits Oz with a fully-charged Aimed Shot, Oz brings down Thundering Retribution, dealing AoE Electro DMG equal to 152.7% of the arrow's DMG.

// If your current active character triggers an Electro-related Elemental Reaction when Oz is on the field,
// the opponent shall be stricken with Thundering Retribution that deals Electro DMG equal to 80% of Fischl's ATK.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// Hyperbloom comes from a gadget so it doesn't ignore gadgets
	//nolint:unparam // ignoring for now, event refactor should get rid of bool return of event sub
	a4cb := func(args ...any) bool {
		ae := args[1].(*info.AttackEvent)

		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		// do nothing if oz not on field
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}
		active := c.Core.Player.ActiveChar()
		if active.StatusIsActive(a4IcdKey) {
			return false
		}
		active.AddStatus(a4IcdKey, 0.5*60, true)

		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Thundering Retribution (A4)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupFischl,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       0.8,
		}

		// A4 uses Oz Snapshot
		// TODO: this should target closest enemy within 15m of "elemental reaction position"
		c.Core.QueueAttackWithSnap(
			ai,
			c.ozSnapshot.Snapshot,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 0.5),
			4)
		return false
	}

	a4cbNoGadget := func(args ...any) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		return a4cb(args...)
	}

	c.Core.Events.Subscribe(event.OnOverload, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnElectroCharged, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSuperconduct, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSwirlElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnHyperbloom, a4cb, "fischl-a4")
	c.Core.Events.Subscribe(event.OnQuicken, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnAggravate, a4cbNoGadget, "fischl-a4")
}

// Witchcraft bonus:
// While Oz is on the field, if any ally causes Overload, Fischl and the active character gains 22.5% ATK.
// If any ally causes EC or LC, Fischl and the active character gains 90 EM.
func (c *char) witchcraftInit() {
	if !c.witchcraft {
		return
	}
	c.Core.Events.Subscribe(event.OnOverload, func(args ...any) bool {
		// do nothing if oz not on field
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.225
		if c.StatusIsActive(witchcraftBonusCKey) {
			m[attributes.ATKP] *= 2.0
		}

		// TODO: Does buff apply to all characters or just Fischl + active?
		// Fischl self buff
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(witchcraftAtkKey, 10*60),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		// If Fischl is not the active char, buff them as well
		if c.Core.Player.Active() != c.Index() {
			c.Core.Player.ActiveChar().AddStatMod(character.StatMod{
				Base:         modifier.NewBase(witchcraftAtkKey, 10*60),
				AffectedStat: attributes.ATKP,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}

		return false
	}, "fischl-witchcraft-atk%")

	c.Core.Events.Subscribe(event.OnElectroCharged, func(args ...any) bool {
		// do nothing if oz not on field
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 90
		if c.StatusIsActive(witchcraftBonusCKey) {
			m[attributes.EM] *= 2.0
		}

		// Fischl self buff
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(witchcraftEmKey, 10*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		// If Fischl is not the active char, buff them as well
		if c.Core.Player.Active() != c.Index() {
			c.Core.Player.ActiveChar().AddStatMod(character.StatMod{
				Base:         modifier.NewBase(witchcraftEmKey, 10*60),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}

		return false
	}, "fischl-witchcraft-em")
}
