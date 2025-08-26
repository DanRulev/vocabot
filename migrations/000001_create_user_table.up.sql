CREATE TABLE user_words (
    user_id BIGINT NOT NULL,
    word_text VARCHAR(255) NOT NULL,
    translation VARCHAR(255) NOT NULL,
    last_seen TIMESTAMP,
    known BOOLEAN DEFAULT FALSE,
    UNIQUE (user_id, word_text)
);

CREATE TABLE user_quiz_results (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    word VARCHAR(255) NOT NULL,
    translation VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    is_correct BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);