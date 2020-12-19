package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/utils"
)

// Args used to pass names and values of named arguments to queries
type Args map[string]interface{}

// Row represent an arbitrary table row
type Row map[string]interface{}

// DB is an interface that defines database functionality required by Store.
type DB interface {
	Close() error
	Query(query string, args Args) (rows []Row, rowCount int, err error)
	JSONQuery(query string, args Args) (result string, rowCount int, err error)
	Exec(string, Args) (rowsAffected int, lastRowID int64, err error)
}

type Store struct {
	db DB
}

//TODO: inject DB into store to remove dependency on NewDB()
func mustNewStore(dsn string) *Store {
	store, err := newStore(dsn)
	if err != nil {
		log.Fatalf("cannot create database %v", err)
	}
	return store
}

func newStore(dsn string) (*Store, error) {
	db, err := NewDB(dsn)
	if err != nil {
		return nil, err
	}
	return &Store{
		db: db,
	}, nil
}

func (st *Store) CreateUser(creds *credentials) (int64, error) {
	query := `INSERT INTO User (username, email, password) 
		VALUES ($username,$email,$passwordHash)`
	_, id, err := st.db.Exec(query, Args{
		"$username":     creds.User.Username,
		"$email":        creds.User.Email,
		"$passwordHash": utils.HashedPassword(creds.User.Password),
	})
	return id, err
}

func (st *Store) SignInByEmailAndPassword(email, password string) (Row, error) {
	rows, count, err := st.db.Query("select id, password from User where email= $email", Args{"$email": email})
	if err != nil || count == 0 {
		return nil, errors.Errorf("user [%s] not found: %v", email, err)
	}
	hashedPassword := rows[0]["password"].(string) // if it does not work, panic is ok
	if !utils.ValidPassword(hashedPassword, password) {
		return nil, errors.Errorf("invalid password")
	}
	return rows[0], nil
}

func (st *Store) GetUserJSON(uid int64) ([]byte, error) {
	rows, count, err := st.db.Query(`select id, username, email, bio, image, 'useless' as token from User 
	where id= $uid`, Args{"$uid": uid})
	if err != nil || count == 0 {
		return nil, errors.Errorf("user not found: %v", err)
	}
	row := rows[0]
	return json.Marshal(utils.Map{"user": row})
}

func (st *Store) GetUserProfileJSON(userName string) ([]byte, error) {
	rows, count, err := st.db.Query(`select id, username, bio, image, 0 as following from User
	 where username= $username`, Args{"$username": userName})
	if err != nil || count == 0 {
		return nil, errors.Errorf("user not found: %v", err)
	}
	row := rows[0]
	return json.Marshal(utils.Map{"profile": row})
}

//FIXME: handle null tags eg Case or coalsce
const articleQueryCols = `'id', a.id, 'slug', slug, 'title', title, 'description', description,
'body', body, 'favorited', favourited, 'favoritesCount', favouritesCount,
'createdAt', DateTime(createdAt, 'unixepoch'), 'updatedAt', DateTime(updatedAt, 'unixepoch'),
'tagList', COALESCE(json_extract(tags, '$'), json_array()) ,
'author', json_object('username', u.username, 'bio', u.bio, 'image', u.image)`

const articleQueryJoins = `
as JSON 
FROM (select * from Article LIMIT $limit OFFSET $offset) a  
INNER JOIN User AS u ON a.author=u.id 
OUTER LEFT JOIN (SELECT articleID,json_group_array(tag) as tags FROM Tag GROUP BY articleID) t on a.id = t.articleID
 %s ORDER BY a.createdAt DESC;`

const articleQueryList = `SELECT json_object('articles', json_group_array(
	json_object(` + articleQueryCols + `)), 'articlesCount', 
	COUNT(a.id))` + articleQueryJoins

const articleQuerySingle = `SELECT json_object('article',json_object(` + articleQueryCols + `))` + articleQueryJoins

type ListArticlesOptions struct {
	Limit       int
	Offset      int
	Slug        string
	Author      string
	Tag         string
	FavoritedBy string
}

func (st *Store) DefaultListArticlesOptions(slug string) *ListArticlesOptions {
	return &ListArticlesOptions{
		Limit:  20,
		Offset: 0,
		Slug:   slug,
	}
}

func (st *Store) ListArticlesJSON(opt *ListArticlesOptions) ([]byte, error) {
	var query string
	// single article requested
	if opt.Slug != "" {
		where := "WHERE a.slug=" + "'" + opt.Slug + "'"
		query = fmt.Sprintf(articleQuerySingle, where)
	} else { // multiple articles possibly filtered
		where := "WHERE 1=1 "
		if opt.Author != "" {
			where = where + "AND u.username=" + "'" + opt.Author + "'"
		}
		if opt.Tag != "" {
			where = where + "AND a.id IN (SELECT articleID FROM Tag WHERE tag='" + opt.Tag + "')"
		}
		query = fmt.Sprintf(articleQueryList, where)
	}
	result, count, err := st.db.JSONQuery(query, Args{"$limit": opt.Limit, "$offset": opt.Offset})
	if err != nil {
		return nil, errors.Errorf("error retrieving articles: %v", err)
	}
	//FIXME: do we need to return an error?
	if count != 1 {
		return nil, errors.Errorf("no such articles")
	}
	return []byte(result), nil
}

