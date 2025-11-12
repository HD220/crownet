package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// size := 300_000
	// data := make([]Neuron, 0, size)

	// for i := 0; i < size; i++ {
	// 	data = append(data, Neuron{
	// 		Position: Vector3{
	// 			float32(i),
	// 			float32(i + 1),
	// 			float32(i + 2),
	// 		},
	// 		Direction: Vector3{
	// 			float32(i),
	// 			float32(i + 1),
	// 			float32(i + 2),
	// 		},
	// 		Distance: float32(i),
	// 		Radius:   float32(i),
	// 		Velocity: float32(i),
	// 	},
	// 	)
	// 	if i%100 == 0 {
	// 		fmt.Println("Hello, World!", i, data[i])
	// 	}
	// }

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Criar uma tabela
	sqlStmt := `CREATE TABLE IF NOT EXISTS neurons (
		id INTEGER PRIMARY KEY, 
		soma_position_x REAL NOT NULL,
		soma_position_Y REAL NOT NULL,
		soma_position_Z REAL NOT NULL,
		soma_diameter REAL NOT NULL,
		soma_radius REAL NOT NULL,
		axon_position_x REAL NOT NULL,
		axon_position_y REAL NOT NULL,
		axon_position_z REAL NOT NULL,
		axon_radius REAL NOT NULL,
		axon_myelinated INTEGER CHECK(axion_myelinated IN (0,1)) NOT NULL,
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	// Criar uma tabela
	sqlStmt := `CREATE VIRTUAL TABLE neurons_rtree USING rtree(
		id INTEGER PRIMARY KEY, 
		center_x REAL NOT NULL,
		center_Y REAL NOT NULL,
		center_Z REAL NOT NULL,
		soma_diameter REAL NOT NULL,
		soma_radius REAL NOT NULL,
		axon_position_x REAL NOT NULL,
		axon_position_y REAL NOT NULL,
		axon_position_z REAL NOT NULL,
		axon_radius REAL NOT NULL,
		axon_myelinated INTEGER CHECK(axion_myelinated IN (0,1)) NOT NULL,
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	v1 := Vector3{X: 1, Y: 2, Z: 3}
	v2 := Vector3{X: 4, Y: 6, Z: 8}

	// Distância
	fmt.Printf("Distância: %.3f\n", v1.Distance(v2))

	// Direção normalizada
	direction := v1.DirectionTo(v2)
	fmt.Printf("Direção normalizada: (%.3f, %.3f, %.3f)\n", direction.X, direction.Y, direction.Z)

	// Normalização
	normalized := v1.Normalized()
	fmt.Printf("Vetor normalizado: (%.3f, %.3f, %.3f)\n", normalized.X, normalized.Y, normalized.Z)

}
