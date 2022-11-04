package storage

import (
	"context"
	"crypto/rand"
	"fmt"
)

// Signup - Add user with a given email. Returns a token used for confirming
// the email address

func (d *Database) Signup(ctx context.Context, email string) (string, error) {
	token, err := createSecret()
	if err != nil {
		return "", err
	}
	// use some orm??
	query := `
		insert into newsletter_subscribers (email, token)
		values ($1, $2)
		on conflict (email) do update set
			token = excluded.token,
			updated = now()`
	_, err = d.DB.ExecContext(ctx, query, email, token)
	return token, err
}

func createSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", secret), nil
}

/*

And there's more. It's always a good idea to handle database changes in exactly this order:

    - Update your code to work with both the current and the new database state, and deploy it
	  and see it working.
    - Write and run the database migration.
    - Check that your database state is how you want it to be and your app works as expected.
    - After some time, remove the code that works with the old database state, deploy and check
	  that everything still works.

This is a safe way to make database changes. It may seem like a lot of work, but the data in your
database is often the most valuable thing in your app. Companies have gone bankrupt losing their
databases. Don't gamble with your data.

**/
