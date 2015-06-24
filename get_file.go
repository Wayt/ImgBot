package main

import (
	"bytes"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/wayt/happyngine"
	"github.com/wayt/happyngine/s3"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime"
	"path/filepath"
	"time"
)

type getFileAction struct {
	happyngine.Action
}

func newGetFileAction(context *happyngine.Context) happyngine.ActionInterface {

	// Init
	this := &getFileAction{}
	this.Context = context

	this.Form = happyngine.NewForm(context,
		happyngine.NewFormElement("bucket", "invalid_bucket"),
		happyngine.NewFormElement("file", "invalid_file"))

	return this
}

func min(v1, v2 uint) uint {
	if v1 < v2 {
		return v1
	}
	return v2
}

func (this *getFileAction) Run() {

	startTime := time.Now()

	bucket := this.Form.Elem("bucket").FormValue()
	file := this.Form.Elem("file").FormValue()

	data, err := s3.Get(bucket, file)
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			this.AddError(404, `not_found_404`)
			return
		}

		panic(err)
	}

	s3GetTime := time.Now()
	var imgResizeTime, imgEncodeTime time.Time

	ext := filepath.Ext(file)
	mimeType := mime.TypeByExtension(ext)

	width := min(uint(this.GetIntParam("width")), 1920)
	height := min(uint(this.GetIntParam("height")), 1920)
	if width > 0 || height > 0 {

		reader := bytes.NewReader(data)
		img, format, err := image.Decode(reader)
		if err != nil {
			panic(err)
		}

		// If we're asked for two resize vector, we only keep the smalest one to keep image ratio
		if width > 0 && height > 0 {

			// Source image size
			imgWidth := img.Bounds().Size().X
			imgHeight := img.Bounds().Size().Y

			// Strange calculus
			if imgWidth > imgHeight {
				width = 0
			} else if imgHeight > imgWidth {
				height = 0
			} else { // imgHeight == imgWidth

				if width > height {
					width = 0
				} else {
					height = 0
				}
			}
		}

		m := resize.Resize(width, height, img, resize.Lanczos3)

		imgResizeTime = time.Now()

		writer := bytes.NewBuffer(nil)

		if format == "jpeg" {
			if err := jpeg.Encode(writer, m, nil); err != nil {
				panic(err)
			}
		} else if format == "png" {
			if err := png.Encode(writer, m); err != nil {
				panic(err)
			}
		} else if format == "gif" {
			if err := gif.Encode(writer, m, nil); err != nil {
				panic(err)
			}
		} else {
			this.Context.Errorln("Unknown image format:", format, " - ", file)
		}

		imgEncodeTime = time.Now()

		data = writer.Bytes()
	}

	this.Context.Debugln("S3 GET Time:", s3GetTime.Sub(startTime), " - Resize Time:", imgResizeTime.Sub(s3GetTime), " - Encode Time:", imgEncodeTime.Sub(imgResizeTime))

	this.SendByte(200, data, fmt.Sprintf("Content-Type: %s", mimeType))
}
