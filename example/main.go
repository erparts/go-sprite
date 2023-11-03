package main

import (
	"log"

	"github.com/erpart/go-sprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	_ "image/png"
)

type Game struct {
	Sprite    *sprite.File
	AsePlayer *sprite.Player
	Img       *ebiten.Image
}

func NewGame() *Game {
	s, err := sprite.Open("16x16Deliveryman.json")
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := ebitenutil.NewImageFromFile(s.ImagePath)
	if err != nil {
		log.Fatal(err)
	}

	var x float64
	p := s.CreatePlayerWithImage(img)
	p.OnDraw = func(p *sprite.Player, screen, img *ebiten.Image, opts *ebiten.DrawImageOptions) bool {
		if p.CurrentTag.Name != "walk" {
			return false
		}

		opts.GeoM.Translate(x, 0)
		x += 1
		return false
	}

	game := &Game{
		Sprite:    s,
		AsePlayer: p,
	}

	// There are four callback functions that you can use to watch for changes to the internal state of the *File.

	// OnLoop is called when the animation is finished and a loop is completed; for ping-pong, it happens on a full revolution (after going forwards and then backwards).
	// game.Sprite.OnLoop = func() { fmt.Println("loop") }

	// OnFrameChange is called when the sprite's frame changes.
	// game.Sprite.OnFrameChange = func() { fmt.Println("frame change") }

	// OnTagEnter is called when the File enters a new Tag (i.e. if you play an animation of a sword being slashed, you can make this callback watch for a tag that indicates when a corresponding sound should play).
	// game.Sprite.OnTagEnter = func(tag *goaseprite.Tag) { fmt.Println("entered: ", tag.Name) }

	// OnTagExit is called when the File leaves the current Tag.
	// game.Sprite.OnTagExit = func(tag *goaseprite.Tag) { fmt.Println("exited: ", tag.Name) }

	// game.Sprite.PlaySpeed = 2

	ebiten.SetWindowTitle("goaseprite example")
	ebiten.SetWindowResizable(true)

	game.AsePlayer.Play("idle")

	return game

}

func (game *Game) Update() error {

	if ebiten.IsKeyPressed(ebiten.Key1) {
		game.AsePlayer.Play("idle")
	} else if ebiten.IsKeyPressed(ebiten.Key2) {
		game.AsePlayer.Play("walk")
	} else if ebiten.IsKeyPressed(ebiten.Key3) {
		game.AsePlayer.Play("") // Calling Play() with a blank string will play the full animation (similar to playing an animation in Aseprite without any tags selected).
	}

	game.AsePlayer.Update(float32(1.0 / 60.0))

	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	game.AsePlayer.Draw(screen)

}

func (game *Game) Layout(w, h int) (int, int) { return 320, 180 }

func main() {
	game := NewGame()
	ebiten.RunGame(game)

}
