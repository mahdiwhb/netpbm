package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// PGM represents a structure to hold PGM image data and attributes.
type PGM struct {
	data          [][]uint8 // 2D slice to store the pixel values.
	width, height int       // Width and height of the image.
	magicNumber   string    // Magic number indicating PGM format (P2 for ASCII, P5 for Binary).
	max           uint8     // Maximum grayscale value.
}

// ReadPGM reads a PGM file and returns a PGM struct and an error if any.
func ReadPGM(filename string) (*PGM, error) {
	// Open the file for reading.
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
	if magicNumber != "P2" && magicNumber != "P5" {
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
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width and height must be positive")
	}

	// Read and validate max grayscale value.
	maxValue, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading max value: %v", err)
	}
	maxValue = strings.TrimSpace(maxValue)
	var max uint8
	_, err = fmt.Sscanf(maxValue, "%d", &max)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %v", err)
	}

	// Read and store image data based on PGM format.
	data := make([][]uint8, height)
	expectedBytesPerPixel := 1

	// Handle P2 format (ASCII).
	if magicNumber == "P2" {
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			rowData := make([]uint8, width)
			for x, field := range fields {
				if x >= width {
					return nil, fmt.Errorf("index out of range at row %d", y)
				}
				var pixelValue uint8
				_, err := fmt.Sscanf(field, "%d", &pixelValue)
				if err != nil {
					return nil, fmt.Errorf("error parsing pixel value at row %d, column %d: %v", y, x, err)
				}
				rowData[x] = pixelValue
			}
			data[y] = rowData
		}
	} else if magicNumber == "P5" {
		// Handle P5 format (binary).
		for y := 0; y < height; y++ {
			row := make([]byte, width*expectedBytesPerPixel)
			n, err := reader.Read(row)
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected end of file at row %d", y)
				}
				return nil, fmt.Errorf("error reading pixel data at row %d: %v", y, err)
			}
			if n < width*expectedBytesPerPixel {
				return nil, fmt.Errorf("unexpected end of file at row %d, expected %d bytes, got %d", y, width*expectedBytesPerPixel, n)
			}

			rowData := make([]uint8, width)
			for x := 0; x < width; x++ {
				pixelValue := uint8(row[x*expectedBytesPerPixel])
				rowData[x] = pixelValue
			}
			data[y] = rowData
		}
	}

	// Construct and return the PGM struct.
	return &PGM{data, width, height, magicNumber, max}, nil
}

// Size returns the width and height of the PGM image.
func (pgm *PGM) Size() (int, int) {
	return pgm.width, pgm.height
}

// At returns the pixel value at the given coordinates.
func (pgm *PGM) At(x, y int) uint8 {
	if x >= 0 && x < pgm.width && y >= 0 && y < pgm.height {
		return pgm.data[y][x]
	}
	return 0
}

// Set sets the pixel value at the given coordinates.
func (pgm *PGM) Set(x, y int, value uint8) {
	if x >= 0 && x < pgm.width && y >= 0 && y < pgm.height {
		pgm.data[y][x] = value
	}
}

