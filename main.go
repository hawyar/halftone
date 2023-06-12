package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

const (
	defaultThreshold     uint8 = 132
	defaultGrayThreshold uint8 = 12
	defaultDotSize             = 3
	defaultGrayColor           = 128
)

var bayerMatrix = [][]int{
	{1, 9, 3, 11},
	{13, 5, 15, 7},
	{4, 12, 2, 10},
	{16, 8, 14, 6},
}

type options struct {
	inputf        string
	outputf       string
	threshold     uint8
	dotSize       int
	grayThreshold uint8
	grayColor     int
}

type halftone struct {
	input   image.Image
	output  image.Image
	options options
}

func main() {
	app := &cli.App{
		Name:  "halftone",
		Usage: "halftone image",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "output%s",
				Usage:   "output file name",
			},
		},
		Action: Cmd,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Cmd(c *cli.Context) error {
	input := c.Args().Get(0)

	if input == "" {
		return cli.Exit("input file is required", 1)
	}

	path, err := os.Stat(input)

	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	if path.IsDir() {
		return cli.Exit("input path is directory", 1)
	}

	opt := options{
		inputf:  input,
		outputf: c.String("output"),
	}

	if opt.outputf == "output%s" {
		ext := filepath.Ext(opt.inputf)
		opt.outputf = fmt.Sprintf(opt.outputf, ext)
	}

	h := NewHalftone(opt)

	if h == nil {
		return cli.Exit("invalid options", 1)
	}

	err = h.Run()

	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	return nil
}

func NewHalftone(options options) *halftone {
	if options.threshold == 0 {
		options.threshold = defaultThreshold
	}

	if options.dotSize == 0 {
		options.dotSize = defaultDotSize
	}

	if options.grayThreshold == 0 {
		options.grayThreshold = defaultGrayThreshold
	}

	if options.grayColor == 0 {
		options.grayColor = defaultGrayColor
	}

	if options.inputf == "" {
		return nil
	}

	if options.outputf == "" {
		return nil
	}

	return &halftone{
		options: options,
	}
}

func (h *halftone) Run() error {
	var err error

	err = h.Decode()

	if err != nil {
		return err
	}

	bounds := h.input.Bounds()
	gray := image.NewGray(bounds)

	draw.Draw(gray, bounds, h.input, bounds.Min, draw.Src)
	output := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += h.options.dotSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += h.options.dotSize {
			// Calculate the average intensity using Bayer matrix
			sum := 0
			count := 0

			// Iterate over the pixels in the dot area
			for j := 0; j < h.options.dotSize; j++ {
				for i := 0; i < h.options.dotSize; i++ {
					// Get the  h.options.dotSize value of the pixel
					px := gray.GrayAt(x+i, y+j)
					thresholdValue := bayerMatrix[i][j] * 16
					if int(px.Y) > thresholdValue {
						sum += 255
					}
					count++
				}
			}

			// Calculate the average intensity
			averageIntensity := uint8(sum / count)

			// Determine if the dot should be black or white based on the average intensity
			var dotColor color.Color
			if averageIntensity > h.options.threshold {
				dotColor = color.White
			} else if averageIntensity < h.options.threshold-uint8(h.options.grayThreshold) {
				dotColor = color.Black
			} else {
				dotColor = color.Gray{Y: uint8(h.options.grayColor)}
			}

			drawDot(output, x, y, h.options.dotSize, dotColor)
		}
	}

	h.output = output
	err = h.Encode()

	if err != nil {
		return err
	}

	return nil
}

func (h *halftone) Encode() error {
	out, err := os.Create(h.options.outputf)

	if err != nil {
		return err
	}

	defer out.Close()

	err = png.Encode(out, h.output)

	if err != nil {
		return err
	}

	return nil
}

func (h *halftone) Decode() error {
	file, err := os.Open(h.options.inputf)

	if err != nil {
		return err
	}

	defer file.Close()

	img, _, err := image.Decode(file)

	h.input = img

	if err != nil {
		return err

	}
	return nil
}

func drawDot(img draw.Image, x, y, size int, dotColor color.Color) {
	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {
			img.Set(x+i, y+j, dotColor)
		}
	}
}
