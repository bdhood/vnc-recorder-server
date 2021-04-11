package main

import (
	"errors"
	"fmt"
	vnc "github.com/amitbet/vnc2video"
	"github.com/amitbet/vnc2video/encoders"
	"image"
	"image/color"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	guuid "github.com/google/uuid"
)

func encodePPMforRGBA(w io.Writer, img *image.RGBA) error {
	maxvalue := 255
	size := img.Bounds()
	// write ppm header
	_, err := fmt.Fprintf(w, "P6\n%d %d\n%d\n", size.Dx(), size.Dy(), maxvalue)
	if err != nil {
		return err
	}

	if convImage == nil {
		convImage = make([]uint8, size.Dy()*size.Dx()*3)
	}

	rowCount := 0
	for i := 0; i < len(img.Pix); i++ {
		if (i % 4) != 3 {
			convImage[rowCount] = img.Pix[i]
			rowCount++
		}
	}

	if _, err := w.Write(convImage); err != nil {
		return err
	}

	return nil
}

func encodePPMGeneric(w io.Writer, img image.Image) error {
	maxvalue := 255
	size := img.Bounds()
	// write ppm header
	_, err := fmt.Fprintf(w, "P6\n%d %d\n%d\n", size.Dx(), size.Dy(), maxvalue)
	if err != nil {
		return err
	}

	// write the bitmap
	colModel := color.RGBAModel
	row := make([]uint8, size.Dx()*3)
	for y := size.Min.Y; y < size.Max.Y; y++ {
		i := 0
		for x := size.Min.X; x < size.Max.X; x++ {
			color := colModel.Convert(img.At(x, y)).(color.RGBA)
			row[i] = color.R
			row[i+1] = color.G
			row[i+2] = color.B
			i += 3
		}
		if _, err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

var convImage []uint8

func encodePPM(w io.Writer, img image.Image) error {
	if img == nil {
		return errors.New("nil image")
	}
	img1, isRGBImage := img.(*vnc.RGBImage)
	img2, isRGBA := img.(*image.RGBA)
	if isRGBImage {
		return encodePPMforRGBImage(w, img1)
	} else if isRGBA {
		return encodePPMforRGBA(w, img2)
	}
	return encodePPMGeneric(w, img)
}
func encodePPMforRGBImage(w io.Writer, img *vnc.RGBImage) error {
	maxvalue := 255
	size := img.Bounds()
	// write ppm header
	_, err := fmt.Fprintf(w, "P6\n%d %d\n%d\n", size.Dx(), size.Dy(), maxvalue)
	if err != nil {
		return err
	}

	if _, err := w.Write(img.Pix); err != nil {
		return err
	}
	return nil
}

type X264ImageCustomEncoder struct {
	encoders.X264ImageEncoder
	FFMpegBinPath      string
	cmd                *exec.Cmd
	input              io.WriteCloser
	closed             bool
	Framerate          int
	ConstantRateFactor int
}

func (enc *X264ImageCustomEncoder) Init(sessionId guuid.UUID, videoFileName string) {
	if enc.Framerate == 0 {
		enc.Framerate = 12
	}
	cmd := exec.Command(enc.FFMpegBinPath,
		"-f", "image2pipe",
		"-vcodec", "ppm",
		"-an", // no audio
		"-y",
		"-i", "-",
		"-vcodec", "libx264",
		"-preset", "medium",
		"-crf", strconv.Itoa(enc.ConstantRateFactor),
		"-pix_fmt", "yuv420p",
		"-hide_banner",
		"-loglevel", "panic",
		videoFileName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	encInput, err := cmd.StdinPipe()
	enc.input = encInput
	if err != nil {
		session_log(sessionId, "can't get ffmpeg input pipe")
		session_log(sessionId, err.Error())
	}
	enc.cmd = cmd
}
func (enc *X264ImageCustomEncoder) Run(sessionId guuid.UUID, videoFileName string) error {
	if _, err := os.Stat(enc.FFMpegBinPath); os.IsNotExist(err) {
		return err
	}

	enc.Init(sessionId, videoFileName)

	err := enc.cmd.Run()
	if err != nil {
		session_log(sessionId, "error while launching ffmpeg")
		session_log(sessionId, err.Error())
		return err
	}
	return nil
}
func (enc *X264ImageCustomEncoder) Encode(sessionId guuid.UUID, img image.Image) {
	if enc.input == nil || enc.closed {
		return
	}

	err := encodePPM(enc.input, img)
	if err != nil && !strings.Contains(err.Error(), "file already closed") {
		session_log(sessionId,"error while encoding image")
		session_log(sessionId, err.Error())
	}
}

func (enc *X264ImageCustomEncoder) Close(sessionId guuid.UUID) {
	if enc.closed {
		return
	}
	enc.closed = true
	err := enc.input.Close()
	if err != nil {
		session_log(sessionId, "could not close input")
		session_log(sessionId, err.Error())
	}
}