// Save writes the PGM image to a file, converting between P2 and P5 formats if necessary.
func (pgm *PGM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = fmt.Fprintln(writer, pgm.magicNumber)
	if err != nil {
		return fmt.Errorf("error writing magic number: %v", err)
	}

	// Write image dimensions.
	_, err = fmt.Fprintf(writer, "%d %d\n", pgm.width, pgm.height)
	if err != nil {
		return fmt.Errorf("error writing dimensions: %v", err)
	}

	// Write max grayscale value.
	_, err = fmt.Fprintln(writer, pgm.max)
	if err != nil {
		return fmt.Errorf("error writing max value: %v", err)
	}
	for _, row := range pgm.data {
		if len(row) != pgm.width {
			return fmt.Errorf("inconsistent row length in data")
		}
	}

	// Write pixel data in the specified PGM format.
	if pgm.magicNumber == "P2" {
		err = saveP2PGM(writer, pgm)
		if err != nil {
			return err
		}
	} else if pgm.magicNumber == "P5" {
		err = saveP5PGM(writer, pgm)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// saveP2PGM saves the image in P2 format (ASCII) to the provided writer.
func saveP2PGM(file *bufio.Writer, pgm *PGM) error {
	for y := 0; y < pgm.height; y++ {
		for x := 0; x < pgm.width; x++ {
			_, err := fmt.Fprint(file, pgm.data[y][x])
			if err != nil {
				return fmt.Errorf("error writing pixel data at row %d, column %d: %v", y, x, err)
			}
			if x < pgm.width-1 {
				_, err = fmt.Fprint(file, " ")
				if err != nil {
					return fmt.Errorf("error writing space after pixel at row %d, column %d: %v", y, x, err)
				}
			}
		}
		_, err := fmt.Fprintln(file)
		if err != nil {
			return fmt.Errorf("error writing newline after row %d: %v", y, err)
		}
	}
	return nil
}

// saveP5PGM saves the image in P5 format (binary) to the provided writer.
func saveP5PGM(file *bufio.Writer, pgm *PGM) error {
	for y := 0; y < pgm.height; y++ {
		row := make([]byte, pgm.width)
		for x := 0; x < pgm.width; x++ {
			row[x] = byte(pgm.data[y][x])
		}
		_, err := file.Write(row)
		if err != nil {
			return fmt.Errorf("error writing pixel data at row %d: %v", y, err)
		}
	}
	return nil
}

// Invert inverts the colors of the image by subtracting pixel values from the max value.
func (pgm *PGM) Invert() {
	for i := range pgm.data {
		for j := range pgm.data[i] {
			pgm.data[i][j] = uint8(pgm.max) - pgm.data[i][j]
		}
	}
}

// Flip performs a horizontal flip of the image.
func (pgm *PGM) Flip() {
	for i := range pgm.data {
		for j, k := 0, len(pgm.data[i])-1; j < k; j, k = j+1, k-1 {
			pgm.data[i][j], pgm.data[i][k] = pgm.data[i][k], pgm.data[i][j]
		}
	}
}

// Flop performs a vertical flip of the image.
func (pgm *PGM) Flop() {
	for i := 0; i < pgm.height/2; i++ {
		pgm.data[i], pgm.data[pgm.height-i-1] = pgm.data[pgm.height-i-1], pgm.data[i]
	}
}

// SetMagicNumber updates the magic number of the image (P2 or P5).
func (pgm *PGM) SetMagicNumber(magicNumber string) {
	pgm.magicNumber = magicNumber
}

// SetMaxValue updates the max grayscale value and rescales pixel values accordingly.
func (pgm *PGM) SetMaxValue(maxValue uint8) {
	for y := 0; y < pgm.height; y++ {
		for x := 0; x < pgm.width; x++ {
			scaledValue := float64(pgm.data[y][x]) * float64(maxValue) / float64(pgm.max)
			newValue := uint8(scaledValue)
			pgm.data[y][x] = newValue
		}
	}
	pgm.max = maxValue
}

// Rotate90CW rotates the image 90 degrees clockwise.
func (pgm *PGM) Rotate90CW() {
	if pgm.width <= 0 || pgm.height <= 0 {
		return
	}

	newData := make([][]uint8, pgm.width)
	for i := 0; i < pgm.width; i++ {
		newData[i] = make([]uint8, pgm.height)
		for j := 0; j < pgm.height; j++ {
			newData[i][j] = pgm.data[pgm.height-j-1][i]
		}
	}
	pgm.data = newData
	pgm.width, pgm.height = pgm.height, pgm.width
}

// ToPBM converts the PGM image to a PBM (Portable Bitmap) image.
func (pgm *PGM) ToPBM() *PBM {
	pbm := &PBM{
		data:        make([][]bool, pgm.height),
		width:       pgm.width,
		height:      pgm.height,
		magicNumber: "P1",
	}
	for y := 0; y < pgm.height; y++ {
		pbm.data[y] = make([]bool, pgm.width)
		for x := 0; x < pgm.width; x++ {
			pbm.data[y][x] = pgm.data[y][x] < uint8(pgm.max/2)
		}
	}
	return pbm
}

// PrintData prints the pixel data of the image to the console.
func (pgm *PGM) PrintData() {
	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			fmt.Printf("%d ", pgm.data[i][j])
		}
		fmt.Println()
	}
}
