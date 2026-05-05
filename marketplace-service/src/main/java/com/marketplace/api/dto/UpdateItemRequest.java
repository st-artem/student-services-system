package com.marketplace.api.dto;

import com.marketplace.domain.ItemStatus;
import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.Size;
import java.math.BigDecimal;
import java.util.List;

public record UpdateItemRequest(
    @Size(min = 3, max = 120) String title,
    @Size(max = 2000) String description,
    String category,
    @DecimalMin("0.01") BigDecimal price,
    ItemStatus status,
    List<String> tags
) {}