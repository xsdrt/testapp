package data

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	up "github.com/upper/db/v4"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Active    int       `db:"user_active"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Token     Token     `db:"-"`
}

// Table returns the table name assoc. with this model in the database...
func (u *User) Table() string { //this func gives us the ability to overide the users in the database
	return "users" // so anytime refering to the table users in the db can overide; so,if legacy is called customers, overide and will refer to the users in the DB...
}

// GetAll returns a slice of every user...
func (u *User) GetAll() ([]*User, error) {
	collection := upper.Collection(u.Table()) // FYI upper refers to data stored in the DB as collections so will use their naming convetion...

	var all []*User

	res := collection.Find().OrderBy("last_name")
	err := res.All((&all))
	if err != nil {
		return nil, err
	}

	return all, nil
}

// GetByEmail returns (1) one user , by their email...
func (u *User) GetByEmail(email string) (*User, error) {
	var theUser User
	collection := upper.Collection(u.Table())
	res := collection.Find(up.Cond{"email =": email}) // So,find all records in the database where the email is = to email supplied...
	err := res.One(&theUser)
	if err != nil {
		return nil, err
	}

	var token Token
	collection = upper.Collection(token.Table())
	res = collection.Find(up.Cond{"user_id =": theUser.ID, "expiry >": time.Now()}).OrderBy("created_at desc")
	err = res.One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return nil, err
		}
	}

	theUser.Token = token

	return &theUser, nil
}

// Get (1) one user by id...
func (u *User) Get(id int) (*User, error) {
	var theUser User
	collection := upper.Collection(u.Table())
	res := collection.Find(up.Cond{"id =": id})

	err := res.One(&theUser)
	if err != nil {
		return nil, err
	}

	var token Token
	collection = upper.Collection(token.Table())
	res = collection.Find(up.Cond{"user_id =": theUser.ID, "expiry >": time.Now()}).OrderBy("created_at desc")
	err = res.One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return nil, err
		}
	}

	theUser.Token = token

	return &theUser, nil
}

func (u *User) Update(theUser User) error {
	theUser.UpdatedAt = time.Now()
	collection := upper.Collection(u.Table())
	res := collection.Find(theUser.ID)
	err := res.Update(&theUser)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) Delete(id int) error {
	collection := upper.Collection(u.Table())
	res := collection.Find(id)
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}

// Insert a new user, and then return the newly inserted id...
func (u *User) Insert(theUser User) (int, error) {
	newHash, err := bcrypt.GenerateFromPassword([]byte(theUser.Password), 12)
	if err != nil {
		return 0, err
	}

	theUser.CreatedAt = time.Now()
	theUser.UpdatedAt = time.Now()
	theUser.Password = string(newHash)

	collection := upper.Collection(u.Table())
	res, err := collection.Insert(theUser)
	if err != nil {
		return 0, err
	}

	id := getInsertID(res.ID())

	return id, nil
}

func (u *User) ResetPassword(id int, password string) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	theUser, err := u.Get(id)
	if err != nil {
		return err
	}

	u.Password = string(newHash)

	err = theUser.Update(*u)
	if err != nil {
		return err
	}

	return nil
}

// PasswordMatches verifies a supplied password against the hash stored in the database.
// It will return true if it is valid, and subsequintly false if the password does not match,
// or if there is an error.  Should Note: an error is on;y returmned if something goes wrong
// (since an invalid password is not an error... it is just the wrong password!)
func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText)) // lower cased plaintext
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

//Dev some tests, unit and intergration.  Going to spin up a docker image of a postgres DB with the proper tables , then use to run some intergration tests. If passes , dispose of the image...
