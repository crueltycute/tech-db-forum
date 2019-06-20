CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE TABLE IF NOT EXISTS Users
(
    nickname CITEXT        NOT NULL PRIMARY KEY,
    fullname TEXT          NOT NULL,
    email    CITEXT UNIQUE NOT NULL,
    about    TEXT
);

CREATE UNIQUE INDEX email_unique_idx on Users (email);



CREATE TABLE IF NOT EXISTS Forum
(
    slug      CITEXT                             NOT NULL PRIMARY KEY,
    forumUser CITEXT REFERENCES Users (nickname) NOT NULL,
    title     CITEXT                             NOT NULL,
    posts     INTEGER DEFAULT 0,
    threads   INTEGER DEFAULT 0
);

CREATE INDEX forum_slug_hash_idx on Forum USING hash (slug);



CREATE TABLE ForumUser
(
    slug     CITEXT REFERENCES Forum (slug),
    nickname CITEXT COLLATE "POSIX" REFERENCES Users (nickname),
    CONSTRAINT unique_slug_nickname UNIQUE (slug, nickname)
);



CREATE TABLE IF NOT EXISTS Thread
(
    id      SERIAL                             NOT NULL PRIMARY KEY,
    author  CITEXT REFERENCES Users (nickname) NOT NULL,
    forum   CITEXT REFERENCES Forum (slug),
    votes   INTEGER    DEFAULT 0,
    slug    CITEXT UNIQUE,
    title   TEXT                               NOT NULL,
    message TEXT                               NOT NULL,
    created TIMESTAMPTZ DEFAULT now()
);

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



CREATE TABLE IF NOT EXISTS Post
(
    id       SERIAL PRIMARY KEY,
    parent   INTEGER                  DEFAULT 0,
    author   CITEXT    NOT NULL,
    message  TEXT      NOT NULL,
    isEdited BOOLEAN,
    forum    CITEXT,
    created  TIMESTAMP WITH TIME ZONE DEFAULT now(),
    thread   INTEGER,

    path     INTEGER[] NOT NULL
);

CREATE OR REPLACE FUNCTION setforum() RETURNS TRIGGER AS
$body$
BEGIN
    NEW.forum = (SELECT forum FROM Thread WHERE id = NEW.thread);
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER set_forum_trigger
    BEFORE INSERT
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE setforum();


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

CREATE OR REPLACE FUNCTION updatepostcount() RETURNS TRIGGER AS
$body$
BEGIN
    UPDATE Forum
    SET posts = posts + 1
    WHERE slug = NEW.forum;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER update_post_count_trigger
    AFTER INSERT
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE updatepostcount();

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
    ON Post
    FOR EACH ROW
EXECUTE PROCEDURE insertforumuser();

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



CREATE TABLE IF NOT EXISTS Vote
(
    id       SERIAL                             NOT NULL PRIMARY KEY,
    nickname CITEXT REFERENCES Users (nickname) NOT NULL,
    threadID INTEGER                            NOT NULL,
    voice    INT                                NOT NULL,
    CONSTRAINT unique_vote UNIQUE (threadID, nickname)
);

CREATE OR REPLACE FUNCTION addvotecount() RETURNS TRIGGER AS
$voteinsertcount$
BEGIN
    INSERT INTO VoteCount(threadID, count)
    VALUES (new.threadId, new.voice)
    ON CONFLICT(threadId) DO UPDATE
        SET count = VoteCount.count + new.voice;
    RETURN NEW;
END;
$voteinsertcount$ LANGUAGE plpgsql;
CREATE TRIGGER add_vote_trigger
    AFTER INSERT
    ON Vote
    FOR EACH ROW
EXECUTE PROCEDURE addvotecount();

CREATE OR REPLACE FUNCTION updatevotecount() RETURNS TRIGGER AS
$voteupdatecount$
BEGIN
    UPDATE VoteCount
    SET count = count - old.voice + new.voice
    WHERE VoteCount.threadId = new.threadId;
    RETURN NEW;
END;
$voteupdatecount$ LANGUAGE plpgsql;
CREATE TRIGGER update_vote_trigger
    AFTER UPDATE
    ON Vote
    FOR EACH ROW
EXECUTE PROCEDURE updatevotecount();



CREATE TABLE VoteCount
(
    threadID BIGINT REFERENCES Thread (ID) UNIQUE NOT NULL,
    count    BIGINT DEFAULT 0
);