//TODO: add a transaction
func (st *Store) CreateArticle(art *articleModel) ([]byte, error) {
	query := `INSERT INTO Article (slug,title,description,body,author) 
	VALUES ($slug,$title,$description,$body,$author)`
	_, id, err := st.db.Exec(query, Args{
		"$slug":        art.Slug,
		"$title":       art.Title,
		"$description": art.Description,
		"$body":        art.Body,
		"$author":      art.Author,
	})
	query = `INSERT INTO Tag (tag,articleID) VALUES ($tag,$articleID)`
	for _, tag := range art.TagList {
		_, _, err = st.db.Exec(query, Args{"$tag": tag, "$articleID": id})
	}
	if err != nil {
		return nil, errors.Errorf("error creating article (%s): %v", art.Title, err)
	}
	opt := st.DefaultListArticlesOptions(art.Slug)
	json, err := st.ListArticlesJSON(opt)
	if err != nil {
		return nil, errors.Errorf("error retrieving article (%s): %v", art.Title, err)
	}
	return json, nil
}

func (st *Store) FavouriteArticle(slug string, userID int64, favourited bool) ([]byte, error) {
	var query string
	if favourited {
		query = `INSERT INTO Favourite(userID,articleID) SELECT $userID, id FROM Article where
		slug=$slug`
	} else {
		query = `DELETE FROM Favourite WHERE userID=$userID AND articleID IN (SELECT id FROM Article where slug=$slug)`
	}
	_, _, err := st.db.Exec(query, Args{"$userID": userID, "$slug": slug})
	if err != nil {
		return nil, errors.Errorf("error changing favourite status for article (%s): %v", slug, err)
	}
	opt := st.DefaultListArticlesOptions(slug)
	json, err := st.ListArticlesJSON(opt)
	if err != nil {
		return nil, errors.Errorf("error retrieving article (%s): %v", slug, err)
	}
	return json, nil
}

func (st *Store) ListTagsJSON() ([]byte, error) {
	const query = `SELECT json_object('tags', json_group_array(tag)) 
		FROM (SELECT DISTINCT tag FROM Tag LIMIT $limit OFFSET $offset)`
	result, count, err := st.db.JSONQuery(query, Args{"$limit": 100, "$offset": 0})
	if err != nil {
		return nil, errors.Errorf("error retrieving tags: %v", err)
	}
	//FIXME: do we need to return an error?
	if count != 1 {
		return nil, errors.Errorf("no such tags")
	}
	return []byte(result), nil
}

func (st *Store) CreateComment(comment *commentModel, slug string) ([]byte, error) {
	query := `INSERT INTO Comment (author, body, articleID) SELECT $author, $body, id FROM Article where
	slug=$slug`
	_, _, err := st.db.Exec(query, Args{
		"$author": comment.Author, "$body": comment.Body, "$slug": slug,
	})
	if err != nil {
		return nil, errors.Errorf("error creating comment (%s): %v", comment.Body, err)
	}
	// opt := st.DefaultListArticlesOptions(slug)
	// json, err := st.ListArticlesJSON(opt)
	// if err != nil {
	// 	return nil, errors.Errorf("error retrieving article (%s): %v", art.Title, err)
	// }
	return nil, nil
}

const commentQueryCols = `'id', c.id, 'body', c.body, 'createdAt', DateTime(c.createdAt, 'unixepoch'), 
'updatedAt', DateTime(c.updatedAt, 'unixepoch'),
'author', json_object('username', u.username, 'bio', u.bio, 'image', u.image)`

const commentQueryJoins = `
as JSON 
FROM Comment c, Article a, User u 
WHERE c.articleID= a.id AND c.author=u.id AND a.slug = $slug
ORDER BY c.createdAt DESC;`

const commentQueryList = `SELECT json_object('comments', json_group_array(
	json_object(` + commentQueryCols + `)))` + commentQueryJoins

const commentQuerySingle = `SELECT json_object('comment',json_object(` + commentQueryCols + `))` + commentQueryJoins

func (st *Store) ListArticleCommentsJSON(slug string, commentID int64) ([]byte, error) {
	query := commentQueryList
	//, "$commentID": commentID
	result, count, err := st.db.JSONQuery(query, Args{"$slug": slug})
	if err != nil {
		return nil, errors.Errorf("error retrieving comments: %v", err)
	}
	//FIXME: do we need to return an error?
	if count != 1 {
		return nil, errors.Errorf("no such comments")
	}
	return []byte(result), nil
}
