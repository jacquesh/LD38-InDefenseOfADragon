package main

import (
    "bytes"
    "fmt"
    "image"
    _ "image/png"
    "image/color"
    "log"
    "math"
    "os"

    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
    screenWidth  = 320
    screenHeight = 240
    aspectRatio = float64(screenWidth)/float64(screenHeight)
    deltaTime = 1.0/60.0

    pathSegmentLength = 25.0
    towerSize = 15.0
    towerAttackRange = 25.0
    projectileSize = 4.0
    enemySize = 10.0
)

var (
    camera Rect

    pixelImg *ebiten.Image
    circleImg *ebiten.Image
    backgroundImg *ebiten.Image
    towerImg [3]*ebiten.Image
    towerCanBuildImg *ebiten.Image
    towerNoCanBuildImg *ebiten.Image
    projectileImg *ebiten.Image
    pathSegmentImg *ebiten.Image
    pathCornerImg *ebiten.Image
    pathStartImg *ebiten.Image
    pathEndImg *ebiten.Image
    enemyImg [6]*ebiten.Image

    pathBoundingBox Rect
    pathEndDirection Vec2
    pathEndLocation Vec2
    pathEndIndex int

    waypoints []Vec2
    targetWaypointCount int
    waypointSpawnInterval float64
    timeTillNewWaypoint float64
    waypointsReady bool

    enemies []*Enemy
    towers []*Tower
    projectiles []*Projectile

    enemySpeed float64
    projectileSpeed float64

    enemyHealth int
    enemiesPerWave int
    enemySpawnInterval float64
    enemyBounty int

    lives int
    credits int
    currentWave int
    waveInProgress bool
    waveEnemiesRemaining int
    timeTillEnemySpawn float64

    blackoutOpacity float64

    ghostTowerVisible bool
    ghostTower *Tower

    keyWasDown [ebiten.KeyMax]bool
    mousePressed [3]bool
)

func screen2WorldLoc(screenLoc Vec2) Vec2 {
    result := screenLoc
    result.x *= camera.size.x/screenWidth
    result.y *= camera.size.y/screenHeight
    result = result.Add(camera.MinXY())
    return result
}

func world2ScreenLoc(worldLoc Vec2) Vec2 {
    result := worldLoc
    result = result.Sub(camera.MinXY())
    result.x *= screenWidth/camera.size.x;
    result.y *= screenHeight/camera.size.y;
    return result
}

func startRound() {
    currentWave++
    enemiesPerWave += currentWave
    enemySpawnInterval = 10.0/float64(enemiesPerWave)
    enemySpeed *= 1.8
    projectileSpeed = enemySpeed*3.0
    ghostTower.cost = int(1.5 * float64(ghostTower.cost))

    if currentWave%2 == 1 {
        enemyHealth += 1
    }
    if currentWave%4 == 1 {
        enemyBounty += 1
    }

    waveEnemiesRemaining = enemiesPerWave
    timeTillEnemySpawn = 0.0
    waveInProgress = true
}

func endRound() {
    waveInProgress = false
    credits += currentWave

    targetWaypointCount = int(float64(len(waypoints))*1.6)
    newWaypointCount := targetWaypointCount - len(waypoints)
    waypointSpawnInterval = 2.0/float64(newWaypointCount)
    timeTillNewWaypoint = 0.0
    waypointsReady = false
}

func sendEnemy() {
    newEnemy := &Enemy {
        health: enemyHealth,
        currentWaypoint: 0,
        position: waypoints[0],
    }
    enemies = append(enemies, newEnemy)
}

func addTower(loc Vec2) {
    newTower := &Tower {
        position: loc,
        scale: ghostTower.scale,
    }
    towers = append(towers, newTower)
}

func createProjectile(source *Tower, target *Enemy) {
    newProjectile := &Projectile {
        position: source.position,
        scale: source.scale,
        target: target,
        damage: 1,
    }
    projectiles = append(projectiles, newProjectile)
}


func resizeCameraToContainRect(r Rect) {
    minSize := pathBoundingBox.size.Add(Vec2{ 50, 50 })
    xScaleFactor := minSize.x/camera.size.x
    yScaleFactor := minSize.y/camera.size.y
    scaleFactor := math.Max(xScaleFactor, yScaleFactor)
    if currentWave > 0 {
        // NOTE: We only scale after the first round so that we can just set the size to be what we
        //       want it to look like at the beginning, and then it'll scale from there
        ghostTower.scale *= scaleFactor
    }

    camera.position = pathBoundingBox.position
    camera.size = camera.size.Mul(scaleFactor)
}

