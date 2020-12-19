CREATE TABLE Article
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
CREATE TRIGGER Article_tr_update After Update On Article Begin
  Update Article Set
    updatedAt = strftime('%s', DateTime('Now', 'localtime'))
  Where id = new.id;
End;
CREATE INDEX Artice_ix_author ON Article (author);

CREATE TABLE Favourite
(
  userID              INTEGER NOT NULL,
  articleID           INTEGER NOT NULL,
  PRIMARY KEY (userID,articleID)
);
CREATE TABLE User
(
  id                   INTEGER PRIMARY KEY,
  email                TEXT NOT NULL UNIQUE,
  emailConfirmed       NUMERIC NOT NULL DEFAULT 0,
  passwordHash         TEXT,
  username             TEXT NOT NULL UNIQUE,
  bio                   TEXT NOT NULL DEFAULT "Please, complete your bio",
  image                 TEXT,   
  -- securityStamp        TEXT,
  -- concurrencyStamp     TEXT    NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  -- phoneNumber          TEXT,
  -- phoneNumberConfirmed NUMERIC NOT NULL DEFAULT 0,
  -- twoFactorEnabled     NUMERIC NOT NULL DEFAULT 0,
  -- lockoutEnd           TEXT,
  -- lockoutEnabled       NUMERIC NOT NULL DEFAULT 0,
  accessFailedCount    INTEGER NOT NULL DEFAULT 0,
  -- Constraints
  CONSTRAINT User_ck_emailConfirmed CHECK (emailConfirmed IN (0, 1))
  -- CONSTRAINT User_ck_phoneNumberConfirmed CHECK (phoneNumberConfirmed IN (0, 1))
  -- CONSTRAINT User_ck_twoFactorEnabled CHECK (twoFactorEnabled IN (0, 1)),
  -- CONSTRAINT User_ck_lockoutEnabled CHECK (lockoutEnabled IN (0, 1))
);
CREATE INDEX User_ix_email ON User (email);
CREATE TABLE Follow (
  userID INTEGER NOT NULL,
  followingID INTEGER NOT NULL,
  PRIMARY KEY (userID,followingID)
  -- CONSTRAINT `FOLLOW_userID` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  -- CONSTRAINT `FOLLOW_usrFollowID` FOREIGN KEY (`user_following_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE TABLE Comment
(
  id                 INTEGER PRIMARY KEY,
  author             INTEGER NOT NULL,
  articleID          INTEGER NOT NULL,
  body               TEXT NOT NULL DEFAULT "Please, complete your comment",
  createdAt          INTEGER NOT NULL default (strftime('%s','now')),
  updatedAt          INTEGER NOT NULL default (strftime('%s','now'))
);
