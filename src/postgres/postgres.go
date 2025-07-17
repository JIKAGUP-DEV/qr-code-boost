package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"qr-code-boost/src/config"

	_ "github.com/lib/pq"
)

func ConnectionPostgres() (*sql.DB, error) {

	databaseURL, err := config.GetEnvVariable("DATABASE_URL")

	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", databaseURL)

	if err != nil {
		log.Fatal("Erro ao conectar:", err)
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		log.Fatal("Erro ao testar conexão:", err)
		return nil, err
	}

	fmt.Println("------------------------")
	fmt.Println("Conectado ao PostgreSQL!")
	fmt.Println("------------------------")

	return db, nil
}

type Band struct {
	id   string `bson:"id,omitempty"`
	name string `bson:"name"`
}

func GetBands(db *sql.DB) ([]Band, error) {
	var bands []Band

	var query = `
		SELECT * FROM bands
	`

	rows, err := db.Query(query)

	if err != nil {
		log.Printf("Erro na query de busca: %v", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var band Band

		err := rows.Scan(&band.id, &band.name)
		if err != nil {
			log.Printf("Erro ao escanear linha.")
			return nil, err
		}

		bands = append(bands, band)
	}

	return bands, nil
}
