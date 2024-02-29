package data

import (
	"time"

	up "github.com/upper/db/v4"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Active    string    `db:"user_active"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Token     Token     `db:"-"`
}

func (u *User) Table() string { //this func gives us the ability to overide the users in the database
	return "users" // so anytime refering to the table users in the db can overide; so,if legacy is called customers, overide and will refer to the users in the DB...
}

func (u *User) GetAll(condition up.Cond) ([]*User, error) {
	collection := upper.Collection(u.Table()) // FYI upper refers to data stored in the DB as collections so will use their naming convetion...

	var all []*User

	res := collection.Find(condition)
	err := res.All((&all))
	if err != nil {
		return nil, err
	}

	return all, nil
}

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
	res = collection.Find(up.Cond{"user_id =": theUser.ID, "expiry <": time.Now()}).OrderBy("created_at desc")
	err = res.One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return nil, err
		}
	}

	theUser.Token = token

	return &theUser, nil
}
