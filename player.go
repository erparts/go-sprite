// Package sprite is an Aseprite JSON loader written in Golang.
package sprite

import (
	"errors"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	ErrNoTagByName = errors.New("no tags by name")
)

// Player is an animation player for Aseprite files.
type Player struct {
	File           *File
	PlaySpeed      float32 // The playback speed; altering this can be used to globally slow down or speed up animation playback.
	CurrentTag     *Tag    // The currently playing animation.
	FrameIndex     int     // The current frame of the File's animation / tag playback.
	PrevFrameIndex int     // The previous frame in the playback.
	frameCounter   float32

	prevUVX float64
	prevUVY float64

	// OnLoop gets called when the playing animation / tag does a complete loop. For a ping-pong
	// animation, this is a full forward + back cycle.
	OnLoop func(p *Player)
	// OnFrameChange gets called when the playing animation / tag changes frames.
	OnFrameChange func(p *Player, frame int)
	// OnTagEnter gets called when entering a tag from "outside" of it (i.e. if not playing a
	//tag and then it gets played, this gets called, or if you're playing a tag and you pass
	// through another tag).
	OnTagEnter func(p *Player, t *Tag)
	OnTagExit  func(p *Player, t *Tag)

	// OnDraw runs just before drawing the sprite; returning true aborts the draw.
	OnDraw func(p *Player, screen, img *ebiten.Image, opts *ebiten.DrawImageOptions) bool

	playDirection int
	img           *ebiten.Image
}

// CreatePlayer returns a new animation player that plays animations from a given Aseprite file.
func (f *File) CreatePlayer() *Player {
	return &Player{
		File:      f,
		PlaySpeed: 1,
	}
}

func (f *File) CreatePlayerWithImage(img *ebiten.Image) *Player {
	return &Player{
		File:      f,
		PlaySpeed: 1,
		img:       img,
	}
}

// Clone clones the Player.
func (p *Player) Clone() *Player {
	newPlayer := p.File.CreatePlayer()
	newPlayer.PlaySpeed = p.PlaySpeed
	newPlayer.CurrentTag = p.CurrentTag
	newPlayer.FrameIndex = p.FrameIndex
	newPlayer.frameCounter = p.frameCounter

	newPlayer.OnLoop = p.OnLoop
	newPlayer.OnFrameChange = p.OnFrameChange
	newPlayer.OnTagEnter = p.OnTagEnter
	newPlayer.OnTagExit = p.OnTagExit

	return newPlayer
}

func (p *Player) Draw(screen *ebiten.Image) error {
	return p.DrawWithOptions(screen, nil)
}

// DrawWithOptions draws the current frame with caller-supplied transforms. It
// does not retain or modify the caller's options. OnDraw still runs afterwards.
func (p *Player) DrawWithOptions(screen *ebiten.Image, options *ebiten.DrawImageOptions) error {
	if p.img == nil {
		return errors.New("sprite player has no image")
	}
	if _, ok := p.CurrentFrame(); !ok {
		return errors.New("sprite player has no active frame")
	}
	opts := &ebiten.DrawImageOptions{}
	if options != nil {
		*opts = *options
	}
	sub := p.img.SubImage(image.Rect(p.CurrentFrameCoords())).(*ebiten.Image)

	if p.OnDraw != nil {
		if stop := p.OnDraw(p, screen, sub, opts); stop {
			return nil
		}
	}

	screen.DrawImage(sub, opts)
	return nil
}

// Play sets the specified tag name up to be played back. A tagName of "" will play back the entire file.
func (p *Player) Play(tagName string) error {
	t, ok := p.File.Tags[tagName]
	if !ok {
		return ErrNoTagByName

	}

	if p.CurrentTag == t {
		return nil
	}

	if p.CurrentTag == nil {
		p.PrevFrameIndex = -1
	} else {
		p.PrevFrameIndex = p.FrameIndex
	}

	p.CurrentTag = t
	p.frameCounter = 0

	if t.Direction == PlayBackward {
		p.playDirection = -1
		p.FrameIndex = p.CurrentTag.End
	} else {
		p.playDirection = 1
		p.FrameIndex = p.CurrentTag.Start
	}

	p.pollTagChanges()
	return nil
}

