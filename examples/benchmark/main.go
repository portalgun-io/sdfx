package main

import . "github.com/deadsy/sdfx/sdf"

func main() {
	s2d := Circle2D(5)
	BenchmarkSDF2("circle SDF2", s2d)

	s2d = FlatFlankCam2D(30, 20, 5)
	BenchmarkSDF2("cam1 SDF2", s2d)

	s2d = ThreeArcCam2D(30, 20, 5, 200)
	BenchmarkSDF2("cam2 SDF2", s2d)

	s2d = Polygon2D(Nagon(6, 10.0))
	BenchmarkSDF2("poly6 SDF2", s2d)

	s2d = Polygon2D(Nagon(12, 10.0))
	BenchmarkSDF2("poly12 SDF2", s2d)

	s3d := Box3D(V3{10, 20, 30}, 1)
	BenchmarkSDF3("box SDF3", s3d)
}
