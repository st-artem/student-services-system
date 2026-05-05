package com.marketplace.api.dto;

import com.marketplace.domain.ItemStatus;
import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import java.util.UUID;

public record ItemResponse(
    UUID id,
    String title,
    String description,
    String category,
    BigDecimal price,
    String currency,
    String sellerReference,
    ItemStatus status,
    List<String> tags,
    Instant createdAt,
    Instant updatedAt
) {}