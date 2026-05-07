package com.marketplace.infrastructure.messaging;

import com.marketplace.domain.ProcessedEvent;
import com.marketplace.repository.ProcessedEventRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.UUID;

@Component
@RequiredArgsConstructor
@Slf4j
public class MarketplaceConsumer {
    private final ProcessedEventRepository processedRepo;

    @KafkaListener(topics = "marketplace.items.events", groupId = "marketplace-audit-consumer")
    @Transactional
    public void consume(ConsumerRecord<String, String> record) {
        UUID eventId = UUID.fromString(record.key());

        // Дедуплікація
        if (processedRepo.existsById(eventId)) {
            log.info("Duplicate event skipped. EventId: {}", eventId);
            return;
        }

        // Обробка події
        log.info("Processing event: {}", record.value());

        // Запис про успішну обробку
        processedRepo.save(new ProcessedEvent(eventId, "audit-consumer", Instant.now()));
    }
}