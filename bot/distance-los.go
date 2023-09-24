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

	// distance to obstacles used for line of sight
	distanceToFirstObstacle := make(map[float32]float32)
	// map for resulting map positions
	resultMap := make(map[swagger.DungeonsandtrollsPosition]MapCellExt)
	// map for original map positions
	for _, objects := range currentMap.Objects {
		resultMap[*coordsToPosition(*objects.Position)] = MapCellExt{
			mapObjects:  objects,
			distance:    math.MaxInt32,
			lineOfSight: false,
		}
	}
	b.Logger.Infow("Original map -> (player: A, no data / free: ' ', wall: w, spawn: *, stairs: s, unknown: ?)")
	for y := int32(0); y < currentMap.Height; y++ {
		row := ""
		for x := int32(0); x < currentMap.Width; x++ {
			cell, found := resultMap[makePosition(x, y)]
			if makePosition(x, y) == *b.GameState.CurrentPosition {
				row += "A"
			} else if !found {
				row += " "
			} else if cell.mapObjects.IsSpawn {
				row += "*"
			} else if cell.mapObjects.IsStairs {
				row += "s"
			} else if cell.mapObjects.IsFree {
				row += " "
			} else if cell.mapObjects.IsWall {
				row += "w"
			} else {
				row += "?"
			}
		}
		b.Logger.Infof("Map row: %s (y = %d)", row, y)
	}

	// standard BFS stuff
	visited := make(map[swagger.DungeonsandtrollsPosition]bool)
	queue := []swagger.DungeonsandtrollsPosition{}

	currentPosition := *b.GameState.CurrentPosition
	// add current node to queue and add its distance to final map
	queue = append(queue, currentPosition)
	cell, found := resultMap[currentPosition]
	mapObjects := swagger.DungeonsandtrollsMapObjects{}
	if found {
		mapObjects = cell.mapObjects
	}
	resultMap[currentPosition] = MapCellExt{
		mapObjects:  mapObjects,
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
				cell, found := resultMap[neighbor]
				// must be in bounds
				// must not be visited
				// must be free
				if b.isInBounds(level, neighbor) && !visited[neighbor] && (!found || cell.mapObjects.IsFree) {
					mapObjects := swagger.DungeonsandtrollsMapObjects{
						IsFree: true,
					}
					if found {
						mapObjects = cell.mapObjects
					}
					distance := resultMap[node].distance + 1
					lineOfSight := b.getLoS(level, resultMap, distanceToFirstObstacle, currentPosition, neighbor)
					resultMap[neighbor] = MapCellExt{
						mapObjects:  mapObjects,
						distance:    distance,
						lineOfSight: lineOfSight,
					}
					b.Logger.Debugw("Setting distance and line of sight",
						"position", neighbor,
						"distance", distance,
						"lineOfSight", lineOfSight,
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

	b.Logger.Infow("Map with distances -> (player: A, no data: !, not reachable: ~, distance < 10: 0-9, distance >= 10: +)")
	for y := int32(0); y < currentMap.Height; y++ {
		row := ""
		for x := int32(0); x < currentMap.Width; x++ {
			cell, found := resultMap[makePosition(x, y)]
			if makePosition(x, y) == currentPosition {
				row += "A"
			} else if !found {
				row += "!"
			} else if cell.distance < 10 {
				row += fmt.Sprintf("%d", cell.distance)
			} else if cell.distance == math.MaxInt32 {
				row += "~"
			} else {
				row += "+"
			}
		}
		b.Logger.Infof("Map row: %s (y = %d)", row, y)
	}
	b.Logger.Infow("Map with line of sight -> (player: A, no data: !, line of sight: ' ', wall: w, no line of sight: ~)")
	for y := int32(0); y < currentMap.Height; y++ {
		row := ""
		for x := int32(0); x < currentMap.Width; x++ {
			cell, found := resultMap[makePosition(x, y)]
			if makePosition(x, y) == currentPosition {
				row += "A"
			} else if !found {
				row += "!"
			} else if cell.lineOfSight {
				row += " "
			} else if cell.mapObjects.IsWall {
				row += "w"
			} else {
				row += "~"
			}
		}
		b.Logger.Infof("Map row: %s (y = %d)", row, y)
	}
	return resultMap
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

func (b *Bot) getLoS(level int32, resultMap map[swagger.DungeonsandtrollsPosition]MapCellExt, distanceToFirstObstacle map[float32]float32, pos1 swagger.DungeonsandtrollsPosition, pos2 swagger.DungeonsandtrollsPosition) bool {
	// get the center of the cell
	x1 := float32(pos1.PositionX) + 0.5
	y1 := float32(pos1.PositionY) + 0.5
	x2 := float32(pos2.PositionX) + 0.5
	y2 := float32(pos2.PositionY) + 0.5

	// slope := float32(y2-y1) / float32(x2-x1)
	distance := math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)))
	// angle in radians
	slope := float32(math.Atan2(float64(y2-y1), float64(x2-x1)))
	// angleDegrees := angleRadians * 180 / math.Pi

	// TODO: somehow round the value to prevent cache misses
	losDist, found := distanceToFirstObstacle[slope]
	if found {
		b.Logger.Infow("LoS: found in cache",
			"playerPosition", pos1,
			"position", pos2,
			"slope", slope,
			"distance", distance,
			"lineOfSightDistance", losDist,
			"lineOfSight", distance < float64(losDist),
		)
		return distance < float64(losDist)
	}
	losDist = b.rayTrace(level, resultMap, slope, x1, y1, x2, y2)
	distanceToFirstObstacle[slope] = losDist
	b.Logger.Infow("LoS: calculated",
		"playerPosition", pos1,
		"position", pos2,
		"slope", slope,
		"distance", distance,
		"lineOfSightDistance", losDist,
		"lineOfSight", distance < float64(losDist),
	)
	return distance < float64(losDist)
}

func (b *Bot) rayTrace(level int32, resultMap map[swagger.DungeonsandtrollsPosition]MapCellExt, slope float32, x1 float32, y1 float32, x2 float32, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1

	// Calculate absolute values of dx and dy
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	// Determine the sign of movement along x and y
	sx := float32(1)
	sy := float32(1)
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}

	// Initialize error variables
	e := dx - dy
	x := x1
	y := y1

	for {
		// TODO: any mapping needed here?
		pos := getPositionsForFloatCoords(x, y)
		// Check the current cell for obstacles or objects
		cell, found := resultMap[pos]

		// obstacle hit if end of map OR not free
		if !b.isInBounds(level, pos) || (found && !cell.mapObjects.IsFree) {
			dist := math.Sqrt(float64((x-x1)*(x-x1) + (y-y1)*(y-y1)))
			return float32(dist)
		}

		// Calculate the next step
		e2 := 2 * e
		if e2 > -dy {
			e -= dy
			x += sx
		}
		if e2 < dx {
			e += dx
			y += sy
		}
	}
}

func getPositionsForFloatCoords(x float32, y float32) swagger.DungeonsandtrollsPosition {
	return swagger.DungeonsandtrollsPosition{
		PositionX: int32(x),
		PositionY: int32(y),
	}
}
