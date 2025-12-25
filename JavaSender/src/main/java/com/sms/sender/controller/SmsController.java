package com.sms.sender.controller;

import com.sms.sender.model.SmsRequest;
import com.sms.sender.model.SmsResponse;
import com.sms.sender.service.SmsService;
import jakarta.validation.Valid;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.BindingResult;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * REST controller for SMS sending operations.
 * Provides the public API endpoint for SMS requests.
 */
@RestController
@RequestMapping("/v0/sms")
@Slf4j
public class SmsController {

    private final SmsService smsService;

    public SmsController(SmsService smsService) {
        this.smsService = smsService;
    }

    /**
     * Send SMS endpoint.
     * Accepts SMS request, validates input, and processes through SMS service.
     * 
     * @param request SMS request containing phone number and message
     * @param bindingResult Validation results
     * @return ResponseEntity with SmsResponse and appropriate HTTP status
     */
    @PostMapping("/send")
    public ResponseEntity<SmsResponse> sendSms(
            @Valid @RequestBody SmsRequest request,
            BindingResult bindingResult) {
        
        log.info("Received SMS send request for phone number: {}", request.getPhoneNumber());
        
        // Check for validation errors
        if (bindingResult.hasErrors()) {
            String errorMessage = bindingResult.getFieldErrors().stream()
                    .map(error -> error.getField() + ": " + error.getDefaultMessage())
                    .collect(Collectors.joining(", "));
            
            log.warn("Validation failed for SMS request: {}", errorMessage);
            
            SmsResponse errorResponse = SmsResponse.builder()
                    .status("FAILED")
                    .message("Validation failed: " + errorMessage)
                    .phoneNumber(request.getPhoneNumber())
                    .timestamp(LocalDateTime.now())
                    .build();
            
            return ResponseEntity.badRequest().body(errorResponse);
        }
        
        // Process SMS request
        SmsResponse response = smsService.sendSms(request);
        
        // Determine HTTP status code based on response status
        HttpStatus httpStatus = switch (response.getStatus()) {
            case "SUCCESS" -> HttpStatus.OK;
            case "BLOCKED" -> HttpStatus.FORBIDDEN;
            case "FAILED" -> HttpStatus.INTERNAL_SERVER_ERROR;
            default -> HttpStatus.OK;
        };
        
        log.info("SMS request completed with status: {} for phone number: {}", 
                response.getStatus(), request.getPhoneNumber());
        
        return ResponseEntity.status(httpStatus).body(response);
    }

    /**
     * Health check endpoint for monitoring.
     * 
     * @return Simple health status
     */
    @GetMapping("/health")
    public ResponseEntity<Map<String, Object>> health() {
        Map<String, Object> health = new HashMap<>();
        health.put("status", "UP");
        health.put("timestamp", LocalDateTime.now());
        health.put("service", "sms-sender");
        health.put("blockedNumbers", smsService.getBlockedNumbersCount());
        return ResponseEntity.ok(health);
    }

    /**
     * Global exception handler for unexpected errors.
     * 
     * @param e Exception that was thrown
     * @return Error response with appropriate status
     */
    @ExceptionHandler(Exception.class)
    public ResponseEntity<SmsResponse> handleException(Exception e) {
        log.error("Unexpected error in SMS controller: {}", e.getMessage(), e);
        
        SmsResponse errorResponse = SmsResponse.builder()
                .status("FAILED")
                .message("Internal server error: " + e.getMessage())
                .phoneNumber("N/A")
                .timestamp(LocalDateTime.now())
                .build();
        
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(errorResponse);
    }
}
