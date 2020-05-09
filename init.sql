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



CREATE INDEX on votes(nickname,threadID);



CREATE INDEX forum_slug_idx ON forum (slug);
CREATE INDEX users_nick_idx ON "user" (nickname);
CREATE INDEX users_nick_lower_idx ON "user" (lower(nickname)); --delete if not work
-- CREATE INDEX users_nick_email_idx ON "user" (nickname,email); --delete if not work
-- CREATE INDEX users_email_idx ON "user" (email); --delete if not work



CREATE INDEX fpi_idx ON posts ((posts.parents[1]), thread);
CREATE INDEX pid_idx ON posts ((posts.parents[1]), id);
CREATE INDEX parents_idx ON posts ((posts.parents[1]));
CREATE INDEX parents_all_idx ON posts ((posts.parents)); --delete if not work
CREATE INDEX thread_idx ON posts (thread);
CREATE INDEX pare_idx ON posts ((posts.parent));

create index if not exists slug_id on threads (slug);
create index if not exists f_created_idx on threads (forum, created);
-- create index if not exists t_author_idx on threads (author, forum);




CREATE TABLE forum_user
(
    user_nickname citext NOT NULL,
    forum_slug    citext NOT NULL,
    FOREIGN KEY (user_nickname) REFERENCES "user" (nickname),
    FOREIGN KEY (forum_slug) REFERENCES forum (Slug),
    UNIQUE (user_nickname, forum_slug)
);

CREATE OR REPLACE FUNCTION uForum_updater() RETURNS TRIGGER AS
$update_users_forum$
BEGIN
    INSERT INTO forum_user (user_nickname, forum_slug) VALUES (NEW.author, NEW.forum) on conflict do nothing;
    return NEW;
end
$update_users_forum$ LANGUAGE plpgsql;



CREATE TRIGGER thread_insert_user_forum
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE uForum_updater();
CREATE TRIGGER post_insert_user_forum
    AFTER INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE uForum_updater();


CREATE INDEX forum_user_index ON forum_user (forum_slug, lower(user_nickname));

