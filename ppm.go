package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

type PPM struct {
	data          [][]Pixel
	width, height int
	magicNumber   string
	max           uint8
}

type Pixel struct {
	R, G, B uint8
}

// ReadPPM reads a PPM image from a file and returns a struct that represents the image.
func ReadPPM(filename string) (*PPM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read magic number
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber)
	if magicNumber != "P3" && magicNumber != "P6" {
		return nil, fmt.Errorf("invalid magic number: %s", magicNumber)
	}

	// Read dimensions
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

	// Read max value
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

	// Read image data
	data := make([][]Pixel, height)
	expectedBytesPerPixel := 3

	if magicNumber == "P3" {
		// Read P3 format (ASCII)
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {
				if x*3+2 >= len(fields) {
					return nil, fmt.Errorf("index out of range at row %d, column %d", y, x)
				}
				var pixel Pixel
				_, err := fmt.Sscanf(fields[x*3], "%d", &pixel.R)
				if err != nil {
					return nil, fmt.Errorf("error parsing Red value at row %d, column %d: %v", y, x, err)
				}
				_, err = fmt.Sscanf(fields[x*3+1], "%d", &pixel.G)
				if err != nil {
					return nil, fmt.Errorf("error parsing Green value at row %d, column %d: %v", y, x, err)
				}
				_, err = fmt.Sscanf(fields[x*3+2], "%d", &pixel.B)
				if err != nil {
					return nil, fmt.Errorf("error parsing Blue value at row %d, column %d: %v", y, x, err)
				}
				rowData[x] = pixel
			}
			data[y] = rowData
		}
	} else if magicNumber == "P6" {
		// Read P6 format (binary)
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

			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {
				pixel := Pixel{R: row[x*expectedBytesPerPixel], G: row[x*expectedBytesPerPixel+1], B: row[x*expectedBytesPerPixel+2]}
				rowData[x] = pixel
			}
			data[y] = rowData
		}
	}

	// Return the PPM struct
	return &PPM{data, width, height, magicNumber, max}, nil
}

func (ppm *PPM) PrintPPM() {
	fmt.Printf("Magic Number: %s\n", ppm.magicNumber)
	fmt.Printf("Width: %d\n", ppm.width)
	fmt.Printf("Height: %d\n", ppm.height)
	fmt.Printf("Max Value: %d\n", ppm.max)

	fmt.Println("Pixel Data:")
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			pixel := ppm.data[y][x]
			fmt.Printf("(%d, %d, %d) ", pixel.R, pixel.G, pixel.B)
		}
		fmt.Println()
	}
}

func (ppm *PPM) Size() (int, int) {
	return ppm.width, ppm.height
}

func (ppm *PPM) At(x, y int) Pixel {
	// Vérification des limites pour éviter les erreurs d'index
	if x < 0 || x >= ppm.width || y < 0 || y >= ppm.height {
		// Vous pouvez également gérer cela différemment, comme renvoyer une valeur par défaut ou une erreur.
		panic("Index out of bounds")
	}

	return ppm.data[y][x]
}

func (ppm *PPM) Set(x, y int, value Pixel) {
	// Vérification des limites pour éviter les erreurs d'index
	if x < 0 || x >= ppm.width || y < 0 || y >= ppm.height {
		// Vous pouvez également gérer cela différemment, comme renvoyer une valeur par défaut ou une erreur.
		panic("Index out of bounds")
	}

	ppm.data[y][x] = value
}

func (ppm *PPM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if ppm.magicNumber == "P6" || ppm.magicNumber == "P3" {
		fmt.Fprintf(file, "%s\n%d %d\n%d\n", ppm.magicNumber, ppm.width, ppm.height, ppm.max)
	} else {
		err = fmt.Errorf("magic number error")
		return err
	}

	//bytesPerPixel := 3 // Nombre d'octets par pixel pour P6

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			pixel := ppm.data[y][x]
			if ppm.magicNumber == "P6" {
				// Conversion inverse des pixels
				file.Write([]byte{pixel.R, pixel.G, pixel.B})
			} else if ppm.magicNumber == "P3" {
				// Conversion inverse des pixels
				fmt.Fprintf(file, "%d %d %d ", pixel.R, pixel.G, pixel.B)
			}
		}
		if ppm.magicNumber == "P3" {
			fmt.Fprint(file, "\n")
		}
	}

	return nil
}

