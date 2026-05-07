package com.marketplace.domain;

import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.*;
import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "processed_events")
@Getter @Setter @NoArgsConstructor @AllArgsConstructor
public class ProcessedEvent {
    @Id private UUID eventId;
    private String consumerName;
    private Instant processedAt;
}