func addPathSegment() {
    pathEndLocation = pathEndLocation.Add(pathEndDirection.Mul(pathSegmentLength))
    waypoints = append(waypoints, pathEndLocation)

    // NOTE: Bit hackery to check the turn direction, from https://rosettacode.org/wiki/Dragon_curve
    turnLowMask := pathEndIndex^(pathEndIndex-1)
    turnCheckBit := pathEndIndex & (turnLowMask+1)
    shouldTurnCCW := (turnCheckBit != 0)
    if(!shouldTurnCCW) {
        pathEndDirection = pathEndDirection.Rotate90CCW()
    } else {
        pathEndDirection = pathEndDirection.Rotate90CW()
    }
    pathEndIndex++

    if !pathBoundingBox.ContainsPoint(pathEndLocation) {
        minX := math.Min(pathEndLocation.x, pathBoundingBox.MinX())
        maxX := math.Max(pathEndLocation.x, pathBoundingBox.MaxX())
        minY := math.Min(pathEndLocation.y, pathBoundingBox.MinY())
        maxY := math.Max(pathEndLocation.y, pathBoundingBox.MaxY())

        pathBoundingBox.position.x = (minX+maxX)/2.0
        pathBoundingBox.position.y = (minY+maxY)/2.0
        pathBoundingBox.size.x = maxX-minX
        pathBoundingBox.size.y = maxY-minY

        resizeCameraToContainRect(pathBoundingBox)
    }
}

