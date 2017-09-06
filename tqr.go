package tqr

import (
	"image"
	"image/color"
	"math"
	"sort"
	"errors"
)

const (
	pad = 0.5
)

var QrParseErr = errors.New("qr parse error")

type e struct {
	dx int
	dy int
	len int
}

type line []e

type finder struct {
	x int
	y int
}

func dest(a, b finder) float64 {
	x := a.x - b.x
	y := a.y - b.y

	s := x*x + y*y

	return math.Sqrt(float64(s))
}

func trim(img *image.Gray, bounds image.Rectangle, minX, maxX, minY, maxY, padX, padY int) *image.Gray {
	ret := image.NewGray(bounds)
	var sx, sy int
	if minX - padX < 0 {
		sx = 0
	} else {
		sx = minX - padX
	}
	if minY - padY < 0 {
		sy = 0
	} else {
		sy = minY - padY
	}

	for x := 0 ; sx+x < maxX + padX || sx+x < img.Bounds().Max.X; x ++ {
		for y := 0 ; sy+y < maxY + padY || sy+y < img.Bounds().Max.Y; y ++ {
			ret.SetGray(x, y, img.GrayAt(sx+x, sy+y))
		}
	}
	return ret
}

func rgbaToGray(img image.Image) *image.Gray {
	var (
		bounds = img.Bounds()
		gray   = image.NewGray(bounds)
		bin = image.NewGray(bounds)
	)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			var rgba = img.At(x, y)
			gray.Set(x, y, rgba)
		}
	}
	for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x ++ {
		for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y ++ {
			if 0xff/  2 < gray.GrayAt(x, y).Y {
				bin.Set(x, y, color.Gray{0xff})
			} else {
				bin.Set(x, y, color.Gray{0x00})
			}
		}
	}
	return bin
}

func cmp(a, b, p int) bool {
	if 0 < p {
		if a > b {
			return float64(p)*(1-pad) < float64(a)/float64(b) && float64(a)/float64(b) < float64(p) * (1+pad)
		}
		return false
	}
	if a < b {
		return float64(-1*p)*(1-pad) < float64(b)/float64(a) && float64(b)/float64(a) < float64(-1*p) * (1+pad)
	}
	return false
}

func Tqr(qr image.Image) (*image.Gray, error) {
	gray := rgbaToGray(qr)
	f := []finder{}

lop:
	for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x ++ {
		l := line{}
		el := e{dx:x,dy:0,len:1}
		bc := gray.GrayAt(x, 0).Y
		for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y ++ {
			if y == 0 {
				continue
			}
			if gray.GrayAt(x, y).Y != bc {
				l = append(l, el)
				el = e{dx:x, dy:y, len: 1}
				bc = gray.GrayAt(x, y).Y
				if len(l) == 5 {
					if !(cmp(l[0].len, l[1].len, 1) || cmp(l[0].len, l[1].len, -1)) {
						l = l[1:]
						continue
					}

					b := cmp(l[0].len, l[1].len, 1) || cmp(l[0].len, l[1].len, -1)
					if !b {
						l = l[1:]
						continue
					}
					for i, _ := range l {
						b = b && l[i].len < 80
					}
					b = b && gray.GrayAt(l[0].dx, l[0].dy).Y == 0x00
					b = b && gray.GrayAt(l[1].dx, l[1].dy).Y == 0xff
					b = b && gray.GrayAt(l[2].dx, l[2].dy).Y == 0x00
					b = b && gray.GrayAt(l[3].dx, l[3].dy).Y == 0xff
					b = b && gray.GrayAt(l[4].dx, l[4].dy).Y == 0x00
					b = b && cmp(l[0].len, l[2].len, -3)
					b = b && cmp(l[1].len, l[2].len, -3)
					b = b && cmp(l[2].len, l[3].len, 3)
					b = b && cmp(l[2].len, l[4].len, 3)
					b = b && (cmp(l[3].len, l[4].len, 1) || cmp(l[3].len, l[4].len, -1))
					if !b {
						l = l[1:]
						continue
					}
					if len(f) == 0 {
						f = append(f, finder{x,y})
					} else {
						n := false
						for _, p := range f {
							d := dest(p, finder{x,y})
							if  d < 200{
								n = true
								break
							}
						}
						if !n {
							f = append(f, finder{x,y})
						}
					}
					l = line{}
					if len(f) == 4 {
						break lop
					}
				}
			} else {
				el.len ++
			}

		}
	}
	if len(f) < 3 {
		return nil, QrParseErr
	}
	xa := []int{}
	ya := []int{}

	for _, e := range f {
		xa = append(xa, e.x)
		ya = append(ya, e.y)
	}
	sort.Ints(xa)
	sort.Ints(ya)

	minX := xa[0]
	maxX := xa[len(xa)-1]
	minY := ya[0]
	maxY := ya[len(ya)-1]

	padX := int(float64(maxX - minX) * 0.2)
	padY := int(float64(maxY - minY) * 0.2)
	bounds := image.Rect(0, 0, padX*2+(maxX-minX), padY*2+(maxY-minY))

	return trim(gray, bounds, minX, maxX, minY, maxY, padX, padY), nil
}