package com.marketplace.domain;

import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.*;
import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "outbox_events")
@Getter @Setter @Builder @NoArgsConstructor @AllArgsConstructor
public class OutboxEvent {
    @Id private UUID id;
    private UUID aggregateId;
    private String eventType;
    private String payload;
    private String correlationId;
    private String status;
    private Instant createdAt;
}