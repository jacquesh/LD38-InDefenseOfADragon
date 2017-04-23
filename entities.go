package main

import "math"

type Enemy struct {
    health int
    position Vec2
    currentWaypoint int

    animFrame int
    animFrameDuration float64
}

func (e *Enemy) Update() {
    simTime := deltaTime
    for (simTime > 0) && (e.currentWaypoint < len(waypoints)) {
        moveDist := enemySpeed * simTime
        offset := waypoints[e.currentWaypoint].Sub(e.position)
        offsetDist := offset.Magnitude()
        if offsetDist > moveDist {
            e.position = e.position.Add(offset.Normalized().Mul(moveDist))
            simTime = 0.0
        } else {
            timeToWaypoint := offsetDist/enemySpeed
            e.position = waypoints[e.currentWaypoint]
            e.currentWaypoint++
            simTime -= timeToWaypoint
        }
    }

    e.animFrameDuration -= deltaTime
    if e.animFrameDuration < 0 {
        e.animFrameDuration += 0.40/float64(6)
        e.animFrame = (e.animFrame+1)%6
    }
}

type Tower struct {
    position Vec2
    scale float64
    cost int

    animFrame int
    animFrameDuration float64

    timeTillAttack float64
    attackRange float64
    currentTarget *Enemy
}

func (t *Tower) Update() {
    t.timeTillAttack -= deltaTime
    if t.currentTarget != nil {
        if (t.timeTillAttack <= 0.0) {
            t.timeTillAttack = 1.5
            createProjectile(t, t.currentTarget)
        }
        if t.currentTarget.health <= 0 {
            t.currentTarget = nil
        }
    }

    t.animFrameDuration -= deltaTime
    if t.animFrameDuration < 0 {
        t.animFrameDuration += 0.35/float64(3)
        t.animFrame = (t.animFrame+1)%3
    }
}

type Projectile struct {
    position Vec2
    scale float64
    target *Enemy
    damage int
    isDead bool

    rotation float64
}

func (p *Projectile) Update() {
    speed := projectileSpeed * deltaTime
    offset := p.target.position.Sub(p.position)
    if offset.Magnitude() > speed {
        p.rotation = math.Atan2(offset.y, offset.x)
        p.position = p.position.Add(offset.Normalized().Mul(speed))
    } else {
        p.isDead = true
        p.target.health -= p.damage
    }
}
