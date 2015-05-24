package main

import (
	"flag"
	"fmt"
	"github.com/akavel/polyclip-go"
	"github.com/swill/svgo" // you need the 'float' branch for this src
	"math"
	"os"
)

// a function to make a dynamic circle polygon.
// 's' is the number of segments per 1/4 circle.
func circle_polygon(cx, cy, r float64, s int) polyclip.Contour {
	// make a circle
	circle := func(x, y, r float64, s int, ps *polyclip.Contour) {
		radians := func(deg float64) float64 { // degrees to radians
			return (deg * math.Pi) / 180
		}
		n := float64(s)
		p := &polyclip.Point{x, y}
		*ps = append(*ps, *p)
		la := radians(90 - (90.0 / (2 * n))) // angle to determine the segment length
		for j := 1.0; j < 4*n; j++ {
			sa := radians(90 - ((90.0 / (2 * n)) * (2*j - 1))) // angle of the vector of each successive segment
			p.X += 2 * r * math.Cos(la) * math.Sin(sa)
			p.Y += 2 * r * math.Cos(la) * math.Cos(sa)
			*ps = append(*ps, *p)
		}
	}
	// draw the circle
	pts := make(polyclip.Contour, 0)
	circle(cx, cy-r, r, s, &pts)
	return pts
}

// prepare contours so 'github.com/swill/svgo' can draw them with PolygonF()
func prepare(pts polyclip.Contour) ([]float64, []float64) {
	xs := make([]float64, 0)
	ys := make([]float64, 0)
	for i := range pts {
		xs = append(xs, pts[i].X)
		ys = append(ys, pts[i].Y)
	}
	return xs, ys
}

// configureble flags
var (
	segments         = flag.Int("segments", 1, "the number of segments per 1/4 circle at the ends of the slot")
	radius           = flag.Float64("radius", 8, "the radius of the circles at the ends of the slot")
	add_right_circle = flag.Bool("add_right_circle", false, "add an additional circle on the right side of the rectangle")
)

// lets get to testing...
func main() {
	flag.Parse()
	poly_style := "fill:none;stroke:black;stroke-width:0.2"
	text_style := "fill:black;font-size:1.3mm;"
	poly_x := 30
	text_x := 60
	start_y := 15
	step_y := 25

	// create a dynamicly sized slot
	center := polyclip.Point{float64(poly_x), float64(start_y)}
	width := float64(20) // static width of the slot (assumes right circle is drawn)
	sw := width - *radius
	l := polyclip.Point{center.X - sw/2, center.Y}
	r := polyclip.Point{center.X + sw/2, center.Y}
	s_pts := polyclip.Contour{{l.X, l.Y - *radius}, {r.X, r.Y - *radius}, {r.X, r.Y + *radius}, {l.X, l.Y + *radius}}
	l_pts := circle_polygon(l.X, l.Y, *radius, *segments)
	r_pts := make(polyclip.Contour, 0)
	if *add_right_circle {
		r_pts = circle_polygon(r.X, r.Y, *radius, *segments)
	}

	// print the contour details in the terminal
	fmt.Println("\n-- Rectangle Contour --")
	fmt.Println(s_pts)

	fmt.Println("\n-- Left Circle Contour --")
	fmt.Println(l_pts)

	if *add_right_circle {
		fmt.Println("\n-- Right Circle Contour --")
		fmt.Println(r_pts)
	}

	// create an SVG file to show the result of the operations
	file, err := os.Create("polycliptest.svg")
	if err != nil {
		panic("ERROR: Unable to create the 'polycliptest.svg' file...")
	}
	defer file.Close()

	canvas := svg.New(file)
	canvas.FloatDecimals = 3
	canvas.StartviewUnitF(100, 150, "mm", 0, 0, 100, 150)

	// prepare the source contours to be added to the SVG file
	Xs, Ys := prepare(s_pts)
	Xl, Yl := prepare(l_pts)
	Xr, Yr := make([]float64, 0), make([]float64, 0)
	if *add_right_circle {
		Xr, Yr = prepare(r_pts)
	}

	// draw the source contours in the SVG file
	canvas.Text(text_x, start_y, "Contours", text_style)
	canvas.PolygonF(Xs, Ys, poly_style+";fill:red;fill-opacity:0.5;")
	canvas.PolygonF(Xl, Yl, poly_style+";fill:blue;fill-opacity:0.5;")
	if *add_right_circle {
		canvas.PolygonF(Xr, Yr, poly_style+";fill:green;fill-opacity:0.5;")
	}

	// setup the operation map to control what we test
	ops := []string{"Union", "Intersection", "Difference", "XOR"}
	op_map := map[string]polyclip.Op{
		"Union":        polyclip.UNION,
		"Intersection": polyclip.INTERSECTION,
		"XOR":          polyclip.XOR,
		"Difference":   polyclip.DIFFERENCE,
	}

	// create a test for each operation
	for i := range ops {
		name := ops[i]     // op name
		op := op_map[name] // polyclip.Op

		// do the operation on the contours
		op_poly := polyclip.Polygon{s_pts}
		op_poly = op_poly.Construct(op, polyclip.Polygon{l_pts})
		if *add_right_circle {
			op_poly = op_poly.Construct(op, polyclip.Polygon{r_pts})
		}

		// print the polygon details in the terminal
		fmt.Printf("\n== %s Polygon ==\n", name)
		fmt.Println(op_poly)

		if len(op_poly) > 0 {
			// draw the polygon in the SVG file
			canvas.Text(text_x, start_y+(i+1)*step_y, name, text_style)
			for p := range op_poly {
				Xp, Yp := prepare(op_poly[p])
				if len(Xp) > 0 && len(Yp) > 0 {
					canvas.TranslateF(0, float64((i+1)*step_y)) // translate to a new row
					canvas.PolygonF(Xp, Yp, poly_style+";fill:grey;fill-opacity:0.5;")
					canvas.Gend()
				}
			}
		} else {
			fmt.Println(name, "operation failed OR the result was empty...")
		}
	}

	canvas.End() // close the SVG file
}
