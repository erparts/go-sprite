package sprite

import (
	"image"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

const atlasJSON = `{
 "frames":[
  {"filename":"corner","frame":{"x":0,"y":0,"w":8,"h":8},"duration":100},
  {"filename":"chain","frame":{"x":16,"y":0,"w":12,"h":32},"duration":300}
 ],
 "meta":{"image":"atlas.png","size":{"w":64,"h":32},"frameTags":[
  {"name":"corner","from":0,"to":0,"direction":"forward"},
  {"name":"chain","from":1,"to":1,"direction":"forward"}
 ]}}
`

func TestAsepriteVariableFramesAndIndependentPlayers(t *testing.T) {
	f, err := ReadAseprite([]byte(atlasJSON))
	if err != nil {
		t.Fatal(err)
	}
	p, q := f.CreatePlayer(), f.CreatePlayer()
	if err := p.Play("corner"); err != nil {
		t.Fatal(err)
	}
	if err := q.Play("chain"); err != nil {
		t.Fatal(err)
	}
	x, y, x2, y2 := q.CurrentFrameCoords()
	if image.Rect(x, y, x2, y2) != image.Rect(16, 0, 28, 32) {
		t.Fatal("per-frame dimensions were lost")
	}
	if p.FrameIndex != 0 || p.File != q.File {
		t.Fatal("players must share data, not playback state")
	}
	if frame, ok := f.FrameByName("chain"); !ok || frame.Width != 12 || frame.Height != 32 {
		t.Fatal("named frame missing")
	}
	if f.Width != 64 || f.Height != 32 {
		t.Fatal("atlas dimensions were replaced with frame dimensions")
	}
}

func TestAsepriteHashWithoutNumericNames(t *testing.T) {
	data := `{"frames":{"wide":{"frame":{"x":0,"y":0,"w":20,"h":8},"duration":50},"short":{"frame":{"x":20,"y":0,"w":5,"h":8},"duration":60}},"meta":{"size":{"w":32,"h":8},"image":"test.png"}}`
	f, err := ReadAseprite([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if f.Frames[0].Name != "short" || f.Frames[1].Name != "wide" {
		t.Fatal("hash order is unstable")
	}
	for _, name := range []string{"test 2.aseprite", "test 10.aseprite", "plain", ""} {
		asepriteFrameNumber(name)
	}
}

func TestVariableDurationsAndLegacyCoordinates(t *testing.T) {
	f, err := ReadAseprite([]byte(atlasJSON))
	if err != nil {
		t.Fatal(err)
	}
	p := f.CreatePlayer()
	_ = p.Play("")
	p.Update(.35)
	if p.FrameIndex != 1 {
		t.Fatal("second frame's longer duration was ignored")
	}
	p.Update(.1)
	if p.FrameIndex != 0 {
		t.Fatal("animation did not loop after both durations")
	}
	legacy, err := ReadSpritesheet(strings.NewReader("idle/0 = 4 5 6 7"), "test.txt", .1)
	if err != nil {
		t.Fatal(err)
	}
	q := legacy.CreatePlayer()
	_ = q.Play("idle")
	x, y, x2, y2 := q.CurrentFrameCoords()
	if image.Rect(x, y, x2, y2) != image.Rect(4, 5, 10, 12) {
		t.Fatal("legacy uniform sheets changed")
	}
}

func TestRejectInvalidAsepriteData(t *testing.T) {
	for _, data := range []string{"bad", `{}`, strings.Replace(atlasJSON, `"w":12`, `"w":99`, 1), strings.Replace(atlasJSON, `"duration":300`, `"duration":0`, 1), strings.Replace(atlasJSON, `"to":1`, `"to":4`, 1), strings.Replace(atlasJSON, `"filename":"chain"`, `"filename":"corner"`, 1), strings.Replace(atlasJSON, `"filename":"chain"`, `"trimmed":true,"filename":"chain"`, 1)} {
		if _, err := ReadAseprite([]byte(data)); err == nil {
			t.Fatalf("invalid atlas accepted: %s", data)
		}
	}
}

func TestBundledAsepriteExport(t *testing.T) {
	f, err := OpenAseprite("example/16x16Deliveryman.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Frames) == 0 || f.FrameWidth != 16 || f.FrameHeight != 16 {
		t.Fatal("bundled export changed")
	}
}

func TestSeekMatchesPlaybackAndClampsWithoutCallbacks(t *testing.T) {
	for _, direction := range []Direction{PlayForward, PlayBackward, PlayPingPong} {
		f := &File{Frames: []Frame{{Duration: .125}, {Duration: .25}, {Duration: .5}}}
		f.Tags = map[string]*Tag{"test": {Name: "test", Start: 0, End: 2, Direction: direction, File: f}}
		p, q := f.CreatePlayer(), f.CreatePlayer()
		_ = p.Play("test")
		_ = q.Play("test")
		q.OnLoop = func(*Player) { t.Fatal("seek fired OnLoop") }
		q.OnFrameChange = func(*Player, int) { t.Fatal("seek fired OnFrameChange") }
		q.OnTagEnter = func(*Player, *Tag) { t.Fatal("seek fired OnTagEnter") }
		for step := 1; step <= 200; step++ {
			p.Update(.03125)
			if err := q.Seek(float32(step)*.03125, true); err != nil {
				t.Fatal(err)
			}
			if p.FrameIndex != q.FrameIndex || p.FrameProgress() != q.FrameProgress() {
				t.Fatalf("%s at %d: update=%d/%f seek=%d/%f", direction, step, p.FrameIndex, p.FrameProgress(), q.FrameIndex, q.FrameProgress())
			}
		}
		if err := q.Seek(100, false); err != nil {
			t.Fatal(err)
		}
		want := map[Direction]int{PlayForward: 2, PlayBackward: 0, PlayPingPong: 1}[direction]
		if q.FrameIndex != want || q.FrameProgress() != 1 {
			t.Fatalf("%s did not clamp", direction)
		}
		if err := q.Seek(-10, false); err != nil {
			t.Fatal(err)
		}
		want = 0
		if direction == PlayBackward {
			want = 2
		}
		if q.FrameIndex != want || q.FrameProgress() != 0 {
			t.Fatal("negative time did not rewind")
		}
	}
}

func TestSingleFramePingPong(t *testing.T) {
	f := &File{Frames: []Frame{{Duration: .125}}}
	f.Tags = map[string]*Tag{"idle": {Start: 0, End: 0, Direction: PlayPingPong, File: f}}
	p := f.CreatePlayer()
	_ = p.Play("idle")
	p.Update(1)
	if p.FrameIndex != 0 {
		t.Fatal("single frame stepped outside its tag")
	}
	if err := p.Seek(1000, true); err != nil {
		t.Fatal(err)
	}
}

func TestDrawOptionsKeepNativeRegionAndCallerState(t *testing.T) {
	f, err := ReadAseprite([]byte(atlasJSON))
	if err != nil {
		t.Fatal(err)
	}
	texture := ebiten.NewImage(64, 32)
	defer texture.Deallocate()
	p := f.CreatePlayerWithImage(texture)
	if err := p.Play("chain"); err != nil {
		t.Fatal(err)
	}
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(10, 20)
	called := false
	p.OnDraw = func(_ *Player, _, frame *ebiten.Image, draw *ebiten.DrawImageOptions) bool {
		called = true
		if frame.Bounds() != image.Rect(16, 0, 28, 32) {
			t.Fatal("draw ignored per-frame bounds")
		}
		if x, y := draw.GeoM.Apply(0, 0); x != 10 || y != 20 {
			t.Fatal("transform lost")
		}
		draw.GeoM.Translate(5, 5)
		return true
	}
	if err := p.DrawWithOptions(nil, &opts); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("draw callback not called")
	}
	if x, y := opts.GeoM.Apply(0, 0); x != 10 || y != 20 {
		t.Fatal("caller options mutated")
	}
	if err := f.CreatePlayer().Draw(nil); err == nil {
		t.Fatal("missing image was accepted")
	}
}
