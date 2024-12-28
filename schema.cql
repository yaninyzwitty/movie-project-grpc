CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT,
    alias_name TEXT,
);

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    name TEXT,
    description TEXT
);

CREATE TABLE IF NOT EXISTS movies_by_user (
    user_id UUID,
    movie_id TIMEUUID,
    category_id UUID,
    name TEXT,
    banner_url TEXT,
    movie_url TEXT,
    description TEXT,  
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY ((user_id), category_id, created_at, movie_id)
) WITH CLUSTERING ORDER BY (category_id ASC, created_at DESC, movie_id ASC);


CREATE CUSTOM INDEX list_movies_by_names ON movie_db.movies_by_user (name) 
  USING 'StorageAttachedIndex' 