CREATE TABLE IF NOT EXISTS urls (
                                    id BIGSERIAL PRIMARY KEY,
                                    original_url TEXT NOT NULL,
                                    short_code VARCHAR(10) UNIQUE NOT NULL,
                                  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                  expires_at TIMESTAMP WITH TIME ZONE
                             );

CREATE INDEX idx_short_code ON urls(short_code);