package main

import (
	"image"
	"image/color"
	"math"
	"watermark/lib"

	"golang.org/x/exp/errors/fmt"
)

const (
	PIXCEL_PER_BLOCK = 256
	STRENGTH         = 200
)

func main() {
	sourceImage := lib.InputImage("./Lenna.png")
	embedText := "KazukiHasegawa"

	imageSize := sourceImage.Bounds()
	outputImage := image.NewRGBA(imageSize)
	var OneDimensionalImage = make([]complex128, 0)

	//画像のRGBを格納
	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			_, _, b, _ := sourceImage.At(y, x).RGBA()
			b = b >> 8
			OneDimensionalImage = append(OneDimensionalImage, complex(float64(b), 0))
		}
	}

	//埋め込むテキストを2進数で2bit毎に変換
	bitTexts := lib.MakeBitArray(embedText)
	fmt.Println(bitTexts)
	fmt.Println(len(bitTexts))

	var embedPixcels = make([]complex128, 0)
	for i := 0; i < len(OneDimensionalImage)/PIXCEL_PER_BLOCK; i++ {
		var targetPixcels []complex128
		if (i+1)*PIXCEL_PER_BLOCK > len(OneDimensionalImage) {
			targetPixcels = OneDimensionalImage[i*PIXCEL_PER_BLOCK:]
		} else {
			targetPixcels = OneDimensionalImage[i*PIXCEL_PER_BLOCK : (i+1)*PIXCEL_PER_BLOCK]
		}
		targetPixcels = lib.SmearTransform(targetPixcels)

		var bitText string
		if len(bitTexts) > 0 {
			bitText = bitTexts[0]
			bitTexts = bitTexts[1:]
		} else {
			bitText = "00"
		}
		var condition int
		switch bitText {
		case "00":
			condition = 0
		case "01":
			condition = 1
		case "10":
			condition = 2
		case "11":
			condition = 3
		}

		var maxValue float64 = 0
		var targetMaxValue float64 = 0
		embedIndex := 0
		for index, targetPixcel := range targetPixcels {
			if real(targetPixcel) > maxValue {
				maxValue = real(targetPixcel)
			}

			if real(targetPixcel) >= targetMaxValue && index%4 == condition {
				targetMaxValue = real(targetPixcel)
				embedIndex = index
			}
		}

		targetPixcels[embedIndex] += complex(maxValue-targetMaxValue+STRENGTH, 0)

		targetPixcels = lib.DesmearTransform(targetPixcels)

		for index, targetPixcel := range targetPixcels {
			targetPixcel += complex(0.5, 0)
			if real(targetPixcel) > 255 {
				targetPixcels[index] = 255
			} else if real(targetPixcel) < 0 {
				targetPixcels[index] = 0
			} else {
				targetPixcels[index] = complex(math.Floor(real(targetPixcel)), 0)
			}
		}

		embedPixcels = append(embedPixcels, targetPixcels...)
	}

	fmt.Println(imageSize.Max.Y * imageSize.Max.X)
	fmt.Println(len(embedPixcels))
	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			r, g, b, a := sourceImage.At(y, x).RGBA()
			if len(embedPixcels) > 0 {
				b = uint32(real(embedPixcels[0]))
				embedPixcels = embedPixcels[1:]
			}

			color := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
			outputImage.Set(y, x, color)
		}
	}

	lib.OutputImage(outputImage, "./result.png")
	main2()
}

func main2() {
	sourceImage := lib.InputImage("./result.png")

	imageSize := sourceImage.Bounds()
	var embedPixcels = make([]complex128, 0)

	//画像のRGBを格納
	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			_, _, b, _ := sourceImage.At(y, x).RGBA()
			b = b >> 8
			embedPixcels = append(embedPixcels, complex(float64(b), 0))
		}
	}

	decodeBitText := make([]string, 0)
	for i := 0; i < len(embedPixcels)/PIXCEL_PER_BLOCK; i++ {
		targetPixcels := embedPixcels[i*PIXCEL_PER_BLOCK : (i+1)*PIXCEL_PER_BLOCK]
		targetPixcels = lib.SmearTransform(targetPixcels)

		var maxValue float64 = 0
		embedIndex := 0
		for index, targetPixcels := range targetPixcels {
			if real(targetPixcels) >= maxValue {
				maxValue = real(targetPixcels)
				embedIndex = index
			}
		}

		var bit string
		switch embedIndex % 4 {
		case 0:
			bit = "00"
		case 1:
			bit = "01"
		case 2:
			bit = "10"
		case 3:
			bit = "11"
		}
		decodeBitText = append(decodeBitText, bit)
	}

	fmt.Println(decodeBitText)
	fmt.Println(len(decodeBitText))

	// for row := imageSize.Min.Y; row < imageSize.Max.Y; row++ {
	// 	for col := imageSize.Min.X; col < imageSize.Max.X; col++ {
	// 		_, g, b, a := sourceImage.At(col, row).RGBA()
	// 		r := targetPixcels[y/partitionY][x/partitionX][0]
	// 		targetPixcels[y/partitionY][x/partitionX] = targetPixcels[col/16][row/16][1:]

	// 		color := color.RGBA{R: uint8(real(r)), G: uint8(g), B: uint8(b), A: uint8(a)}
	// 		outputImage.Set(col, row, color)
	// 	}
	// }
}
