package main

import (
	"flag"
	"fmt"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/canvas"
	"golang.org/x/exp/rand"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var rory image.Image

var points []sketchy.IndexPoint

var tree *sketchy.QuadTree

//var tree *sketchy.KDTree
var line []sketchy.Point
var count int
var currentPoint sketchy.IndexPoint
var isAnimating bool
var animateStart int64
var animateEnd int64

func between(a, b, c float64) bool {
	return c >= a && c <= b
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	needsUpdate := false
	for _, s := range s.Sliders[0:2] {
		if s.DidJustChange {
			needsUpdate = true
			break
		}
	}
	if needsUpdate || count == 0 {
		s.Rand.SetSeed(42)
		isAnimating = false
		points = []sketchy.IndexPoint{}
		line = []sketchy.Point{}
		tree = sketchy.NewQuadTree(s.CanvasRect())
		//tree = sketchy.NewKDTree(s.CanvasRect())
		count = 0
		size := rory.Bounds().Size()
		luminosity := s.Slider("luminosity")
		probability := s.Slider("probability")
		dx := s.Width() / float64(size.X)
		dy := s.Height() / float64(size.Y)
		for x := 0; x < size.X; x++ {
			for y := 0; y < size.Y; y++ {
				pixel := rory.At(x, y)
				pixelColor, _ := colorful.MakeColor(pixel)
				_, _, l := pixelColor.Hsl()
				if l < luminosity && rand.Float64() < probability {
					point := sketchy.IndexPoint{
						Index: count,
						Point: sketchy.Point{
							X: float64(x) * dx,
							Y: s.Height() - float64(y)*dy,
						},
					}
					points = append(points, point)
					tree.Insert(point)
					count++
				}
			}
		}
		fmt.Println("Points: ", count)
	}
	if isAnimating {
		advance(int(s.Slider("pointsPerTick")))
	}
	if s.Toggle("Activate Jetpack Goat!") {
		isAnimating = true
		animateStart = time.Now().UnixMilli()
		line = []sketchy.Point{}
		currentPoint = sketchy.IndexPoint{
			Index: -1,
			Point: sketchy.Point{X: 0, Y: 0},
		}
	}
}

func advance(n int) {
	i := 0
	for i < n && len(line) < tree.Size() {
		neighbor := tree.NearestNeighbors(currentPoint, 1)
		if len(neighbor) == 0 {
			break
		}
		//fmt.Println(currentPoint.Point.String(), currentPoint.Index, neighbor[0].Point.String(), neighbor[0].Index)
		if currentPoint.Point.IsEqual(neighbor[0].Point) {
			fmt.Println("points are same")
			log.Fatalln("exiting")
		}
		line = append(line, neighbor[0].Point)
		q := tree.UpdateIndex(neighbor[0], -1)
		if q == nil {
			fmt.Println("something went wrong")
		}
		currentPoint = *q
		i++
	}
	if len(line) == tree.Size() {
		isAnimating = false
		animateEnd = time.Now().UnixMilli()
		seconds := float64(animateEnd-animateStart) / 1000.0
		rate := float64(len(line)) / seconds
		fmt.Printf("%d points in %.2f seconds\n Points/second: %.2f\nms/point: %.2f", len(line), seconds, rate, 1000.0/rate)
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	if s.Toggle("Show Image") {
		c.DrawImage(0, 0, rory, canvas.DefaultResolution)
	}
	if s.Toggle("Show Points") {
		c.SetFillColor(color.Transparent)
		c.SetStrokeColor(canvas.Blue)
		c.SetStrokeWidth(0.3)
		for _, p := range points {
			p.Draw(0.1, c)
		}
	}
	if len(line) > 0 {
		c.SetStrokeColor(canvas.Magenta)
		c.SetStrokeWidth(0.3)
		curve := sketchy.Curve{
			Points: line,
			Closed: false,
		}
		curve.Draw(c)
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "sketch", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	s.Prefix = prefix
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	f, err := os.Open("rory.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalln(err)
	}
	rory = img
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
