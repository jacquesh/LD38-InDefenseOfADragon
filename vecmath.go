package main

import (
    "math"
)

type Vec2 struct {
    x float64
    y float64
}

func (v Vec2) Add(u Vec2) Vec2 {
    result := Vec2 {
        x: v.x + u.x,
        y: v.y + u.y,
    }
    return result
}

func (v Vec2) Sub(u Vec2) Vec2 {
    result := Vec2 {
        x: v.x - u.x,
        y: v.y - u.y,
    }
    return result
}

func (v Vec2) Mul(u float64) Vec2 {
    result := Vec2 {
        x: v.x * u,
        y: v.y * u,
    }
    return result
}

func (v Vec2) Normalized() Vec2 {
    mag := v.Magnitude()
    result := Vec2 {
        x: v.x/mag,
        y: v.y/mag,
    }
    return result
}

func (v Vec2) Magnitude() float64 {
    return math.Sqrt(v.x*v.x + v.y*v.y)
}

func (v Vec2) Rotate90CCW() Vec2 {
    result := Vec2 {
        x: -v.y,
        y: v.x,
    }
    return result
}

func (v Vec2) Rotate90CW() Vec2 {
    result := Vec2 {
        x: v.y,
        y: -v.x,
    }
    return result
}

type Rect struct {
    position Vec2 // NOTE: position defines the *centre* of the Rect
    size Vec2
}

func (r *Rect) ContainsPoint(v Vec2) bool {
    if((v.x >= r.MinX()) && (v.x <= r.MaxX()) &&
        (v.y >= r.MinY()) && (v.y <= r.MaxY())) {
        return true
    }
    return false
}

func (r *Rect) ContainsRect(other *Rect) bool {
    if((r.MinX() <= other.MinX()) && (r.MaxX() >= other.MaxX()) &&
        (r.MinY() <= other.MinY()) && (r.MaxY() >= other.MaxY())) {
        return true
    }
    return false
}

func (r *Rect) MinX() float64 {
    return r.position.x - r.size.x/2.0
}
func (r *Rect) MaxX() float64 {
    return r.position.x + r.size.x/2.0
}
func (r *Rect) MinY() float64 {
    return r.position.y - r.size.y/2.0
}
func (r *Rect) MaxY() float64 {
    return r.position.y + r.size.y/2.0
}

func (r *Rect) MinXY() Vec2 {
    return Vec2 {
        x: r.MinX(),
        y: r.MinY(),
    }
}

