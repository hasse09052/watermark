package main

import (
	"image"
	"image/color"
	"math"
	"watermark/lib"

	"golang.org/x/exp/errors/fmt"
)

const (
	SQRT_AREA int = 16 //透かしを埋め込む領域数の平方根
	STRENGTH      = 100
)

/*
8bit * 64文字 = 512it
512 / 2 = 256[area]
256領域必要つまり16×16領域に分ければいい
*/

func main() {
	sourceImage := lib.InputImage("./Lenna.png")
	embedText := "KazukiHasegawa"

	imageSize := sourceImage.Bounds()
	partitionY := imageSize.Max.Y / SQRT_AREA
	partitionX := imageSize.Max.X / SQRT_AREA
	outputImage := image.NewRGBA(imageSize)

	var targetPixcels = make([][][]complex128, SQRT_AREA)
	for i := range targetPixcels {
		targetPixcels[i] = make([][]complex128, SQRT_AREA)
	}

	//画像をSQRT_AREA * SQRT_AREAに分割
	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			red, _, _, _ := sourceImage.At(y, x).RGBA()
			red = red >> 8
			targetPixcels[y/partitionY][x/partitionX] = append(targetPixcels[y/partitionY][x/partitionX], complex(float64(red), 0))
		}
	}
	fmt.Println(targetPixcels[0][0][:20])

	//埋め込むテキストを2進数で2bit毎に変換
	bitTexts := lib.MakeBitArray(embedText, int(math.Pow(float64(SQRT_AREA), 2)*2/8))
	fmt.Println(bitTexts)
	fmt.Println(len(bitTexts))

	for y := 0; y < len(targetPixcels); y++ {
		for x := 0; x < len(targetPixcels[y]); x++ {
			var maxValue float64 = 0
			var targetValue float64 = 0
			embedIndex := 0
			for index, value := range targetPixcels[y][x] {
				if real(value) >= maxValue {
					maxValue = real(value)
				}

				var condition int
				switch bitTexts[0] {
				case "00":
					condition = 0
				case "01":
					condition = 1
				case "10":
					condition = 2
				case "11":
					condition = 3
				}
				if index%4 == condition && real(value) >= targetValue {
					embedIndex = index
					targetValue = real(value)
				}
			}
			if len(bitTexts) != 0 {
				bitTexts = bitTexts[1:]
			}

			impulse := make([]complex128, len(targetPixcels[y][x]))
			impulse[embedIndex] = complex(maxValue-targetValue+STRENGTH, 0)
			targetPixcels[y][x] = lib.SmearTransform(targetPixcels[y][x])

			for i := range targetPixcels[y][x] {
				targetPixcels[y][x][i] += impulse[i]
			}
			targetPixcels[y][x] = lib.DesmearTransform(targetPixcels[y][x])

			//正規化
			for i := range targetPixcels[y][x] {
				normalization := real(targetPixcels[y][x][i]) + 0.5
				if normalization < 0 {
					targetPixcels[y][x][i] = 0
				} else if normalization > 255 {
					targetPixcels[y][x][i] = 255
				} else {
					targetPixcels[y][x][i] = complex(math.Floor(normalization), 0)
				}
			}
		}
	}

	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			_, g, b, a := sourceImage.At(y, x).RGBA()
			r := targetPixcels[y/partitionY][x/partitionX][0]
			targetPixcels[y/partitionY][x/partitionX] = targetPixcels[y/partitionY][x/partitionX][1:]

			color := color.RGBA{R: uint8(real(r)), G: uint8(g), B: uint8(b), A: uint8(a)}
			outputImage.Set(y, x, color)
		}
	}

	lib.OutputImage(outputImage, "./result.png")
	main2()
}

func main2() {
	sourceImage := lib.InputImage("./result.png")

	imageSize := sourceImage.Bounds()
	partitionY := imageSize.Max.Y / SQRT_AREA
	partitionX := imageSize.Max.X / SQRT_AREA

	var targetPixcels = make([][][]complex128, SQRT_AREA)
	for i := range targetPixcels {
		targetPixcels[i] = make([][]complex128, SQRT_AREA)
	}

	//画像をSQRT_AREA * SQRT_AREAに分割
	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			red, _, _, _ := sourceImage.At(y, x).RGBA()
			red = red >> 8
			targetPixcels[y/partitionY][x/partitionX] = append(targetPixcels[y/partitionY][x/partitionX], complex(float64(red), 0))
		}
	}

	decodeBitText := make([]string, 0)
	for y := 0; y < len(targetPixcels); y++ {
		for x := 0; x < len(targetPixcels[y]); x++ {
			targetPixcels[y][x] = lib.SmearTransform(targetPixcels[y][x])

			var maxValue float64 = 0
			embedIndex := 0
			for index, value := range targetPixcels[y][x] {
				if real(value) >= maxValue {
					maxValue = real(value)
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
