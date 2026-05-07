package com.marketplace.infrastructure.messaging;

import com.marketplace.domain.OutboxEvent;
import com.marketplace.repository.OutboxEventRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import java.util.List;

@Component
@RequiredArgsConstructor
public class OutboxPublisher {
    private final OutboxEventRepository outboxRepo;
    private final KafkaTemplate<String, String> kafkaTemplate;

    @Scheduled(fixedDelay = 5000)
    @Transactional
    public void publishEvents() {
        List<OutboxEvent> events = outboxRepo.findByStatus("NEW");
        for (OutboxEvent event : events) {
            try {
                // Публікуємо в Kafka
                kafkaTemplate.send("marketplace.items.events", event.getId().toString(), event.getPayload());
                event.setStatus("PUBLISHED");
            } catch (Exception e) {
                event.setStatus("FAILED");
            }
        }
        outboxRepo.saveAll(events);
    }
}