package economy

// ResourceType represents a type of game resource.
type ResourceType string

const (
	ResourceGold    ResourceType = "gold"
	ResourceWood    ResourceType = "wood"
	ResourceStone   ResourceType = "stone"
	ResourceFood    ResourceType = "food"
	ResourceIron    ResourceType = "iron"
	ResourceCoal    ResourceType = "coal"
	ResourceCrystal ResourceType = "crystal"
)

// ResourceSet is a collection of resources with amounts.
type ResourceSet map[ResourceType]int

// Add adds another resource set to this one.
func (rs ResourceSet) Add(other ResourceSet) {
	for typ, amount := range other {
		rs[typ] += amount
	}
}

// Sub subtracts another resource set, returns false if any would go negative.
func (rs ResourceSet) Sub(other ResourceSet) bool {
	// First check that we have enough
	for typ, amount := range other {
		if rs[typ] < amount {
			return false
		}
	}
	// Then subtract
	for typ, amount := range other {
		rs[typ] -= amount
	}
	return true
}

// Total returns the total quantity of all resources.
func (rs ResourceSet) Total() int {
	sum := 0
	for _, amount := range rs {
		sum += amount
	}
	return sum
}

// Clone creates a copy of the resource set.
func (rs ResourceSet) Clone() ResourceSet {
	clone := make(ResourceSet)
	for typ, amount := range rs {
		clone[typ] = amount
	}
	return clone
}