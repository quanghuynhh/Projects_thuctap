package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"

	"gocv.io/x/gocv"
)

func main() {
	// Đặt tên file video và thư mục lưu các ảnh
	videoFile := "video_datthung_combo3thung.mp4"
	imageFolder := "./texluu/images/"

	// Mở video
	capture, err := gocv.VideoCaptureFile(videoFile)
	if err != nil {
		fmt.Printf("Error opening video file: %v\n", videoFile)
		return
	}
	defer capture.Close()

	// Tạo thư mục để lưu ảnh
	err = os.MkdirAll(imageFolder, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating image folder: %v\n", imageFolder)
		return
	}
	// Lấy số khung hình trên giây (FPS) của video
	frameFPS := int(capture.Get(gocv.VideoCaptureFPS))
	fmt.Println("FPS của video:", frameFPS)

	if frameFPS == 0 {
		fmt.Println("Không thể lấy thông tin FPS của video.")
		return
	}
	// Tạo ma trận hình ảnh
	img1 := gocv.NewMat()
	defer img1.Close()
	img2 := gocv.NewMat()
	defer img2.Close()

	capture.Read(&img1)
	imagePath := fmt.Sprintf("%sframe_%04d.jpg", imageFolder, 0)
	gocv.IMWrite(imagePath, img1)
	//số frame 2 giây
	frameInterval := frameFPS * 2
	frameIndex := 0

	for {
		enoughFrames := true

		// Đọc khung hình tiếp theo từ video
		for i := 0; i < frameInterval; i++ {
			if ok := capture.Read(&img2); !ok {
				enoughFrames = false
				break
			}
		}

		if !enoughFrames {
			break
		}

		soA := 45
		decodedImg1 := matToImage(img1)
		decodedImg2 := matToImage(img2)

		similarity := compareImages(decodedImg1, decodedImg2, soA) * 100
		fmt.Printf("Phần trăm giống nhau theo Similarity %04d : %.2f%%\n", frameIndex, similarity)

		if similarity < 90 {
			img1 = img2.Clone()

			imagePath := fmt.Sprintf("%sframe_%04d.jpg", imageFolder, frameIndex)
			gocv.IMWrite(imagePath, img2)
		}

		frameIndex += frameInterval
	}
}

func compareImages(resizedImg1, resizedImg2 image.Image, soA int) float64 {
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

			// Convert soA to an float64
			soAInt := float64(soA)

			if diffr >= soAInt || diffg >= soAInt || diffb >= soAInt || diffa >= soAInt {
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
func matToImage(mat gocv.Mat) image.Image {
	img, _ := mat.ToImage()
	return img
}
