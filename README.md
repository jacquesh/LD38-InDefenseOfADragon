## In Defense of a Dragon

IDoaD is a simple tower defense created for Ludum Dare 38, where the path followed by the enemies grows after each wave. The game's name comes from the fact that this path forms the [dragon curve](https://en.wikipedia.org/wiki/Dragon_curve).

This game is the first thing I've ever done using Go (I went through the Go tour on Friday evening and started working on this on Saturday morning), and took a total of around 18 hours to do.
I used the [Ebiten](https://github.com/hajimehoshi/ebiten) game engine, which I opted for over [raylib-go](https://github.com/gen2brain/raylib-go) because it's web build supports audio. Except I didn't end up getting a chance to do any audio, and the web build's rendering is a bit broken for some reason. :(
