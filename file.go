package sprite

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

type Direction string

const (
	// PlayForward plays animations forward
	PlayForward Direction = "forward"
	// PlayBackward plays animations backwards
	PlayBackward Direction = "reverse"
	// PlayPingPong plays animation forward then backward
	PlayPingPong Direction = "pingpong"
)

// File contains all properties of an exported aseprite file. ImagePath is the absolute path to the image as reported by the exported
// Aseprite JSON data. Path is the string used to open the File if it was opened with the Open() function; otherwise, it's blank.
type File struct {
	Path                    string          // Path to the file (exampleSprite.json); blank if the *File was loaded using Read().
	ImagePath               string          // Path to the image associated with the Aseprite file (exampleSprite.png).
	Width, Height           int32           // Overall width and height of the File.
	FrameWidth, FrameHeight int32           // Width and height of the frames in the File.
	Frames                  []Frame         // The animation Frames present in the File.
	Tags                    map[string]*Tag // A map of Tags, with their names being the keys.
	Layers                  []Layer         // A slice of Layers.
	Slices                  []Slice         // A slice of the Slices present in the file.
}

// OpenAseprite will use os.ReadFile() to open the Aseprite JSON file path specified to parse the data. Returns a *goaseprite.File.
// This can be your starting point. Files created with Open() will put the JSON filepath used in the Path field.
func OpenAseprite(jsonPath string) (*File, error) {
	fileData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	f, err := ReadAseprite(fileData)
	if err != nil {
		return nil, err
	}

	f.Path = jsonPath
	return f, nil
}

// ReadAseprite returns a *goaseprite.File for a given sequence of bytes read from an Aseprite JSON file.
func ReadAseprite(data []byte) (*File, error) {
	f := &File{}
	return f, f.decode(data)

}

func (f *File) decode(data []byte) error {
	json := string(data)

	f.ImagePath = filepath.Clean(gjson.Get(json, "meta.image").String())

	frameNames := []string{}

	f.Width = int32(gjson.Get(json, "meta.size.w").Num)
	f.Height = int32(gjson.Get(json, "meta.size.h").Num)

	for _, key := range gjson.Get(json, "meta.layers").Array() {
		f.Layers = append(f.Layers, Layer{Name: key.Get("name").String(), Opacity: uint8(key.Get("opacity").Int()), BlendMode: key.Get("blendMode").String()})
	}

	for key := range gjson.Get(json, "frames").Map() {
		frameNames = append(frameNames, key)
	}

	sort.Slice(frameNames, func(i, j int) bool {
		x := frameNames[i]
		y := frameNames[j]
		xfi := strings.LastIndex(x, " ") + 1
		xli := strings.LastIndex(x, ".")
		xv, _ := strconv.ParseInt(x[xfi:xli], 10, 32)
		yfi := strings.LastIndex(y, " ") + 1
		yli := strings.LastIndex(y, ".")
		yv, _ := strconv.ParseInt(y[yfi:yli], 10, 32)
		return xv < yv
	})

	for _, key := range frameNames {
		frameName := key
		frameName = strings.Replace(frameName, ".", `\.`, -1)
		frameData := gjson.Get(json, "frames."+frameName)

		frame := Frame{}
		frame.X = int(frameData.Get("frame.x").Num)
		frame.Y = int(frameData.Get("frame.y").Num)
		frame.Duration = float32(frameData.Get("duration").Num) / 1000

		f.Frames = append(f.Frames, frame)

		// We want to set it only on the first frame loaded
		if f.FrameWidth == 0 {
			f.FrameWidth = int32(frameData.Get("sourceSize.w").Num)
			f.FrameHeight = int32(frameData.Get("sourceSize.h").Num)
		}
	}

	f.Tags = make(map[string]*Tag, 0)

	// Default ("") animation
	f.Tags[""] = &Tag{
		Name:      "",
		Start:     0,
		End:       len(f.Frames) - 1,
		Direction: PlayForward,
		File:      f,
	}

	for _, anim := range gjson.Get(json, "meta.frameTags").Array() {
		animName := anim.Get("name").Str

		f.Tags[animName] = &Tag{
			Name:      animName,
			Start:     int(anim.Get("from").Num),
			End:       int(anim.Get("to").Num),
			Direction: Direction(anim.Get("direction").Str),
			File:      f,
		}
	}

	for _, sliceData := range gjson.Get(json, "meta.slices").Array() {
		color, _ := strconv.ParseInt("0x"+sliceData.Get("color").Str[1:], 0, 64)

		newSlice := Slice{
			Name:  sliceData.Get("name").Str,
			Data:  sliceData.Get("data").Str,
			Color: color,
		}

		for _, sdKey := range sliceData.Get("keys").Array() {
			newSlice.Keys = append(newSlice.Keys, SliceKey{
				Frame: int32(sdKey.Get("frame").Int()),
				X:     int(sdKey.Get("bounds.x").Int()),
				Y:     int(sdKey.Get("bounds.y").Int()),
				W:     int(sdKey.Get("bounds.w").Int()),
				H:     int(sdKey.Get("bounds.h").Int()),
			})
		}

		f.Slices = append(f.Slices, newSlice)
	}

	return nil
}

// SliceByName returns a Slice that has the name specified and a boolean indicating whether it could be found or not.
// Note that a File can have multiple Slices by the same name.
func (f *File) SliceByName(sliceName string) (Slice, bool) {
	for _, slice := range f.Slices {
		if slice.Name == sliceName {
			return slice, true
		}
	}
	return Slice{}, false
}

// HasSlice returns true if the File has a Slice of the specified name.
func (f *File) HasSlice(sliceName string) bool {
	_, exists := f.SliceByName(sliceName)
	return exists
}

// Frame contains timing and position information for the frame on the spritesheet.
type Frame struct {
	X, Y     int
	Duration float32 // The duration of the frame in seconds.
}

// Slice represents a Slice (rectangle) that was defined in Aseprite and exported in the JSON file.
type Slice struct {
	Name  string     // Name is the name of the Slice, as specified in Aseprite.
	Data  string     // Data is blank by default, but can be specified on export from Aseprite to be whatever you need it to be.
	Keys  []SliceKey // The individual keys (positions and sizes of Slices) according to the Frames they operate on.
	Color int64
}

// SliceKey represents a Slice's size and position in the Aseprite file on a specific frame. An individual Aseprite File can have multiple
// Slices inside, which can also have multiple frames in which the Slice's position and size changes. The SliceKey's Frame indicates which
// frame the key is operating on.
type SliceKey struct {
	Frame      int32
	X, Y, W, H int
}

// Center returns the center X and Y position of the Slice in the current key.
func (k SliceKey) Center() (int, int) {
	return k.X + (k.W / 2), k.Y + (k.H / 2)
}

// Tag contains details regarding each tag or animation from Aseprite.
// Start and End are the starting and ending frame of the Tag. Direction is a string, and can be assigned one of the playback constants.
type Tag struct {
	Name       string
	Start, End int
	Direction  Direction
	File       *File
}

// Layer contains details regarding the layers exported from Aseprite, including the layer's name (string), opacity (0-255), and
// blend mode (string).
type Layer struct {
	Name      string
	Opacity   uint8
	BlendMode string
}
