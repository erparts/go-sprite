# go-sprite

Aseprite JSON loading and sprite playback for Ebitengine. Reads native Aseprite
JSON arrays and hashes, with named regions, animation tags, per-frame sizes and
durations. Derived from [goaseprite](https://github.com/SolarLune/goaseprite).

## Native atlases

Export a PNG sheet and JSON metadata from Aseprite, with frame tags enabled.
Both `json-array` and `json-hash` are accepted. Prefer arrays when animation
order matters: array order is preserved. Hash keys use numeric frame suffixes
when available, then lexical order for named regions. Repeated numeric suffixes
are ambiguous and rejected; export JSON arrays in that case.

Disable trimming and rotation when exporting. Those layouts are explicitly
rejected rather than rendered incorrectly. Each frame can have a different
packed width and height. Named static regions can use single-frame tags in the
same file as animated tags; they share one texture.

```go
f, err := sprite.OpenAseprite("atlas.json")
if err != nil {
    return err
}
// Load f.ImagePath relative to your assets directory into *ebiten.Image.
player := f.CreatePlayerWithImage(texture)
if err := player.Play("walk"); err != nil {
    return err
}

// In Update:
player.Update(1.0 / 60)

// In Draw:
var opts ebiten.DrawImageOptions
opts.GeoM.Translate(100, 80)
if err := player.DrawWithOptions(screen, &opts); err != nil {
    panic(err)
}
```

Import as `sprite "github.com/erparts/go-sprite"`. With embedded files, use
`ReadAseprite(data)` instead of `OpenAseprite`. `File.FrameByName` exposes named
frame metadata without creating another file. Each `Player` has independent
playback state, while its `File` and image can be shared.

`Play("")` selects all frames. `Play` supports forward, reverse, ping-pong and reverse ping-pong
animation. `Update` advances through each frame's own duration and loops.
`Play` on the active tag preserves its position; use `Restart` to rewind it.
`Clone` shares the immutable file and texture while copying playback and callbacks.
Aseprite tag `repeat` is currently ignored: `Update` loops until the caller stops it.
Callbacks on the player report loops, frame changes and tag transitions.

`Seek(seconds, loop)` samples an absolute position without callbacks. With
`loop=false` it clamps to the final frame of one cycle: a ping-pong cycle
omits the duplicated endpoints (0,1,2,1; reverse: 2,1,0,1). This is useful for replay, paused
previews and externally controlled clocks. `FrameProgress()` exposes the
fraction elapsed in the current frame, including 1 at a clamped endpoint.
Seeking does not change the looping behaviour of subsequent `Update` calls.

`Draw` uses default options; `DrawWithOptions` applies caller transforms and
colour without modifying the supplied options. `OnDraw` still runs before the
actual draw; returning true skips it (the existing callback convention).
Drawing without an image or active frame returns an error without invoking
`OnDraw`; set `CurrentTag = nil` to stop playback and skip Draw while hidden.

## Programmatic sheets and compatibility

For runtime-generated textures, create a `File` with native `Frame` and `Tag`
values and pass the texture to `CreatePlayerWithImage`. Give frames explicit
`Width`, `Height` and `Duration`. No intermediate text or JSON format is needed.

The legacy text-sheet reader remains available. Frames with zero width or
height fall back to `File.FrameWidth` / `File.FrameHeight`, preserving uniform
sheet behaviour. Native JSON imports always retain individual packed sizes.

## Validation

The native reader rejects malformed JSON, empty sheets, duplicate names,
out-of-bounds regions, nonpositive durations and invalid tag ranges/directions.

```sh
go test ./...
```

Tests import Ebitengine and therefore need a display. In a headless development
environment, use the project's display wrapper, for example
`ebitenmcp run -- go test ./...` when installed.
