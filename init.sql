CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE Users
(
    nickname CITEXT UNIQUE NOT NULL PRIMARY KEY,
    fullname VARCHAR(255)  NOT NULL,
    about    TEXT,
    email    CITEXT UNIQUE NOT NULL
);

CREATE INDEX IF not exists users_nickname ON Users (nickname);

-- ---------------------------------------------------------------

CREATE TABLE Forum
(
    title   CITEXT                               NOT NULL,
    forumUser CITEXT REFERENCES Users (nickname) NOT NULL,
    slug    CITEXT UNIQUE                        NOT NULL PRIMARY KEY,
    posts   BIGINT DEFAULT 0,
    threads BIGINT DEFAULT 0
);

CREATE INDEX IF not exists forum_slug ON Forum (slug);

-- ---------------------------------------------------------------

CREATE TABLE ForumUser
(
    slug     CITEXT,
    nickname CITEXT COLLATE "POSIX",
    CONSTRAINT unique_slug_nickname UNIQUE (slug, nickname)
);

CREATE INDEX forumUser_slug_idx on ForumUser (slug);
CREATE INDEX forumUser_nickname_idx on ForumUser (nickname);

-- ---------------------------------------------------------------

CREATE TABLE Post
(
    id       BIGSERIAL PRIMARY KEY,
    parent   BIGINT,
    author   CITEXT,
    message  TEXT,
    isEdited BOOLEAN,
    forum    CITEXT,
    thread   BIGINT,
    created  TIMESTAMP WITH TIME ZONE DEFAULT now(),
    path     BIGINT[] NOT NULL
);

CREATE INDEX IF NOT EXISTS post_path_id ON Post (id, (path[1]));
CREATE INDEX IF NOT EXISTS post_path ON Post (path);
CREATE INDEX IF NOT EXISTS post_path_1 ON Post ((path[1]));
CREATE INDEX IF NOT EXISTS post_thread_id ON Post (thread, id);
CREATE INDEX IF NOT EXISTS post_thread ON Post (thread);
CREATE INDEX IF NOT EXISTS post_thread_path_id ON Post (thread, path, id);
CREATE INDEX IF NOT EXISTS post_thread_id_path_parent ON Post (thread, id, (path[1]), parent);
CREATE INDEX IF NOT EXISTS post_author_forum ON Post (author, forum);

CREATE OR REPLACE FUNCTION createpath() RETURNS TRIGGER AS
$postmatpath$
BEGIN
    NEW.path = (SELECT path FROM Post WHERE id = NEW.parent) || NEW.id;
    RETURN NEW;
END;
$postmatpath$ LANGUAGE plpgsql;

CREATE TRIGGER create_path_trigger
    BEFORE INSERT
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE createpath();

CREATE OR REPLACE FUNCTION updateeditedcolumn() RETURNS TRIGGER AS
$body$
BEGIN
    IF NEW.message != OLD.message THEN
        NEW.isEdited = TRUE;
    END IF;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER update_edited_column_trigger
    BEFORE UPDATE
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE updateeditedcolumn();

-- ---------------------------------------------------------------

CREATE TABLE Thread
(
    id      BIGSERIAL PRIMARY KEY,
    title   VARCHAR(255)                         NOT NULL,
    author  CITEXT REFERENCES Users (nickname) NOT NULL,
    forum   CITEXT REFERENCES Forum (slug),
    message TEXT                                 NOT NULL,
    slug    CITEXT UNIQUE                        NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT now(),
    votes   BIGINT                   DEFAULT 0
);

CREATE INDEX IF not exists thread_slug ON Thread (slug);
CREATE INDEX IF NOT EXISTS thread_forum_created ON Thread (forum, created);
CREATE INDEX IF not exists thread_author_forum ON Thread (author, forum);

CREATE OR REPLACE FUNCTION updatethreadcount() RETURNS TRIGGER AS
$body$
BEGIN
    UPDATE Forum
    SET threads = threads + 1
    WHERE slug = NEW.forum;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER update_thread_count_trigger
    AFTER INSERT
    ON Thread
    FOR EACH ROW
EXECUTE PROCEDURE updatethreadcount();

CREATE OR REPLACE FUNCTION insertforumuser() RETURNS TRIGGER AS
$body$
BEGIN
    INSERT INTO ForumUser(slug, nickname)
    VALUES (NEW.forum, NEW.author)
    ON CONFLICT DO NOTHING;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER insert_forum_user_trigger
    AFTER INSERT
    ON Thread
    FOR EACH ROW
EXECUTE PROCEDURE insertforumuser();

-- ---------------------------------------------------------------

CREATE TABLE Vote
(
    ID       BIGSERIAL PRIMARY KEY,
    threadID BIGINT                               NOT NULL,
    nickname CITEXT REFERENCES Users (nickname) NOT NULL,
    voice    SMALLINT                             NOT NULL,
    CONSTRAINT unique_vote UNIQUE (threadID, nickname)
);

CREATE INDEX IF NOT EXISTS vote_nickname_thread ON Vote (nickname, threadID);

CREATE OR REPLACE FUNCTION updatevotecount() RETURNS TRIGGER AS
$voteupdatecount$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE Thread
        SET votes = votes + NEW.voice
        WHERE id = NEW.threadId;
    ELSE
        UPDATE Thread
        SET votes = votes - OLD.voice + NEW.voice
        WHERE id = NEW.threadId;
    END IF;
    RETURN NEW;
END;
$voteupdatecount$ LANGUAGE plpgsql;
CREATE TRIGGER update_vote_trigger
    AFTER UPDATE OR INSERT
    ON Vote
    FOR EACH ROW
EXECUTE PROCEDURE updatevotecount();