package combat

import (
	"math"
	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/logger"
)

// DamageType represents the type of damage (piercing, blunt, siege, etc.)
type DamageType string

const (
	DamageTypeNormal DamageType = "normal"
	DamageTypePierce DamageType = "pierce"
	DamageTypeSiege  DamageType = "siege"
)

// AttackModifier defines multipliers against different armor types.
type AttackModifier struct {
	VsLight   float64
	VsMedium  float64
	VsHeavy   float64
	VsFortified float64
}

// armorModifiers maps damage type to armor type multipliers.
var armorModifiers = map[DamageType]AttackModifier{
	DamageTypeNormal: {VsLight: 1.0, VsMedium: 0.8, VsHeavy: 0.5, VsFortified: 0.3},
	DamageTypePierce: {VsLight: 1.5, VsMedium: 1.0, VsHeavy: 0.7, VsFortified: 0.4},
	DamageTypeSiege:  {VsLight: 0.5, VsMedium: 1.0, VsHeavy: 1.5, VsFortified: 2.0},
}

// ArmorType represents the armor class of a unit or building.
type ArmorType string

const (
	ArmorTypeLight     ArmorType = "light"
	ArmorTypeMedium    ArmorType = "medium"
	ArmorTypeHeavy     ArmorType = "heavy"
	ArmorTypeFortified ArmorType = "fortified"
)

// getUnitArmorType returns the armor type based on unit type.
func getUnitArmorType(unitType UnitType) ArmorType {
	switch unitType {
	case UnitTypeSwordsman:
		return ArmorTypeMedium
	case UnitTypeArcher:
		return ArmorTypeLight
	case UnitTypeCatapult:
		return ArmorTypeHeavy
	default:
		return ArmorTypeLight
	}
}

// getBuildingArmorType returns armor type for buildings (most are fortified).
func getBuildingArmorType(building *internal.Building) ArmorType {
	// Could be based on building type; for simplicity, assume fortified.
	return ArmorTypeFortified
}

// CalculateDamage computes the final damage after all modifiers.
func CalculateDamage(attacker *Unit, targetArmor ArmorType, targetDefense int, isBuilding bool) int {
	// Determine damage type based on attacker
	var dmgType DamageType
	switch attacker.Type {
	case UnitTypeArcher:
		dmgType = DamageTypePierce
	case UnitTypeCatapult:
		dmgType = DamageTypeSiege
	default:
		dmgType = DamageTypeNormal
	}

	mod := armorModifiers[dmgType]
	var multiplier float64
	switch targetArmor {
	case ArmorTypeLight:
		multiplier = mod.VsLight
	case ArmorTypeMedium:
		multiplier = mod.VsMedium
	case ArmorTypeHeavy:
		multiplier = mod.VsHeavy
	case ArmorTypeFortified:
		multiplier = mod.VsFortified
	default:
		multiplier = 1.0
	}

	// Base damage = attacker's attack + siege bonus if building
	base := attacker.Attack
	if isBuilding {
		base += attacker.Siege
	}

	// Apply defense reduction (each point of defense reduces damage by 1%)
	defenseFactor := 1.0 - float64(targetDefense)/100.0
	if defenseFactor < 0.1 {
		defenseFactor = 0.1
	}

	damage := float64(base) * multiplier * defenseFactor
	// Add small random variation ±10%
	variation := 0.9 + (0.2 * (float64(attacker.ID[0]) / 255.0)) // pseudo‑random
	damage *= variation

	// Minimum damage is 1
	result := int(math.Max(1, math.Round(damage)))
	logger.Get().Debug("damage calculated",
		"attacker", attacker.ID,
		"target_armor", targetArmor,
		"damage_type", dmgType,
		"multiplier", multiplier,
		"defense_factor", defenseFactor,
		"result", result,
	)
	return result
}

// PerformUnitAttack is a helper that applies damage to a target unit.
func PerformUnitAttack(attacker, target *Unit) (damage int, killed bool) {
	targetArmor := getUnitArmorType(target.Type)
	damage = CalculateDamage(attacker, targetArmor, target.Defense, false)
	target.Health -= damage
	if target.Health <= 0 {
		target.Health = 0
		killed = true
	}
	return damage, killed
}

// PerformBuildingAttack is a helper that applies damage to a building.
func PerformBuildingAttack(attacker *Unit, building *internal.Building) (damage int, destroyed bool) {
	targetArmor := getBuildingArmorType(building)
	// Building defense is stored as Health? For simplicity, we assume defense = 0.
	damage = CalculateDamage(attacker, targetArmor, 0, true)
	building.Health -= damage
	if building.Health <= 0 {
		building.Health = 0
		destroyed = true
	}
	return damage, destroyed
}