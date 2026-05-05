CREATE TABLE items (
    id UUID PRIMARY KEY,
    title VARCHAR(120) NOT NULL,
    description VARCHAR(2000),
    category VARCHAR(100),
    price DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    seller_reference VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE item_tags (
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL
);

CREATE INDEX idx_items_status ON items(status);
CREATE INDEX idx_items_category ON items(category);