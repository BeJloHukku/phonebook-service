package postgres

import (
	"context"
	"database/sql"
	"errors"

	phonebook "github.com/BeJloHukku/phonebook/internal/models"
)

var ErrNotFound = errors.New("not found")

type UpdateContactParams struct {
	Name       *string
	Surname    *string
	SecondName *string
	Numbers    *[]string
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateContact(
	ctx context.Context,
	contact phonebook.Contact,
) (phonebook.Contact, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return phonebook.Contact{}, err
	}
	defer tx.Rollback()

	queryContacts := `
		INSERT INTO contacts (name, surname, second_name)
    	VALUES ($1, $2, $3)
    	RETURNING id
	`
	err = tx.QueryRowContext(
		ctx,
		queryContacts,
		contact.Name,
		contact.Surname,
		contact.SecondName,
	).Scan(&contact.ID)

	if err != nil {
		return phonebook.Contact{}, err
	}

	queryNumbers := `
    	INSERT INTO numbers (number, contact_id)
    	SELECT number, $2
    	FROM unnest($1::text[]) AS n(number)
	`

	_, err = tx.ExecContext(
		ctx,
		queryNumbers,
		contact.Numbers,
		contact.ID,
	)
	if err != nil {
		return phonebook.Contact{}, err
	}

	err = tx.Commit()
	if err != nil {
		return phonebook.Contact{}, err
	}

	return contact, nil
}

func (s *Store) GetContactByID(ctx context.Context, id int) (phonebook.Contact, error) {
	query := `
		SELECT c.id, 
			   c.name, 
			   COALESCE(c.surname, ''),
			   COALESCE(c.second_name, ''),
			   n.number
		FROM numbers n
		JOIN contacts c ON n.contact_id = c.id
		WHERE c.id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return phonebook.Contact{}, err
	}
	defer rows.Close()

	var c phonebook.Contact
	c.Numbers = make([]string, 0)
	for rows.Next() {
		var number string
		err := rows.Scan(&c.ID, &c.Name, &c.Surname, &c.SecondName, &number)
		if err != nil {
			return phonebook.Contact{}, err
		}
		c.Numbers = append(c.Numbers, number)
	}

	if err := rows.Err(); err != nil {
		return phonebook.Contact{}, err
	}

	return c, nil
}

func (s *Store) DeleteContactByID(ctx context.Context, id int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	delNumbersQuery := `DELETE FROM numbers WHERE contact_id = $1`
	delContactQuery := `DELETE FROM contacts WHERE id = $1`

	result, err := tx.ExecContext(ctx, delNumbersQuery, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	_, err = tx.ExecContext(ctx, delContactQuery, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) UpdateContact(ctx context.Context, id int, params UpdateContactParams) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM contacts WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE contacts SET
            name        = COALESCE($2, name),
            surname     = COALESCE($3, surname),
            second_name = COALESCE($4, second_name)
        WHERE id = $1
    `, id, params.Name, params.Surname, params.SecondName)
	if err != nil {
		return err
	}

	if params.Numbers != nil {
		_, err = tx.ExecContext(ctx, `DELETE FROM numbers WHERE contact_id = $1`, id)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
            INSERT INTO numbers (number, contact_id)
            SELECT number, $2
            FROM unnest($1::text[]) AS n(number)
        `, *params.Numbers, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
