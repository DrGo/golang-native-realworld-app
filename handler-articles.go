package main

import (
	"net/http"

	"github.com/drgo/realworld/errors"
	"github.com/drgo/realworld/sessions"
	"github.com/drgo/realworld/utils"
)

// handles routes

type articleModel struct {
	ID              int64    `json:"id"`
	Author          int64    `json:"author"`
	Slug            string   `json:"slug"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Body            string   `json:"body"`
	Favourited      bool     `json:"favorited"`
	FavouritesCount int64    `json:"favoritesCount"`
	TagList         []string `json:"tagList"`
}

type commentModel struct {
	ID        int64  `json:"id"`
	Author    int64  `json:"author"`
	ArticleID int64  `json:"articleID"`
	Body      string `json:"body"`
}

// ServeArticles handles "/api/articles/*"
func ServeArticles(ctx *Ctx) error {
	dx := errors.D(ctx.Req, "ServeArticles")
	// no slug -> GET or POST /api/articles
	if dx.Path == "/" {
		switch dx.Method {
		case "GET":
			return articlesList(ctx, "")
		case "POST":
			if ctx.Authenticated(dx) {
				return articlesCreate(ctx, ctx.Session)
			}
		}
		return errors.E(dx, http.StatusMethodNotAllowed)
	}
	// slug provided -> GET /api/article/slug
	var slug, action string
	slug, ctx.Req.URL.Path = utils.ShiftPath(dx.Path)
	//GET /api/articles/feed
	if slug == "feed" {
		//FIXME: make user-specific
		return articlesList(ctx, "")
	}
	action, ctx.Req.URL.Path = utils.ShiftPath(ctx.Req.URL.Path)
	_ = errors.Debug && errors.Logln(slug, action)
	switch action {
	case "comments": ///api/articles/:slug/comments
		switch dx.Method {
		case "GET":
			return articlesListComments(ctx, slug, 0)
		case "POST":
			if ctx.Authenticated(dx) {
				return articlesCreateComment(ctx, slug, ctx.Session.UserID)
			}
		case "DELETE":
			//FIXME:
			if ctx.Authenticated(dx) {
				return nil
			}
		default:
			return errors.E(dx, http.StatusMethodNotAllowed)
		}
	case "favorite": // POST or DELETE /api/articles/:slug/favorite
		if ctx.Authenticated(dx) {
			return articlesFavourite(ctx, slug, ctx.Session.UserID, dx.Method == "POST")
		}
	case "": // GET OR DELETE /api/articles/:slug
		switch dx.Method {
		case "GET":
			return articlesList(ctx, slug)
		case "DELETE":
			//FIXME:
			return nil
		default:
			return errors.E(dx, http.StatusMethodNotAllowed)
		}
	default:
		return errors.E(dx, http.StatusNotFound)
	}
	return nil // error must have been handled by this point
}

// GET /api/articles and GET /api/articles/:slug
// /api/articles -> Returns most recent articles globally by default, filter results by Query Parameters:
// Filter by tag:  ?tag=AngularJS
// Filter by author:  ?author=jake
// Favorited by user:  ?favorited=jake
// Limit number of articles (default is 20): ?limit=20
// Offset/skip number of articles (default is 0): ?offset=0
// Authentication optional, will return multiple articles, ordered by most recent first
// /api/articles/:slug -> No authentication required, will return single article
func articlesList(ctx *Ctx, slug string) error {
	dx := errors.D(ctx.Req, "articlesGet")
	opt := ctx.Store().DefaultListArticlesOptions(slug)
	// extract filters
	if values, n := ctx.QueryParams("author"); n > 0 {
		opt.Author = values[0] // only use first author parameter
	}
	if values, n := ctx.QueryParams("tag"); n > 0 {
		opt.Tag = values[0] // only use first tag parameter
	}
	if values, n := ctx.QueryParams("favorited"); n > 0 {
		opt.FavoritedBy = values[0] // only use first favorited parameter
	}
	json, err := ctx.Store().ListArticlesJSON(opt)
	if err != nil {
		return errors.E(dx, err, http.StatusNotFound)
	}
	return utils.SendJSON(ctx.Res, http.StatusOK, json)
}

// POST /api/articles
// Example request body:
// {
//   "article": {
//     "title": "How to train your dragon",
//     "description": "Ever wonder how?",
//     "body": "You have to believe",
//     "tagList": ["reactjs", "angularjs", "dragons"]
//   }
// }
// Authentication required, will return an Article
// Required fields: title, description, body
// Optional fields: tagList as an array of Strings
func articlesCreate(ctx *Ctx, session *sessions.Session) error {
	dx := errors.D(ctx.Req, "articlesCreate")
	var payload struct {
		Art *articleModel `json:"article"`
	}
	err := utils.DecodeJSONBody(ctx.Res, ctx.Req, &payload)
	if err != nil {
		return errors.E(dx, err, http.StatusBadRequest)
	}
	art := payload.Art
	art.Author = session.UserID
	//TODO: validate inputs
	//FIXME: make sure slugs are unique
	art.Slug = utils.Slugify(art.Title)
	json, err := ctx.Store().CreateArticle(art)
	if err != nil {
		return errors.E(dx, err, http.StatusInternalServerError)
	}
	return utils.SendJSON(ctx.Res, http.StatusOK, json)
}

// Favorite Article
// POST /api/articles/:slug/favorite
// Unfavorite Article
// DELETE /api/articles/:slug/favorite
// Authentication required, returns the Article
// No additional parameters required
func articlesFavourite(ctx *Ctx, slug string, userID int64, favourited bool) error {
	dx := errors.D(ctx.Req, "articlesFavourite")
	json, err := ctx.Store().FavouriteArticle(slug, userID, favourited)
	if err != nil {
		return errors.E(dx, err, http.StatusInternalServerError)
	}
	return utils.SendJSON(ctx.Res, http.StatusOK, json)
}

// Get Comments from an Article
// GET /api/articles/:slug/comments
// Authentication optional, returns multiple comments
func articlesListComments(ctx *Ctx, slug string, commentID int64) error {
	dx := errors.D(ctx.Req, "articlesListComments")
	json, err := ctx.Store().ListArticleCommentsJSON(slug, 0)
	if err != nil {
		return errors.E(dx, err, http.StatusNotFound)
	}
	return utils.SendJSON(ctx.Res, http.StatusOK, json)
}

func articlesCreateComment(ctx *Ctx, slug string, userID int64) error {
	dx := errors.D(ctx.Req, "articlesCreateComment")
	var payload struct {
		Comment *commentModel `json:"comment"`
	}
	err := utils.DecodeJSONBody(ctx.Res, ctx.Req, &payload)
	if err != nil {
		return errors.E(dx, err, http.StatusBadRequest)
	}
	comment := payload.Comment
	comment.Author = userID
	//TODO: validate inputs
	json, err := ctx.Store().CreateComment(comment, slug)
	if err != nil {
		return errors.E(dx, err, http.StatusInternalServerError)
	}
	return utils.SendJSON(ctx.Res, http.StatusOK, json)
}
