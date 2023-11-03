package sprite

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSpritesheet(t *testing.T) {
	f, err := ReadSpritesheet(strings.NewReader(data), "foo.txt", 67/1000)
	require.NoError(t, err)

	require.Len(t, f.Frames, 30)
	assert.Equal(t, f.ImagePath, "foo.png")
	assert.Equal(t, f.Frames[1].X, 50)
	assert.Equal(t, f.Frames[1].Y, 0)
}

var data = `frame0000 = 0 0 50 50
frame0001 = 50 0 50 50
frame0002 = 100 0 50 50
frame0003 = 150 0 50 50
frame0004 = 200 0 50 50
frame0005 = 250 0 50 50
frame0006 = 300 0 50 50
frame0007 = 350 0 50 50
frame0008 = 400 0 50 50
frame0009 = 450 0 50 50
frame0010 = 500 0 50 50
frame0011 = 550 0 50 50
frame0012 = 600 0 50 50
frame0013 = 650 0 50 50
frame0014 = 700 0 50 50
frame0015 = 750 0 50 50
frame0016 = 800 0 50 50
frame0017 = 850 0 50 50
frame0018 = 900 0 50 50
frame0019 = 950 0 50 50
frame0020 = 1000 0 50 50
frame0021 = 1050 0 50 50
frame0022 = 1100 0 50 50
frame0023 = 1150 0 50 50
frame0024 = 1200 0 50 50
frame0025 = 1250 0 50 50
frame0026 = 1300 0 50 50
frame0027 = 1350 0 50 50
frame0028 = 1400 0 50 50
frame0029 = 1450 0 50 50`
