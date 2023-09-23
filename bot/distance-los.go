package bot

import (
	"fmt"
	"math"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

type MapCellExt struct {
	mapObjects  swagger.DungeonsandtrollsMapObjects
	distance    int
	lineOfSight bool
}

func (b *Bot) calculateDistanceAndLineOfSight(level int32) map[swagger.DungeonsandtrollsPosition]MapCellExt {
	currentMap := b.GameState.Map_.Levels[level]

	// map for final map positions
	final := make(map[swagger.DungeonsandtrollsPosition]MapCellExt)
	// map for original map positions
	original := make(map[swagger.DungeonsandtrollsPosition]swagger.DungeonsandtrollsMapObjects)
	for _, objects := range currentMap.Objects {
		original[*coordsToPosition(*objects.Position)] = objects
	}
	b.Logger.Infow("Original map ->")
	for y := int32(0); y < currentMap.Height; y++ {
		row := ""
		for x := int32(0); x < currentMap.Width; x++ {
			cell, found := original[makePosition(x, y)]
			if !found {
				row += "!"
			} else if cell.IsFree {
				row += " "
			} else if cell.IsWall {
				row += "w"
			} else {
				row += "?"
			}
		}
		b.Logger.Infow("Map row: " + row)
	}

	// standard BFS stuff
	visited := make(map[swagger.DungeonsandtrollsPosition]bool)
	queue := []swagger.DungeonsandtrollsPosition{}

	currentPosition := *b.GameState.CurrentPosition
	// add current node to queue and add its distance to final map
	queue = append(queue, currentPosition)
	final[currentPosition] = MapCellExt{
		mapObjects:  original[currentPosition],
		distance:    0,
		lineOfSight: true,
	}
	b.Logger.Debugw("Setting distance and line of sight (current position)",
		"position", currentPosition,
		"distance", 0,
		"lineOfSight", true,
	)

	for len(queue) > 0 {
		b.Logger.Debugw("Queue",
			"queue", queue,
			"queueLength", len(queue),
		)
		node := queue[0]
		queue = queue[1:]

		nodeVisited, found := visited[node]
		if !found || !nodeVisited {
			b.Logger.Debugw("Visiting node",
				"position", node,
			)
			visited[node] = true

			// Enqueue all unvisited neighbors
			for _, neighbor := range getNeighbors(node) {
				b.Logger.Debugw("Checking neighbor",
					"position", neighbor,
				)
				mapObjects, found := original[neighbor]
				// must be in bounds
				// must not be visited
				// must be free
				if b.isInBounds(level, neighbor) && !visited[neighbor] && (!found || mapObjects.IsFree) {
					distance := final[node].distance + 1
					final[neighbor] = MapCellExt{
						mapObjects:  mapObjects,
						distance:    distance,
						lineOfSight: false, // TODO
					}
					b.Logger.Debugw("Setting distance and line of sight",
						"position", neighbor,
						"distance", distance,
						"lineOfSight", false,
					)
					queue = append(queue, neighbor)
				}
			}
		} else {
			b.Logger.Debugw("Node already visited",
				"position", node,
			)
		}
	}

	b.Logger.Infow("Final map with distances ->")
	for y := int32(0); y < currentMap.Height; y++ {
		row := ""
		for x := int32(0); x < currentMap.Width; x++ {
			cell, found := final[makePosition(x, y)]
			if !found {
				row += "~"
			} else if cell.distance < 10 {
				row += fmt.Sprintf("%d", cell.distance)
			} else {
				row += "+"
			}
		}
		b.Logger.Infow("Map row: " + row)
	}
	return final
}

func makePosition(x int32, y int32) swagger.DungeonsandtrollsPosition {
	return swagger.DungeonsandtrollsPosition{
		PositionX: x,
		PositionY: y,
	}
}

func getNeighbors(pos swagger.DungeonsandtrollsPosition) []swagger.DungeonsandtrollsPosition {
	return []swagger.DungeonsandtrollsPosition{
		makePosition(pos.PositionX-1, pos.PositionY),
		makePosition(pos.PositionX+1, pos.PositionY),
		makePosition(pos.PositionX, pos.PositionY-1),
		makePosition(pos.PositionX, pos.PositionY+1),
	}
}

// manhattan adjacency
func isAdjacent(pos1 swagger.DungeonsandtrollsPosition, pos2 swagger.DungeonsandtrollsPosition) bool {
	absX := math.Abs(float64(pos1.PositionX) - float64(pos2.PositionX))
	absY := math.Abs(float64(pos1.PositionY) - float64(pos2.PositionY))
	return (absX <= 1 && absY == 0) || (absX == 0 && absY <= 1)
}

func (b *Bot) isInBounds(level int32, pos swagger.DungeonsandtrollsPosition) bool {
	currentMap := b.GameState.Map_.Levels[level]
	return pos.PositionX >= 0 && pos.PositionX < currentMap.Width && pos.PositionY >= 0 && pos.PositionY < currentMap.Height
}