func (ppm *PPM) Invert() {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			pixel := &ppm.data[y][x]
			pixel.R = 255 - pixel.R
			pixel.G = 255 - pixel.G
			pixel.B = 255 - pixel.B
		}
	}
}

func (ppm *PPM) Flip() {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width/2; x++ {
			ppm.data[y][x], ppm.data[y][ppm.width-x-1] = ppm.data[y][ppm.width-x-1], ppm.data[y][x]
		}
	}
}

func (ppm *PPM) Flop() {
	for y := 0; y < ppm.height/2; y++ {
		ppm.data[y], ppm.data[ppm.height-y-1] = ppm.data[ppm.height-y-1], ppm.data[y]
	}
}

func (ppm *PPM) SetMagicNumber(magicNumber string) {
	ppm.magicNumber = magicNumber
}

// SetMaxValue updates the maximum pixel value in the PPM structure
// and scales the pixel values in data based on the new max value.
func (ppm *PPM) SetMaxValue(maxValue uint8) {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Scale the RGB values based on the new max value
			ppm.data[y][x].R = uint8(float64(ppm.data[y][x].R) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].G = uint8(float64(ppm.data[y][x].G) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].B = uint8(float64(ppm.data[y][x].B) * float64(maxValue) / float64(ppm.max))
		}
	}

	// Update the max value
	ppm.max = maxValue
}

func (ppm *PPM) Rotate90CW() {
	newPPM := PPM{
		data:        make([][]Pixel, ppm.width),
		width:       ppm.height,
		height:      ppm.width,
		magicNumber: ppm.magicNumber,
		max:         ppm.max,
	}

	for i := range newPPM.data {
		newPPM.data[i] = make([]Pixel, newPPM.width)
	}

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			newPPM.data[x][ppm.height-y-1] = ppm.data[y][x]
		}
	}

	*ppm = newPPM
}

// ToPGM converts the PPM image to a PGM image (grayscale).
func (ppm *PPM) ToPGM() *PGM {
	pgm := &PGM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P2",
		max:         ppm.max,
	}

	pgm.data = make([][]uint8, ppm.height)
	for i := range pgm.data {
		pgm.data[i] = make([]uint8, ppm.width)
	}

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Convert RGB to grayscale
			gray := uint8((int(ppm.data[y][x].R) + int(ppm.data[y][x].G) + int(ppm.data[y][x].B)) / 3)
			pgm.data[y][x] = gray
		}
	}

	return pgm
}

type Point struct {
	X, Y int
}

// rgbToGray converts an RGB color to a grayscale value.
func rgbToGray(color Pixel) uint8 {
	// Use luminosity method for converting RGB to grayscale
	// Gray = 0.299*R + 0.587*G + 0.114*B
	return uint8(0.299*float64(color.R) + 0.587*float64(color.G) + 0.114*float64(color.B))
}

func (ppm *PPM) ToPBM() *PBM {
	pbm := &PBM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P1",
	}

	pbm.data = make([][]bool, ppm.height)
	for i := range pbm.data {
		pbm.data[i] = make([]bool, ppm.width)
	}

	// Set a threshold for binary conversion
	threshold := uint8(ppm.max / 2)

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Calculate the average intensity of RGB values
			average := (uint16(ppm.data[y][x].R) + uint16(ppm.data[y][x].G) + uint16(ppm.data[y][x].B)) / 3
			// Set the binary value based on the threshold
			pbm.data[y][x] = average < uint16(threshold)
		}
	}
	return pbm
}

// pbm.Save("tetconvert.pgm")
// Draw
//
//

// SetPixel sets the color of a pixel at a given point.
func (ppm *PPM) SetPixel(p Point, color Pixel) {
	// Check if the point is within the PPM dimensions.
	if p.X >= 0 && p.X < ppm.width && p.Y >= 0 && p.Y < ppm.height {
		ppm.data[p.Y][p.X] = color
	}
}

// DrawLine draws a line between two points.

