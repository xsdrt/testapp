package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"net/http"
	"strings"
	"time"

	up "github.com/upper/db/v4"
)

type Token struct {
	ID        int       `db:"id,omitempty" json:"id"` //added omitempty as it was failing my integration tests
	UserID    int       `db:"user_id" json:"user_id"`
	FirstName string    `db:"first_name" json:"first_name"`
	Email     string    `db:"email" json:"email"`
	PlainText string    `db:"token" json:"token"`
	Hash      []byte    `db:"token_hash" jason:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Expires   time.Time `db:"expiry" json:"expiry"`
}

// Below are all the database functions for tokens...

func (t *Token) Table() string {
	return "tokens"
}

// Allows us to get a user by token...
func (t *Token) GetUserForToken(token string) (*User, error) {
	var u User
	var theToken Token

	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": token})
	err := res.One(&theToken)
	if err != nil {
		return nil, err
	}

	collection = upper.Collection("users") //refactored missing s in users... may have been affecting my integration tests.
	res = collection.Find(up.Cond{"id": theToken.UserID})
	err = res.One(&u)
	if err != nil {
		return nil, err
	}

	u.Token = theToken

	return &u, nil
}

// get all the tokens for a given user...
func (t *Token) GetTokensForUser(id int) ([]*Token, error) {
	var tokens []*Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"user_id": id})
	err := res.All(&tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Get a token by id...
func (t *Token) Get(id int) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"id": id})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// Get token by plainText...
func (t *Token) GetByToken(plainText string) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainText})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// Delete a token by id, especially when a user logs out...
func (t *Token) Delete(id int) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(id)
	err := res.Delete()
	if err != nil {
		return err
	}

	return nil
}

// Delete by token(plainText)...
func (t *Token) DeleteByToken(plainText string) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainText})
	err := res.Delete()
	if err != nil {
		return err
	}

	return nil
}

// Insert a token into the database...
func (t *Token) Insert(token Token, u User) error {
	collection := upper.Collection(t.Table())

	// Delete existing tokens if inserting a new token(s)...
	res := collection.Find(up.Cond{"user_id": u.ID})
	err := res.Delete()
	if err != nil {
		return err
	}

	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	token.FirstName = u.FirstName
	token.Email = u.Email

	_, err = collection.Insert(token)
	if err != nil {
		return err
	}

	return nil
}

// Need to generate a token before they can be inserterd of course...
func (t *Token) GenerateToken(userID int, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID:  userID,
		Expires: time.Now().Add(ttl),
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes) // Using crypto/rand pkg...
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

// Authenticate the token....
func (t *Token) AuthenticateToken(r *http.Request) (*User, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("invalid token") //no authorization header received
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("invalid token") //no authorization header received
	}

	token := headerParts[1]

	if len(token) != 26 { // every token should be 26 characters long...
		return nil, errors.New("invalid token") //token wrong size
	}

	tkn, err := t.GetByToken(token)
	if err != nil {
		return nil, errors.New("invalid token") //no matching token found
	}

	if tkn.Expires.Before(time.Now()) {
		return nil, errors.New("invalid token") //expired token
	}

	user, err := t.GetUserForToken(token)
	if err != nil {
		return nil, errors.New("invalid token") //no matching user found
	}

	return user, nil
}

// validate incoming token
func (t *Token) ValidToken(token string) (bool, error) {
	user, err := t.GetUserForToken(token)
	if err != nil {
		return false, errors.New("invalid credentials") //no matching user found
	}

	if user.Token.PlainText == "" {
		return false, errors.New("invalid credentials") //no matching token found
	}

	if user.Token.Expires.Before(time.Now()) {
		return false, errors.New("invalid credentials") //expired token
	}

	return true, nil
}
