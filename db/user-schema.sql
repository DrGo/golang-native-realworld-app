-- source: https://github.com/membership/membership.db/tree/master/sqlite
-- In SQLite, INTEGER PRIMARY KEY column is auto-incremented and becomes the rowid
--TODO: make username and email unique
DROP TABLE IF EXISTS User;
CREATE TABLE IF NOT EXISTS User
(
  id                   INTEGER PRIMARY KEY,
  email                TEXT NOT NULL UNIQUE,
  emailConfirmed       NUMERIC NOT NULL DEFAULT 0,
  password             TEXT,
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

CREATE INDEX IF NOT EXISTS User_ix_email ON User (email);

-- INSERT INTO User(email, userName)
-- VALUES 
--     ("t1@t.ca", "t1"),
--     ("t2@t.ca", "t2"),
--     ("t3@t.ca", "t3"),
--     ("t4@t.ca", "t4"),
--     ("t5@t.ca", "t5");

DROP TABLE IF EXISTS Follow;
--TODO: support FK constraints
CREATE TABLE IF NOT EXISTS Follow (
  userID INTEGER NOT NULL,
  followingID INTEGER NOT NULL,
  PRIMARY KEY (userID,followingID)
  -- CONSTRAINT `FOLLOW_userID` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  -- CONSTRAINT `FOLLOW_usrFollowID` FOREIGN KEY (`user_following_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
);



