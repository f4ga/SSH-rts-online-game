package building

// Building represents a constructed building in the game.
type Building struct {
	ID         string
	PlayerID   string
	BlueprintID string
	Level      int
	Health     int
	Position   struct{ X, Y int }
}