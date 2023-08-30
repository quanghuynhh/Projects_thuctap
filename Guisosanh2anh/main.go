package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Nr90/imgsim"
	"github.com/fogleman/gg"
)

var (
	img2Data       image.Image
	dhash1, dhash2 imgsim.Hash
	img1Data       image.Image
	img1Chosen     bool
	img2Chosen     bool
	threshold      float64 = 45
)

func main() {
	a := app.New()
	w := a.NewWindow("So sánh ảnh")
	w.Resize(fyne.NewSize(1200, 350))

	iconFilename := "iconapp.ico"
	if _, err := os.Stat(iconFilename); err == nil {
		iconResource, err := loadResourceFromFilename(iconFilename)
		if err == nil {
			w.SetIcon(iconResource)
		}
	}

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Quit", func() { a.Quit() }),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			dialog.ShowCustom("About", "Close", container.NewVBox(
				widget.NewLabel("So sánh hai ảnh cùng kích thước"),
				widget.NewLabel("Version: v1.0.0"),
			), w)
		}))

	mainMenu := fyne.NewMainMenu(
		fileMenu,
		helpMenu,
	)
	w.SetMainMenu(mainMenu)

	img1 := canvas.NewImageFromResource(nil)
	img1.SetMinSize(fyne.NewSize(250, 250))
	anh1 := widget.NewLabel("image 1")

	btn1 := widget.NewButton("Open Image1", func() {
		fileDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, _ error) {
			var err error
			img1Data, err = openFileAndGetImageData(uc)
			if err != nil {
				fmt.Println("Error decoding image:", err)
				return
			}

			img1.Resource = fyne.NewStaticResource(uc.URI().Name(), dataURIReadCloserToBytes(uc))
			img1.Refresh()

			dhash1 = imgsim.DifferenceHash(img1Data)
			fmt.Println("dhash1:", dhash1)

			if !img1Chosen {
				fmt.Println("img1Data:", true)
				img1Chosen = true
			}
		}, w)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fileDialog.Show()
	})

	anh2 := widget.NewLabel("image 2")
	img2 := canvas.NewImageFromResource(nil)
	img2.SetMinSize(fyne.NewSize(250, 250))

	btn2 := widget.NewButton("Open Image2", func() {
		fileDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, _ error) {
			var err error
			img2Data, err = openFileAndGetImageData(uc)
			if err != nil {
				fmt.Println("Error decoding image:", err)
				return
			}

			img2.Resource = fyne.NewStaticResource(uc.URI().Name(), dataURIReadCloserToBytes(uc))
			img2.Refresh()

			dhash2 = imgsim.DifferenceHash(img2Data)
			fmt.Println("dhash2:", dhash2)

			if !img2Chosen {
				fmt.Println("img2Data:", true)
				img2Chosen = true
			}
		}, w)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fileDialog.Show()
	})

	value := widget.NewLabel("mức độ chênh lệch pixel:")
	slider := widget.NewSlider(1, 255)
	slider.Value = 45

	valueLabel := widget.NewLabel(fmt.Sprintf("Value: %.0f", slider.Value))

	slider.OnChanged = func(value float64) {
		if value != 45 {
			threshold = value
		} else {
			threshold = 45
		}
		valueLabel.SetText(fmt.Sprintf("Value: %.0f", threshold))
	}

	anh3 := widget.NewLabel("Ảnh khác")
	img3 := canvas.NewImageFromResource(nil)
	img3.SetMinSize(fyne.NewSize(250, 250))
	ss3 := widget.NewLabel("")

	btn3 := widget.NewButton("Start Comparison", func() {
		fmt.Println("Ngưỡng chênh lệch pixel:", threshold)

		similarity := compareImages(img1Data, img2Data) * 100
		fmt.Printf("Phần trăm giống nhau theo Similarity: %.2f%%\n", similarity)

		if similarity != 100 {
			resultImg := image.NewRGBA(img1Data.Bounds())
			dc := gg.NewContextForRGBA(resultImg)
			dc.DrawImage(img1Data, 0, 0)
			width := resultImg.Bounds().Dx()
			height := resultImg.Bounds().Dy()

			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					r1, g1, b1, _ := img1Data.At(x, y).RGBA()
					r2, g2, b2, _ := img2Data.At(x, y).RGBA()

					diffr := sqDiffUInt32(r1, r2)
					diffg := sqDiffUInt32(g1, g2)
					diffb := sqDiffUInt32(b1, b2)

					if diffr >= threshold || diffg >= threshold || diffb >= threshold {
						dc.SetColor(color.RGBA{R: 255, A: 255})
						dc.DrawRectangle(float64(x), float64(y), 1, 1)
						dc.Stroke()
					}
				}
			}

			img3.Image = resultImg
			img3.Refresh()

		}

		if similarity != 100 {
			ss3.SetText(fmt.Sprintf("Phần trăm giống nhau theo Similarity: %.2f%%", similarity))
		} else {
			ss3.SetText("Hình ảnh giống nhau.")
		}
	})

	anh4 := widget.NewLabel("Ảnh khác theo dhash")
	img4 := canvas.NewImageFromResource(nil)
	img4.SetMinSize(fyne.NewSize(250, 250))
	ss4 := widget.NewLabel("")

	btn4 := widget.NewButton("Start Comparison", func() {
		diffBits := countDifferentBits(dhash1, dhash2)
		fmt.Println("Number of different bits:", diffBits)

		differencePercentage := float64((64 - diffBits) * 100.0 / 64)
		fmt.Println("Phần trăm giống nhau theo DifferenceHash:", differencePercentage)

		if differencePercentage != 100 {
			resultImg4 := highlightDifferences(img1Data, dhash1, dhash2)
			img4.Image = resultImg4
			img4.Refresh()
		}

		ss4.SetText(fmt.Sprintf("Phần trăm giống nhau theo DifferenceHash: %.2f%%", differencePercentage))

	})

	col1 := container.NewVBox(anh1, img1, btn1)
	col2 := container.NewVBox(anh2, img2, btn2)
	col3 := container.NewVBox(anh3, img3, value, slider, valueLabel, btn3, ss3)
	col4 := container.NewVBox(anh4, img4, btn4, ss4)

	content := container.NewGridWithColumns(4, col1, col2, col3, col4)

	w.SetContent(content)
	w.ShowAndRun()
}

