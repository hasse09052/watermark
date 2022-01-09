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

type pixcelData struct {
	x int
	y int
	r complex128
	g complex128
	b complex128
	a complex128
}

/*
8bit * 64文字 = 512it
512 / 2 = 256[area]
256領域必要つまり16×16領域に分ければいい
*/

func GetFieldArray(pixcelDatas []pixcelData, fieldName string) []complex128 {
	result := make([]complex128, 0)
	for _, pixcelData := range pixcelDatas {
		var fieldValue complex128
		switch fieldName {
		case "R":
			fieldValue = pixcelData.r
		case "G":
			fieldValue = pixcelData.g
		case "B":
			fieldValue = pixcelData.b
		}
		result = append(result, fieldValue)
	}
	return result
}

func normalizationPixcelData(pixcelDatas []pixcelData) []pixcelData {
	for i, pixcelData := range pixcelDatas {
		rgbValues := map[string]float64{"r": real(pixcelData.r), "g": real(pixcelData.g), "b": real(pixcelData.b)}
		for k := range rgbValues {
			rgbValues[k] += 0.5
			if rgbValues[k] < 0 {
				rgbValues[k] = 0
			} else if rgbValues[k] > 255 {
				rgbValues[k] = 255
			} else {
				rgbValues[k] = math.Floor(rgbValues[k])
			}
		}

		pixcelDatas[i].r = complex(rgbValues["r"], 0)
		pixcelDatas[i].g = complex(rgbValues["g"], 0)
		pixcelDatas[i].b = complex(rgbValues["b"], 0)
	}
	return pixcelDatas
}

func main() {
	sourceImage := lib.InputImage("./Lenna.png")
	embedText := "KazukiHasegawa"

	imageSize := sourceImage.Bounds()
	outputImage := image.NewRGBA(imageSize)
	var embedPixcels = make([]pixcelData, 0)

	//画像のRGBを格納
	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			r, g, b, a := sourceImage.At(y, x).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8
			embedPixcels = append(embedPixcels, pixcelData{x, y, complex(float64(r), 0), complex(float64(g), 0), complex(float64(b), 0), complex(float64(a), 0)})
		}
	}

	//埋め込むテキストを2進数で2bit毎に変換
	bitTexts := lib.MakeBitArray(embedText)
	fmt.Println(bitTexts)
	fmt.Println(len(bitTexts))

	for i := 0; i < len(embedPixcels)/PIXCEL_PER_BLOCK; i++ {
		targetPixcels := embedPixcels[i*PIXCEL_PER_BLOCK : (i+1)*PIXCEL_PER_BLOCK]

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
			value := real(targetPixcel.r + targetPixcel.g + targetPixcel.b)
			if value > maxValue {
				maxValue = value
			}

			if value >= targetMaxValue && index%4 == condition {
				targetMaxValue = value
				embedIndex = index
			}
		}

		//RGBへの埋め込み
		for _, fieldName := range []string{"R", "G", "B"} {
			fieldArray := GetFieldArray(targetPixcels, fieldName)
			fieldArray = lib.SmearTransform(fieldArray)
			fieldArray[embedIndex] += complex((maxValue-targetMaxValue)/3+STRENGTH, 0)
			fieldArray = lib.DesmearTransform(fieldArray)
			for i := range targetPixcels {
				switch fieldName {
				case "R":
					targetPixcels[i].r = fieldArray[i]
				case "G":
					targetPixcels[i].g = fieldArray[i]
				case "B":
					targetPixcels[i].b = fieldArray[i]
				}
			}
		}
	}

	//正規化
	embedPixcels = normalizationPixcelData(embedPixcels)

	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			_, _, _, a := sourceImage.At(y, x).RGBA()
			r := real(embedPixcels[0].r)
			g := real(embedPixcels[0].g)
			b := real(embedPixcels[0].b)
			embedPixcels = embedPixcels[1:]

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
	var embedPixcels = make([]pixcelData, 0)

	//画像のRGBを格納
	for y := 0; y < imageSize.Max.Y; y++ {
		for x := 0; x < imageSize.Max.X; x++ {
			r, g, b, a := sourceImage.At(y, x).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8
			embedPixcels = append(embedPixcels, pixcelData{x, y, complex(float64(r), 0), complex(float64(g), 0), complex(float64(b), 0), complex(float64(a), 0)})
		}
	}

	decodeBitText := make([]string, 0)
	for i := 0; i < len(embedPixcels)/PIXCEL_PER_BLOCK; i++ {
		targetPixcels := embedPixcels[i*PIXCEL_PER_BLOCK : (i+1)*PIXCEL_PER_BLOCK]
		fieldArrayR := lib.SmearTransform(GetFieldArray(targetPixcels, "R"))
		fieldArrayG := lib.SmearTransform(GetFieldArray(targetPixcels, "G"))
		fieldArrayB := lib.SmearTransform(GetFieldArray(targetPixcels, "B"))

		var maxValue float64 = 0
		embedIndex := 0
		for i := range targetPixcels {
			value := real(fieldArrayR[i] + fieldArrayG[i] + fieldArrayB[i])
			if value >= maxValue {
				maxValue = value
				embedIndex = i
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
