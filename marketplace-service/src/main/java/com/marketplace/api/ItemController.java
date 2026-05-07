package com.marketplace.api;

import com.marketplace.api.dto.CreateItemRequest;
import com.marketplace.api.dto.ItemResponse;
import com.marketplace.api.dto.UpdateItemRequest;
import com.marketplace.domain.Item;
import com.marketplace.domain.ItemStatus;
import com.marketplace.repository.ItemRepository;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.data.web.PageableDefault;
import org.springframework.http.HttpStatus;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@RestController
@RequestMapping("/api/v1/items")
@RequiredArgsConstructor
public class ItemController {

    private final ItemRepository itemRepository;

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    @Transactional // Важливо!
    public ItemResponse createItem(@Valid @RequestBody CreateItemRequest request) throws Exception {
        Item item = Item.builder().title(request.title()).price(request.price()).status(ItemStatus.DRAFT).build();
        itemRepository.save(item);

        // Записуємо подію в Outbox
        OutboxEvent outboxEvent = OutboxEvent.builder()
                .id(UUID.randomUUID())
                .aggregateId(item.getId())
                .eventType("ItemCreated")
                .payload(objectMapper.writeValueAsString(item))
                .correlationId(MDC.get("correlationId"))
                .status("NEW")
                .createdAt(Instant.now())
                .build();
        outboxEventRepository.save(outboxEvent);

        return mapToResponse(item);
    }

    @GetMapping("/{id}")
    public ItemResponse getItem(@PathVariable UUID id) {
        Item item = itemRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Item not found"));
        return mapToResponse(item);
    }

    @PatchMapping("/{id}")
    @Transactional
    public ItemResponse updateItem(@PathVariable UUID id, @Valid @RequestBody UpdateItemRequest request) {
        Item item = itemRepository.findById(id)
                .orElseThrow(() -> new IllegalArgumentException("Item not found"));

        if (request.status() != null) {
            if (item.getStatus() == ItemStatus.ARCHIVED && request.status() == ItemStatus.ACTIVE) {
                throw new IllegalStateException("Cannot activate an ARCHIVED item");
            }
            item.setStatus(request.status());
        }
        if (request.title() != null) item.setTitle(request.title());
        if (request.description() != null) item.setDescription(request.description());
        if (request.category() != null) item.setCategory(request.category());
        if (request.price() != null) item.setPrice(request.price());
        if (request.tags() != null) item.setTags(request.tags());

        itemRepository.save(item);
        return mapToResponse(item);
    }

    @GetMapping
    public Page<ItemResponse> getItems(
            @RequestParam(required = false) ItemStatus status,
            @PageableDefault(size = 20, sort = "createdAt", direction = Sort.Direction.DESC) Pageable pageable) {
        
        Page<Item> items = status != null 
            ? itemRepository.findByStatus(status, pageable) 
            : itemRepository.findAll(pageable);
            
        return items.map(this::mapToResponse);
    }

    private ItemResponse mapToResponse(Item item) {
        return new ItemResponse(
                item.getId(), item.getTitle(), item.getDescription(), item.getCategory(),
                item.getPrice(), item.getCurrency(), item.getSellerReference(),
                item.getStatus(), item.getTags(), item.getCreatedAt(), item.getUpdatedAt()
        );
    }
}