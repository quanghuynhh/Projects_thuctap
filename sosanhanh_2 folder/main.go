package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"

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

		phantram, err := strconv.Atoi(c.FormValue("phantram"))
		if err != nil {
			// Nếu không có giá trị 'soa' trong POST data, thì giá trị mặc định là 45
			phantram = 45
		}
		//folder 1
		folderPath1 := c.FormValue("folder_path1")
		if folderPath1 == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Vui lòng cung cấp đường dẫn thư mục",
			})
		}

		if _, err := os.Stat(folderPath1); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Thư mục không tồn tại",
			})
		}

		imagePaths1, err := getImagePaths(folderPath1)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Lỗi khi đọc thư mục",
			})
		}
		response1 := fiber.Map{
			"folder_path1": folderPath1,
			"file_info1":   imagePaths1,
		}

		//folder 2
		folderPath2 := c.FormValue("folder_path2")
		if folderPath2 == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Vui lòng cung cấp đường dẫn thư mục",
			})
		}

		if _, err := os.Stat(folderPath2); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Thư mục không tồn tại",
			})
		}

		imagePaths2, err := getImagePaths(folderPath2)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Lỗi khi đọc thư mục",
			})
		}
		response2 := fiber.Map{
			"folder_path2": folderPath2,
			"file_info2":   imagePaths2,
		}

		// Tạo file để ghi dữ liệu
		filePath := "./textluu/result.txt"
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatal("Không thể tạo file:", err)
		}
		defer file.Close()

		// Tạo file để ghi dữ liệu
		filePath2 := "./textluu/tenanh.txt"
		file2, err := os.Create(filePath2)
		if err != nil {
			log.Fatal("Không thể tạo file:", err)
		}
		defer file2.Close()

		for i := 0; i < len(imagePaths1); i++ {
			imagePath1 := imagePaths1[i]
			imagePath2 := imagePaths2[0]

			/// Tên ảnh
			imageName1 := filepath.Base(imagePath1)

			// Xuất tên ảnh và số thứ tự ảnh
			fmt.Printf("Image: %s\n", imageName1)
			fmt.Fprintf(file, "Image %d: %s\n", i+1, imageName1)

			/// Tên ảnh
			imageName2 := filepath.Base(imagePath2)

			// Xuất tên ảnh và số thứ tự ảnh
			fmt.Printf("Image: %s\n", imageName2)
			fmt.Fprintf(file, "Image %d: %s\n", 0, imageName2)

			// calulate
			img1, err := readImage(imagePath1)
			if err != nil {
				log.Printf("Error reading image %s: %v\n", imagePath1, err)
				continue
			}

			img2, err := readImage(imagePath2)
			if err != nil {
				log.Printf("Error reading image %s: %v\n", imagePath2, err)
				continue
			}

			similarity := compareImages(img1, img2, soA) * 100
			fmt.Printf("Phần trăm giỗng nhau theo Similarity: %.2f%%\n", similarity)
			fmt.Fprintf(file, "Similarity: %.2f%%\n", similarity)
			phantram := float64(phantram)
			if similarity >= phantram {
				fmt.Fprintf(file2, "Image %d: %s\n", i+1, imageName1)

			}

		}
		// Combine the responses into a single JSON response
		response := fiber.Map{
			"folder1": response1,
			"folder2": response2,
		}

		return c.JSON(response)

	})

	app.Listen(":3000")
}

func getImagePaths(folderPath string) ([]string, error) {
	var imagePaths []string

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			imagePath := filepath.Join(folderPath, file.Name())
			imagePaths = append(imagePaths, imagePath)
		}
	}

	return imagePaths, nil
}

func isImageFile(fileName string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif"}

	for _, ext := range extensions {
		if filepath.Ext(fileName) == ext {
			return true
		}
	}

	return false
}
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
