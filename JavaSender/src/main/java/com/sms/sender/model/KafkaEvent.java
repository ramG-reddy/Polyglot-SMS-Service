package com.sms.sender.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

/**
 * Event model for Kafka messages.
 * Represents the SMS event that will be published to Kafka topic.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class KafkaEvent {

    /**
     * Unique identifier for the event
     */
    @JsonProperty("eventId")
    private String eventId;

    /**
     * User ID (same as phone number for this system)
     */
    @JsonProperty("userId")
    private String userId;

    /**
     * Phone number that received the SMS
     */
    @JsonProperty("phoneNumber")
    private String phoneNumber;

    /**
     * SMS message content
     */
    @JsonProperty("message")
    private String message;

    /**
     * Status of the SMS operation (SUCCESS/FAILED)
     */
    @JsonProperty("status")
    private String status;

    /**
     * Timestamp when the event was created
     */
    @JsonProperty("createdAt")
    private LocalDateTime createdAt;

    /**
     * Helper method to create a Kafka event from request and status
     */
    public static KafkaEvent from(SmsRequest request, String status) {
        return KafkaEvent.builder()
                .eventId(java.util.UUID.randomUUID().toString())
                .userId(request.getPhoneNumber())
                .phoneNumber(request.getPhoneNumber())
                .message(request.getMessage())
                .status(status)
                .createdAt(LocalDateTime.now())
                .build();
    }
}
