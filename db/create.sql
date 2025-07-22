
CREATE DATABASE snipdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE snipdb;

-- Create the snippets table

CREATE TABLE snippets (
                          id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT, title VARCHAR(100) NOT NULL,
                          content TEXT NOT NULL,
                          created DATETIME NOT NULL,
                          expires DATETIME NOT NULL
);

CREATE INDEX idx_snippets_created ON snippets(created);


-- Add some dummy records (which we'll use in the next couple of chapters).
INSERT INTO snippets (title, content, created, expires) VALUES (
                                                                   'An old silent pond',
                                                                   'An old silent pond...\nA frog jumps into the pond,\nsplash! Silence again.\n\n– Matsuo Bashō', UTC_TIMESTAMP(),
                                                                   DATE_ADD(UTC_TIMESTAMP(), INTERVAL 365 DAY)
                                                               );
INSERT INTO snippets (title, content, created, expires) VALUES (
                                                                   'Over the wintry forest',
                                                                   'Over the wintry\nforest, winds howl in rage\nwith no leaves to blow.\n\n– Natsume Soseki', UTC_TIMESTAMP(),
                                                                   DATE_ADD(UTC_TIMESTAMP(), INTERVAL 365 DAY)
                                                               );
INSERT INTO snippets (title, content, created, expires) VALUES (
                                                                   'First autumn morning',
                                                                   'First autumn morning\nthe mirror I stare into\nshows my father''s face.\n\n– Murakami Kijo', UTC_TIMESTAMP(),
                                                                   DATE_ADD(UTC_TIMESTAMP(), INTERVAL 7 DAY)
                                                               );

-- For MySQL, use the following:
CREATE USER 'snipuser'@'localhost' IDENTIFIED BY 'snippassword';
GRANT SELECT, INSERT, UPDATE, DELETE ON snipdb.* TO 'snipuser'@'localhost';

-- For SQL Server, use the following instead:
-- CREATE LOGIN snipuser WITH PASSWORD = 'snippassword';
-- USE snipdb;
-- CREATE USER snipuser FOR LOGIN snipuser;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON SCHEMA :: dbo TO snipuser;

CREATE TABLE sessions (
                          token CHAR(43) PRIMARY KEY, data BLOB NOT NULL,
                          expiry TIMESTAMP(6) NOT NULL
);
CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE users (
                       id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT, name VARCHAR(255) NOT NULL,
                       email NVARCHAR(255) NOT NULL,
                       hashed_password NVARCHAR(60) NOT NULL,
                       created DATETIME NOT NULL
);
ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);