// Update updates the currently playing animation. dt is the delta value between the previous frame and the current frame.
func (p *Player) Update(dt float32) {
	if p.CurrentTag == nil {
		return
	}

	t := p.CurrentTag
	p.frameCounter += dt * p.PlaySpeed
	p.prevUVX, p.prevUVY = p.CurrentUVCoords()

	for {
		frameDur := p.File.Frames[p.FrameIndex].Duration
		if frameDur <= 0 || p.frameCounter < frameDur {
			break
		}
		p.frameCounter -= frameDur
		p.PrevFrameIndex = p.FrameIndex
		p.FrameIndex += p.playDirection

		if t.Direction == PlayPingPong && t.Start != t.End {
			if p.FrameIndex > t.End {
				p.FrameIndex = t.End - 1
				p.playDirection *= -1
			} else if p.FrameIndex < t.Start {
				p.FrameIndex = t.Start + 1
				p.playDirection *= -1
				if p.OnLoop != nil {
					p.OnLoop(p)
				}
			}

		} else if p.playDirection > 0 && p.FrameIndex > t.End {
			p.FrameIndex -= t.End - t.Start + 1
			if p.OnLoop != nil {
				p.OnLoop(p)
			}
		} else if p.playDirection < 0 && p.FrameIndex < t.Start {
			p.FrameIndex += t.End - t.Start + 1
			if p.OnLoop != nil {
				p.OnLoop(p)
			}
		}

		if p.FrameIndex != p.PrevFrameIndex && p.OnFrameChange != nil {
			p.OnFrameChange(p, p.FrameIndex)
		}

		p.pollTagChanges()
	}
}

// Seek samples the current tag at an absolute time in seconds without firing
// callbacks. Negative times select the beginning. If loop is false, times past
// the end select the final frame with progress 1. This does not change Update's
// looping policy. PlaySpeed is not applied to an absolute time.
func (p *Player) Seek(seconds float32, loop bool) error {
	if p.CurrentTag == nil {
		return errors.New("sprite player has no active tag")
	}
	if math.IsNaN(float64(seconds)) || math.IsInf(float64(seconds), 0) {
		return errors.New("invalid seek time")
	}
	t := p.CurrentTag
	n := t.End - t.Start + 1
	if n <= 0 || t.Start < 0 || t.End >= len(p.File.Frames) {
		return errors.New("invalid animation range")
	}
	count := n
	if t.Direction == PlayPingPong && n > 1 {
		count = 2*n - 2
	}
	indexAt := func(step int) int {
		if t.Direction == PlayBackward {
			return t.End - step
		}
		if step >= n {
			return t.End - (step - n + 1)
		}
		return t.Start + step
	}
	total := float64(0)
	for step := 0; step < count; step++ {
		d := float64(p.File.Frames[indexAt(step)].Duration)
		if d <= 0 || math.IsNaN(d) || math.IsInf(d, 0) {
			return errors.New("invalid frame duration")
		}
		total += d
	}
	remaining := math.Max(0, float64(seconds))
	if loop {
		remaining = math.Mod(remaining, total)
	} else {
		remaining = math.Min(remaining, total)
	}
	p.prevUVX, p.prevUVY = p.CurrentUVCoords()
	p.PrevFrameIndex = p.FrameIndex
	for step := 0; step < count; step++ {
		index := indexAt(step)
		d := float64(p.File.Frames[index].Duration)
		if remaining < d || step == count-1 {
			p.FrameIndex = index
			p.frameCounter = float32(math.Min(remaining, d))
			p.playDirection = 1
			if t.Direction == PlayBackward || (t.Direction == PlayPingPong && step >= n-1 && n > 1) {
				p.playDirection = -1
			}
			return nil
		}
		remaining -= d
	}
	return nil
}

// FrameProgress returns elapsed time within the current frame, from 0 to 1.
func (p *Player) FrameProgress() float32 {
	frame, ok := p.CurrentFrame()
	if !ok || frame.Duration <= 0 {
		return 0
	}
	return max(0, min(1, p.frameCounter/frame.Duration))
}

// TouchingTags returns the tags currently being touched by the Player (tag).
func (p *Player) TouchingTags() []*Tag {
	var tags []*Tag
	for _, t := range p.File.Tags {
		if p.FrameIndex >= t.Start && p.FrameIndex <= t.End {
			tags = append(tags, t)
		}
	}

	return tags
}