func countDifferentBits(dhash1, dhash2 imgsim.Hash) int {
	diffBits := 0
	for i := 0; i < 64; i++ {
		bit1 := (dhash1 >> uint(63-i)) & 1
		bit2 := (dhash2 >> uint(63-i)) & 1
		if bit1 != bit2 {
			diffBits++
		}
	}
	return diffBits
}

func compareImages(resizedImg1, resizedImg2 image.Image) float64 {
	bounds1 := resizedImg1.Bounds()
	bounds2 := resizedImg2.Bounds()

	if bounds1 != bounds2 {
		log.Fatal("The images have different bounds.")
	}

	totalPixels := (bounds1.Max.X - bounds1.Min.X) * (bounds1.Max.Y - bounds1.Min.Y)
	diffPixels := 0

	for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
		for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
			r1, g1, b1, a1 := resizedImg1.At(x, y).RGBA()
			r2, g2, b2, a2 := resizedImg2.At(x, y).RGBA()

			diffr := sqDiffUInt32(r1, r2)
			diffg := sqDiffUInt32(g1, g2)
			diffb := sqDiffUInt32(b1, b2)
			diffa := sqDiffUInt32(a1, a2)

			if diffr >= threshold || diffg >= threshold || diffb >= threshold || diffa >= threshold {
				diffPixels++
			}
		}
	}

	difference := float64(diffPixels) / float64(totalPixels)
	similarity := 1.0 - difference
	return similarity
}

func sqDiffUInt32(x, y uint32) float64 {
	x >>= 8
	y >>= 8
	return math.Abs(float64(y) - float64(x))
}

func loadResourceFromFilename(path string) (*fyne.StaticResource, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	resource := fyne.NewStaticResource(filepath.Base(path), data)
	return resource, nil
}

func openFileAndGetImageData(fileURI fyne.URIReadCloser) (image.Image, error) {
	data, err := os.ReadFile(fileURI.URI().Path())
	if err != nil {
		return nil, err
	}

	decodedImg, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return decodedImg, nil
}

func dataURIReadCloserToBytes(dataURIReadCloser fyne.URIReadCloser) []byte {
	var buf bytes.Buffer
	io.Copy(&buf, dataURIReadCloser)
	data := buf.Bytes()
	dataURIReadCloser.Close()
	return data
}

func highlightDifferences(baseImage image.Image, dhash1, dhash2 imgsim.Hash) *image.RGBA {
	resultImg := image.NewRGBA(baseImage.Bounds())
	dc := gg.NewContextForRGBA(resultImg)
	dc.DrawImage(baseImage, 0, 0)

	width := resultImg.Bounds().Dx() / 8
	height := resultImg.Bounds().Dy() / 8

	for y := 7; y >= 0; y-- {
		for x := 7; x >= 0; x-- {
			i := (7-y)*8 + (7 - x)
			bit1 := (dhash1 >> uint(63-i)) & 1
			bit2 := (dhash2 >> uint(63-i)) & 1

			if bit1 != bit2 {
				rectX := x * width
				rectY := y * height
				dc.SetColor(color.RGBA{R: 255, A: 255})
				dc.DrawRectangle(float64(rectX), float64(rectY), float64(width), float64(height))
				dc.Stroke()
			}
		}
	}

	return resultImg
}
