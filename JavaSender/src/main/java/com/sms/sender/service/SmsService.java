package com.sms.sender.service;

import com.sms.sender.model.KafkaEvent;
import com.sms.sender.model.SmsRequest;
import com.sms.sender.model.SmsResponse;
import com.sms.sender.kafka.SmsKafkaProducer;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.Random;
import java.util.concurrent.TimeUnit;

/**
 * Core business service for SMS sending operations.
 * Handles block list validation, vendor API calls, and Kafka event production.
 */
@Service
@Slf4j
public class SmsService {

    private final BlockListService blockListService;
    private final SmsKafkaProducer kafkaProducer;
    private final Random random;

    @Value("${app.mock-vendor.min-delay-ms}")
    private int minDelayMs;

    @Value("${app.mock-vendor.max-delay-ms}")
    private int maxDelayMs;

    @Value("${app.mock-vendor.failure-rate}")
    private double failureRate;

    public SmsService(BlockListService blockListService, SmsKafkaProducer kafkaProducer) {
        this.blockListService = blockListService;
        this.kafkaProducer = kafkaProducer;
        this.random = new Random();
    }

    /**
     * Process SMS sending request.
     * Validates against block list, calls vendor API, and produces Kafka event.
     * 
     * @param request SMS request containing phone number and message
     * @return SmsResponse with operation status
     */
    public SmsResponse sendSms(SmsRequest request) {
        String phoneNumber = request.getPhoneNumber();
        String message = request.getMessage();
        
        log.info("Processing SMS request for phone number: {}", phoneNumber);
        
        try {
            // Step 1: Check if phone number is blocked
            if (blockListService.isBlocked(phoneNumber)) {
                log.warn("SMS rejected: Phone number {} is in block list", phoneNumber);
                return SmsResponse.blocked(phoneNumber);
            }
            
            // Step 2: Call mock 3rd party vendor API
            boolean vendorSuccess = callMockVendorApi(phoneNumber, message);
            
            // Step 3: Determine status and prepare response
            String status = vendorSuccess ? "SUCCESS" : "FAILED";
            
            // Step 4: Produce Kafka event (synchronous for v0)
            KafkaEvent event = KafkaEvent.from(request, status);
            boolean kafkaSent = kafkaProducer.sendSmsEventSync(event);
            
            if (!kafkaSent) {
                log.error("Failed to send Kafka event for phone number: {}", phoneNumber);
                // return failure response if Kafka send fails
                return SmsResponse.failed(phoneNumber, "Kafka: Failed to log SMS event");
            }
            
            // Step 5: Return response based on vendor result
            if (vendorSuccess) {
                log.info("SMS sent successfully to phone number: {}", phoneNumber);
                return SmsResponse.success(phoneNumber, "SMS sent successfully");
            } else {
                log.warn("SMS failed for phone number: {} - Vendor error", phoneNumber);
                return SmsResponse.failed(phoneNumber, "Vendor:  API returned error");
            }
            
        } catch (Exception e) {
            log.error("Unexpected error processing SMS for phone number: {} - {}", 
                    phoneNumber, e.getMessage(), e);
            return SmsResponse.failed(phoneNumber, "Internal server error: " + e.getMessage());
        }
    }

    /**
     * Mock 3rd party vendor API call.
     * Simulates network latency with random delay.
     * Returns random success/failure based on configured failure rate.
     * 
     * @param phoneNumber Phone number to send SMS to
     * @param message SMS message content
     * @return true if mock vendor call succeeds, false otherwise
     */
    private boolean callMockVendorApi(String phoneNumber, String message) {
        log.debug("Calling mock vendor API for phone number: {}", phoneNumber);
        
        try {
            // Simulate network latency with random delay
            int delay = minDelayMs + random.nextInt(maxDelayMs - minDelayMs + 1);
            log.debug("Simulating vendor API delay of {}ms", delay);
            TimeUnit.MILLISECONDS.sleep(delay);
            
            // Randomly determine success/failure based on configured failure rate
            boolean success = random.nextDouble() > failureRate;
            
            if (success) {
                log.debug("Mock vendor API returned SUCCESS for phone number: {}", phoneNumber);
            } else {
                log.debug("Mock vendor API returned FAILURE for phone number: {}", phoneNumber);
            }
            
            return success;
            
        } catch (InterruptedException e) {
            log.error("Mock vendor API call interrupted for phone number: {}", phoneNumber, e);
            Thread.currentThread().interrupt();
            return false;
        }
    }

    /**
     * Get statistics about block list size.
     * Useful for monitoring and debugging.
     * 
     * @return Number of blocked phone numbers
     */
    public long getBlockedNumbersCount() {
        return blockListService.getBlockListSize();
    }
}
