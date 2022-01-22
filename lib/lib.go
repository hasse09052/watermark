package lib

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"math/cmplx"
	"os"

	"github.com/mjibson/go-dsp/fft"
)

func InputImage(filePath string) image.Image {
	inputFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println("open error")
	}

	image, _, err := image.Decode(inputFile)

	return image
}

func OutputImage(image image.Image, filePath string) {
	outputFile, err := os.Create(filePath)
	defer outputFile.Close()
	if err != nil {
		fmt.Println("create error")
	}

	err = png.Encode(outputFile, image)
	if err != nil {
		fmt.Println("encode error")
	}
}

func MatrixDot(a []complex128, b [][]complex128) []complex128 {
	// a := []complex128{1, 2, 3}
	// b := [][]complex128{
	// 	{1, 2, 3},
	// 	{4, 5, 6},
	// 	{7, 8, 9},
	// }

	result := make([]complex128, 0)

	for i := 0; i < len(a); i++ {
		var sum complex128 = 0
		for j := 0; j < len(b); j++ {
			sum += a[j] * b[j][i]
		}
		result = append(result, sum)
	}

	return result
}

func SmearTransform(data []complex128) []complex128 {
	alpha := 1
	pi := math.Pi
	L := len(data)
	var LL int

	if L%2 == 0 {
		LL = L / 2
	} else {
		LL = (L - 1) / 2
	}

	theta := make([]complex128, L)
	for k := 0; k < LL; k++ {
		theta[k] = complex(2*pi*float64(alpha)*(1/float64(L))*math.Pow(float64(k), 2), 0)
	}
	for k := LL; k < L; k++ {
		theta[k] = complex(-2*pi*float64(alpha)*(1/float64(L))*math.Pow(float64((L-k)), 2), 0)
	}

	s := make([][]complex128, L)
	for k := 0; k < L; k++ {
		s[k] = make([]complex128, L)
		s[k][k] = cmplx.Exp(1i * theta[k])
	}

	return fft.IFFT(MatrixDot(fft.FFT(data), s))
}

func DesmearTransform(data []complex128) []complex128 {
	alpha := 1
	pi := math.Pi
	L := len(data)
	var LL int

	if L%2 == 0 {
		LL = L / 2
	} else {
		LL = (L - 1) / 2
	}

	theta := make([]complex128, L)
	for k := 0; k < LL; k++ {
		theta[k] = complex(2*pi*float64(alpha)*(1/float64(L))*math.Pow(float64(k), 2), 0)
	}
	for k := LL; k < L; k++ {
		theta[k] = complex(-2*pi*float64(alpha)*(1/float64(L))*math.Pow(float64((L-k)), 2), 0)
	}

	s := make([][]complex128, L)
	for k := 0; k < L; k++ {
		s[k] = make([]complex128, L)
		s[k][k] = cmplx.Exp(-1i * theta[k])
	}

	return fft.IFFT(MatrixDot(fft.FFT(data), s))
}

//textを2進数に変換後、2bitづつに分解
func MakeBitArray(text string) []string {
	bitTexts := make([]string, 0)
	for _, character := range text {
		bitTexts = append(bitTexts, fmt.Sprintf("%08b", character))
	}
	fmt.Println(bitTexts)

	result := make([]string, 0)
	for _, v := range bitTexts {
		for i := 0; i < 4; i++ {
			result = append(result, v[:2])
			v = v[2:]
		}
	}

	return result
}
