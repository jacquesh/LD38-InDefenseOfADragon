package main

import (
    "math"

    "github.com/hajimehoshi/ebiten"
)

func transformForCamera(opts *ebiten.DrawImageOptions) {
    opts.GeoM.Translate(-camera.MinX(), -camera.MinY())
    opts.GeoM.Scale(screenWidth/camera.size.x, screenHeight/camera.size.y)
}

func drawLine(screen *ebiten.Image, from, to Vec2, width float64, clr ebiten.ColorM) {
    spriteWidth,spriteHeight := pathSegmentImg.Size()
    spriteYScale := width/float64(spriteHeight)
    offset := to.Sub(from)
    offsetDir := offset.Normalized()

    opts := ebiten.DrawImageOptions{}
    opts.GeoM.Translate(-0.5*float64(spriteHeight), -0.5*float64(spriteWidth))
    opts.GeoM.Scale(spriteYScale, spriteYScale)
    opts.GeoM.Rotate(math.Atan2(offset.y, offset.x))
    opts.GeoM.Translate(from.x, from.y)
    transformForCamera(&opts)
    opts.ColorM = clr
    screen.DrawImage(pathCornerImg, &opts)

    segmentLoc := from.Add(offsetDir.Mul(0.5*width))
    segmentXScale := (offset.Magnitude()-width)/float64(spriteWidth)
    opts = ebiten.DrawImageOptions{}
    opts.GeoM.Translate(0, -0.5*float64(spriteHeight))
    opts.GeoM.Scale(segmentXScale, spriteYScale)
    opts.GeoM.Rotate(math.Atan2(offset.y, offset.x))
    opts.GeoM.Translate(segmentLoc.x, segmentLoc.y)
    transformForCamera(&opts)
    opts.ColorM = clr
    screen.DrawImage(pathSegmentImg, &opts)
}

func drawSquare(screen *ebiten.Image, position Vec2, size float64, clr ebiten.ColorM) {
    opts := ebiten.DrawImageOptions{}
    opts.GeoM.Translate(-0.5, -0.5)
    opts.GeoM.Scale(size, size)
    opts.GeoM.Translate(position.x, position.y)
    transformForCamera(&opts)
    opts.ColorM = clr
    screen.DrawImage(pixelImg, &opts)
}

func drawCircle(screen *ebiten.Image, position Vec2, radius float64, clr ebiten.ColorM) {
    circleImgSizePx, _ := circleImg.Size()
    circleImgSize := float64(circleImgSizePx)
    opts := ebiten.DrawImageOptions{}
    opts.GeoM.Translate(-0.5*circleImgSize, -0.5*circleImgSize)
    opts.GeoM.Scale(2.0*radius/circleImgSize, 2.0*radius/circleImgSize)
    opts.GeoM.Translate(position.x, position.y)
    transformForCamera(&opts)
    opts.ColorM = clr
    screen.DrawImage(circleImg, &opts)
}

func drawSprite(screen *ebiten.Image,
                position Vec2,
                drawSize, rotation float64,
                sprite *ebiten.Image,
                clr ebiten.ColorM) {
    spriteSizePx, _ := sprite.Size()
    spriteSize := float64(spriteSizePx)
    opts := ebiten.DrawImageOptions{}
    opts.GeoM.Translate(-0.5*spriteSize, -0.5*spriteSize)
    opts.GeoM.Scale(drawSize/spriteSize, drawSize/spriteSize)
    opts.GeoM.Rotate(rotation)
    opts.GeoM.Translate(position.x, position.y)
    transformForCamera(&opts)
    opts.ColorM = clr
    screen.DrawImage(sprite, &opts)
}