// DrawLine uses Bresenham's line algorithm to draw a line between two points.
// Bresenham's algorithm efficiently rasterizes a line on a grid of pixels.
func (ppm *PPM) DrawLine(p1, p2 Point, color Pixel) {
	// Bresenham's line algorithm

	// Extract coordinates of the two points.
	x1, y1 := p1.X, p1.Y
	x2, y2 := p2.X, p2.Y

	// Calculate differences in x and y coordinates.
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)

	// Determine the direction of the line along the x-axis.
	var sx, sy int
	if x1 < x2 {
		sx = 1
	} else {
		sx = -1
	}

	// Determine the direction of the line along the y-axis.
	if y1 < y2 {
		sy = 1
	} else {
		sy = -1
	}

	// Initialize the error term.
	err := dx - dy

	// Iterate through the points along the line using Bresenham's algorithm.
	for {
		// Set the pixel at the current point on the line.
		ppm.SetPixel(Point{x1, y1}, color)

		// Check if the end point of the line is reached.
		if x1 == x2 && y1 == y2 {
			break
		}

		// Calculate the doubled error term.
		e2 := 2 * err

		// Update the error term based on the decision parameter.
		if e2 > -dy {
			err -= dy
			x1 += sx
		}

		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DrawRectangle draws a rectangle.
func (ppm *PPM) DrawRectangle(p1 Point, width, height int, color Pixel) {
	// Draw the four sides of the rectangle using DrawLine.
	p2 := Point{p1.X + width, p1.Y}
	p3 := Point{p1.X + width, p1.Y + height}
	p4 := Point{p1.X, p1.Y + height}

	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p4, color)
	ppm.DrawLine(p4, p1, color)
}

func (ppm *PPM) DrawFilledRectangle(p1 Point, width, height int, color Pixel) {
	// Ensure positive width and height.
	if width <= 0 || height <= 0 {
		return
	}

	for w := width; w > 0; w-- {
		// Draw a rectangle with reduced width.
		ppm.DrawRectangle(p1, w, height, color)

		// Move the starting point for the next iteration.
		p1.X++
	}
}

func (ppm *PPM) setPixel(x, y int, color Pixel) {
	if x >= 0 && x < ppm.width && y >= 0 && y < ppm.height {
		ppm.data[y][x] = color
	}
}

func (ppm *PPM) DrawCircle(center Point, radius int, color Pixel) {
	// converts polar coordinates to Cartesian coordinates

	// Ensure non-negative radius.
	if radius < 0 {
		return
	}
	// -0.01 because 1 fucking pixel is not correct place
	for theta := -0.01; theta <= 1.99*math.Pi; theta += (1.0 / float64(radius)) {
		x := center.X + int(float64(radius)*math.Cos(theta))
		y := center.Y + int(float64(radius)*math.Sin(theta))

		ppm.setPixel(x, y, color)
	}
}

func (ppm *PPM) DrawFilledCircle(center Point, radius int, color Pixel) {
	// Assurez-vous que le rayon est non négatif.
	if radius < 0 {
		return
	}

	// Remplir le point central du cercle
	ppm.setPixel(center.X, center.Y, color)

	// -0.01 parce qu'un putain de pixel n'est pas à la bonne place
	for theta := -0.01; theta <= 1.99*math.Pi; theta += (1.0 / float64(radius)) {
		x := center.X + int(float64(radius)*math.Cos(theta))
		y := center.Y + int(float64(radius)*math.Sin(theta))

		// Remplir la ligne horizontale du centre aux bords du cercle
		for xi := x; xi < center.X; xi++ {
			ppm.setPixel(xi, y, color)
			ppm.setPixel(center.X*2-xi, y, color)
		}

		// Remplir la ligne verticale du centre aux bords du cercle
		for yi := y; yi < center.Y; yi++ {
			ppm.setPixel(x, yi, color)
			ppm.setPixel(x, center.Y*2-yi, color)
		}
	}
}
func (ppm *PPM) drawHorizontalLine(x1, x2, y int, color Pixel) {
	// Ensure valid y-coordinate.
	if y < 0 || y >= ppm.height {
		return
	}

	// Ensure x1 is less than or equal to x2.
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	// Clip x-coordinates to the image bounds.
	x1 = clamp(x1, 0, ppm.width-1)
	x2 = clamp(x2, 0, ppm.width-1)

	for x := x1; x <= x2; x++ {
		ppm.setPixel(x, y, color)
	}
}
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// DrawTriangle draws a triangle.
func (ppm *PPM) DrawTriangle(p1, p2, p3 Point, color Pixel) {
	// Draw the three sides of the triangle using DrawLine.
	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p1, color)
}