func update(screen *ebiten.Image) error {
    if ebiten.IsKeyPressed(ebiten.KeyEscape) {
        os.Exit(0)
    }

    screen.Fill(color.Black)

    sPressed := ebiten.IsKeyPressed(ebiten.KeyS)
    if sPressed && !keyWasDown[ebiten.KeyS] {
        if (len(enemies) == 0) && (waveEnemiesRemaining == 0) && (lives > 0) && waypointsReady {
            startRound()
        }
    }
    keyWasDown[ebiten.KeyS] = sPressed

    gPressed := ebiten.IsKeyPressed(ebiten.KeyG)
    if gPressed && !keyWasDown[ebiten.KeyG] {
        ghostTowerVisible = !ghostTowerVisible
    }
    keyWasDown[ebiten.KeyG] = gPressed

    rPressed := ebiten.IsKeyPressed(ebiten.KeyR)
    if rPressed && (lives == 0) {
        reset()
    }

    mouseX, mouseY := ebiten.CursorPosition()
    mouseScreenLoc := Vec2 { float64(mouseX), float64(mouseY) }
    mouseWorldLoc := screen2WorldLoc(mouseScreenLoc)

    ghostTower.position = mouseWorldLoc

    leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
    if leftPressed && !mousePressed[ebiten.MouseButtonLeft] {
        if ghostTowerVisible {
            if (credits >= ghostTower.cost) && waypointsReady {
                addTower(ghostTower.position)
                credits -= ghostTower.cost
            }
        } else {
            ghostTowerVisible = true
        }
    }
    mousePressed[ebiten.MouseButtonLeft] = leftPressed

    if waveEnemiesRemaining > 0 {
        timeTillEnemySpawn -= deltaTime
        for (timeTillEnemySpawn < 0.0) && (waveEnemiesRemaining > 0) {
            timeTillEnemySpawn += enemySpawnInterval
            waveEnemiesRemaining--
            sendEnemy()
        }
    }


    if !waypointsReady && (len(waypoints) < targetWaypointCount) {
        timeTillNewWaypoint -= deltaTime
        for timeTillNewWaypoint <= 0.0 {
            addPathSegment()
            timeTillNewWaypoint += waypointSpawnInterval
        }
        if len(waypoints) == targetWaypointCount {
            waypointsReady = true
        }
    }

    for index,enemy := range enemies {
        if enemy == nil {
            continue
        }
        enemy.Update()
        if enemy.health <= 0 {
            credits += enemyBounty
            enemies[index] = enemies[len(enemies)-1]
            enemies[len(enemies)-1] = nil
            enemies = enemies[:len(enemies)-1]
            continue
        }
        if enemy.currentWaypoint == len(waypoints) {
            lives--
            if lives == 0 {
                waveInProgress = false
                blackoutOpacity = 0.0
            } else if lives < 0 {
                lives = 0
            }
            enemy.health = -1
            enemies[index] = enemies[len(enemies)-1]
            enemies[len(enemies)-1] = nil
            enemies = enemies[:len(enemies)-1]

            continue
        }
    }
    for _,tower := range towers {
        tower.Update()

        if tower.currentTarget != nil {
            targetOffset := tower.currentTarget.position.Sub(tower.position)
            if targetOffset.Magnitude() > towerAttackRange*tower.scale {
                tower.currentTarget = nil
            }
        }
        if (tower.currentTarget == nil) {
            for _,enemy := range enemies {
                offset := enemy.position.Sub(tower.position)
                if offset.Magnitude() < towerAttackRange*tower.scale {
                    tower.currentTarget = enemy
                    break
                }
            }
        }
    }
    for index,projectile := range projectiles {
        if projectile == nil {
            continue // NOTE: I mean this works, but is there a less-hacky way of removing inside a range?
        }
        projectile.Update()
        if projectile.isDead {
            projectiles[index] = projectiles[len(projectiles)-1]
            projectiles[len(projectiles)-1] = nil
            projectiles = projectiles[:len(projectiles)-1]
            continue
        }
    }

    if waveInProgress && (len(enemies) == 0) && (waveEnemiesRemaining == 0) && (lives > 0) {
        endRound()
    }

    if ebiten.IsRunningSlowly() {
        return nil
    }

    bgOpts := ebiten.DrawImageOptions{}
    screen.DrawImage(backgroundImg, &bgOpts)
    pathSize := 10.0
    white := ebiten.ColorM{}

    for i := 1; i<len(waypoints); i++ {
        prevWaypoint := waypoints[i-1]
        nextWaypoint := waypoints[i]

        drawLine(screen, prevWaypoint, nextWaypoint, pathSize, white)
    }

    pathStartDir := waypoints[1].Sub(waypoints[0])
    pathStartAngle := math.Atan2(pathStartDir.y, pathStartDir.x)
    pathEndDir := waypoints[len(waypoints)-1].Sub(waypoints[len(waypoints)-2])
    pathEndAngle := math.Atan2(pathEndDir.y, pathEndDir.x)
    drawSprite(screen, waypoints[0], pathSize, pathStartAngle, pathStartImg, white)
    drawSprite(screen, waypoints[len(waypoints)-1], pathSize, pathEndAngle, pathEndImg, white)

    for _,enemy := range enemies {
        drawSprite(screen, enemy.position, enemySize, 0, enemyImg[enemy.animFrame], white)
    }
    rangeClr := ebiten.ScaleColor(1,1,1,0.3)
    for _,tower := range towers {
        if mouseWorldLoc.Sub(tower.position).Magnitude() < 12.0 {
            drawCircle(screen, tower.position, towerAttackRange*tower.scale, rangeClr)
        }
        drawSprite(screen, tower.position, tower.scale*towerSize, 0, towerImg[tower.animFrame], white)
    }
    for _,proj := range projectiles {
        drawSprite(screen, proj.position, projectileSize*proj.scale, proj.rotation, projectileImg, white)
    }

    ghostTowerClr := white
    ghostTowerClr.Scale(1,1,1,0.5)
    ghostRangeClr := rangeClr
    ghostRangeClr.Scale(1,1,1,0.5)
    if ghostTowerVisible {
        if credits >= ghostTower.cost {
            drawSprite(screen, ghostTower.position, ghostTower.scale*towerSize, 0, towerCanBuildImg, ghostTowerClr)
        } else {
            drawSprite(screen, ghostTower.position, ghostTower.scale*towerSize, 0, towerNoCanBuildImg, ghostTowerClr)
        }
        drawCircle(screen, ghostTower.position, towerAttackRange*ghostTower.scale, ghostRangeClr)
    }

    if (lives == 0) {
        blackoutOpacity = math.Min(1.0, blackoutOpacity + 1.0*deltaTime)
        blackoutClr := ebiten.ScaleColor(0,0,0,blackoutOpacity)

        opts := ebiten.DrawImageOptions{}
        opts.GeoM.Scale(screenWidth, screenHeight)
        opts.ColorM = blackoutClr
        screen.DrawImage(pixelImg, &opts)
    }

    var msg string
    if (lives == 0) {
        msg = fmt.Sprintf("You lost on wave %d :(\nPress R to restart", currentWave)

    } else if (len(enemies) == 0) && (waveEnemiesRemaining == 0) {
        if currentWave == 0 {
            msg = fmt.Sprintf(
                "Lives: %d\n" +
                "Credits: %d\n" +
                "Tower cost: %d\n" +
                "Press S to start the wave %d\n" +
                "Press Left mouse to place a tower (will show the cursor instead, if its hidden)\n" +
                "Mouse-over an existing tower to see its attack range\n" +
                "Press G to toggle the place-tower cursor\n" +
                "Press Esc to quit at any time",
                lives, credits, ghostTower.cost, currentWave+1)
        } else {
            msg = fmt.Sprintf(
                "Lives: %d\n" +
                "Credits: %d\n" +
                "Tower cost: %d\n" +
                "Press S to start the wave %d",
                lives, credits, ghostTower.cost, currentWave+1)
        }

    } else {
        msg = fmt.Sprintf(
            "Lives: %d\n" +
            "Credits: %d\n" +
            "Tower cost: %d",
            lives, credits, ghostTower.cost)
    }
    ebitenutil.DebugPrint(screen, msg)
    return nil
}

