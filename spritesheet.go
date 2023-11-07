package sprite

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func OpenSpritesheet(filename string, duration float32) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return ReadSpritesheet(f, filename, duration)
}

func ReadSpritesheet(r io.Reader, filename string, duration float32) (*File, error) {
	i := &spridesheetImporter{
		filename: filename,
		duration: duration,
	}

	return i.loadFile(r)
}

type spridesheetImporter struct {
	filename string
	duration float32
}

func (i *spridesheetImporter) loadFile(r io.Reader) (*File, error) {
	file := &File{}

	file.Tags = make(map[string]*Tag)

	var count int
	s := bufio.NewScanner(r)
	for s.Scan() {
		f, tag, w, h := i.parseLine(s.Text())
		file.Frames = append(file.Frames, *f)
		file.Width = int32(w)
		file.Height = int32(h)
		file.FrameWidth = int32(w)
		file.FrameHeight = int32(h)

		if tag == "" {
			continue
		}

		if _, ok := file.Tags[tag]; !ok {
			file.Tags[tag] = &Tag{
				Name:      tag,
				Start:     count,
				End:       count - 1,
				Direction: "fordward",
				File:      file,
			}
		}

		file.Tags[tag].End++
		count++

	}

	name := strings.Split(i.filename, ".")
	file.ImagePath = fmt.Sprintf("%s.png", name[0])

	file.Tags[""] = &Tag{
		Name:      "",
		Start:     0,
		End:       len(file.Frames) - 1,
		Direction: "fordward",
		File:      file,
	}

	return file, s.Err()
}

func (i *spridesheetImporter) parseLine(line string) (*Frame, string, int, int) {
	parts := strings.Split(line, "=")

	frame := strings.TrimSpace(parts[0])
	names := strings.Split(frame, "/")

	var tag string
	if len(names) == 2 {
		tag = strings.TrimSpace(names[0])
	}

	var w, h int
	f := &Frame{}
	for i, n := range strings.Split(strings.TrimSpace(parts[1]), " ") {
		num, _ := strconv.Atoi(n)
		switch i {
		case 0:
			f.X = num
		case 1:
			f.Y = num
		case 2:
			w = num
		case 3:
			h = num
		}
	}

	f.Duration = i.duration
	return f, tag, w, h
}
