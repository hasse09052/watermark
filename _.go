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
	STRENGTH      = 400
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
	outputImage := image.NewRGBA(imageSize)

	var targetPixcels = make([][][]complex128, SQRT_AREA)
	for i := range targetPixcels {
		targetPixcels[i] = make([][]complex128, SQRT_AREA)
	}

	//画像をSQRT_AREA * SQRT_AREAに分割
	for row := imageSize.Min.Y; row < imageSize.Max.Y; row++ {
		for col := imageSize.Min.X; col < imageSize.Max.X; col++ {
			red, _, _, _ := sourceImage.At(col, row).RGBA()
			red = red >> 8
			targetPixcels[col/SQRT_AREA][row/SQRT_AREA] = append(targetPixcels[col/SQRT_AREA][row/SQRT_AREA], complex(float64(red), 0))
		}
	}
	fmt.Println(targetPixcels[0][0][:20])

	//埋め込むテキストを2進数で2bit毎に変換
	bitTexts := lib.MakeBitArray(embedText, int(math.Pow(float64(SQRT_AREA), 2)*2/8))
	fmt.Println(bitTexts)
	fmt.Println(len(bitTexts))

	for row := 0; row < len(targetPixcels); row++ {
		for col := 0; col < len(targetPixcels[row]); col++ {
			var maxValue float64 = 0
			var targetValue float64 = 0
			embedIndex := 0
			for index, value := range targetPixcels[row][col] {
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

			impulse := make([]complex128, len(targetPixcels[row][col]))
			impulse[embedIndex] = complex(maxValue-targetValue+STRENGTH, 0)
			impulse = lib.SmearTransform(impulse)

			for i := 0; i < len(targetPixcels[row][col]); i++ {
				targetPixcels[row][col][i] += impulse[i]
			}
		}
	}

	for row := imageSize.Min.Y; row < imageSize.Max.Y; row++ {
		for col := imageSize.Min.X; col < imageSize.Max.X; col++ {
			_, g, b, a := sourceImage.At(col, row).RGBA()
			r := targetPixcels[col/SQRT_AREA][row/SQRT_AREA][0]
			targetPixcels[col/SQRT_AREA][row/SQRT_AREA] = targetPixcels[col/16][row/16][1:]

			color := color.RGBA{R: uint8(real(r)), G: uint8(g), B: uint8(b), A: uint8(a)}
			outputImage.Set(col, row, color)
		}
	}

	lib.OutputImage(outputImage, "./result.png")
	main2()
}

func main2() {
	sourceImage := lib.InputImage("./result.png")

	imageSize := sourceImage.Bounds()

	var targetPixcels = make([][][]complex128, SQRT_AREA)
	for i := range targetPixcels {
		targetPixcels[i] = make([][]complex128, SQRT_AREA)
	}

	//画像をSQRT_AREA * SQRT_AREAに分割
	for row := imageSize.Min.Y; row < imageSize.Max.Y; row++ {
		for col := imageSize.Min.X; col < imageSize.Max.X; col++ {
			red, _, _, _ := sourceImage.At(col, row).RGBA()
			red = red >> 8
			targetPixcels[col/SQRT_AREA][row/SQRT_AREA] = append(targetPixcels[col/SQRT_AREA][row/SQRT_AREA], complex(float64(red), 0))
		}
	}

	decodeBitText := make([]string, 0)
	for row := 0; row < len(targetPixcels); row++ {
		for col := 0; col < len(targetPixcels[row]); col++ {
			targetPixcels[row][col] = lib.DesmearTransform(targetPixcels[row][col])

			var maxValue float64 = 0
			embedIndex := 0
			for index, value := range targetPixcels[row][col] {
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
	// 		r := targetPixcels[col/SQRT_AREA][row/SQRT_AREA][0]
	// 		targetPixcels[col/SQRT_AREA][row/SQRT_AREA] = targetPixcels[col/16][row/16][1:]

	// 		color := color.RGBA{R: uint8(real(r)), G: uint8(g), B: uint8(b), A: uint8(a)}
	// 		outputImage.Set(col, row, color)
	// 	}
	// }
}