func reset() {
    enemies = enemies[:0]
    towers = towers[:0]
    projectiles = projectiles[:0]

    projectileSpeed = 300.0
    enemySpeed = 15.0
    enemiesPerWave = 0
    enemyBounty = 1
    enemyHealth = 1
    currentWave = 0
    lives = 10

    cameraWidth := float64(screenWidth)
    cameraHeight := float64(screenHeight)
    camera = Rect {
        position: Vec2 { cameraWidth/2.0, cameraHeight/2.0 },
        size: Vec2 { float64(cameraWidth), float64(cameraHeight) },
    }

    credits = 2

    ghostTowerVisible = true
    ghostTower.scale = 1.0
    ghostTower.attackRange = 25.0
    ghostTower.cost = 2

    pathBoundingBox = Rect {}
    pathEndDirection = Vec2 { -1.0, 0.0 }
    pathEndLocation = Vec2 { 0.0, 0.0 }
    pathEndIndex = 1
    waypoints = waypoints[:1]
    waypoints[0] = pathEndLocation
    for i:=1; i<8; i++ {
        addPathSegment()
    }
    waypointsReady = true
}

func loadImage(path string) *ebiten.Image {
    data, err := Asset(path)
    if err != nil {
        log.Fatal(err)
    }

    reader := bytes.NewReader(data)
    image, _, err := image.Decode(reader)
    if err != nil {
        log.Fatal(err)
    }

    result, err := ebiten.NewImageFromImage(image, ebiten.FilterNearest)
    if err != nil {
        log.Fatal(err)
    }
    return result
}

func main() {
    circleImg = loadImage("_resources/circle.png")
    backgroundImg = loadImage("_resources/background.png")
    towerImg[0] = loadImage("_resources/tower_1.png")
    towerImg[1] = loadImage("_resources/tower_2.png")
    towerImg[2] = loadImage("_resources/tower_3.png")
    towerCanBuildImg = loadImage("_resources/tower_canbuild.png")
    towerNoCanBuildImg = loadImage("_resources/tower_nocanbuild.png")
    projectileImg = loadImage("_resources/droplet.png")
    pathSegmentImg  = loadImage("_resources/pathsegment.png")
    pathCornerImg = loadImage("_resources/pathsegment_end.png")
    pathStartImg = loadImage("_resources/pathsegment_first.png")
    pathEndImg = loadImage("_resources/pathsegment_last.png")
    enemyImg[0] = loadImage("_resources/enemy_1.png")
    enemyImg[1] = loadImage("_resources/enemy_1b.png")
    enemyImg[2] = loadImage("_resources/enemy_2.png")
    enemyImg[3] = loadImage("_resources/enemy_3.png")
    enemyImg[4] = loadImage("_resources/enemy_4.png")
    enemyImg[5] = loadImage("_resources/enemy_4b.png")

    pixelImg,_ = ebiten.NewImage(1,1, ebiten.FilterNearest)
    pixelImg.Fill(color.White)

    enemies = make([]*Enemy, 0)
    towers = make([]*Tower, 0)
    projectiles = make([]*Projectile, 0)
    waypoints = make([]Vec2, 1)

    ghostTower = &Tower{}

    reset()

    if err := ebiten.Run(update, screenWidth, screenHeight, 2, "In Defence of a Dragon"); err != nil {
        log.Fatal(err)
    }
}
