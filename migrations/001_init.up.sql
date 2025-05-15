CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    frequency VARCHAR(16) NOT NULL,
    confirmed BOOLEAN DEFAULT FALSE,
    confirm_token VARCHAR(255) NOT NULL,
    unsubscribe_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_subscriptions_email ON subscriptions(email);
CREATE INDEX idx_subscriptions_confirm_token ON subscriptions(confirm_token);
CREATE INDEX idx_subscriptions_unsubscribe_token ON subscriptions(unsubscribe_token);
