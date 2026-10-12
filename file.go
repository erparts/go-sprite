package sprite

import (
	"fmt"
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
	// PlayPingPongReverse starts at the last frame and returns forward.
	PlayPingPongReverse Direction = "pingpong_reverse"
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
	if err := f.decode(data); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *File) decode(data []byte) error {
	json := string(data)
	if !gjson.Valid(json) {
		return fmt.Errorf("invalid Aseprite JSON")
	}

	f.ImagePath = filepath.Clean(gjson.Get(json, "meta.image").String())

	f.Width = int32(gjson.Get(json, "meta.size.w").Num)
	f.Height = int32(gjson.Get(json, "meta.size.h").Num)
	if f.Width <= 0 || f.Height <= 0 {
		return fmt.Errorf("invalid Aseprite atlas size")
	}

	for _, key := range gjson.Get(json, "meta.layers").Array() {
		f.Layers = append(f.Layers, Layer{Name: key.Get("name").String(), Opacity: uint8(key.Get("opacity").Int()), BlendMode: key.Get("blendMode").String()})
	}

	frames := gjson.Get(json, "frames")
	var names []string
	var entries []gjson.Result
	if frames.IsArray() {
		entries = frames.Array()
		for _, entry := range entries {
			names = append(names, entry.Get("filename").String())
		}
	} else if frames.IsObject() {
		byName := frames.Map()
		for name := range byName {
			names = append(names, name)
		}
		// Preserve conventional Aseprite numeric ordering, with a deterministic
		// lexical fallback for named atlas regions and names without extensions.
		sort.Slice(names, func(i, j int) bool {
			a, aok := asepriteFrameNumber(names[i])
			b, bok := asepriteFrameNumber(names[j])
			if aok != bok {
				return aok
			}
			if aok && bok && a != b {
				return a < b
			}
			return names[i] < names[j]
		})
		for _, name := range names {
			entries = append(entries, byName[name])
		}
	} else {
		return fmt.Errorf("Aseprite frames must be an array or object")
	}
	if len(entries) == 0 {
		return fmt.Errorf("Aseprite atlas has no frames")
	}
	seen := map[string]bool{}
	for i, entry := range entries {
		name := names[i]
		if name == "" || seen[name] {
			return fmt.Errorf("empty or duplicate frame name %q", name)
		}
		seen[name] = true
		if entry.Get("rotated").Bool() || entry.Get("trimmed").Bool() {
			return fmt.Errorf("frame %q: rotated/trimmed frames are not supported; export without rotation or trimming", name)
		}
		frame := Frame{Name: name, X: int(entry.Get("frame.x").Int()), Y: int(entry.Get("frame.y").Int()),
			Width: int(entry.Get("frame.w").Int()), Height: int(entry.Get("frame.h").Int()),
			Duration: float32(entry.Get("duration").Float() / 1000)}
		if frame.Width <= 0 || frame.Height <= 0 || frame.X < 0 || frame.Y < 0 ||
			frame.X > int(f.Width)-frame.Width || frame.Y > int(f.Height)-frame.Height || frame.Duration <= 0 {
			return fmt.Errorf("frame %q: invalid bounds or duration", name)
		}
		f.Frames = append(f.Frames, frame)
		if i == 0 {
			f.FrameWidth, f.FrameHeight = int32(frame.Width), int32(frame.Height)
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
		if _, exists := f.Tags[animName]; exists {
			return fmt.Errorf("duplicate Aseprite tag %q", animName)
		}
		from, to := int(anim.Get("from").Int()), int(anim.Get("to").Int())
		direction := Direction(anim.Get("direction").Str)
		if direction == "" {
			direction = PlayForward
		}
		if animName == "" || from < 0 || to < from || to >= len(f.Frames) ||
			(direction != PlayForward && direction != PlayBackward && direction != PlayPingPong && direction != PlayPingPongReverse) {
			return fmt.Errorf("invalid Aseprite tag %q", animName)
		}

		f.Tags[animName] = &Tag{
			Name:      animName,
			Start:     from,
			End:       to,
			Direction: direction,
			File:      f,
		}
	}

	for _, sliceData := range gjson.Get(json, "meta.slices").Array() {
		color, _ := strconv.ParseInt(strings.TrimPrefix(sliceData.Get("color").Str, "#"), 16, 64)

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

func asepriteFrameNumber(name string) (int64, bool) {
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	n, err := strconv.ParseInt(stem[strings.LastIndex(stem, " ")+1:], 10, 64)
	return n, err == nil
}

// FrameByName returns a named region from an Aseprite export.
func (f *File) FrameByName(name string) (Frame, bool) {
	for _, frame := range f.Frames {
		if frame.Name == name {
			return frame, true
		}
	}
	return Frame{}, false
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
	X, Y          int
	Duration      float32 // The duration of the frame in seconds.
	Name          string  // The exported filename/region name, when available.
	Width, Height int     // Packed dimensions. Zero uses File.FrameWidth/FrameHeight for legacy files.
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
