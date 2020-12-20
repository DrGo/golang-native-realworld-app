-- sqlite does not have a date/time type. so using INTEGER and entering Unix time which
-- is likely efficient for storage (4 bytes) and for sorting and finding ranges
-- apps should use int64 to store and retrive dates
DROP TABLE IF EXISTS Article;
CREATE TABLE IF NOT EXISTS Article
(
  id                  INTEGER PRIMARY KEY,
  author              INTEGER NOT NULL,
  slug                TEXT NOT NULL,
  title               TEXT NOT NULL DEFAULT "Please, enter article title",
  description         TEXT,
  body                TEXT NOT NULL DEFAULT "Please, complete your article",
  favourited          NUMERIC NOT NULL DEFAULT 0,
  favouritesCount    NUMERIC NOT NULL DEFAULT 0,
  createdAt          INTEGER NOT NULL default (strftime('%s','now')),
  updatedAt          INTEGER NOT NULL default (strftime('%s','now')),
  CONSTRAINT favourited CHECK (favourited IN (0, 1))
);

-- automatically update 
Drop Trigger IF EXISTS Article_tr_update;
Create Trigger Article_tr_update After Update On Article Begin
  Update Article Set
    updatedAt = strftime('%s', DateTime('Now', 'localtime'))
  Where id = new.id;
End;

CREATE INDEX IF NOT EXISTS Artice_ix_author ON Article (author);
CREATE INDEX IF NOT EXISTS Artice_ix_tag ON Article (tag);

-- convert Unix-Times to DateTimes so not every single query needs to do so
-- convert Integer(4) (treating it as Unix-Time)--   to YYYY-MM-DD HH:MM:SS
Drop View IF EXISTS ArticleList;
Create View If Not Exists ArticleList As 
  Select a.id, slug, title, description, body, favourited, favouritesCount,
    DateTime(createdAt, 'unixepoch') As createdAt, 
    DateTime(updatedAt, 'unixepoch') As updatedAt,
    u.username, u.bio, u.image 
From Article As a INNER JOIN User as u on a.author=u.id;

INSERT INTO Article(author, slug, title)
VALUES 
    (1, "user1art1", "user 1 article 1"),
    (1, "user1art2", "user 1 article 2"),
    (2, "user2art1", "user 2 article 1"),
    (2, "user2art2", "user 2 article 2")
    ;

DROP TABLE IF EXISTS Comment;
CREATE TABLE IF NOT EXISTS Comment
(
  id                 INTEGER PRIMARY KEY,
  author             INTEGER NOT NULL,
  articleID          INTEGER NOT NULL,
  body               TEXT NOT NULL DEFAULT "Please, complete your comment",
  createdAt          INTEGER NOT NULL default (strftime('%s','now')),
  updatedAt          INTEGER NOT NULL default (strftime('%s','now'))
);

DROP TABLE IF EXISTS "Tag";
CREATE TABLE "Tag" (
	"tag"	TEXT NOT NULL,
	"articleID"	INTEGER NOT NULL,
	-- FOREIGN KEY("articleID") REFERENCES ,
	PRIMARY KEY("tag","articleID")
);
CREATE INDEX IF NOT EXISTS Tag_ix_tag ON Tag (tag);
CREATE INDEX IF NOT EXISTS Tag_ix_articleID ON Tag (articleID);

INSERT INTO "Tag" ("tag", "articleID") VALUES 
('js', '1'),
('go', '1'),
('es', '1'),
('golang', '1'),
('web', '2'),
('golang', '2')
;

DROP TABLE IF EXISTS Favourite;
CREATE TABLE IF NOT EXISTS Favourite
(
  userID              INTEGER NOT NULL,
  articleID           INTEGER NOT NULL,
  PRIMARY KEY (userID,articleID)
)