//-----------------------------------------------------------------------------
/*

Render an SDF

*/
//-----------------------------------------------------------------------------

package sdf

import "fmt"

//-----------------------------------------------------------------------------

// Render an SDF3 as a STL file.
func RenderSTL(
	s SDF3, //sdf3 to render
	mesh_cells int, //number of cells on the longest axis. e.g 200
	path string, //path to filename
) {
	// work out the region we will sample
	bb0 := s.BoundingBox()
	bb0_size := bb0.Size()
	mesh_inc := bb0_size.MaxComponent() / float64(mesh_cells)
	bb1_size := bb0_size.DivScalar(mesh_inc)
	bb1_size = bb1_size.Ceil().AddScalar(1)
	cells := bb1_size.ToV3i()
	bb1_size = bb1_size.MulScalar(mesh_inc)
	bb := NewBox3(bb0.Center(), bb1_size)

	fmt.Printf("rendering %s (%dx%dx%d)\n", path, cells[0], cells[1], cells[2])

	m := MarchingCubes(s, bb, mesh_inc)
	err := SaveSTL(path, m)
	if err != nil {
		fmt.Printf("%s", err)
	}
}

//-----------------------------------------------------------------------------

func mesh_sink(m *[]*Triangle3) chan<- *Triangle3 {
	c := make(chan *Triangle3)
	go func() {
		for t := range c {
			*m = append(*m, t)
		}
	}()
	return c
}

// Render an SDF3 as a STL file.
func RenderSTL_New(
	s SDF3, //sdf3 to render
	mesh_cells int, //number of cells on the longest axis. e.g 200
	path string, //path to filename
) {

	resolution := s.BoundingBox().Size().MaxComponent() / float64(mesh_cells)
	//resolution := 0.5

	fmt.Printf("rendering %s (resolution %.3f)\n", path, resolution)

	var m []*Triangle3

	output := mesh_sink(&m)
	MarchingCubesX(s, resolution, output)
	// stop the STL writer reading on the channel
	close(output)

	err := SaveSTL(path, m)
	if err != nil {
		fmt.Printf("%s", err)
	}
}

//-----------------------------------------------------------------------------

// Render an SDF2 as a DXF file.
func RenderDXF(
	s SDF2, //sdf2 to render
	mesh_cells int, //number of cells on the longest axis. e.g 200
	path string, //path to filename
) {
	// work out the region we will sample
	bb0 := s.BoundingBox()
	bb0_size := bb0.Size()
	mesh_inc := bb0_size.MaxComponent() / float64(mesh_cells)
	bb1_size := bb0_size.DivScalar(mesh_inc)
	bb1_size = bb1_size.Ceil().AddScalar(1)
	cells := bb1_size.ToV2i()
	bb1_size = bb1_size.MulScalar(mesh_inc)
	bb := NewBox2(bb0.Center(), bb1_size)

	fmt.Printf("rendering %s (%dx%d)\n", path, cells[0], cells[1])

	m := MarchingSquares(s, bb, mesh_inc)
	err := SaveDXF(path, m)
	if err != nil {
		fmt.Printf("%s", err)
	}
}

//-----------------------------------------------------------------------------
