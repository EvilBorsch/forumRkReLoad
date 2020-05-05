CREATE EXTENSION IF NOT EXISTS CITEXT;
CREATE TABLE "user"
(
    nickname CITEXT PRIMARY KEY,
    fullname text,
    about    text,
    email    CITEXT UNIQUE
);


CREATE TABLE forum
(
    title         text,
    user_nickname CITEXT,
    foreign key (user_nickname) references "user" (nickname) ON DELETE CASCADE,
    slug          CITEXT PRIMARY KEY,
    posts         int DEFAULT 0,
    threads       int DEFAULT 0
);

CREATE TABLE threads
(
    id      BIGSERIAL PRIMARY KEY,
    author  citext,
    foreign key (author) references "user" (nickname) ON DELETE CASCADE,
    created timestamp with time zone,
    forum   citext,
    foreign key (forum) references forum (slug) ON DELETE CASCADE,
    message text,
    slug    citext UNIQUE,
    title   text,
    votes   int DEFAULT 0

);

CREATE TABLE posts
(
    id       BIGSERIAL primary key,
    author   citext references "user" (nickname),
    created  timestamp with time zone,
    forum    citext references forum (slug),
    isEdited bool,
    message  text,
    parent   int references posts (id),
    thread   int references threads (id),
    parents  bigint[] default array []::INTEGER[]
);

CREATE TABLE votes
(
    threadID   int references threads (id),
    threadSlug citext references threads (slug),
    nickname   citext references "user" (nickname),
    voice      int
);


CREATE OR REPLACE FUNCTION parents_change() RETURNS TRIGGER AS
$parents_change$
DECLARE
    temp_arr     BIGINT[];
    first_parent posts;
BEGIN
    IF (NEW.parent IS NULL) THEN
        NEW.parents := array_append(new.parents, new.id);
    ELSE
        SELECT parents FROM posts WHERE id = new.parent INTO temp_arr;
        SELECT * FROM posts WHERE id = temp_arr[1] INTO first_parent;
        IF NOT FOUND OR first_parent.thread != NEW.thread THEN
            RAISE EXCEPTION 'bad parent' USING ERRCODE = '00409';
        end if;

        NEW.parents := NEW.parents || temp_arr || new.id;
    end if;
    RETURN new;
end
$parents_change$ LANGUAGE plpgsql;

create trigger inserter
    before insert
    on posts
    for each row
execute procedure parents_change();
