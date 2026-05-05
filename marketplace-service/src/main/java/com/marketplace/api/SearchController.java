package com.marketplace.api;

import com.marketplace.api.dto.ItemResponse;
import com.marketplace.domain.Item;
import com.marketplace.domain.ItemStatus;
import com.marketplace.repository.ItemRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;

@RestController
@RequestMapping("/api/v1/search/items")
@RequiredArgsConstructor
public class SearchController {

    private final ItemRepository itemRepository;

    @GetMapping
    public Page<ItemResponse> searchItems(
            @RequestParam(required = false) String q,
            @RequestParam(required = false) String category,
            @RequestParam(required = false) BigDecimal minPrice,
            @RequestParam(required = false) BigDecimal maxPrice,
            @RequestParam(required = false) ItemStatus status,
            Pageable pageable) {

        if (minPrice != null && maxPrice != null && minPrice.compareTo(maxPrice) > 0) {
            throw new IllegalArgumentException("minPrice cannot be greater than maxPrice");
        }

        Page<Item> items = itemRepository.searchItems(q, category, minPrice, maxPrice, status, pageable);
        
        return items.map(item -> new ItemResponse(
                item.getId(), item.getTitle(), item.getDescription(), item.getCategory(),
                item.getPrice(), item.getCurrency(), item.getSellerReference(),
                item.getStatus(), item.getTags(), item.getCreatedAt(), item.getUpdatedAt()
        ));
    }
}