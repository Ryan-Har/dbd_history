package main

import "fmt"

type Character struct {
	ID    int
	Name  string
	Image string
}

func getCharacters() ([]Character, error) {
	rows, err := db.Query(`
        SELECT id, name
        FROM characters
        ORDER BY id ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	characters := []Character{}

	for rows.Next() {
		var character Character

		if err := rows.Scan(
			&character.ID,
			&character.Name,
		); err != nil {
			return nil, err
		}

		character.Image = fmt.Sprintf("/images/%d", character.ID)

		characters = append(characters, character)
	}

	return characters, rows.Err()
}

func addCharacter(name string, image []byte, imageType string) error {
	_, err := db.Exec(`
        INSERT INTO characters (name, image, image_type)
        VALUES (?, ?, ?)
    `, name, image, imageType)

	return err
}

func deleteCharacter(id int) error {
	_, err := db.Exec(`
        DELETE FROM characters
        WHERE id = ?
    `, id)

	return err
}

func updateCharacter(id int, name string, image []byte, imageType string) error {
	_, err := db.Exec(`
        UPDATE characters
        SET name = ?, image = ?, image_type = ?
        WHERE id = ?
    `, name, image, imageType, id)

	return err
}
