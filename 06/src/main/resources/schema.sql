create table books
(
  id    bigint primary key auto_increment,
  title varchar(30)
);

create table authors
(
  id        bigint primary key auto_increment,
  first_name varchar(30),
  last_name  varchar(30)
);

create table genres
(
  id   int primary key auto_increment,
  name varchar(30)
);

create table book_authors
(
  book_id   bigint,
  author_id bigint,
  primary key (book_id, author_id)
);

create table book_genres
(
  book_id  bigint,
  genre_id int,
  primary key (book_id, genre_id)
);

create table book_comments
(
    book_id bigint,
    comment_id bigint primary key,
    comment varchar(100)
)
