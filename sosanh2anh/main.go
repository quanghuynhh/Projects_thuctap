package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Nr90/imgsim"
	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Fiber instance
	app := fiber.New(fiber.Config{
		// Set the maximum request body size to 10MB (or any size you prefer)
		BodyLimit: 10 * 1024 * 1024,
	})
	app.Use(cors.New())
	app.Static("/hinhanh", "./textluu")

	// Routes
	app.Post("/", func(c *fiber.Ctx) error {

		soA, err := strconv.Atoi(c.FormValue("thamso"))
		if err != nil {
			// Nếu không có giá trị 'soa' trong POST data, thì giá trị mặc định là 45
			soA = 45
		}

		// Kiểm tra giới hạn cho số A (1 <= soA <= 100)
		if soA < 1 || soA > 255 {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid input. 'soA' must be between 1 and 100.",
			})
		}

		// Get first file from form field "imagePath1":
		file1, err := c.FormFile("imagePath1")
		if err != nil {
			return err
		}

		// Get second file from form field "imagePath2":
		file2, err := c.FormFile("imagePath2")
		if err != nil {
			return err
		}

		// Save the uploaded files
		savePath1 := "./textluu/" + file1.Filename
		if err := c.SaveFile(file1, savePath1); err != nil {
			log.Println("Failed to save file1:", err)
			return handleError(c, "Failed to save file1", http.StatusInternalServerError)
		}

		savePath2 := "./textluu/" + file2.Filename
		if err := c.SaveFile(file2, savePath2); err != nil {
			log.Println("Failed to save file2:", err)
			return handleError(c, "Failed to save file2", http.StatusInternalServerError)
		}

		// Read the uploaded images
		img1, err := readImage(savePath1)
		if err != nil {
			log.Println("Failed to read image1:", err)
			return handleError(c, "Failed to read image1", http.StatusInternalServerError)
		}

		img2, err := readImage(savePath2)
		if err != nil {
			log.Println("Failed to read image2:", err)
			return handleError(c, "Failed to read image2", http.StatusInternalServerError)
		}

		// Compare the images
		dhash1 := imgsim.DifferenceHash(img1)
		fmt.Println("dhash1:", dhash1)

		dhash2 := imgsim.DifferenceHash(img2)
		fmt.Println("dhash2:", dhash2)

		diffBits := countDifferentBits(dhash1, dhash2)
		fmt.Println("Number of different bits:", diffBits)

		differencePercentage := float64((64 - diffBits) * 100.0 / 64)
		fmt.Println("Phần trăm giống nhau theo DifferenceHash:", differencePercentage)

		similarity := compareImages(img1, img2, soA) * 100
		fmt.Printf("Phần trăm giống nhau theo Similarity: %.2f%%\n", similarity)

		var resultData map[string]interface{}

		if similarity != 100 {
			// Tạo hình ảnh mới để vẽ các vùng khác biệt
			resultImg := image.NewRGBA(img1.Bounds())
			dc := gg.NewContextForRGBA(resultImg)
			dc.DrawImage(img1, 0, 0)
			width := resultImg.Bounds().Dx()
			height := resultImg.Bounds().Dy()

			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					r1, g1, b1, _ := img1.At(x, y).RGBA()
					r2, g2, b2, _ := img2.At(x, y).RGBA()

					diffr := sqDiffUInt32(r1, r2)
					diffg := sqDiffUInt32(g1, g2)
					diffb := sqDiffUInt32(b1, b2)

					soAInt := float64(soA)

					if diffr >= soAInt || diffg >= soAInt || diffb >= soAInt {
						dc.SetColor(color.RGBA{R: 255, A: 255})
						dc.DrawRectangle(float64(x), float64(y), 1, 1)
						dc.Stroke()
					}
				}
			}

			outputPath := "./textluu/simi_result.png"
			if err := saveImage(outputPath, resultImg); err != nil {
				log.Println("Failed to save result image:", err)
				return handleError(c, "Failed to save result image", http.StatusInternalServerError)
			}

			fmt.Println("Result image saved successfully.")
			resultImageUrl := "/hinhanh/simi_result.png"
			resizedImg1Url := fmt.Sprintf("/hinhanh/%s", file1.Filename)
			resizedImg2Url := fmt.Sprintf("/hinhanh/%s", file2.Filename)
			similarityPercentage := fmt.Sprintf("%.2f", similarity)
			resultData = map[string]interface{}{
				"Phần trăm giống nhau theo DifferenceHash": differencePercentage,
				"Phần trăm giống nhau theo Similarity":     similarityPercentage,
				"sai số màu Similarity":                    soA,
				"ảnh khác theo Similarity ":                resultImageUrl,
				"ảnh 1":                                    resizedImg1Url,
				"ảnh 2":                                    resizedImg2Url,
			}
		}

		if differencePercentage != 100 {
			resultImg := highlightDifferences(img1, dhash1, dhash2)
			diffImagePath := "./textluu/diff_result.png"
			if err := saveImage(diffImagePath, resultImg); err != nil {
				return handleError(c, "Failed to save result image", http.StatusInternalServerError)
			}
			diffImageUrl := "/hinhanh/diff_result.png"
			resultData["ảnh khác theo DifferenceHash"] = diffImageUrl
		}

		if resultData == nil {
			resizedImg1Url := fmt.Sprintf("/hinhanh/%s", file1.Filename)
			resizedImg2Url := fmt.Sprintf("/hinhanh/%s", file2.Filename)
			similarityPercentage := fmt.Sprintf("%.2f", similarity)
			resultData = map[string]interface{}{
				"Phần trăm giống nhau theo DifferenceHash": differencePercentage,
				"Phần trăm giống nhau theo Similarity":     similarityPercentage,
				"sai số màu Similarity":                    soA,
				"thông báo":                                "Hai ảnh giống nhau",
				"ảnh 1":                                    resizedImg1Url,
				"ảnh 2":                                    resizedImg2Url,
			}
		}

		// Return JSON response with combined data from both algorithms
		return c.JSON(fiber.Map{
			"status":  http.StatusOK,
			"message": "Images compared successfully",
			"data":    resultData,
		})
	})

	// Start server
	log.Fatal(app.Listen(":3000"))
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

// Read an image from the file path
func readImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// Save an image to a specific path
func saveImage(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch filepath.Ext(path) {
	case ".jpeg", ".jpg":
		return jpeg.Encode(file, img, nil)
	case ".png":
		return png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported image format")
	}
}

func sqDiffUInt32(x, y uint32) float64 {

	x >>= 8
	y >>= 8
	return math.Abs(float64(y) - float64(x))
}
func handleError(c *fiber.Ctx, message string, statusCode int) error {
	log.Println(message)
	return c.JSON(fiber.Map{
		"status":  statusCode,
		"message": "Server error",
		"data":    nil,
	})
}

func countDifferentBits(dhash1, dhash2 imgsim.Hash) int {
	diffBits := 0
	for i := 0; i < 64; i++ {
		bit1 := (dhash1 >> uint(63-i)) & 1
		bit2 := (dhash2 >> uint(63-i)) & 1
		fmt.Printf("Bit1: %d, Bit2: %d\n", bit1, bit2) // In giá trị của bit1 và bit2
		if bit1 != bit2 {
			diffBits++
		}
	}
	return diffBits
}

// ...

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
