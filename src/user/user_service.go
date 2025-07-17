package user

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	id        string     `bson:"id"`
	Name      string     `bson:"name"`
	Email     string     `bson:"email"`
	deletedAt *time.Time `bson:"deletedAt,omitempty"`
}

func FindById(id string, db *sql.DB) (*User, error) {
	var user User

	query := `
		SELECT 
				id,
				name, 
				deleted_at
		FROM
		 		users
		WHERE
				id = $1
	`

	err := db.QueryRow(query, id).Scan(
		&user.id,
		&user.Name,
		&user.deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &user, nil
		}

		fmt.Printf("Erro encontrado: %v", err)

		return &user, err
	}

	return &user, nil
}

// func findAll(db *sql.DB) ([]User, error) {
// 	var users []User

// 	var query = `
// 	SELECT name, email FROM users
// 	`

// 	rows, err := db.Query(query)

// 	if err != nil {
// 		log.Printf("Erro na query de busca: %v", err)
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		var user User

// 		err := rows.Scan(&user.Name, &user.Email)

// 		if err != nil {
// 			log.Printf("Erro na query de busca: %v", err)
// 			return nil, err
// 		}

// 		users = append(users, user)
// 	}

// 	return users, nil
// }
