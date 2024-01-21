package Netpbm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// PBM represents a structure to hold PBM image data and attributes.
type PBM struct {
	data          [][]bool // 2D slice to store the pixel values.
	width, height int      // Width and height of the image.
	magicNumber   string   // Magic number indicating PBM format (P1 for ASCII, P4 for Binary).
}

// ReadPBM reads a PBM file and returns a PBM struct and an error if any.
func ReadPBM(filename string) (*PBM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read and validate the magic number.
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber)
	if magicNumber != "P1" && magicNumber != "P4" {
		return nil, fmt.Errorf("invalid magic number: %s", magicNumber)
	}

	// Read and parse image dimensions.
	dimensions, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading dimensions: %v", err)
	}
	var width, height int
	_, err = fmt.Sscanf(strings.TrimSpace(dimensions), "%d %d", &width, &height)
	if err != nil {
		return nil, fmt.Errorf("invalid dimensions: %v", err)
	}

	data := make([][]bool, height)

	for i := range data {
		data[i] = make([]bool, width)
	}

	// Handle P1 format (ASCII).
	if magicNumber == "P1" {
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			for x, field := range fields {
				if x >= width {
					return nil, fmt.Errorf("index out of range at row %d", y)
				}
				data[y][x] = field == "1"
			}
		}

	} else if magicNumber == "P4" {
		// Handle P4 format (binary).
		expectedBytesPerRow := (width + 7) / 8

		for y := 0; y < height; y++ {
			row := make([]byte, expectedBytesPerRow)
			n, err := reader.Read(row)
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected end of file at row %d", y)
				}
				return nil, fmt.Errorf("error reading pixel data at row %d: %v", y, err)
			}

			if n < expectedBytesPerRow {
				return nil, fmt.Errorf("unexpected end of file at row %d, expected %d bytes, got %d", y, expectedBytesPerRow, n)
			}

			for x := 0; x < width; x++ {
				byteIndex := x / 8
				bitIndex := 7 - (x % 8)
				bitValue := (int(row[byteIndex]) >> bitIndex) & 1
				data[y][x] = bitValue != 0
			}
		}
	}

	return &PBM{data, width, height, magicNumber}, nil
}

// Size returns the width and height of the PBM image.
func (pbm *PBM) Size() (int, int) {
	return pbm.width, pbm.height
}

// At returns the value of the pixel at the given coordinates.
func (pbm *PBM) At(x, y int) bool {
	return pbm.data[y][x]
}

// Set sets the value of the pixel at the given coordinates.
func (pbm *PBM) Set(x, y int, value bool) {
	pbm.data[y][x] = value
}

// Save writes the PBM image to a file in the specified format (P1 or P4).
func (pbm *PBM) Save(filename string) error {
	if pbm == nil {
		return errors.New("cannot save a nil PBM")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write magic number, width, and height.
	fmt.Fprintf(file, "%s\n%d %d\n", pbm.magicNumber, pbm.width, pbm.height)

	// Save in the appropriate format based on the magic number.
	if pbm.magicNumber == "P1" {
		return pbm.saveP1(file)
	} else if pbm.magicNumber == "P4" {
		return pbm.saveP4(file)
	} else {
		return fmt.Errorf("unsupported magic number: %s", pbm.magicNumber)
	}
}

// saveP1 saves the PBM image in P1 format (ASCII).
func (pbm *PBM) saveP1(file *os.File) error {
	for y := 0; y < pbm.height; y++ {
		for x := 0; x < pbm.width; x++ {
			if pbm.data[y][x] {
				fmt.Fprint(file, "1")
			} else {
				fmt.Fprint(file, "0")
			}
			if x < pbm.width-1 {
				fmt.Fprint(file, " ")
			}
		}
		fmt.Fprintln(file)
	}
	return nil
}

// saveP4 saves the PBM image in P4 format (binary).
func (pbm *PBM) saveP4(file *os.File) error {
	expectedBytesPerRow := (pbm.width + 7) / 8
	for y := 0; y < pbm.height; y++ {
		row := make([]byte, expectedBytesPerRow)
		for x := 0; x < pbm.width; x++ {
			byteIndex := x / 8
			bitIndex := 7 - (x % 8)
			if pbm.data[y][x] {
				row[byteIndex] |= 1 << bitIndex
			}
		}
		_, err := file.Write(row)
		if err != nil {
			return fmt.Errorf("error writing pixel data at row %d: %v", y, err)
		}
	}
	return nil
}

// Invert inverts the colors of the PBM image by flipping pixel values.
func (pbm *PBM) Invert() {
	for y := 0; y < pbm.height; y++ {
		for x := 0; x < pbm.width; x++ {
			pbm.data[y][x] = !pbm.data[y][x]
		}
	}
}

// Flip performs a horizontal flip of the PBM image.
func (pbm *PBM) Flip() {
	for y := 0; y < pbm.height; y++ {
		for x := 0; x < pbm.width/2; x++ {
			pbm.data[y][x], pbm.data[y][pbm.width-x-1] = pbm.data[y][pbm.width-x-1], pbm.data[y][x]
		}
	}
}

// Flop performs a vertical flip of the PBM image.
func (pbm *PBM) Flop() {
	for y := 0; y < pbm.height/2; y++ {
		pbm.data[y], pbm.data[pbm.height-y-1] = pbm.data[pbm.height-y-1], pbm.data[y]
	}
}

// SetMagicNumber updates the magic number of the PBM image (P1 or P4).
func (pbm *PBM) SetMagicNumber(magicNumber string) {
	pbm.magicNumber = magicNumber
}
