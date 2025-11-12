package main

type Neuron struct {
	SomaPosition  Vector3
	SomaRadius    float64
	AxionPosition Vector3
	AxionRadius   float64
}

func (n *Neuron) Move() {
	n.Position.X += n.Direction.X * n.Velocity
	n.Position.Y += n.Direction.Y * n.Velocity
	n.Position.Z += n.Direction.Z * n.Velocity
}

func (n *Neuron) Fire() string {
	return "Pulso disparado na direção ⚡"
}