func (ppm *PPM) DrawFilledTriangle(p1, p2, p3 Point, color Pixel) {
	var Corners [3]Point
	// Sort the vertices based on Y-coordinate.
	if p1.Y <= p2.Y && p1.Y <= p3.Y {
		Corners[0], Corners[1], Corners[2] = p1, p2, p3
	} else if p2.Y <= p1.Y && p2.Y <= p3.Y {
		Corners[0], Corners[1], Corners[2] = p2, p1, p3
	} else {
		Corners[0], Corners[1], Corners[2] = p3, p1, p2
	}
	// Calculate slopes for the two edges of the triangle.
	slope1 := float64(Corners[2].X-Corners[0].X) / float64(Corners[2].Y-Corners[0].Y)
	slope2 := float64(Corners[2].X-Corners[1].X) / float64(Corners[2].Y-Corners[1].Y)

	x1 := float64(Corners[0].X)
	x2 := float64(Corners[1].X)

	for y := Corners[0].Y; y <= Corners[1].Y; y++ {
		ppm.DrawLine(Point{int(x1 + 0.5), y}, Point{int(x2 + 0.5), y}, color)
		x1 += slope1
		x2 += slope2
	}
	x2 = float64(Corners[1].X)
	for y := Corners[1].Y + 1; y <= Corners[2].Y; y++ {
		ppm.DrawLine(Point{int(x1 + 0.5), y}, Point{int(x2 + 0.5), y}, color)
		x1 += slope1
		x2 += slope2
	}
}

// DrawPolygon draws a polygon.
func (ppm *PPM) DrawPolygon(points []Point, color Pixel) {
	// Draw the sides of the polygon using DrawLine.
	for i := 0; i < len(points)-1; i++ {
		ppm.DrawLine(points[i], points[i+1], color)
	}
	// Connect the last and first points to close the polygon.
	ppm.DrawLine(points[len(points)-1], points[0], color)
}

func (ppm *PPM) DrawFilledPolygon(points []Point, color Pixel) {
	minY := points[0].Y
	maxY := points[0].Y

	for _, point := range points {
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	xco := make([][]int, maxY-minY+1)

	for i := 0; i < len(points); i++ {
		p1 := points[i]
		p2 := points[(i+1)%len(points)]

		var start, end Point
		if p1.Y <= p2.Y {
			start, end = p1, p2
		} else {
			start, end = p2, p1
		}

		slope := float64(end.X-start.X) / float64(end.Y-start.Y)

		x := float64(start.X)

		for y := start.Y; y <= end.Y; y++ {
			index := y - minY
			xco[index] = append(xco[index], int(x+0.5))
			x += slope
		}
	}

	// Loop through odd rows (starting from index 0) to draw lines between pairs of points.
	for i := 0; i < len(xco); i += 2 {
		// Draw lines between pairs of points for even rows.
		for j := 0; j < len(xco[i])-1; j += 2 {
			// Use the Even-Odd Fill Algorithm to draw lines between pairs of points.
			// The algorithm ensures that the interior of the polygon is correctly filled.
			// It determines whether a point is inside or outside the polygon by counting
			// the number of intersections with polygon edges along the scanline.
			ppm.DrawLine(Point{xco[i][j], i + minY},
				Point{xco[i][j+1], i + minY}, color)
		}

		// If there's a next row, draw lines between pairs of points for odd rows.
		if i+1 < len(xco) {
			for j := 0; j < len(xco[i+1])-1; j += 1 {
				// Use the Even-Odd Fill Algorithm to draw lines between pairs of points.
				// Similar to the even rows, it ensures correct filling of the polygon.
				ppm.DrawLine(Point{xco[i+1][j], i + minY + 1},
					Point{xco[i+1][j+1], i + minY + 1}, color)
			}
		}
	}

}
