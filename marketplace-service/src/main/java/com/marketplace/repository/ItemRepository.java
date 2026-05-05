package com.marketplace.repository;

import com.marketplace.domain.Item;
import com.marketplace.domain.ItemStatus;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.math.BigDecimal;
import java.util.UUID;

@Repository
public interface ItemRepository extends JpaRepository<Item, UUID> {
    
    Page<Item> findByStatus(ItemStatus status, Pageable pageable);

    // Простий механізм пошуку для SearchController
    @Query("SELECT i FROM Item i WHERE " +
           "(:q IS NULL OR LOWER(i.title) LIKE LOWER(CONCAT('%', :q, '%')) OR LOWER(i.description) LIKE LOWER(CONCAT('%', :q, '%'))) " +
           "AND (:category IS NULL OR i.category = :category) " +
           "AND (:minPrice IS NULL OR i.price >= :minPrice) " +
           "AND (:maxPrice IS NULL OR i.price <= :maxPrice) " +
           "AND (:status IS NULL OR i.status = :status)")
    Page<Item> searchItems(
            @Param("q") String q,
            @Param("category") String category,
            @Param("minPrice") BigDecimal minPrice,
            @Param("maxPrice") BigDecimal maxPrice,
            @Param("status") ItemStatus status,
            Pageable pageable);
}