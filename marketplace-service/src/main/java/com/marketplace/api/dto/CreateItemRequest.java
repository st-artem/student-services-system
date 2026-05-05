package com.marketplace.api.dto;

import jakarta.validation.constraints.*;
import java.math.BigDecimal;
import java.util.List;

public record CreateItemRequest(
    @NotBlank @Size(min = 3, max = 120) String title,
    @Size(max = 2000) String description,
    String category,
    @NotNull @DecimalMin("0.01") BigDecimal price,
    @Pattern(regexp = "^(UAH|USD|EUR)$") String currency,
    @NotBlank String sellerReference,
    List<@NotBlank String> tags
) {}