// TouchingTagByName returns if a tag by the given name is being touched by the Player (tag).
func (p *Player) TouchingTagByName(tagName string) bool {
	for _, t := range p.File.Tags {
		if t.Name == tagName && p.FrameIndex >= t.Start && p.FrameIndex <= t.End {
			return true
		}
	}

	return false
}

// pollTagChanges polls the File for tag changes (entering or exiting Tags).
func (p *Player) pollTagChanges() {
	if p.OnTagExit != nil {
		for _, tag := range p.File.Tags {
			if (p.PrevFrameIndex >= tag.Start && p.PrevFrameIndex <= tag.End) && (p.FrameIndex < tag.Start || p.FrameIndex > tag.End) {
				p.OnTagExit(p, tag)
			}
		}
	}

	if p.OnTagEnter != nil {
		for _, tag := range p.File.Tags {
			if (p.PrevFrameIndex < tag.Start || p.PrevFrameIndex > tag.End) && (p.FrameIndex >= tag.Start && p.FrameIndex <= tag.End) {
				p.OnTagEnter(p, tag)
			}
		}
	}

}

// CurrentFrame returns the current frame for the currently playing Tag in the File and a boolean indicating if the Player is playing a Tag or not.
func (p *Player) CurrentFrame() (Frame, bool) {
	if p.CurrentTag == nil {
		return Frame{}, false
	}

	return p.File.Frames[p.FrameIndex], true
}

// CurrentFrameCoords returns the four corners of the current frame, of format (x1, y1, x2, y2). If File.CurrentFrame() is nil, it will instead
// return all -1's.
func (p *Player) CurrentFrameCoords() (int, int, int, int) {
	frame, ok := p.CurrentFrame()
	if !ok {
		return -1, -1, -1, -1
	}

	w, h := frame.Width, frame.Height
	if w == 0 {
		w = int(p.File.FrameWidth)
	}
	if h == 0 {
		h = int(p.File.FrameHeight)
	}
	return frame.X, frame.Y, frame.X + w, frame.Y + h

}

// CurrentUVCoords returns the top-left corner of the current frame, of format (x, y). If File.CurrentFrame() is nil, it will instead
// return (-1, -1).
func (p *Player) CurrentUVCoords() (float64, float64) {
	frame, ok := p.CurrentFrame()
	if !ok {
		return -1, -1
	}

	return float64(frame.X) / float64(p.File.Width), float64(frame.Y) / float64(p.File.Height)

}

// CurrentUVCoordsDelta returns the current UV Coords as a coordinate movement delta.
// For example, if an animation were to return the X-axis UV coordinates of :
// [ 0, 0, 0, 0, 0.5, 0.5, 0.5, 0.5 ],
// CurrentUVCoordsDelta would return [ 0, 0, 0, 0, 0.5, 0, 0, 0 ], as the UV coordinate
// only changes on that one frame in the middle, from 0 to 0.5. Once it goes to the end,
// it would return -0.5 to return back to the starting frame.
func (p *Player) CurrentUVCoordsDelta() (float64, float64) {
	currentX, currentY := p.CurrentUVCoords()
	return currentX - p.prevUVX, currentY - p.prevUVY
}

// SetFrameIndex sets the currently visible frame to frameIndex, using the playing animation as the range.
// This means calling SetFrameIndex with a frameIndex of 2 would set it to the third frame of the animation that is currently playing.
func (p *Player) SetFrameIndex(frameIndex int) {
	if p.CurrentTag == nil {
		return
	}

	p.FrameIndex = p.CurrentTag.Start + frameIndex
	if p.FrameIndex > p.CurrentTag.End {
		p.FrameIndex = p.CurrentTag.End
	}
	p.frameCounter = 0
}

// FrameIndexInAnimation returns the currently visible frame index, using the playing animation as the range.
// This means that a FrameIndexInAnimation of 0 would be the first frame in the currently playing animation,
// regardless of what frame in the sprite strip that is).
// If no animation is being played, this function will return -1.
func (p *Player) FrameIndexInAnimation() int {
	if p.CurrentTag == nil {
		return -1
	}

	return p.FrameIndex - p.CurrentTag.Start
}
