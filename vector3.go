package main

import "math"

type Vector3 struct {
	X, Y, Z float32
}

func (v *Vector3) Dot(other Vector3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v *Vector3) Sum(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

func (v *Vector3) Normalized() Vector3 {
	magnitude := float32(math.Sqrt(float64(v.Dot(*v))))

	if magnitude == 0 {
		return Vector3{0, 0, 0}
	}

	return Vector3{
		X: v.X / magnitude,
		Y: v.Y / magnitude,
		Z: v.Z / magnitude,
	}
}

func (v *Vector3) Difference(other Vector3) Vector3 {
	return Vector3{
		X: other.X - v.X,
		Y: other.Y - v.Y,
		Z: other.Z - v.Z,
	}
}

func (v *Vector3) DirectionTo(other Vector3) Vector3 {
	diff := v.Difference(other)
	return diff.Normalized()
}

func (v *Vector3) Distance(other Vector3) float32 {
	diff := v.Difference(other)

	return float32(math.Sqrt(float64(diff.Dot(diff))))
}
