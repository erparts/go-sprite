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

func TestReadSpritesheetWithTags(t *testing.T) {
	f, err := ReadSpritesheet(strings.NewReader(dataTags), "foo.txt", 67/1000)
	require.NoError(t, err)

	require.Len(t, f.Frames, 66)
	assert.Equal(t, f.ImagePath, "foo.png")
	assert.Equal(t, f.Frames[1].X, 106)
	assert.Equal(t, f.Frames[1].Y, 0)

	require.Len(t, f.Tags, 9)
	assert.Equal(t, f.Tags[""].Start, 0)
	assert.Equal(t, f.Tags[""].End, 65)
	assert.Equal(t, f.Tags["attack_A"].Start, 0)
	assert.Equal(t, f.Tags["attack_A"].End, 5)
	assert.Equal(t, f.Tags["attack_A_charge"].Start, 6)
	assert.Equal(t, f.Tags["attack_A_charge"].End, 13)
	assert.Equal(t, f.Tags["idle"].Start, 60)
	assert.Equal(t, f.Tags["idle"].End, 65)
}

var dataTags = `attack_A/frame0000 = 0 0 106 104
attack_A/frame0001 = 106 0 106 104
attack_A/frame0002 = 212 0 106 104
attack_A/frame0003 = 318 0 106 104
attack_A/frame0004 = 424 0 106 104
attack_A/frame0005 = 530 0 106 104
attack_A_charge/frame0000 = 0 104 106 104
attack_A_charge/frame0001 = 106 104 106 104
attack_A_charge/frame0002 = 212 104 106 104
attack_A_charge/frame0003 = 318 104 106 104
attack_A_charge/frame0004 = 424 104 106 104
attack_A_charge/frame0005 = 530 104 106 104
attack_A_charge/frame0006 = 636 104 106 104
attack_A_charge/frame0007 = 742 104 106 104
attack_A_start/frame0000 = 0 208 106 104
attack_A_start/frame0001 = 106 208 106 104
attack_A_start/frame0002 = 212 208 106 104
attack_B/frame0000 = 0 312 106 104
attack_B/frame0001 = 106 312 106 104
attack_B/frame0002 = 212 312 106 104
attack_B/frame0003 = 318 312 106 104
attack_B/frame0004 = 424 312 106 104
attack_B/frame0005 = 530 312 106 104
attack_B/frame0006 = 636 312 106 104
attack_B/frame0007 = 742 312 106 104
attack_B/frame0008 = 848 312 106 104
attack_B/frame0009 = 954 312 106 104
attack_B/frame0010 = 1060 312 106 104
die/frame0000 = 0 416 106 104
die/frame0001 = 106 416 106 104
die/frame0002 = 212 416 106 104
die/frame0003 = 318 416 106 104
die/frame0004 = 424 416 106 104
die/frame0005 = 530 416 106 104
die/frame0006 = 636 416 106 104
die/frame0007 = 742 416 106 104
die/frame0008 = 848 416 106 104
die/frame0009 = 954 416 106 104
die/frame0010 = 1060 416 106 104
die/frame0011 = 1166 416 106 104
die/frame0012 = 1272 416 106 104
die/frame0013 = 1378 416 106 104
die/frame0014 = 1484 416 106 104
die/frame0015 = 1590 416 106 104
die/frame0016 = 1696 416 106 104
fly/frame0000 = 0 520 106 104
fly/frame0001 = 106 520 106 104
fly/frame0002 = 212 520 106 104
fly/frame0003 = 318 520 106 104
fly/frame0004 = 424 520 106 104
fly/frame0005 = 530 520 106 104
fly/frame0006 = 636 520 106 104
fly/frame0007 = 742 520 106 104
get_hit/frame0000 = 0 624 106 104
get_hit/frame0001 = 106 624 106 104
get_hit/frame0002 = 212 624 106 104
get_hit/frame0003 = 318 624 106 104
get_hit/frame0004 = 424 624 106 104
get_hit/frame0005 = 530 624 106 104
get_hit/frame0006 = 636 624 106 104
idle/frame0000 = 0 728 106 104
idle/frame0001 = 106 728 106 104
idle/frame0002 = 212 728 106 104
idle/frame0003 = 318 728 106 104
idle/frame0004 = 424 728 106 104
idle/frame0005 = 530 728 106 104`
